package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Merith-TK/packwiz-wrapper/internal/build"
)

// CmdBuild provides enhanced build/export operations
func CmdBuild() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"build", "export"},
		"Export pack to various formats (CurseForge, Modrinth, MultiMC, etc.)",
		`Usage:
  pw build <format> [options]

Formats: cf (curseforge), mr (modrinth), mmc (multimc), technic, server, all

Options:
  --local, -l         - Use local path for MultiMC (default: remote URL)
  -o <file>           - Specify output filename

Examples:
  pw build cf         - Export CurseForge
  pw build mmc -l     - Export MultiMC with local path
  pw build all        - Export all formats`,
		func(args []string) error {
			if len(args) == 0 {
				fmt.Println("Please specify a build target: cf, mr, mmc, technic, server, all")
				return nil
			}

			packDir, _ := os.Getwd()

			// Parse flags
			var useLocal bool
			var buildTarget string

			for _, arg := range args {
				switch arg {
				case "-l", "--local":
					useLocal = true
				default:
					if !strings.HasPrefix(arg, "-") && buildTarget == "" {
						buildTarget = arg
					}
				}
			}

			if buildTarget == "" {
				return fmt.Errorf("no build target specified")
			}

			// Ensure .build directory exists
			buildDir := filepath.Join(packDir, ".build")
			if err := os.MkdirAll(buildDir, 0755); err != nil {
				return fmt.Errorf("failed to create .build directory: %w", err)
			}

			// Get pack name from directory
			packName := filepath.Base(packDir)

			switch buildTarget {
			case "curseforge", "cf":
				return build.ExportCurseForge(packDir, packName)
			case "modrinth", "mr":
				return build.ExportModrinth(packDir, packName)
			case "multimc", "mmc":
				return build.ExportMultiMC(packDir, packName, useLocal)
			case "technic":
				return build.ExportTechnic(packDir, packName)
			case "server":
				return build.ExportServer(packDir, packName)
			case "all":
				fmt.Println("Exporting all formats...")
				formats := []string{"curseforge", "modrinth", "multimc", "technic", "server"}
				for _, format := range formats {
					fmt.Printf("\n=== Exporting %s ===\n", format)
					if err := executeBuildFormat(format, packDir, packName, useLocal); err != nil {
						fmt.Printf("Warning: Failed to export %s: %v\n", format, err)
					}
				}
				return nil
			default:
				return fmt.Errorf("unknown build target: %s", buildTarget)
			}
		}
}

func executeBuildFormat(format, packDir, packName string, useLocal bool) error {
	switch format {
	case "curseforge", "cf":
		return build.ExportCurseForge(packDir, packName)
	case "modrinth", "mr":
		return build.ExportModrinth(packDir, packName)
	case "multimc", "mmc":
		return build.ExportMultiMC(packDir, packName, useLocal)
	case "technic":
		return build.ExportTechnic(packDir, packName)
	case "server":
		return build.ExportServer(packDir, packName)
	case "all":
		// Export all formats
		fmt.Println("Exporting all formats...")
		formats := []string{"curseforge", "modrinth", "multimc", "technic", "server"}
		var lastErr error
		for _, exportFormat := range formats {
			fmt.Printf("\n=== Exporting %s ===\n", exportFormat)
			if err := executeBuildFormat(exportFormat, packDir, packName, useLocal); err != nil {
				fmt.Printf("Warning: Failed to export %s: %v\n", exportFormat, err)
				lastErr = err
			}
		}
		return lastErr // Return last error, or nil if all succeeded
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}
