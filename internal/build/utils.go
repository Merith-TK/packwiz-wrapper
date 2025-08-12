package build

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// CreateZipFromDir creates a zip file from a directory
func CreateZipFromDir(srcDir, zipPath string) error {
	// Create the zip file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	// Create a new zip writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Walk through the source directory
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path for the zip entry
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Convert Windows paths to Unix paths for zip compatibility
		relPath = filepath.ToSlash(relPath)

		if info.IsDir() {
			// Create directory entry (with trailing slash)
			_, err := zipWriter.Create(relPath + "/")
			return err
		} else {
			// Create file entry
			return addFileToZip(zipWriter, path, relPath)
		}
	})
}

// addFileToZip adds a file to the zip archive
func addFileToZip(zipWriter *zip.Writer, filePath, zipPath string) error {
	// Open the file to be added
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create entry in zip
	writer, err := zipWriter.Create(zipPath)
	if err != nil {
		return err
	}

	// Copy file contents to zip
	_, err = io.Copy(writer, file)
	return err
}

// CreateZipFromFiles creates a zip file from a list of files with custom paths
func CreateZipFromFiles(zipPath string, files map[string]string) error {
	// files map: local path -> zip path
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for localPath, zipEntryPath := range files {
		info, err := os.Stat(localPath)
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", localPath, err)
		}

		// Convert to Unix path for zip compatibility
		zipEntryPath = filepath.ToSlash(zipEntryPath)

		if info.IsDir() {
			// Add directory recursively
			err = filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				relPath, err := filepath.Rel(localPath, path)
				if err != nil {
					return err
				}

				fullZipPath := zipEntryPath
				if relPath != "." {
					fullZipPath = filepath.ToSlash(filepath.Join(zipEntryPath, relPath))
				}

				if info.IsDir() {
					if fullZipPath != zipEntryPath {
						_, err := zipWriter.Create(fullZipPath + "/")
						return err
					}
					return nil
				} else {
					return addFileToZip(zipWriter, path, fullZipPath)
				}
			})
			if err != nil {
				return err
			}
		} else {
			// Add single file
			if err := addFileToZip(zipWriter, localPath, zipEntryPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetPackNameFromDir extracts a reasonable pack name from the directory path
func GetPackNameFromDir(packDir string) string {
	name := filepath.Base(packDir)

	// Clean up the name
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ToLower(name)

	// Remove invalid characters
	validName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return -1
	}, name)

	if validName == "" {
		return "modpack"
	}

	return validName
}

// DownloadPackwizInstaller downloads the latest packwiz-installer-bootstrap.jar from GitHub
func DownloadPackwizInstaller(destinationPath string) error {
	// Get the latest release info from GitHub API
	resp, err := http.Get("https://api.github.com/repos/packwiz/packwiz-installer-bootstrap/releases/latest")
	if err != nil {
		return fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release struct {
		Assets []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release info: %w", err)
	}

	// Find the JAR file
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == "packwiz-installer-bootstrap.jar" {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("packwiz-installer-bootstrap.jar not found in latest release")
	}

	// Download the file
	resp, err = http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download installer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create the destination file
	outFile, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer outFile.Close()

	// Copy the downloaded content
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write downloaded file: %w", err)
	}

	fmt.Printf("✅ Downloaded packwiz-installer-bootstrap.jar\n")
	return nil
}
