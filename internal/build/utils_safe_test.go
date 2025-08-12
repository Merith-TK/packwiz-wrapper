package build

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateZipFromDir(t *testing.T) {
	// Create test directory with files
	srcDir := t.TempDir()
	testFile := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create zip
	dstDir := t.TempDir()
	zipPath := filepath.Join(dstDir, "test.zip")

	err := CreateZipFromDir(srcDir, zipPath)
	if err != nil {
		t.Fatalf("createZipFromDir failed: %v", err)
	}

	// Verify zip exists
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		t.Error("Zip file was not created")
	}

	// Verify zip contains our file
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("Failed to open zip: %v", err)
	}
	defer zipReader.Close()

	found := false
	for _, file := range zipReader.File {
		if strings.Contains(file.Name, "test.txt") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Test file not found in zip")
	}
}

func TestSanitizeIconName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"icon.png", "icon.png"},
		{"icon with spaces.png", "icon_with_spaces.png"},
		{"icon/with/slashes.png", "icon_with_slashes.png"},
		{"icon\\with\\backslashes.png", "icon_with_backslashes.png"},
		{"icon:with:colons.png", "icon_with_colons.png"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeIconName(tt.input)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	// Create source file
	srcDir := t.TempDir()
	srcFile := filepath.Join(srcDir, "source.txt")
	content := "test content"
	if err := os.WriteFile(srcFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy to destination
	dstDir := t.TempDir()
	dstFile := filepath.Join(dstDir, "destination.txt")

	err := copyFile(srcFile, dstFile)
	if err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	// Verify copy
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(dstContent) != content {
		t.Errorf("Content mismatch: expected '%s', got '%s'", content, string(dstContent))
	}
}

func TestCopyFileNonexistent(t *testing.T) {
	// Test copying nonexistent file
	srcFile := filepath.Join(t.TempDir(), "nonexistent.txt")
	dstFile := filepath.Join(t.TempDir(), "destination.txt")

	err := copyFile(srcFile, dstFile)
	if err == nil {
		t.Error("Expected error when copying nonexistent file")
	}
}
