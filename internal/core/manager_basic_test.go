package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewManagerBasic(t *testing.T) {
	manager := NewManager(nil)
	if manager == nil {
		t.Fatal("NewManager should not return nil")
	}
}

func TestManagerGetPackInfoSimple(t *testing.T) {
	// Create temporary directory with simple pack.toml
	tmpDir, err := os.MkdirTemp("", "manager_simple")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create minimal pack.toml
	packToml := `name = "Test"
pack-format = "packwiz:1.1.0"
`

	err = os.WriteFile(filepath.Join(tmpDir, "pack.toml"), []byte(packToml), 0644)
	if err != nil {
		t.Fatalf("Failed to write pack.toml: %v", err)
	}

	// Test GetPackInfo
	manager := NewManager(nil)
	packInfo, err := manager.GetPackInfo(tmpDir)
	if err != nil {
		t.Fatalf("GetPackInfo failed: %v", err)
	}

	if packInfo == nil {
		t.Fatal("PackInfo should not be nil")
	}

	if packInfo.Name != "Test" {
		t.Errorf("Expected name 'Test', got '%s'", packInfo.Name)
	}
}

func TestManagerGetPackInfoMissing(t *testing.T) {
	// Test with empty directory
	tmpDir, err := os.MkdirTemp("", "manager_empty")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	manager := NewManager(nil)
	packInfo, err := manager.GetPackInfo(tmpDir)

	if err == nil {
		t.Error("Should return error for missing pack.toml")
	}

	if packInfo != nil {
		t.Error("PackInfo should be nil for missing pack")
	}
}

func TestNoOpLoggerBasic(t *testing.T) {
	logger := &NoOpLogger{}

	// These should not panic
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")
	logger.Debug("test")
}
