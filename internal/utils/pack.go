package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/Merith-TK/packwiz-wrapper/internal/packwiz"
)

// FindPackToml finds the pack.toml file in the given directory or its parents.
// It checks both the current directory and .minecraft subdirectory (common modpack pattern).
// Returns the directory containing pack.toml, or empty string if not found.
func FindPackToml(startDir string) string {
	dir := startDir
	for {
		// Check current directory
		packTomlPath := filepath.Join(dir, "pack.toml")
		if _, err := os.Stat(packTomlPath); err == nil {
			return dir
		}

		// Check .minecraft subdirectory (common modpack pattern)
		minecraftDir := filepath.Join(dir, ".minecraft")
		packTomlPath = filepath.Join(minecraftDir, "pack.toml")
		if _, err := os.Stat(packTomlPath); err == nil {
			return minecraftDir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}
	return ""
}

// LoadPackConfig loads pack configuration from the current directory or parent directories
func LoadPackConfig(packDir string) (*packwiz.PackToml, string, error) {
	// Find pack.toml in current or parent directories
	packLocation := packDir
	for {
		packTomlPath := filepath.Join(packLocation, "pack.toml")
		if _, err := os.Stat(packTomlPath); err == nil {
			break
		}

		parent := filepath.Dir(packLocation)
		if parent == packLocation {
			return nil, "", fmt.Errorf("pack.toml not found in current directory or parent directories")
		}
		packLocation = parent
	}

	packTomlPath := filepath.Join(packLocation, "pack.toml")
	data, err := os.ReadFile(packTomlPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read pack.toml: %w", err)
	}

	var packToml packwiz.PackToml
	if err := toml.Unmarshal(data, &packToml); err != nil {
		return nil, "", fmt.Errorf("failed to parse pack.toml: %w", err)
	}

	return &packToml, packLocation, nil
}
