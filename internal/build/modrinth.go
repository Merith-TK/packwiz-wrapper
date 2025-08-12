package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Merith-TK/packwiz-wrapper/internal/utils"
)

// ExportModrinth exports the pack as a Modrinth mrpack file
func ExportModrinth(packDir, packName string) error {
	fmt.Println("💚 Exporting Modrinth pack...")

	// Find pack.toml location
	packLocation := utils.FindPackToml(packDir)
	if packLocation == "" {
		return fmt.Errorf("pack.toml not found")
	}

	// Create .build directory
	buildDir := filepath.Join(packDir, ".build")
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create .build directory: %w", err)
	}

	// Create output filename with timestamp
	timestamp := time.Now().Format("_01-02_15-04-05")
	outputFilename := fmt.Sprintf("%s-modrinth%s.mrpack", packName, timestamp)
	outputPath := filepath.Join(buildDir, outputFilename)

	// Run packwiz modrinth export using self-execution with output specified
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	packTomlPath := filepath.Join(packLocation, "pack.toml")
	cmd := exec.Command(executable, "modrinth", "export", "--pack-file", packTomlPath, "-o", outputPath)
	cmd.Dir = packDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("packwiz modrinth export failed: %w", err)
	}

	fmt.Printf("✅ Exported Modrinth pack to .build/%s\n", outputFilename)
	return nil
}
