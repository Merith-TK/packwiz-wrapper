package build

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Merith-TK/packwiz-wrapper/internal/packwiz"
	"github.com/Merith-TK/packwiz-wrapper/internal/utils"
)

// MultiMCComponent represents a component in mmc-pack.json
type MultiMCComponent struct {
	UID     string `json:"uid"`
	Version string `json:"version"`
}

// MultiMCPack represents the mmc-pack.json structure
type MultiMCPack struct {
	Components    []MultiMCComponent `json:"components"`
	FormatVersion int                `json:"formatVersion"`
}

// ExportMultiMC exports the pack as a MultiMC instance
func ExportMultiMC(packDir, packName string, useLocal bool) error {
	fmt.Println("🎮 Exporting MultiMC pack...")

	// Find pack.toml location
	packLocation := utils.FindPackToml(packDir)
	if packLocation == "" {
		return fmt.Errorf("pack.toml not found")
	}

	// Read pack.toml to get version information
	packToml, err := readPackToml(packLocation)
	if err != nil {
		return fmt.Errorf("failed to read pack.toml: %w", err)
	}

	// Create temporary build directory
	tempDir := filepath.Join(packDir, ".mmc-temp")
	defer os.RemoveAll(tempDir)

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Create instance.cfg
	if err := createInstanceCfg(tempDir, packName, packLocation, useLocal); err != nil {
		return fmt.Errorf("failed to create instance.cfg: %w", err)
	}

	// Create mmc-pack.json
	if err := createMMCPack(tempDir, packToml); err != nil {
		return fmt.Errorf("failed to create mmc-pack.json: %w", err)
	}

	// Copy icon if it exists
	iconName := sanitizeIconName(packName)
	iconPath := filepath.Join(packLocation, "icon.png")
	if _, err := os.Stat(iconPath); err == nil {
		destIcon := filepath.Join(tempDir, iconName+"_icon.png")
		if err := copyFile(iconPath, destIcon); err != nil {
			fmt.Printf("Warning: Failed to copy icon: %v\n", err)
		}
	}

	// Ensure packwiz-installer-bootstrap.jar is included
	if err := ensurePackwizInstaller(tempDir, packLocation); err != nil {
		return fmt.Errorf("failed to ensure packwiz installer: %w", err)
	}

	// Copy .minecraft directory (excluding unnecessary files)
	minecraftDir := filepath.Join(tempDir, ".minecraft")
	if err := copyMinecraftDir(packLocation, minecraftDir); err != nil {
		return fmt.Errorf("failed to copy .minecraft directory: %w", err)
	}

	// Create the zip file
	buildDir := filepath.Join(packDir, ".build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create .build directory: %w", err)
	}

	zipPath := filepath.Join(buildDir, packName+"-multimc.zip")
	if err := CreateZipFromDir(tempDir, zipPath); err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}

	fmt.Printf("✅ Created MultiMC pack: .build/%s-multimc.zip\n", packName)
	return nil
}

// readPackToml reads and parses the pack.toml file
func readPackToml(packLocation string) (*packwiz.PackToml, error) {
	packTomlPath := filepath.Join(packLocation, "pack.toml")
	data, err := os.ReadFile(packTomlPath)
	if err != nil {
		return nil, err
	}

	var packToml packwiz.PackToml
	if err := toml.Unmarshal(data, &packToml); err != nil {
		return nil, err
	}

	return &packToml, nil
}

// createInstanceCfg creates the instance.cfg file for MultiMC
func createInstanceCfg(tempDir, packName, packLocation string, useLocal bool) error {
	iconName := sanitizeIconName(packName)

	// Get pack URL for the pre-launch command
	packURL := getPackURL(packLocation, useLocal)

	cfg := fmt.Sprintf(`[General]
InstanceType=OneSix
iconKey=%s_icon
name=%s
OverrideCommands=true
PreLaunchCommand="$INST_JAVA" -jar packwiz-installer-bootstrap.jar %s
`, iconName, packName, packURL)

	return os.WriteFile(filepath.Join(tempDir, "instance.cfg"), []byte(cfg), 0644)
}

// createMMCPack creates the mmc-pack.json file
func createMMCPack(tempDir string, packToml *packwiz.PackToml) error {
	var components []MultiMCComponent

	// Add Minecraft component
	mcVersion := packToml.Versions.Minecraft
	if mcVersion == "" {
		mcVersion = packToml.McVersion
	}
	if mcVersion != "" {
		components = append(components, MultiMCComponent{
			UID:     "net.minecraft",
			Version: mcVersion,
		})
	}

	// Add LWJGL component (usually 3.3.3 for modern versions)
	components = append(components, MultiMCComponent{
		UID:     "org.lwjgl3",
		Version: "3.3.3",
	})

	// Add mod loader components
	if packToml.Versions.Fabric != "" {
		components = append(components, MultiMCComponent{
			UID:     "net.fabricmc.fabric-loader",
			Version: packToml.Versions.Fabric,
		})
	}

	if packToml.Versions.Forge != "" {
		components = append(components, MultiMCComponent{
			UID:     "net.minecraftforge",
			Version: packToml.Versions.Forge,
		})
	}

	if packToml.Versions.Quilt != "" {
		components = append(components, MultiMCComponent{
			UID:     "org.quiltmc.quilt-loader",
			Version: packToml.Versions.Quilt,
		})
	}

	pack := MultiMCPack{
		Components:    components,
		FormatVersion: 1,
	}

	data, err := json.MarshalIndent(pack, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(tempDir, "mmc-pack.json"), data, 0644)
}

// sanitizeIconName converts pack name to a valid icon name
func sanitizeIconName(packName string) string {
	// Replace non-alphanumeric characters with underscores
	result := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, packName)

	return result
}

// getPackURL tries to get the pack URL, falls back to local path
func getPackURL(packLocation string, useLocal bool) string {
	if useLocal {
		// Return local pack.toml path
		return filepath.Join(packLocation, "pack.toml")
	}

	// Try to detect remote URL from git if available
	remoteURL, err := utils.DetectRemotePackURL(packLocation)
	if err != nil || remoteURL == "" {
		// Fall back to local path if remote detection fails
		return filepath.Join(packLocation, "pack.toml")
	}

	return remoteURL
}

// copyMinecraftDir copies the .minecraft directory excluding unnecessary files
func copyMinecraftDir(srcDir, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Skip certain files/directories
		if shouldSkipFile(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		destPath := filepath.Join(destDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		} else {
			return copyFile(path, destPath)
		}
	})
}

// shouldSkipFile determines if a file should be skipped during MultiMC export
func shouldSkipFile(relPath string) bool {
	skipPatterns := []string{
		"mods", // Skip mods directory - will be downloaded by packwiz
		".build",
		".git",
		".temp",
		".mmc-temp", // Skip our own temp directory
		".technic",  // Skip other temp directories
		".server",
		"packwiz-installer-bootstrap.jar", // Will be included separately
	}

	for _, pattern := range skipPatterns {
		if strings.HasPrefix(relPath, pattern) || relPath == pattern {
			return true
		}
	}

	return false
}

// copyFile copies a file from src to dest
func copyFile(src, dest string) error {
	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dest, data, 0644)
}

// ensurePackwizInstaller ensures the packwiz-installer-bootstrap.jar is present in the temp directory
func ensurePackwizInstaller(tempDir, packLocation string) error {
	installerName := "packwiz-installer-bootstrap.jar"
	installerPath := filepath.Join(tempDir, installerName)

	// Check if installer already exists in pack location
	existingInstaller := filepath.Join(packLocation, installerName)
	if _, err := os.Stat(existingInstaller); err == nil {
		// Copy existing installer
		return copyFile(existingInstaller, installerPath)
	}

	// Download installer from GitHub releases
	fmt.Println("📥 Downloading packwiz-installer-bootstrap.jar...")
	return DownloadPackwizInstaller(installerPath)
}
