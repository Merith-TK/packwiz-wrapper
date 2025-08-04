package commands

import (
	"fmt"
	"os"
	"path/filepath"
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
				return exportCurseForge(packDir, packName)
			case "modrinth", "mr":
				return exportModrinth(packDir, packName)
			case "multimc", "mmc":
				return exportMultiMC(packDir, packName)
			case "technic":
				return exportTechnic(packDir, packName)
			case "server":
				return exportServer(packDir, packName)
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
		return exportCurseForge(packDir, packName)
	case "modrinth":
		return exportModrinth(packDir, packName)
	case "multimc":
		return exportMultiMC(packDir, packName)
	case "technic":
		return exportTechnic(packDir, packName)
	case "server":
		return exportServer(packDir, packName)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func exportCurseForge(packDir, packName string) error {
	fmt.Println("Exporting CurseForge pack...")

	if err := ExecuteSelfCommand([]string{"curseforge", "export"}, packDir); err != nil {
		return fmt.Errorf("packwiz curseforge export failed: %w", err)
	}

	return moveBuildFiles("zip", packDir, packName+"-curseforge")
}

func exportModrinth(packDir, packName string) error {
	fmt.Println("Exporting Modrinth pack...")

	if err := ExecuteSelfCommand([]string{"modrinth", "export"}, packDir); err != nil {
		return fmt.Errorf("packwiz modrinth export failed: %w", err)
	}

	return moveBuildFiles("mrpack", packDir, packName+"-modrinth")
}

func exportMultiMC(packDir, packName string) error {
	fmt.Println("Exporting MultiMC pack...")
	fmt.Println("Note: MultiMC export requires additional implementation")
	return nil
}

func exportTechnic(packDir, packName string) error {
	fmt.Println("Exporting Technic pack...")
	fmt.Println("Note: Technic export requires additional implementation")
	return nil
}

func exportServer(packDir, packName string) error {
	fmt.Println("Exporting Server pack...")
	fmt.Println("Note: Server export requires additional implementation")
	return nil
}

func moveBuildFiles(extension, packDir, baseName string) error {
	packTomlDir := findPackToml(packDir)
	if packTomlDir == "" {
		return fmt.Errorf("pack.toml not found")
	}

	baseDir := packTomlDir
	buildDir := filepath.Join(packDir, ".build")

	return filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		if filepath.Ext(path) == "."+extension {
			// Create filename with timestamp
			timestamp := info.ModTime().Format("_01-02_15-04-05")
			filename := fmt.Sprintf("%s%s.%s", baseName, timestamp, extension)
			newPath := filepath.Join(buildDir, filename)

			// Check if file already exists
			if _, err := os.Stat(newPath); err == nil {
				return fmt.Errorf("file already exists: %s", newPath)
			}

			// Move the file
			if err := os.Rename(path, newPath); err != nil {
				return fmt.Errorf("failed to move %s to %s: %w", path, newPath, err)
			}

			fmt.Printf("Moved %s to %s\n", filepath.Base(path), filename)
		}
		return nil
	})
}
