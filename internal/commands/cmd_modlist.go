package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
  pw modlist print        - Only prints to terminal, does not save to a file
  pw modlist help         - Show this help

Examples:
  pw modlist              - Generate formatted modlist.md
  pw modlist raw          - Show raw mod names only
  pw modlist versions     - Generate modlist with version info
  pw modlist print        - Only print modlist to terminal`,
		func(args []string) error {
			rawOutput := false
			showVersions := false
			onlyPrint := false

			// Parse arguments
			for _, arg := range args {
				switch arg {
				case "raw":
					rawOutput = true
				case "versions":
					showVersions = true
				case "print":
					onlyPrint = true
				case "help":
					fmt.Println("Usage: pw modlist [options]")
					fmt.Println("Options:")
					fmt.Println("  raw      - Output raw modlist without markdown formatting")
					fmt.Println("  versions - Show mod versions")
					fmt.Println("  print    - Only print modlist to terminal")
					fmt.Println("  help     - Show this help")
					return nil
				}
			}

			return generateModlist(rawOutput, showVersions, onlyPrint)
		}
}

func generateModlist(rawOutput bool, showVersions bool, onlyPrint bool) error {
	packDir, _ := os.Getwd()

	// Find pack directory
	packLocation := findPackToml(packDir)
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

	// Prepare output file (only if not raw output or print only)
	var outputFile *os.File
	outputPath := filepath.Join(packDir, "modlist.md")

	if (!rawOutput) && (!onlyPrint) {
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
	var modlist []packwiz.ModToml
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

		// Set mod.Parse.ModID to the last part of the path without the .pw.toml extension
		// This is needed for CurseForge URLs
		modID := strings.TrimSuffix(filepath.Base(modFilePath), ".pw.toml")
		mod.Parse.ModID = modID

		modlist = append(modlist, mod)
	}

	// Sort mods by side
	var clientMods []packwiz.ModToml
	var serverMods []packwiz.ModToml
	var sharedMods []packwiz.ModToml

	for _, mod := range modlist {
		switch mod.Side {
		case "client":
			clientMods = append(clientMods, mod)
		case "server":
			serverMods = append(serverMods, mod)
		default: // "both" or empty - treat as shared
			sharedMods = append(sharedMods, mod)
		}
	}

	// Write sections in order: Client, Shared, Server
	if rawOutput {
		// For raw output, just print all mods with their URLs
		allMods := append(append(clientMods, sharedMods...), serverMods...)
		for _, mod := range allMods {
			modURL := getModURL(mod, showVersions)
			fmt.Printf("%s\n%s\n\n", mod.Name, modURL)
		}
	} else {
		totalMods := len(clientMods) + len(serverMods) + len(sharedMods)
		fmt.Printf("Found %d mods (%d client, %d shared, %d server)\n",
			totalMods, len(clientMods), len(sharedMods), len(serverMods))

		writeSection("## Client Mods\n\n", clientMods, outputFile, showVersions)
		writeSection("## Shared Mods\n\n", sharedMods, outputFile, showVersions)
		writeSection("## Server Mods\n\n", serverMods, outputFile, showVersions)

		if outputFile != nil {
			fmt.Printf("Modlist written to modlist.md\n")
		}
	}

	return nil
}

func writeSection(header string, mods []packwiz.ModToml, f *os.File, showVersions bool) {
	if len(mods) == 0 {
		return
	}

	// Write header to console and file
	fmt.Print(header)
	if f != nil {
		f.WriteString(header)
	}

	// Write each mod
	for _, mod := range mods {
		writeMod(mod, f, showVersions)
	}

	// Write newline separator
	fmt.Println()
	if f != nil {
		f.WriteString("\n")
	}
}

func writeMod(mod packwiz.ModToml, f *os.File, showVersions bool) {
	modURL := getModURL(mod, showVersions)
	line := fmt.Sprintf("- [%s](%s)\n", mod.Name, modURL)

	// Write to console and file
	fmt.Print(line)
	if f != nil {
		f.WriteString(line)
	}
}

func getModURL(mod packwiz.ModToml, showVersions bool) string {
	var modURL string

	if mod.Update.Modrinth.ModID != "" {
		modURL = "https://modrinth.com/mod/" + mod.Update.Modrinth.ModID
		if showVersions && mod.Update.Modrinth.Version != "" {
			modURL += "/version/" + mod.Update.Modrinth.Version
		}
	} else if mod.Update.Curseforge.ProjectID != 0 {
		modURL = "https://www.curseforge.com/minecraft/mc-mods/"
		if mod.Parse.ModID != "" {
			modURL += mod.Parse.ModID
		} else {
			modURL += strconv.Itoa(mod.Update.Curseforge.ProjectID)
		}
		if showVersions && mod.Update.Curseforge.FileID != 0 {
			modURL += "/files/" + strconv.Itoa(mod.Update.Curseforge.FileID)
		}
	} else if mod.Download.URL != "" {
		modURL = mod.Download.URL
	} else {
		modURL = "#"
	}

	return modURL
}
