// Package java provides Java version management utilities for PackWrap.
//
// Download and installation functionality.
package java

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/Merith-TK/utils/pkg/archive"
)

// InstallJava downloads and extracts a Java runtime with progress reporting
func InstallJava(majorVersion int, progress ProgressCallback) (string, error) {
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
	if err := downloadFile(downloadURL, zipPath, progress); err != nil {
		return "", fmt.Errorf("failed to download Java: %w", err)
	}

	reportProgress("Extracting Java...")
	// Extract Java
	if err := extractJavaZip(zipPath, versionDir, progress); err != nil {
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

// downloadFile downloads a file from URL to the specified path with progress reporting
func downloadFile(url, filepath string, progress ProgressCallback) error {
	reportProgress := func(msg string) {
		if progress != nil {
			progress(msg)
		} else {
			fmt.Println(msg)
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

// extractJavaZip extracts a Java JRE zip file with progress reporting
func extractJavaZip(zipPath, extractDir string, progress ProgressCallback) error {
	reportProgress := func(msg string) {
		if progress != nil {
			progress(msg)
		} else {
			fmt.Println(msg)
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
