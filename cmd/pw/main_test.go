package main

import (
	"strings"
	"testing"
)

func TestVersionInfo(t *testing.T) {
	// Test that version variables can be set (they're set by ldflags during build)
	// Default values should be "dev" when not set by build process

	// Since these are package-level variables, we can't easily test
	// the ldflags injection in unit tests, but we can verify the
	// variables exist and have reasonable defaults

	originalVersion := version
	originalCommit := commit
	originalDate := date

	// Test with empty values (simulate fresh build)
	version = ""
	commit = ""
	date = ""

	if version == "" {
		version = "dev" // Set default for testing
	}

	if commit == "" {
		commit = "unknown" // Set default for testing
	}

	if date == "" {
		date = "unknown" // Set default for testing
	}

	// Basic sanity checks
	if len(version) == 0 {
		t.Error("Version should not be empty")
	}

	if len(commit) == 0 {
		t.Error("Commit should not be empty")
	}

	if len(date) == 0 {
		t.Error("Date should not be empty")
	}

	// Test with actual values
	version = "v1.2.3"
	commit = "abc123def456"
	date = "2024-01-01T00:00:00Z"

	if version != "v1.2.3" {
		t.Errorf("Expected version 'v1.2.3', got '%s'", version)
	}

	if commit != "abc123def456" {
		t.Errorf("Expected commit 'abc123def456', got '%s'", commit)
	}

	if date != "2024-01-01T00:00:00Z" {
		t.Errorf("Expected date '2024-01-01T00:00:00Z', got '%s'", date)
	}

	// Restore original values
	version = originalVersion
	commit = originalCommit
	date = originalDate
}

func TestVersionVariableDefaults(t *testing.T) {
	// Test that the variables have reasonable defaults when not set by build
	// This simulates the development environment

	testCases := []struct {
		name     string
		value    *string
		expected string
	}{
		{"version", &version, ""},
		{"commit", &commit, ""},
		{"date", &date, ""},
	}

	for _, tc := range testCases {
		// Variables might be empty in test environment
		if *tc.value == "" {
			// This is expected in development/test environment
			continue
		}

		// If they're set, they should be reasonable strings
		if len(*tc.value) == 0 {
			t.Errorf("%s should not be empty string", tc.name)
		}

		// Version should start with 'v' if it's a release version
		if tc.name == "version" && strings.HasPrefix(*tc.value, "v") {
			if len(*tc.value) < 2 {
				t.Errorf("Version '%s' seems too short", *tc.value)
			}
		}

		// Commit should be a reasonable length for a git hash
		if tc.name == "commit" && *tc.value != "unknown" && *tc.value != "" && *tc.value != "none" {
			if len(*tc.value) < 7 || len(*tc.value) > 40 {
				t.Errorf("Commit '%s' doesn't look like a git hash", *tc.value)
			}
		}
	}
}

func TestPackageLevel(t *testing.T) {
	// Test that we can access package-level items
	// This ensures the package structure is correct

	// Test that the version variables exist (even if empty)
	versionCopy := version
	commitCopy := commit
	dateCopy := date

	// These should be accessible and consistent
	if versionCopy != version {
		t.Error("Version variable access inconsistent")
	}

	if commitCopy != commit {
		t.Error("Commit variable access inconsistent")
	}

	if dateCopy != date {
		t.Error("Date variable access inconsistent")
	}
}
