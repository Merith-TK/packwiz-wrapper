// Package utils provides Java version management utilities for PackWrap.
//
// Main functionalities:
// - Java version detection and validation
// - Automatic download and installation of Java from Adoptium
// - Minecraft version to Java version mapping
// - Managed Java installation discovery
//
// Supported Java versions: 8, 17, 21
package utils

import (
	"github.com/Merith-TK/packwiz-wrapper/internal/utils/java"
)

// Re-export types and functions from the java subpackage for backward compatibility

// Supported Java versions
const (
	Java8  = java.Java8
	Java17 = java.Java17
	Java21 = java.Java21
)

var SupportedJavaVersions = java.SupportedJavaVersions

// ProgressCallback is a function type for reporting progress during operations
type ProgressCallback = java.ProgressCallback

// JavaVersion represents a detected Java installation
type JavaVersion = java.JavaVersion

// ============================================================================
// PUBLIC API FUNCTIONS
// ============================================================================

// IsValidJavaVersion checks if a Java version is supported
func IsValidJavaVersion(version int) bool {
	return java.IsValidJavaVersion(version)
}

// GetRequiredJavaVersion returns the required Java version for a given Minecraft version
func GetRequiredJavaVersion(mcVersion string) int {
	return java.GetRequiredJavaVersion(mcVersion)
}

// GetStrictJavaVersion returns the strict minimum Java version
func GetStrictJavaVersion(mcVersion string) int {
	return java.GetStrictJavaVersion(mcVersion)
}

// FindJavaInstallations finds all available Java installations
func FindJavaInstallations() ([]JavaVersion, error) {
	return java.FindJavaInstallations()
}

// FindCompatibleJava finds a Java installation compatible with the given Minecraft version
func FindCompatibleJava(mcVersion string) (*JavaVersion, error) {
	return java.FindCompatibleJava(mcVersion)
}

// ValidateJava checks if Java is available and compatible
func ValidateJava(mcVersion string) error {
	return java.ValidateJava(mcVersion)
}

// EnsureJava ensures a compatible Java version is available, downloading if necessary
func EnsureJava(mcVersion string) (*JavaVersion, error) {
	return java.EnsureJava(mcVersion)
}

// EnsureJavaWithProgress ensures a compatible Java version is available with progress reporting
func EnsureJavaWithProgress(mcVersion string, progress ProgressCallback) (*JavaVersion, error) {
	return java.EnsureJavaWithProgress(mcVersion, progress)
}

// DownloadAndInstallJava downloads and extracts a Java runtime
func DownloadAndInstallJava(majorVersion int) (string, error) {
	return java.DownloadAndInstallJava(majorVersion)
}

// DownloadAndInstallJavaWithProgress downloads and extracts a Java runtime with progress reporting
func DownloadAndInstallJavaWithProgress(majorVersion int, progress ProgressCallback) (string, error) {
	return java.DownloadAndInstallJavaWithProgress(majorVersion, progress)
}

// DetectJavaVersion detects the version of a Java executable
func DetectJavaVersion(javaCmd string) (JavaVersion, error) {
	return java.DetectJavaVersion(javaCmd)
}
