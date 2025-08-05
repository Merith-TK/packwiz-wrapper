package utils

import (
	"os"
	"path/filepath"
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
