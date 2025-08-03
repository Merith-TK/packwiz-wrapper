package packwrap

import (
	"context"
)

// PackManager provides high-level operations for modpack management
type PackManager interface {
	// Pack operations
	GetPackInfo(packDir string) (*PackInfo, error)
	RefreshPack(packDir string) error

	// Mod operations
	ListMods(packDir string) ([]*ModInfo, error)
	AddMod(packDir string, modRef string) error
	RemoveMod(packDir string, modID string) error
	UpdateMod(packDir string, modID string) error

	// Import/Export operations
	ImportFromFile(packDir string, filename string) error
	ImportFromURLs(packDir string, urls []string) error
	ExportPack(packDir string, format ExportFormat) (string, error)

	// Batch operations
	BatchOperation(ctx context.Context, dirs []string, operation BatchOp) error

	// Server operations
	StartTestServer(packDir string) error
	SetupServer(packDir string) error
	CleanServer(packDir string) error
}

// PackInfo represents basic pack information
type PackInfo struct {
	Name        string `json:"name"`
	Author      string `json:"author"`
	McVersion   string `json:"mc_version"`
	PackFormat  string `json:"pack_format"`
	Description string `json:"description"`
	ModCount    int    `json:"mod_count"`
	PackDir     string `json:"pack_dir"`
}

// ModInfo represents information about a mod
type ModInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Filename    string `json:"filename"`
	Side        string `json:"side"`
	DownloadURL string `json:"download_url"`
	Version     string `json:"version"`
	Platform    string `json:"platform"` // modrinth, curseforge, url
}

// ExportFormat represents different export formats
type ExportFormat string

const (
	ExportCurseForge ExportFormat = "curseforge"
	ExportModrinth   ExportFormat = "modrinth"
	ExportMultiMC    ExportFormat = "multimc"
	ExportTechnic    ExportFormat = "technic"
	ExportServer     ExportFormat = "server"
	ExportAll        ExportFormat = "all"
)

// BatchOp represents a batch operation
type BatchOp struct {
	Name string
	Args []string
}

// ProgressCallback represents a progress callback function
type ProgressCallback func(current, total int, message string)

// Logger interface for different logging implementations
type Logger interface {
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}
