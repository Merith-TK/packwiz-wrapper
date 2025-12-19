// Package java provides Java version management utilities for PackWrap.
//
// Shared types and constants.
package java

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Supported Java versions
const (
	Java8  = 8
	Java17 = 17
	Java21 = 21
)

var SupportedJavaVersions = []int{Java8, Java17, Java21}

// Adoptium repository mappings for downloading Java
var adoptiumRepos = map[int]string{
	Java8:  "adoptium/temurin8-binaries",
	Java17: "adoptium/temurin17-binaries",
	Java21: "adoptium/temurin21-binaries",
}

// ProgressCallback is a function type for reporting progress during operations
type ProgressCallback func(message string)

// JavaVersion represents a detected Java installation
type JavaVersion struct {
	Path    string // Path to java executable
	Version string // Version string (e.g., "21.0.1")
	Major   int    // Major version number (e.g., 21)
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

// getJavaExecutablePath returns the path to the java executable for a given installation directory
func getJavaExecutablePath(javaDir string) string {
	javaExe := filepath.Join(javaDir, "bin", "java")
	if runtime.GOOS == "windows" {
		javaExe = filepath.Join(javaDir, "bin", "java.exe")
	}
	return javaExe
}

// getManagedJavaPath returns the installation directory for a managed Java version
func getManagedJavaPath(majorVersion int) string {
	dataDir := getDataDirectory()
	return filepath.Join(dataDir, "java", fmt.Sprintf("java-%d", majorVersion))
}

// getDataDirectory returns the platform-specific data directory for storing Java installations
func getDataDirectory() string {
	var dataDir string

	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			dataDir = filepath.Join(appData, "PackWrap")
		} else {
			dataDir = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "PackWrap")
		}
	case "darwin":
		if home := os.Getenv("HOME"); home != "" {
			dataDir = filepath.Join(home, "Library", "Application Support", "PackWrap")
		}
	case "linux":
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			dataDir = filepath.Join(xdgData, "packwrap")
		} else if home := os.Getenv("HOME"); home != "" {
			dataDir = filepath.Join(home, ".local", "share", "packwrap")
		}
	}

	if dataDir == "" {
		// Fallback to current directory
		dataDir = ".packwrap"
	}

	return dataDir
}
