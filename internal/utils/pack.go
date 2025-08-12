package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	// Use the centralized pack.toml finding logic
	packLocation := FindPackToml(packDir)
	if packLocation == "" {
		return nil, "", fmt.Errorf("pack.toml not found in current directory, .minecraft subdirectory, or parent directories")
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

// DetectRemotePackURL tries to detect the remote pack URL from git
// Returns the raw URL to pack.toml in the remote repository
func DetectRemotePackURL(packLocation string) (string, error) {
	// Change to pack directory for git commands
	originalDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(packLocation); err != nil {
		return "", err
	}

	// Get git remote URL
	remote, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git remote: %w", err)
	}
	remoteString := strings.TrimSpace(string(remote))
	remoteString = strings.TrimSuffix(remoteString, ".git")

	// Get current branch
	branch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		// Fallback: try GITHUB_HEAD_REF environment variable (GitHub Actions)
		if envBranch := os.Getenv("GITHUB_HEAD_REF"); envBranch != "" {
			branch = []byte(envBranch)
		} else {
			return "", fmt.Errorf("failed to get git branch: %w", err)
		}
	}
	branchString := strings.TrimSpace(string(branch))

	// Get relative path to pack.toml from git root
	gitRoot, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git root: %w", err)
	}
	gitRootString := strings.TrimSpace(string(gitRoot))

	relPath, err := filepath.Rel(gitRootString, filepath.Join(packLocation, "pack.toml"))
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}
	relPath = strings.ReplaceAll(relPath, "\\", "/") // Normalize path separators for URLs

	// Parse remote URL to determine the hosting service
	var urlString string
	if strings.Contains(remoteString, "github.com") {
		// Convert github.com URL to raw.githubusercontent.com
		remoteString = strings.Replace(remoteString, "github.com", "raw.githubusercontent.com", 1)
		urlString = strings.Join([]string{remoteString, branchString, relPath}, "/")
	} else if strings.Contains(remoteString, "gitlab.com") {
		// GitLab raw URL format
		urlString = remoteString + "/-/raw/" + branchString + "/" + relPath
	} else {
		// Generic git hosting (Gitea, etc.)
		urlString = remoteString + "/raw/branch/" + branchString + "/" + relPath
	}

	return urlString, nil
}
