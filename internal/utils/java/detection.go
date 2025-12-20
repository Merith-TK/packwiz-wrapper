// Package java provides Java version management utilities for PackWrap.
//
// Detection and validation functionality.
package java

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// ============================================================================
// PUBLIC API FUNCTIONS
// ============================================================================

// IsValidJavaVersion checks if a Java version is supported
func IsValidJavaVersion(version int) bool {
	for _, v := range SupportedJavaVersions {
		if v == version {
			return true
		}
	}
	return false
}

// GetRequiredJavaVersion returns the required Java version for a given Minecraft version
func GetRequiredJavaVersion(mcVersion string) int {
	version := parseMinecraftVersion(mcVersion)

	if version.Compare("1.20.5") >= 0 {
		return Java21 // Java 21 for 1.20.5+
	} else if version.Compare("1.17") >= 0 {
		return Java17 // Java 17 for 1.17-1.20.4
	} else if version.Compare("1.13") >= 0 {
		return Java17 // Java 17 recommended for 1.13-1.16.5 (can use 8)
	} else {
		return Java8 // Java 8 for 1.12.2 and below
	}
}

// GetMinimumJavaVersion returns the strict minimum Java version
func GetMinimumJavaVersion(mcVersion string) int {
	version := parseMinecraftVersion(mcVersion)

	if version.Compare("1.17") >= 0 {
		return Java17 // Java 17 required for 1.17+
	} else {
		return Java8 // Java 8 for older versions
	}
}

// FindJavaInstallations finds all available Java installations
func FindJavaInstallations() ([]JavaVersion, error) {
	var installations []JavaVersion

	// Try common Java commands
	javaCmds := []string{"java", "java.exe"}

	for _, cmd := range javaCmds {
		if version, err := DetectJavaVersion(cmd); err == nil {
			installations = append(installations, version)
		}
	}

	// Check managed Java installations
	managedJavas, err := findManagedJavaInstallations()
	if err == nil {
		installations = append(installations, managedJavas...)
	}

	// TODO: Add registry scanning for Windows Java installations
	// TODO: Add common path scanning (/usr/lib/jvm, etc.)

	return installations, nil
}

// FindCompatibleJava finds a Java installation compatible with the given Minecraft version
func FindCompatibleJava(mcVersion string) (*JavaVersion, error) {
	required := GetRequiredJavaVersion(mcVersion)
	strict := GetMinimumJavaVersion(mcVersion)

	installations, err := FindJavaInstallations()
	if err != nil {
		return nil, fmt.Errorf("failed to find Java installations: %w", err)
	}

	if len(installations) == 0 {
		return nil, fmt.Errorf("no Java installations found")
	}

	// First try to find exact match
	for _, java := range installations {
		if java.Major == required {
			return &java, nil
		}
	}

	// Then try to find compatible version (>= strict minimum)
	for _, java := range installations {
		if java.Major >= strict {
			return &java, nil
		}
	}

	return nil, fmt.Errorf("no compatible Java found (need Java %d+, found: %v)",
		strict, getVersionList(installations))
}

// ValidateJava checks if Java is available and compatible
func ValidateJava(mcVersion string) error {
	java, err := FindCompatibleJava(mcVersion)
	if err != nil {
		return err
	}

	required := GetRequiredJavaVersion(mcVersion)
	if java.Major < required {
		return fmt.Errorf("java %d recommended for Minecraft %s (found Java %d)",
			required, mcVersion, java.Major)
	}

	return nil
}

// GetOrInstallJavaWithProgress ensures a compatible Java version is available with progress reporting
func GetOrInstallJavaWithProgress(mcVersion string, progress ProgressCallback) (*JavaVersion, error) {
	// First try to find existing Java
	if java, err := FindCompatibleJava(mcVersion); err == nil {
		return java, nil
	}

	reportProgress := func(msg string) {
		if progress != nil {
			progress(msg)
		} else {
			fmt.Println(msg)
		}
	}

	// No compatible Java found, try to download it
	requiredVersion := GetRequiredJavaVersion(mcVersion)
	reportProgress(fmt.Sprintf("No compatible Java found for Minecraft %s", mcVersion))
	reportProgress(fmt.Sprintf("Downloading Java %d...", requiredVersion))

	javaPath, err := InstallJavaWithProgress(requiredVersion, progress)
	if err != nil {
		return nil, fmt.Errorf("failed to download Java %d: %w", requiredVersion, err)
	}

	// Detect the downloaded Java version
	javaExe := getJavaExecutablePath(javaPath)

	java, err := DetectJavaVersion(javaExe)
	if err != nil {
		return nil, fmt.Errorf("failed to validate downloaded Java: %w", err)
	}

	reportProgress(fmt.Sprintf("✅ Java %d downloaded and ready at: %s", requiredVersion, javaPath))
	return &java, nil
}

// ============================================================================
// DETECTION FUNCTIONS
// ============================================================================

// findManagedJavaInstallations finds Java installations managed by this tool
func findManagedJavaInstallations() ([]JavaVersion, error) {
	var installations []JavaVersion

	for _, version := range SupportedJavaVersions {
		installPath := getManagedJavaPath(version)
		javaExe := getJavaExecutablePath(installPath)

		if version, err := DetectJavaVersion(javaExe); err == nil {
			installations = append(installations, version)
		}
	}

	return installations, nil
}

// DetectJavaVersion detects the version of a Java executable
func DetectJavaVersion(javaCmd string) (JavaVersion, error) {
	cmd := exec.Command(javaCmd, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return JavaVersion{}, fmt.Errorf("failed to run %s -version: %w", javaCmd, err)
	}

	// Parse version from output
	version := parseJavaVersionString(string(output))
	if version == "" {
		return JavaVersion{}, fmt.Errorf("could not parse Java version from output")
	}

	major := parseJavaMajorVersion(version)

	return JavaVersion{
		Path:    javaCmd,
		Version: version,
		Major:   major,
	}, nil
}

// parseJavaVersionString extracts version string from java -version output
func parseJavaVersionString(output string) string {
	// Look for version patterns like:
	// openjdk version "21.0.1" 2023-10-17
	// java version "1.8.0_391"

	re := regexp.MustCompile(`version "([^"]+)"`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// parseJavaMajorVersion extracts major version number from Java version string
func parseJavaMajorVersion(version string) int {
	// Handle both old (1.8.0_391) and new (21.0.1) formats
	if strings.HasPrefix(version, "1.") {
		// Old format: 1.8.0_391 -> 8
		return parseVersionPart(version, 1) // Get second part after "1."
	}

	// New format: 21.0.1 -> 21
	return parseVersionPart(version, 0) // Get first part
}

// parseVersionPart extracts a specific part of a version string
func parseVersionPart(version string, partIndex int) int {
	parts := strings.Split(version, ".")
	if len(parts) > partIndex {
		if major, err := strconv.Atoi(parts[partIndex]); err == nil {
			return major
		}
	}
	return 0
}

// getVersionList returns a list of major versions for error messages
func getVersionList(installations []JavaVersion) []int {
	var versions []int
	for _, java := range installations {
		versions = append(versions, java.Major)
	}
	return versions
}

// ============================================================================
// MINECRAFT VERSION HANDLING
// ============================================================================

// MinecraftVersion represents a parsed Minecraft version for comparison
type MinecraftVersion struct {
	Major int
	Minor int
	Patch int
	Raw   string
}

// parseMinecraftVersion parses a Minecraft version string
func parseMinecraftVersion(version string) MinecraftVersion {
	parts := strings.Split(version, ".")
	mv := MinecraftVersion{Raw: version}

	if len(parts) >= 1 {
		mv.Major, _ = strconv.Atoi(parts[0])
	}
	if len(parts) >= 2 {
		mv.Minor, _ = strconv.Atoi(parts[1])
	}
	if len(parts) >= 3 {
		mv.Patch, _ = strconv.Atoi(parts[2])
	}

	return mv
}

// Compare compares this version with another version string
// Returns: -1 if this < other, 0 if equal, 1 if this > other
func (mv MinecraftVersion) Compare(other string) int {
	otherVersion := parseMinecraftVersion(other)

	if mv.Major != otherVersion.Major {
		if mv.Major < otherVersion.Major {
			return -1
		}
		return 1
	}

	if mv.Minor != otherVersion.Minor {
		if mv.Minor < otherVersion.Minor {
			return -1
		}
		return 1
	}

	if mv.Patch != otherVersion.Patch {
		if mv.Patch < otherVersion.Patch {
			return -1
		}
		return 1
	}

	return 0
}
