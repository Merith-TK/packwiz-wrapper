package utils

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
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

// EnsureUTF8 checks if a file is UTF-8 encoded and converts it if needed
func EnsureUTF8(filename string) error {
	// Read the file to check encoding
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Strip UTF-8 BOM if present (EF BB BF)
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		fmt.Println("  Stripping UTF-8 BOM...")
		return os.WriteFile(filename, data[3:], 0644)
	}

	// Detect encoding using chardet
	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(data)
	if err != nil {
		// If detection fails, assume it's already UTF-8 or ASCII
		return nil
	}

	// Check if conversion is needed
	encoding := result.Charset
	confidence := result.Confidence

	// If confidence is low and encoding is ISO-8859-1, it's likely already UTF-8
	// (ISO-8859-1 is often a false positive for UTF-8 text with low confidence)
	if encoding == "ISO-8859-1" && confidence < 50 {
		return nil
	}

	if encoding == "UTF-8" || encoding == "ASCII" {
		// Already UTF-8 or ASCII compatible
		return nil
	}

	// Convert to UTF-8
	fmt.Printf("  Detected %s encoding (confidence: %d%%), converting to UTF-8...\n",
		encoding, confidence)

	var decoder transform.Transformer
	switch encoding {
	case "UTF-16LE":
		decoder = unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
	case "UTF-16BE":
		decoder = unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewDecoder()
	default:
		// For other encodings, try to decode as UTF-16 if it has a BOM
		if len(data) >= 2 {
			if data[0] == 0xFF && data[1] == 0xFE {
				decoder = unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
			} else if data[0] == 0xFE && data[1] == 0xFF {
				decoder = unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewDecoder()
			}
		}
	}

	if decoder == nil {
		return fmt.Errorf("unsupported encoding: %s", encoding)
	}

	// Decode to UTF-8
	reader := transform.NewReader(bytes.NewReader(data), decoder)
	utf8Data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to convert from %s to UTF-8: %w", encoding, err)
	}

	// Write back to file
	return os.WriteFile(filename, utf8Data, 0644)
}
