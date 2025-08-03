package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Merith-TK/packwiz-wrapper/internal/packwiz"
	"github.com/pelletier/go-toml"
)

// CmdModlist provides mod listing functionality
func CmdModlist() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"modlist", "list-mods", "mods"},
		"Generate and display mod list",
		`Modlist Commands:
  pw modlist              - Generate modlist.md file
  pw modlist raw          - Output raw mod list (no markdown)
  pw modlist versions     - Include mod versions in output
  pw modlist help         - Show this help

Examples:
  pw modlist              - Generate formatted modlist.md
  pw modlist raw          - Show raw mod names only
  pw modlist versions     - Generate modlist with version info`,
		func(args []string) error {
			rawOutput := false
			showVersions := false

			// Parse arguments
			for _, arg := range args {
				switch arg {
				case "raw":
					rawOutput = true
				case "versions":
					showVersions = true
				case "help":
					fmt.Println("Usage: pw modlist [options]")
					fmt.Println("Options:")
					fmt.Println("  raw      - Output raw modlist without markdown formatting")
					fmt.Println("  versions - Show mod versions")
					fmt.Println("  help     - Show this help")
					return nil
				}
			}

			return generateModlist(rawOutput, showVersions)
		}
}

func generateModlist(rawOutput, showVersions bool) error {
	packDir, _ := os.Getwd()
	client := packwiz.NewClient(packDir)

	// Find pack directory
	packLocation := client.GetPackDir()
	if packLocation == "" {
		return fmt.Errorf("pack.toml not found")
	}

	// Read index.toml
	indexFile := filepath.Join(packLocation, "index.toml")
	indexFileHandler, err := os.Open(indexFile)
	if err != nil {
		return fmt.Errorf("failed to open index.toml: %w", err)
	}
	defer indexFileHandler.Close()

	var index packwiz.IndexToml
	if err := toml.NewDecoder(indexFileHandler).Decode(&index); err != nil {
		return fmt.Errorf("failed to decode index.toml: %w", err)
	}

	// Prepare output file (only if not raw output)
	var outputFile *os.File
	if !rawOutput {
		outputPath := filepath.Join(packDir, "modlist.md")
		os.Remove(outputPath) // Remove existing file

		outputFile, err = os.OpenFile(outputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to create modlist.md: %w", err)
		}
		defer outputFile.Close()

		// Write header
		if _, err := outputFile.WriteString("# Modlist\n\n"); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
	}

	// Process mod files
	var mods []packwiz.ModToml
	for _, file := range index.Files {
		if !file.Metafile {
			continue
		}

		modFilePath := filepath.Join(packLocation, file.File)
		modFile, err := os.Open(modFilePath)
		if err != nil {
			fmt.Printf("Warning: failed to open %s: %v\n", file.File, err)
			continue
		}

		var mod packwiz.ModToml
		if err := toml.NewDecoder(modFile).Decode(&mod); err != nil {
			fmt.Printf("Warning: failed to decode %s: %v\n", file.File, err)
			modFile.Close()
			continue
		}
		modFile.Close()

		mods = append(mods, mod)
	}

	// Sort mods by name (simple alphabetical sort)
	for i := 0; i < len(mods)-1; i++ {
		for j := i + 1; j < len(mods); j++ {
			if mods[i].Name > mods[j].Name {
				mods[i], mods[j] = mods[j], mods[i]
			}
		}
	}

	// Output mods
	fmt.Printf("Found %d mods:\n", len(mods))
	for _, mod := range mods {
		var line string
		if rawOutput {
			if showVersions {
				version := getModVersion(mod)
				if version != "" {
					line = fmt.Sprintf("%s %s", mod.Name, version)
				} else {
					line = mod.Name
				}
			} else {
				line = mod.Name
			}
			fmt.Println(line)
		} else {
			if showVersions {
				version := getModVersion(mod)
				if version != "" {
					line = fmt.Sprintf("- **%s** (v%s)", mod.Name, version)
				} else {
					line = fmt.Sprintf("- **%s**", mod.Name)
				}
			} else {
				line = fmt.Sprintf("- **%s**", mod.Name)
			}

			// Add side information if available
			if mod.Side != "" && mod.Side != "both" {
				line += fmt.Sprintf(" (%s-side)", mod.Side)
			}

			line += "\n"

			// Write to console and file
			fmt.Print(line)
			if outputFile != nil {
				if _, err := outputFile.WriteString(line); err != nil {
					return fmt.Errorf("failed to write to modlist.md: %w", err)
				}
			}
		}
	}

	if !rawOutput && outputFile != nil {
		fmt.Printf("\nModlist written to modlist.md\n")
	}

	return nil
}

func getModVersion(mod packwiz.ModToml) string {
	// Try to get version from update sources
	if mod.Update.Modrinth.Version != "" {
		return mod.Update.Modrinth.Version
	}

	// For CurseForge, we don't have a direct version field,
	// but we could potentially extract it from the filename
	if mod.Filename != "" {
		// This is a simple heuristic - could be improved
		return ""
	}

	return ""
}
