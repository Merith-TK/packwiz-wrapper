package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// HTTPDownloader provides utilities for downloading files
type HTTPDownloader struct{}

// DownloadFile downloads a file from URL to local path
func (d *HTTPDownloader) DownloadFile(url, filePath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return NewFailedToError("perform HTTP request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return NewFailedToError("create file", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return NewFailedToError("write file", err)
	}

	return nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return NewFailedToError("open source file", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return NewFailedToError("create destination file", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return NewFailedToError("copy file contents", err)
	}

	return nil
}
