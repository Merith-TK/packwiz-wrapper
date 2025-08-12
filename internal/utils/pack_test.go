package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindPackToml(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Test case 1: pack.toml in current directory
	packTomlPath := filepath.Join(tmpDir, "pack.toml")
	if err := os.WriteFile(packTomlPath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test pack.toml: %v", err)
	}

	result := FindPackToml(tmpDir)
	if result != tmpDir {
		t.Errorf("Expected to find pack.toml in %s, got %s", tmpDir, result)
	}

	// Test case 2: pack.toml in .minecraft subdirectory
	tmpDir2 := t.TempDir()
	minecraftDir := filepath.Join(tmpDir2, ".minecraft")
	if err := os.MkdirAll(minecraftDir, 0755); err != nil {
		t.Fatalf("Failed to create .minecraft directory: %v", err)
	}

	packTomlPath2 := filepath.Join(minecraftDir, "pack.toml")
	if err := os.WriteFile(packTomlPath2, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test pack.toml in .minecraft: %v", err)
	}

	result2 := FindPackToml(tmpDir2)
	if result2 != minecraftDir {
		t.Errorf("Expected to find pack.toml in %s, got %s", minecraftDir, result2)
	}

	// Test case 3: No pack.toml found
	tmpDir3 := t.TempDir()
	result3 := FindPackToml(tmpDir3)
	if result3 != "" {
		t.Errorf("Expected empty string when no pack.toml found, got %s", result3)
	}
}

func TestLoadPackConfig(t *testing.T) {
	// Create a temporary directory with a valid pack.toml
	tmpDir := t.TempDir()
	packTomlContent := `name = "Test Pack"
author = "Test Author"
version = "1.0.0"
pack-format = "packwiz:1.1.0"

[index]
file = "index.toml"
hash-format = "sha256"

[versions]
minecraft = "1.20.1"
`

	packTomlPath := filepath.Join(tmpDir, "pack.toml")
	if err := os.WriteFile(packTomlPath, []byte(packTomlContent), 0644); err != nil {
		t.Fatalf("Failed to create test pack.toml: %v", err)
	}

	packToml, location, err := LoadPackConfig(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load pack config: %v", err)
	}

	if location != tmpDir {
		t.Errorf("Expected location %s, got %s", tmpDir, location)
	}

	if packToml.Name != "Test Pack" {
		t.Errorf("Expected pack name 'Test Pack', got '%s'", packToml.Name)
	}

	if packToml.Author != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%s'", packToml.Author)
	}

	if packToml.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", packToml.Version)
	}
}

func TestLoadPackConfigNotFound(t *testing.T) {
	// Test with directory that has no pack.toml
	tmpDir := t.TempDir()

	_, _, err := LoadPackConfig(tmpDir)
	if err == nil {
		t.Error("Expected error when pack.toml not found, got nil")
	}

	if !strings.Contains(err.Error(), "pack.toml not found") {
		t.Errorf("Expected 'pack.toml not found' error, got: %v", err)
	}
}

func TestLoadPackConfigInvalidToml(t *testing.T) {
	// Test with invalid TOML content
	tmpDir := t.TempDir()
	packTomlContent := `name = "Invalid TOML"
[invalid toml syntax
`

	packTomlPath := filepath.Join(tmpDir, "pack.toml")
	if err := os.WriteFile(packTomlPath, []byte(packTomlContent), 0644); err != nil {
		t.Fatalf("Failed to create invalid pack.toml: %v", err)
	}

	_, _, err := LoadPackConfig(tmpDir)
	if err == nil {
		t.Error("Expected error when pack.toml is invalid, got nil")
	}

	if !strings.Contains(err.Error(), "failed to parse pack.toml") {
		t.Errorf("Expected 'failed to parse pack.toml' error, got: %v", err)
	}
}

func TestDetectRemotePackURLNoGit(t *testing.T) {
	// Test with directory that's not a git repo
	tmpDir := t.TempDir()

	_, err := DetectRemotePackURL(tmpDir)
	if err == nil {
		t.Error("Expected error when not in git repo, got nil")
	}
}

// Helper function to run git commands in tests
func runCmd(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.Run()
}
