package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Merith-TK/packwiz-wrapper/internal/build"
)

// CmdBuild provides enhanced build/export operations
func CmdBuild() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"build", "export"},
		"Export pack to various formats (CurseForge, Modrinth, MultiMC, etc.)",
		`Build Commands:
  pw build curseforge|cf  - Export CurseForge pack
  pw build modrinth|mr    - Export Modrinth pack
  pw build multimc|mmc    - Export MultiMC pack
  pw build technic        - Export Technic pack
  pw build server         - Export server pack
  pw build all            - Export all supported formats

Examples:
  pw build cf             - Quick CurseForge export
  pw export modrinth      - Export to Modrinth (using alias)
  pw build all            - Export everything`,
		func(args []string) error {
			if len(args) == 0 {
				fmt.Println("Please specify a build target: cf, mr, mmc, technic, server, all")
				return nil
			}

			packDir, _ := os.Getwd()

			// Ensure .build directory exists
			buildDir := filepath.Join(packDir, ".build")
			if err := os.MkdirAll(buildDir, 0755); err != nil {
				return fmt.Errorf("failed to create .build directory: %w", err)
			}

			// Get pack name from directory
			packName := filepath.Base(packDir)

			switch args[0] {
			case "curseforge", "cf":
				return build.ExportCurseForge(packDir, packName)
			case "modrinth", "mr":
				return build.ExportModrinth(packDir, packName)
			case "multimc", "mmc":
				return build.ExportMultiMC(packDir, packName)
			case "technic":
				return build.ExportTechnic(packDir, packName)
			case "server":
				return build.ExportServer(packDir, packName)
			case "all":
				fmt.Println("Exporting all formats...")
				formats := []string{"curseforge", "modrinth", "multimc", "technic", "server"}
				for _, format := range formats {
					fmt.Printf("\n=== Exporting %s ===\n", format)
					if err := executeBuildFormat(format, packDir, packName); err != nil {
						fmt.Printf("Warning: Failed to export %s: %v\n", format, err)
					}
				}
				return nil
			default:
				return fmt.Errorf("unknown build target: %s", args[0])
			}
		}
}

func executeBuildFormat(format, packDir, packName string) error {
	switch format {
	case "curseforge":
		return build.ExportCurseForge(packDir, packName)
	case "modrinth":
		return build.ExportModrinth(packDir, packName)
	case "multimc":
		return build.ExportMultiMC(packDir, packName)
	case "technic":
		return build.ExportTechnic(packDir, packName)
	case "server":
		return build.ExportServer(packDir, packName)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}
