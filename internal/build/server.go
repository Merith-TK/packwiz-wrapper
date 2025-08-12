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

// ExportServer exports the pack as a server pack
func ExportServer(packDir, packName string) error {
	fmt.Println("🖥️  Exporting Server pack...")

	// Find pack.toml location
	packLocation := utils.FindPackToml(packDir)
	if packLocation == "" {
		return fmt.Errorf("pack.toml not found")
	}

	// Create temporary server directory
	serverDir := filepath.Join(packDir, ".server")
	defer os.RemoveAll(serverDir)

	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return fmt.Errorf("failed to create .server directory: %w", err)
	}

	// Copy server-relevant files
	if err := copyServerFiles(packLocation, serverDir); err != nil {
		return fmt.Errorf("failed to copy server files: %w", err)
	}

	// Create server icon
	if err := createServerIcon(packLocation, serverDir); err != nil {
		fmt.Printf("Warning: Failed to create server icon: %v\n", err)
	}

	// Install server-side mods using packwiz installer
	if err := installServerMods(serverDir, packLocation); err != nil {
		return fmt.Errorf("failed to install server mods: %w", err)
	}

	// Create server startup scripts
	if err := createServerScripts(serverDir); err != nil {
		return fmt.Errorf("failed to create server scripts: %w", err)
	}

	// Create the zip file
	buildDir := filepath.Join(packDir, ".build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create .build directory: %w", err)
	}

	zipPath := filepath.Join(buildDir, packName+"-server.zip")
	if err := CreateZipFromDir(serverDir, zipPath); err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}

	fmt.Printf("✅ Created Server pack: .build/%s-server.zip\n", packName)
	return nil
}

// copyServerFiles copies server-relevant files from pack location
func copyServerFiles(packLocation, serverDir string) error {
	return filepath.Walk(packLocation, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(packLocation, path)
		if err != nil {
			return err
		}

		// Skip client-only files and directories
		if shouldSkipServerFile(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		destPath := filepath.Join(serverDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		} else {
			return copyFile(path, destPath)
		}
	})
}

// shouldSkipServerFile determines if a file should be skipped for server export
func shouldSkipServerFile(relPath string) bool {
	skipPatterns := []string{
		".build",
		".git",
		".temp",
		"resourcepacks", // Client-only
		"shaderpacks",   // Client-only
		"screenshots",   // Client-only
		"logs",          // Runtime generated
		"crash-reports", // Runtime generated
		"saves",         // Client-only
		"options.txt",   // Client-only
		"optionsof.txt", // OptiFine client settings
		// Note: mods directory and .pw.toml files are needed by packwiz installer
	}

	// Allow mod metadata files (needed by packwiz installer)
	if strings.HasSuffix(relPath, ".pw.toml") {
		return false
	}

	// Skip actual mod jar files (will be downloaded by installer with server filtering)
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

// createServerIcon creates a server-icon.png from the pack icon
func createServerIcon(packLocation, serverDir string) error {
	iconPath := filepath.Join(packLocation, "icon.png")
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		return nil // No icon to copy
	}

	destPath := filepath.Join(serverDir, "server-icon.png")
	return copyFile(iconPath, destPath)
}

// installServerMods uses packwiz installer to download server-side mods
func installServerMods(serverDir, packLocation string) error {
	fmt.Println("📦 Installing server-side mods...")

	// Find packwiz installer
	installerPath := filepath.Join(packLocation, "packwiz-installer-bootstrap.jar")
	if _, err := os.Stat(installerPath); os.IsNotExist(err) {
		// Try to download it
		fmt.Println("⬇️  Downloading packwiz installer...")
		if err := DownloadPackwizInstaller(installerPath); err != nil {
			return fmt.Errorf("failed to download packwiz installer: %w", err)
		}
	}

	// Copy pack.toml to server directory
	packTomlSrc := filepath.Join(packLocation, "pack.toml")
	packTomlDest := filepath.Join(serverDir, "pack.toml")
	if err := copyFile(packTomlSrc, packTomlDest); err != nil {
		return fmt.Errorf("failed to copy pack.toml: %w", err)
	}

	// Copy index.toml to server directory
	indexTomlSrc := filepath.Join(packLocation, "index.toml")
	indexTomlDest := filepath.Join(serverDir, "index.toml")
	if err := copyFile(indexTomlSrc, indexTomlDest); err != nil {
		return fmt.Errorf("failed to copy index.toml: %w", err)
	}

	// Copy installer to server directory
	serverInstallerPath := filepath.Join(serverDir, "packwiz-installer-bootstrap.jar")
	if err := copyFile(installerPath, serverInstallerPath); err != nil {
		return fmt.Errorf("failed to copy installer: %w", err)
	}

	// Find compatible Java for the pack
	packToml, _, err := utils.LoadPackConfig(packLocation)
	if err != nil {
		return fmt.Errorf("failed to load pack config: %w", err)
	}

	mcVersion := getMinecraftVersionServer(packToml)
	java, err := utils.FindCompatibleJava(mcVersion)
	if err != nil {
		return fmt.Errorf("no compatible Java found for Minecraft %s: %w", mcVersion, err)
	}

	fmt.Printf("Using Java %s for mod installation\n", java.Version)

	// Run packwiz installer with server flag and no-gui mode
	cmd := exec.Command(java.Path, "-jar", "packwiz-installer-bootstrap.jar", "pack.toml", "-s", "server", "-g")
	cmd.Dir = serverDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Don't attach stdin to prevent hanging on input prompts
	cmd.Stdin = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("packwiz installer failed: %w", err)
	}

	fmt.Println("✅ Server mods installed successfully")
	return nil
}

// createServerScripts creates startup scripts for the server
func createServerScripts(serverDir string) error {
	// Create start.bat for Windows
	batScript := `@echo off
title Minecraft Server
echo Starting Minecraft Server...
java -Xmx4G -Xms1G -jar server.jar nogui
pause`

	if err := os.WriteFile(filepath.Join(serverDir, "start.bat"), []byte(batScript), 0644); err != nil {
		return fmt.Errorf("failed to create start.bat: %w", err)
	}

	// Create start.sh for Unix systems
	shScript := `#!/bin/bash
echo "Starting Minecraft Server..."
java -Xmx4G -Xms1G -jar server.jar nogui`

	if err := os.WriteFile(filepath.Join(serverDir, "start.sh"), []byte(shScript), 0755); err != nil {
		return fmt.Errorf("failed to create start.sh: %w", err)
	}

	// Create eula.txt (user needs to accept EULA)
	eulaContent := `# By changing the setting below to TRUE you are indicating your agreement to our EULA (https://aka.ms/MinecraftEULA).
# You must accept the EULA to run the server.
eula=false`

	if err := os.WriteFile(filepath.Join(serverDir, "eula.txt"), []byte(eulaContent), 0644); err != nil {
		return fmt.Errorf("failed to create eula.txt: %w", err)
	}

	// Create server.properties with basic settings
	serverProps := `# Minecraft server properties
server-port=25565
gamemode=survival
difficulty=normal
max-players=20
motd=A Minecraft Server
online-mode=true
spawn-protection=16
level-name=world
level-type=minecraft\:normal`

	if err := os.WriteFile(filepath.Join(serverDir, "server.properties"), []byte(serverProps), 0644); err != nil {
		return fmt.Errorf("failed to create server.properties: %w", err)
	}

	return nil
}

// getMinecraftVersionServer gets the Minecraft version from pack.toml
func getMinecraftVersionServer(packToml *packwiz.PackToml) string {
	// Prefer versions.minecraft, fallback to mc-version
	if packToml.Versions.Minecraft != "" {
		return packToml.Versions.Minecraft
	}
	return packToml.McVersion
}
