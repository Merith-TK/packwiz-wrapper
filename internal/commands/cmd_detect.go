package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	packLocation := findPackToml(packDir)
	if packLocation == "" {
		return fmt.Errorf("pack.toml not found")
	}

	// Convert to relative path for URL construction
	relPath := strings.TrimPrefix(packLocation, packDir)
	relPath = strings.TrimPrefix(relPath, string(filepath.Separator))
	relPath = filepath.ToSlash(relPath)
	if !strings.HasSuffix(relPath, "/") && relPath != "" {
		relPath = relPath + "/"
	}
	relPath = relPath + "pack.toml"

	if localPath {
		// Return absolute local path
		absPath, err := filepath.Abs(filepath.Join(packDir, relPath))
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}
		absPath = filepath.ToSlash(absPath)
		absPath = strings.Replace(absPath, "/..minecraft", "/.minecraft", 1)
		fmt.Println(absPath)
		return nil
	}

	// Get git remote URL
	remote, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return fmt.Errorf("failed to get git remote: %w", err)
	}
	remoteString := strings.TrimSpace(string(remote))
	remoteString = strings.TrimSuffix(remoteString, ".git")

	// Get current branch
	branch, err := exec.Command("git", "branch", "--show-current").Output()
	if err != nil {
		return fmt.Errorf("failed to get git branch: %w", err)
	}
	branchString := strings.TrimSpace(string(branch))

	// If branch is empty, try GITHUB_HEAD_REF environment variable
	if branchString == "" {
		if envBranch := os.Getenv("GITHUB_HEAD_REF"); envBranch != "" {
			branchString = envBranch
		} else {
			return fmt.Errorf("could not determine branch name")
		}
	}

	// Parse remote URL to determine the hosting service
	var urlString string
	if strings.Contains(remoteString, "github.com") {
		// GitHub: use raw.githubusercontent.com
		remoteString = strings.Replace(remoteString, "github.com", "raw.githubusercontent.com", 1)
		urlString = strings.Join([]string{remoteString, branchString, relPath}, "/")
	} else if strings.Contains(remoteString, "gitlab.com") {
		// GitLab: use /-/raw/ pattern
		urlString = remoteString + "/-/raw/" + branchString + "/" + relPath
	} else {
		// Default (Gitea/other): use /raw/branch/ pattern
		urlString = remoteString + "/raw/branch/" + branchString + "/" + relPath
	}

	fmt.Println(urlString)
	return nil
}
