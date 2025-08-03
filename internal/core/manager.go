package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Merith-TK/packwiz-wrapper/internal/packwiz"
	"github.com/Merith-TK/packwiz-wrapper/pkg/packwrap"
	"github.com/pelletier/go-toml"
)

// Manager implements the PackManager interface
type Manager struct {
	logger packwrap.Logger
}

// NewManager creates a new core manager
func NewManager(logger packwrap.Logger) *Manager {
	if logger == nil {
		logger = &NoOpLogger{}
	}
	return &Manager{
		logger: logger,
	}
}

// GetPackInfo retrieves information about a pack
func (m *Manager) GetPackInfo(packDir string) (*packwrap.PackInfo, error) {
	client := packwiz.NewClient(packDir)
	packLocation := client.GetPackDir()
	if packLocation == "" {
		return nil, fmt.Errorf("pack.toml not found in %s", packDir)
	}

	// Read pack.toml
	packTomlPath := filepath.Join(packLocation, "pack.toml")
	packFile, err := os.Open(packTomlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open pack.toml: %w", err)
	}
	defer packFile.Close()

	var packToml packwiz.PackToml
	if err := toml.NewDecoder(packFile).Decode(&packToml); err != nil {
		return nil, fmt.Errorf("failed to decode pack.toml: %w", err)
	}

	// Count mods
	modCount, err := m.countMods(packLocation)
	if err != nil {
		m.logger.Warn("Failed to count mods: %v", err)
		modCount = 0
	}

	return &packwrap.PackInfo{
		Name:        packToml.Name,
		Author:      packToml.Author,
		McVersion:   packToml.McVersion,
		PackFormat:  packToml.PackFormat,
		Description: packToml.Description,
		ModCount:    modCount,
		PackDir:     packDir,
	}, nil
}

// RefreshPack refreshes the pack
func (m *Manager) RefreshPack(packDir string) error {
	client := packwiz.NewClient(packDir)
	return client.Execute([]string{"refresh"})
}

// ListMods lists all mods in the pack
func (m *Manager) ListMods(packDir string) ([]*packwrap.ModInfo, error) {
	client := packwiz.NewClient(packDir)
	packLocation := client.GetPackDir()
	if packLocation == "" {
		return nil, fmt.Errorf("pack.toml not found")
	}

	// Read index.toml
	indexFile := filepath.Join(packLocation, "index.toml")
	indexFileHandler, err := os.Open(indexFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open index.toml: %w", err)
	}
	defer indexFileHandler.Close()

	var index packwiz.IndexToml
	if err := toml.NewDecoder(indexFileHandler).Decode(&index); err != nil {
		return nil, fmt.Errorf("failed to decode index.toml: %w", err)
	}

	var mods []*packwrap.ModInfo
	for _, file := range index.Files {
		if !file.Metafile {
			continue
		}

		modFilePath := filepath.Join(packLocation, file.File)
		modFile, err := os.Open(modFilePath)
		if err != nil {
			m.logger.Warn("Failed to open %s: %v", file.File, err)
			continue
		}

		var mod packwiz.ModToml
		if err := toml.NewDecoder(modFile).Decode(&mod); err != nil {
			m.logger.Warn("Failed to decode %s: %v", file.File, err)
			modFile.Close()
			continue
		}
		modFile.Close()

		// Determine platform
		platform := "url"
		if mod.Update.Modrinth.ModID != "" {
			platform = "modrinth"
		} else if mod.Update.Curseforge.ProjectID != 0 {
			platform = "curseforge"
		}

		// Get version
		version := ""
		if mod.Update.Modrinth.Version != "" {
			version = mod.Update.Modrinth.Version
		}

		mods = append(mods, &packwrap.ModInfo{
			ID:          strings.TrimSuffix(filepath.Base(modFilePath), ".pw.toml"),
			Name:        mod.Name,
			Filename:    mod.Filename,
			Side:        mod.Side,
			DownloadURL: mod.Download.URL,
			Version:     version,
			Platform:    platform,
		})
	}

	return mods, nil
}

// AddMod adds a mod to the pack
func (m *Manager) AddMod(packDir string, modRef string) error {
	client := packwiz.NewClient(packDir)

	// Parse mod reference and build appropriate command
	source, slug, version := m.parseModIdentifier(modRef)

	var args []string
	switch source {
	case "modrinth", "mr":
		args = []string{"modrinth", "add", slug}
		if version != "" {
			args = append(args, "--version", version)
		}
	case "curseforge", "cf":
		args = []string{"curseforge", "add", slug}
		if version != "" {
			args = append(args, "--file", version)
		}
	case "url":
		// Direct URL
		args = []string{"add", modRef}
	default:
		// Try modrinth first, then curseforge
		if err := client.Execute([]string{"modrinth", "add", modRef}); err != nil {
			return client.Execute([]string{"curseforge", "add", modRef})
		}
		return nil
	}

	return client.Execute(args)
}

// RemoveMod removes a mod from the pack
func (m *Manager) RemoveMod(packDir string, modID string) error {
	client := packwiz.NewClient(packDir)
	return client.Execute([]string{"remove", modID})
}

// UpdateMod updates a mod in the pack
func (m *Manager) UpdateMod(packDir string, modID string) error {
	client := packwiz.NewClient(packDir)
	return client.Execute([]string{"update", modID})
}

// ImportFromFile imports mods from a file
func (m *Manager) ImportFromFile(packDir string, filename string) error {
	// This would use the existing import logic from cmd_import.go
	// For now, we'll use a simplified approach

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open import file: %w", err)
	}
	defer file.Close()

	// Implementation would go here - simplified for now
	m.logger.Info("Import from file functionality needs full implementation")
	return fmt.Errorf("import from file not yet implemented in core")
}

// ImportFromURLs imports mods from URLs
func (m *Manager) ImportFromURLs(packDir string, urls []string) error {
	client := packwiz.NewClient(packDir)

	for _, url := range urls {
		if err := client.Execute([]string{"add", url}); err != nil {
			m.logger.Error("Failed to import %s: %v", url, err)
			return err
		}
	}

	return nil
}

// ExportPack exports the pack to the specified format
func (m *Manager) ExportPack(packDir string, format packwrap.ExportFormat) (string, error) {
	client := packwiz.NewClient(packDir)

	switch format {
	case packwrap.ExportCurseForge:
		return "", client.Execute([]string{"curseforge", "export"})
	case packwrap.ExportModrinth:
		return "", client.Execute([]string{"modrinth", "export"})
	default:
		return "", fmt.Errorf("export format %s not yet implemented", format)
	}
}

// BatchOperation performs a batch operation across multiple directories
func (m *Manager) BatchOperation(ctx context.Context, dirs []string, operation packwrap.BatchOp) error {
	// Implementation would be extracted from cmd_batch.go
	m.logger.Info("Batch operation functionality needs full implementation")
	return fmt.Errorf("batch operations not yet implemented in core")
}

// StartTestServer starts a test server
func (m *Manager) StartTestServer(packDir string) error {
	// Implementation would be extracted from cmd_server.go
	m.logger.Info("Start test server functionality needs full implementation")
	return fmt.Errorf("test server not yet implemented in core")
}

// SetupServer sets up a test server
func (m *Manager) SetupServer(packDir string) error {
	// Implementation would be extracted from cmd_server.go
	m.logger.Info("Setup server functionality needs full implementation")
	return fmt.Errorf("server setup not yet implemented in core")
}

// CleanServer cleans server files
func (m *Manager) CleanServer(packDir string) error {
	// Implementation would be extracted from cmd_server.go
	m.logger.Info("Clean server functionality needs full implementation")
	return fmt.Errorf("server cleanup not yet implemented in core")
}

// Helper methods

func (m *Manager) countMods(packLocation string) (int, error) {
	indexFile := filepath.Join(packLocation, "index.toml")
	indexFileHandler, err := os.Open(indexFile)
	if err != nil {
		return 0, err
	}
	defer indexFileHandler.Close()

	var index packwiz.IndexToml
	if err := toml.NewDecoder(indexFileHandler).Decode(&index); err != nil {
		return 0, err
	}

	count := 0
	for _, file := range index.Files {
		if file.Metafile {
			count++
		}
	}
	return count, nil
}

func (m *Manager) parseModIdentifier(identifier string) (source, slug, version string) {
	// Check if it's a URL
	if strings.HasPrefix(identifier, "http://") || strings.HasPrefix(identifier, "https://") {
		return "url", identifier, ""
	}

	// Parse source:slug:version format
	parts := strings.Split(identifier, ":")

	switch len(parts) {
	case 1:
		// Just slug, auto-detect
		return "auto", parts[0], ""
	case 2:
		// source:slug
		return parts[0], parts[1], ""
	case 3:
		// source:slug:version
		return parts[0], parts[1], parts[2]
	default:
		return "auto", identifier, ""
	}
}

// NoOpLogger is a logger that does nothing
type NoOpLogger struct{}

func (l *NoOpLogger) Info(msg string, args ...interface{})  {}
func (l *NoOpLogger) Warn(msg string, args ...interface{})  {}
func (l *NoOpLogger) Error(msg string, args ...interface{}) {}
func (l *NoOpLogger) Debug(msg string, args ...interface{}) {}
