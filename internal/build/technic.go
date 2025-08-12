package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Merith-TK/packwiz-wrapper/internal/packwiz"
	"github.com/Merith-TK/packwiz-wrapper/internal/utils"
)

// ExportTechnic exports the pack as a Technic pack
func ExportTechnic(packDir, packName string) error {
	fmt.Println("⚡ Exporting Technic pack...")

	// Find pack.toml location
	packLocation := utils.FindPackToml(packDir)
	if packLocation == "" {
		return fmt.Errorf("pack.toml not found")
	}

	// Create temporary technic directory
	technicDir := filepath.Join(packDir, ".technic")
	defer os.RemoveAll(technicDir)

	if err := os.MkdirAll(technicDir, 0755); err != nil {
		return fmt.Errorf("failed to create .technic directory: %w", err)
	}

	// Copy .minecraft contents to .technic
	if err := copyTechnicFiles(packLocation, technicDir); err != nil {
		return fmt.Errorf("failed to copy files: %w", err)
	}

	// Download mods using packwiz installer
	if err := installModsForTechnic(technicDir, packLocation); err != nil {
		return fmt.Errorf("failed to install mods: %w", err)
	}

	// Clean up packwiz files
	if err := cleanupTechnicFiles(technicDir); err != nil {
		return fmt.Errorf("failed to cleanup files: %w", err)
	}

	// Create the zip file
	buildDir := filepath.Join(packDir, ".build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create .build directory: %w", err)
	}

	zipPath := filepath.Join(buildDir, packName+"-technic.zip")
	if err := CreateZipFromDir(technicDir, zipPath); err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}

	fmt.Printf("✅ Created Technic pack: .build/%s-technic.zip\n", packName)
	return nil
}

// copyTechnicFiles copies necessary files from pack location to technic directory
func copyTechnicFiles(packLocation, technicDir string) error {
	return filepath.Walk(packLocation, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(packLocation, path)
		if err != nil {
			return err
		}

		// Skip certain files for technic
		if shouldSkipTechnicFile(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		destPath := filepath.Join(technicDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		} else {
			return copyFile(path, destPath)
		}
	})
}

// shouldSkipTechnicFile determines if a file should be skipped for Technic export
func shouldSkipTechnicFile(relPath string) bool {
	skipPatterns := []string{
		".build",
		".git",
		".temp",
		// Note: mods directory and .pw.toml files are needed by packwiz installer
	}

	// Allow mod metadata files (needed by packwiz installer)
	if strings.HasSuffix(relPath, ".pw.toml") {
		return false
	}

	// Skip actual mod jar files (will be downloaded by installer)
	if strings.HasPrefix(relPath, "mods"+string(filepath.Separator)) && strings.HasSuffix(relPath, ".jar") {
		return true
	}

	for _, pattern := range skipPatterns {
		if strings.HasPrefix(relPath, pattern) {
			return true
		}
	}

	return false
}

// installModsForTechnic uses packwiz installer to download mods
func installModsForTechnic(technicDir, packLocation string) error {
	fmt.Println("📦 Installing mods for Technic pack...")

	// Find packwiz installer
	installerPath := filepath.Join(packLocation, "packwiz-installer-bootstrap.jar")
	if _, err := os.Stat(installerPath); os.IsNotExist(err) {
		// Try to download it
		fmt.Println("⬇️  Downloading packwiz installer...")
		if err := DownloadPackwizInstaller(installerPath); err != nil {
			return fmt.Errorf("failed to download packwiz installer: %w", err)
		}
	}

	// Copy installer to technic directory
	technicInstallerPath := filepath.Join(technicDir, "packwiz-installer-bootstrap.jar")
	if err := copyFile(installerPath, technicInstallerPath); err != nil {
		return fmt.Errorf("failed to copy installer: %w", err)
	}

	// Copy pack.toml
	packTomlSrc := filepath.Join(packLocation, "pack.toml")
	packTomlDest := filepath.Join(technicDir, "pack.toml")
	if err := copyFile(packTomlSrc, packTomlDest); err != nil {
		return fmt.Errorf("failed to copy pack.toml: %w", err)
	}

	// Copy index.toml
	indexTomlSrc := filepath.Join(packLocation, "index.toml")
	indexTomlDest := filepath.Join(technicDir, "index.toml")
	if err := copyFile(indexTomlSrc, indexTomlDest); err != nil {
		return fmt.Errorf("failed to copy index.toml: %w", err)
	}

	// Find compatible Java for the pack
	packToml, _, err := utils.LoadPackConfig(packLocation)
	if err != nil {
		return fmt.Errorf("failed to load pack config: %w", err)
	}

	mcVersion := getMinecraftVersionTechnic(packToml)
	java, err := utils.FindCompatibleJava(mcVersion)
	if err != nil {
		return fmt.Errorf("no compatible Java found for Minecraft %s: %w", mcVersion, err)
	}

	fmt.Printf("Using Java %s for mod installation\n", java.Version)

	// Run packwiz installer (without server flag to get all mods) with no-gui mode
	cmd := exec.Command(java.Path, "-jar", "packwiz-installer-bootstrap.jar", "pack.toml", "-g")
	cmd.Dir = technicDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Don't attach stdin to prevent hanging on input prompts
	cmd.Stdin = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("packwiz installer failed: %w", err)
	}

	fmt.Println("✅ Mods installed successfully for Technic pack")
	return nil
}

// cleanupTechnicFiles removes packwiz-specific files after mod installation
func cleanupTechnicFiles(technicDir string) error {
	filesToRemove := []string{
		"packwiz-installer-bootstrap.jar",
		"pack.toml",
		"index.toml",
	}

	for _, file := range filesToRemove {
		filePath := filepath.Join(technicDir, file)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: Failed to remove %s: %v\n", file, err)
		}
	}

	// Remove .pw.toml files from mods directory
	modsDir := filepath.Join(technicDir, "mods")
	if _, err := os.Stat(modsDir); err == nil {
		return filepath.Walk(modsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if strings.HasSuffix(path, ".pw.toml") {
				if err := os.Remove(path); err != nil {
					fmt.Printf("Warning: Failed to remove %s: %v\n", path, err)
				}
			}

			return nil
		})
	}

	return nil
}

// getMinecraftVersionTechnic gets the Minecraft version from pack.toml
func getMinecraftVersionTechnic(packToml *packwiz.PackToml) string {
	// Prefer versions.minecraft, fallback to mc-version
	if packToml.Versions.Minecraft != "" {
		return packToml.Versions.Minecraft
	}
	return packToml.McVersion
}
