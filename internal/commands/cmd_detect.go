package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Merith-TK/packwiz-wrapper/internal/utils"
)

// CmdDetect provides pack URL detection from git remotes
func CmdDetect() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"detect", "detect-url", "url"},
		"Detect pack URL from git remote and branch",
		`Detect Commands:
  pw detect               - Show remote pack URL
  pw detect local         - Show local pack path
  pw detect --local       - Show local pack path (flag version)

Examples:
  pw detect               - Get raw.githubusercontent.com URL
  pw detect local         - Get local filesystem path`,
		func(args []string) error {
			localPath := false

			// Check for local flag
			for _, arg := range args {
				if arg == "local" || arg == "--local" {
					localPath = true
					break
				}
			}

			return detectPackURL(localPath)
		}
}

func detectPackURL(localPath bool) error {
	packDir, _ := os.Getwd()

	// Find pack.toml location
	packLocation := utils.FindPackToml(packDir)
	if packLocation == "" {
		return fmt.Errorf("pack.toml not found")
	}

	if localPath {
		// Return absolute local path
		packTomlPath := filepath.Join(packLocation, "pack.toml")
		absPath, err := filepath.Abs(packTomlPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}
		absPath = filepath.ToSlash(absPath)
		absPath = strings.Replace(absPath, "/..minecraft", "/.minecraft", 1)
		fmt.Println(absPath)
		return nil
	}

	// Use shared remote URL detection
	remoteURL, err := utils.DetectRemotePackURL(packLocation)
	if err != nil {
		return fmt.Errorf("failed to detect remote URL: %w", err)
	}

	fmt.Println(remoteURL)
	return nil
}
