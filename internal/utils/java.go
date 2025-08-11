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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/Merith-TK/utils/pkg/archive"
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
// PUBLIC API FUNCTIONS
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

// GetStrictJavaVersion returns the strict minimum Java version
func GetStrictJavaVersion(mcVersion string) int {
	version := parseMinecraftVersion(mcVersion)

	if version.Compare("1.20.5") >= 0 {
		return Java21 // Java 21 required for 1.20.5+
	} else if version.Compare("1.17") >= 0 {
		return Java17 // Java 17 required for 1.17+
	} else {
		return Java8 // Java 8 minimum for everything else
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
	strict := GetStrictJavaVersion(mcVersion)

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

// EnsureJava ensures a compatible Java version is available, downloading if necessary
func EnsureJava(mcVersion string) (*JavaVersion, error) {
	return EnsureJavaWithProgress(mcVersion, nil)
}

// EnsureJavaWithProgress ensures a compatible Java version is available with progress reporting
func EnsureJavaWithProgress(mcVersion string, progress ProgressCallback) (*JavaVersion, error) {
	// First try to find existing Java
	if java, err := FindCompatibleJava(mcVersion); err == nil {
		return java, nil
	}

	reportProgress := func(msg string) {
		if progress != nil {
			progress(msg)
		} else {
			fmt.Printf(msg + "\n")
		}
	}

	// No compatible Java found, try to download it
	requiredVersion := GetRequiredJavaVersion(mcVersion)
	reportProgress(fmt.Sprintf("No compatible Java found for Minecraft %s", mcVersion))
	reportProgress(fmt.Sprintf("Downloading Java %d...", requiredVersion))

	javaPath, err := DownloadAndInstallJavaWithProgress(requiredVersion, progress)
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
// JAVA INSTALLATION AND DOWNLOAD
// ============================================================================

// DownloadAndInstallJava downloads and extracts a Java runtime
func DownloadAndInstallJava(majorVersion int) (string, error) {
	return DownloadAndInstallJavaWithProgress(majorVersion, nil)
}

// DownloadAndInstallJavaWithProgress downloads and extracts a Java runtime with progress reporting
func DownloadAndInstallJavaWithProgress(majorVersion int, progress ProgressCallback) (string, error) {
	if !IsValidJavaVersion(majorVersion) {
		return "", fmt.Errorf("unsupported Java version: %d (supported: %v)", majorVersion, SupportedJavaVersions)
	}

	reportProgress := func(msg string) {
		if progress != nil {
			progress(msg)
		} else {
			fmt.Println(msg)
		}
	}

	// Get storage directory
	dataDir := getDataDirectory()
	javaDir := filepath.Join(dataDir, "java")

	if err := os.MkdirAll(javaDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create Java directory: %w", err)
	}

	// Check if already installed
	versionDir := getManagedJavaPath(majorVersion)
	if _, err := os.Stat(versionDir); err == nil {
		return versionDir, nil
	}

	reportProgress(fmt.Sprintf("Getting download information for Java %d...", majorVersion))

	// Download Java
	downloadURL, filename, err := getJavaDownloadURL(majorVersion)
	if err != nil {
		return "", fmt.Errorf("failed to get download URL: %w", err)
	}

	zipPath := filepath.Join(javaDir, filename)
	reportProgress("Downloading Java...")
	if err := downloadFileWithProgress(downloadURL, zipPath, progress); err != nil {
		return "", fmt.Errorf("failed to download Java: %w", err)
	}

	reportProgress("Extracting Java...")
	// Extract Java
	if err := extractJavaZipWithProgress(zipPath, versionDir, progress); err != nil {
		os.Remove(zipPath) // Clean up on failure
		return "", fmt.Errorf("failed to extract Java: %w", err)
	}

	// Clean up zip file
	os.Remove(zipPath)

	reportProgress("Java installation completed successfully!")
	return versionDir, nil
}

// getJavaDownloadURL constructs the download URL for the specified Java version
func getJavaDownloadURL(majorVersion int) (string, string, error) {
	if !IsValidJavaVersion(majorVersion) {
		return "", "", fmt.Errorf("unsupported Java version: %d (supported: %v)", majorVersion, SupportedJavaVersions)
	}

	repo := adoptiumRepos[majorVersion]
	arch := getArchitecture()
	hostOS := getHostOS()

	// Get latest release info from GitHub API
	releaseURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	resp, err := http.Get(releaseURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", fmt.Errorf("failed to parse release info: %w", err)
	}

	// Find the appropriate asset
	pattern := fmt.Sprintf("OpenJDK%dU-jre_%s_%s_hotspot_.*\\.zip", majorVersion, arch, hostOS)
	regex := regexp.MustCompile(pattern)

	for _, asset := range release.Assets {
		if regex.MatchString(asset.Name) {
			return asset.BrowserDownloadURL, asset.Name, nil
		}
	}

	return "", "", fmt.Errorf("no matching Java %d asset found for %s %s", majorVersion, hostOS, arch)
}

// getArchitecture returns the architecture string for download URLs
func getArchitecture() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x64"
	case "386":
		return "x86-32"
	case "arm":
		return "arm"
	case "arm64":
		return "aarch64"
	default:
		return "x64" // Default fallback
	}
}

// getHostOS returns the host OS string for download URLs
func getHostOS() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "linux":
		return "linux"
	case "darwin":
		return "mac" // Note: might need adjustment based on actual release naming
	default:
		return "windows" // Default fallback
	}
}

// downloadFileWithProgress downloads a file from URL to the specified path with progress reporting
func downloadFileWithProgress(url, filepath string, progress ProgressCallback) error {
	reportProgress := func(msg string) {
		if progress != nil {
			progress(msg)
		} else {
			fmt.Printf(msg + "\n")
		}
	}

	reportProgress(fmt.Sprintf("Downloading from: %s", url))

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	reportProgress(fmt.Sprintf("Downloaded: %s", filepath))
	return nil
}

// extractJavaZipWithProgress extracts a Java JRE zip file with progress reporting
func extractJavaZipWithProgress(zipPath, extractDir string, progress ProgressCallback) error {
	reportProgress := func(msg string) {
		if progress != nil {
			progress(msg)
		} else {
			fmt.Printf(msg + "\n")
		}
	}

	reportProgress(fmt.Sprintf("Extracting Java to: %s", extractDir))

	// Create a temporary directory for extraction
	tempDir := extractDir + "-temp"
	defer os.RemoveAll(tempDir) // Clean up temp directory

	// Extract to temporary directory first
	if err := archive.Unzip(zipPath, tempDir); err != nil {
		return fmt.Errorf("failed to extract zip: %w", err)
	}

	// Find the Java installation directory (should be the only subdirectory)
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return fmt.Errorf("failed to read temp directory: %w", err)
	}

	var javaDir string
	for _, entry := range entries {
		if entry.IsDir() {
			javaDir = filepath.Join(tempDir, entry.Name())
			break
		}
	}

	if javaDir == "" {
		return fmt.Errorf("no Java directory found in archive")
	}

	// Move the Java directory contents to the final location
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("failed to create extract directory: %w", err)
	}

	// Move all contents from javaDir to extractDir
	javaEntries, err := os.ReadDir(javaDir)
	if err != nil {
		return fmt.Errorf("failed to read Java directory: %w", err)
	}

	for _, entry := range javaEntries {
		src := filepath.Join(javaDir, entry.Name())
		dst := filepath.Join(extractDir, entry.Name())
		if err := os.Rename(src, dst); err != nil {
			return fmt.Errorf("failed to move %s to %s: %w", src, dst, err)
		}
	}

	reportProgress("✅ Java extracted successfully")
	return nil
}

// findManagedJavaInstallations finds Java installations managed by this tool
func findManagedJavaInstallations() ([]JavaVersion, error) {
	var installations []JavaVersion

	// Get storage directory
	dataDir := getDataDirectory()
	javaDir := filepath.Join(dataDir, "java")

	if _, err := os.Stat(javaDir); os.IsNotExist(err) {
		return installations, nil // No managed installations
	}

	entries, err := os.ReadDir(javaDir)
	if err != nil {
		return installations, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if this looks like a Java installation
		installDir := filepath.Join(javaDir, entry.Name())
		javaExe := getJavaExecutablePath(installDir)

		if version, err := DetectJavaVersion(javaExe); err == nil {
			installations = append(installations, version)
		}
	}

	return installations, nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// getDataDirectory returns the data directory for storing managed Java installations
// Uses Fyne-compatible location for consistency
func getDataDirectory() string {
	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "xyz.merith.packwrap")
		}
		return filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "xyz.merith.packwrap")
	case "darwin":
		if home := os.Getenv("HOME"); home != "" {
			return filepath.Join(home, "Library", "Application Support", "xyz.merith.packwrap")
		}
	case "linux":
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			return filepath.Join(xdgData, "xyz.merith.packwrap")
		}
		if home := os.Getenv("HOME"); home != "" {
			return filepath.Join(home, ".local", "share", "xyz.merith.packwrap")
		}
	}

	// Fallback to current directory
	return filepath.Join(".", ".packwrap-data")
}

// detectJavaVersion detects the version of a Java executable
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
