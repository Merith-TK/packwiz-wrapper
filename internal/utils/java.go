package utils

import (
	"archive/zip"
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
)

// JavaVersion represents a detected Java installation
type JavaVersion struct {
	Path    string // Path to java executable
	Version string // Version string (e.g., "21.0.1")
	Major   int    // Major version number (e.g., 21)
}

// GetRequiredJavaVersion returns the required Java version for a given Minecraft version
func GetRequiredJavaVersion(mcVersion string) int {
	// Parse version to determine Java requirement
	version := parseMinecraftVersion(mcVersion)
	
	if version.Compare("1.20.5") >= 0 {
		return 21 // Java 21 for 1.20.5+
	} else if version.Compare("1.17") >= 0 {
		return 17 // Java 17 for 1.17-1.20.4
	} else if version.Compare("1.13") >= 0 {
		return 17 // Java 17 recommended for 1.13-1.16.5 (can use 8)
	} else {
		return 8 // Java 8 for 1.12.2 and below
	}
}

// GetStrictJavaVersion returns the strict minimum Java version
func GetStrictJavaVersion(mcVersion string) int {
	version := parseMinecraftVersion(mcVersion)
	
	if version.Compare("1.20.5") >= 0 {
		return 21 // Java 21 required for 1.20.5+
	} else if version.Compare("1.17") >= 0 {
		return 17 // Java 17 required for 1.17+
	} else {
		return 8 // Java 8 minimum for everything else
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
	// First try to find existing Java
	if java, err := FindCompatibleJava(mcVersion); err == nil {
		return java, nil
	}
	
	// No compatible Java found, try to download it
	requiredVersion := GetRequiredJavaVersion(mcVersion)
	fmt.Printf("No compatible Java found for Minecraft %s\n", mcVersion)
	fmt.Printf("Downloading Java %d...\n", requiredVersion)
	
	javaPath, err := DownloadAndInstallJava(requiredVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to download Java %d: %w", requiredVersion, err)
	}
	
	// Detect the downloaded Java version
	javaExe := filepath.Join(javaPath, "bin", "java")
	if runtime.GOOS == "windows" {
		javaExe = filepath.Join(javaPath, "bin", "java.exe")
	}
	
	java, err := DetectJavaVersion(javaExe)
	if err != nil {
		return nil, fmt.Errorf("failed to validate downloaded Java: %w", err)
	}
	
	fmt.Printf("✅ Java %d downloaded and ready at: %s\n", requiredVersion, javaPath)
	return &java, nil
}

// DownloadAndInstallJava downloads and extracts a Java runtime
func DownloadAndInstallJava(majorVersion int) (string, error) {
	// Get storage directory (Fyne data directory for consistency)
	dataDir := getDataDirectory()
	javaDir := filepath.Join(dataDir, "java")
	
	if err := os.MkdirAll(javaDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create Java directory: %w", err)
	}
	
	// Check if already installed
	versionDir := filepath.Join(javaDir, fmt.Sprintf("java-%d", majorVersion))
	if _, err := os.Stat(versionDir); err == nil {
		return versionDir, nil
	}
	
	// Download Java
	downloadURL, filename, err := getJavaDownloadURL(majorVersion)
	if err != nil {
		return "", fmt.Errorf("failed to get download URL: %w", err)
	}
	
	zipPath := filepath.Join(javaDir, filename)
	if err := downloadFile(downloadURL, zipPath); err != nil {
		return "", fmt.Errorf("failed to download Java: %w", err)
	}
	
	// Extract Java
	if err := extractJavaZip(zipPath, versionDir); err != nil {
		os.Remove(zipPath) // Clean up on failure
		return "", fmt.Errorf("failed to extract Java: %w", err)
	}
	
	// Clean up zip file
	os.Remove(zipPath)
	
	return versionDir, nil
}

// getJavaDownloadURL constructs the download URL for the specified Java version
func getJavaDownloadURL(majorVersion int) (string, string, error) {
	arch := getArchitecture()
	hostOS := getHostOS()
	
	var repo string
	switch majorVersion {
	case 8:
		repo = "adoptium/temurin8-binaries"
	case 17:
		repo = "adoptium/temurin17-binaries"
	case 21:
		repo = "adoptium/temurin21-binaries"
	default:
		return "", "", fmt.Errorf("unsupported Java version: %d", majorVersion)
	}
	
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

// downloadFile downloads a file from URL to the specified path
func downloadFile(url, filepath string) error {
	fmt.Printf("Downloading from: %s\n", url)
	
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
	
	fmt.Printf("Downloaded: %s\n", filepath)
	return nil
}

// extractJavaZip extracts a Java JRE zip file to the specified directory
func extractJavaZip(zipPath, extractDir string) error {
	fmt.Printf("Extracting Java to: %s\n", extractDir)
	
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer reader.Close()
	
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return fmt.Errorf("failed to create extract directory: %w", err)
	}
	
	// Find the root directory in the zip (usually something like jdk-21.0.1+12-jre)
	var rootDir string
	for _, file := range reader.File {
		if file.FileInfo().IsDir() && strings.Count(file.Name, "/") == 1 {
			rootDir = strings.TrimSuffix(file.Name, "/")
			break
		}
	}
	
	for _, file := range reader.File {
		// Skip the root directory itself and extract contents directly
		if file.Name == rootDir+"/" {
			continue
		}
		
		// Remove root directory from path
		relativePath := strings.TrimPrefix(file.Name, rootDir+"/")
		if relativePath == "" {
			continue
		}
		
		path := filepath.Join(extractDir, relativePath)
		
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, file.FileInfo().Mode()); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", path, err)
			}
			continue
		}
		
		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory for %s: %w", path, err)
		}
		
		// Extract file
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("failed to open file in zip: %w", err)
		}
		
		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
		if err != nil {
			rc.Close()
			return fmt.Errorf("failed to create output file %s: %w", path, err)
		}
		
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		
		if err != nil {
			return fmt.Errorf("failed to extract file %s: %w", path, err)
		}
	}
	
	fmt.Printf("✅ Java extracted successfully\n")
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
		javaExe := filepath.Join(javaDir, entry.Name(), "bin", "java")
		if runtime.GOOS == "windows" {
			javaExe = filepath.Join(javaDir, entry.Name(), "bin", "java.exe")
		}
		
		if version, err := DetectJavaVersion(javaExe); err == nil {
			installations = append(installations, version)
		}
	}
	
	return installations, nil
}

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

// parseJavaMajorVersion extracts major version number
func parseJavaMajorVersion(version string) int {
	// Handle both old (1.8.0_391) and new (21.0.1) formats
	if strings.HasPrefix(version, "1.") {
		// Old format: 1.8.0_391 -> 8
		parts := strings.Split(version, ".")
		if len(parts) >= 2 {
			if major, err := strconv.Atoi(parts[1]); err == nil {
				return major
			}
		}
	} else {
		// New format: 21.0.1 -> 21
		parts := strings.Split(version, ".")
		if len(parts) >= 1 {
			if major, err := strconv.Atoi(parts[0]); err == nil {
				return major
			}
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
