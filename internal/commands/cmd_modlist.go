package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Merith-TK/packwiz-wrapper/internal/commands/modlist"
	"github.com/Merith-TK/packwiz-wrapper/internal/modformat"
	"github.com/Merith-TK/packwiz-wrapper/internal/packwiz"
	"github.com/Merith-TK/packwiz-wrapper/internal/utils"
	"github.com/pelletier/go-toml"
)

// CmdModlist provides mod listing functionality
func CmdModlist() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"modlist", "list-mods", "mods"},
		"Generate and display mod list",
		`Usage:
  pw modlist [options]

Options:
  raw                 - Output raw list (no markdown formatting)
  versions            - Include mod versions
  print               - Print to terminal only (don't save file)
  authors             - Fetch and display mod authors (requires API calls)
  platform            - Show platform (Modrinth/CurseForge)

Examples:
  pw modlist               - Generate modlist.md
  pw modlist raw           - Raw output without markdown
  pw modlist versions      - Include version numbers
  pw modlist raw versions  - Raw output with version numbers`,
		func(args []string) error {
			rawOutput := false
			showVersions := false
			onlyPrint := false
			showAuthors := false
			showPlatform := false

			// Parse arguments
			for _, arg := range args {
				switch arg {
				case "raw":
					rawOutput = true
				case "versions":
					showVersions = true
				case "print":
					onlyPrint = true
				case "authors":
					showAuthors = true
				case "platform":
					showPlatform = true
				case "help":
					fmt.Println("Usage: pw modlist [options]")
					fmt.Println("Options:")
					fmt.Println("  raw      - Output raw modlist without markdown formatting")
					fmt.Println("  versions - Show mod versions")
					fmt.Println("  print    - Only print modlist to terminal")
					fmt.Println("  authors  - Fetch and display mod authors (requires API calls)")
					fmt.Println("  platform - Show platform (Modrinth/CurseForge)")
					fmt.Println("  help     - Show this help")
					return nil
				}
			}

			opts := modformat.ModlistOptions{
				ShowVersions: showVersions,
				ShowAuthors:  showAuthors,
				ShowPlatform: showPlatform,
				RawOutput:    rawOutput,
				OnlyPrint:    onlyPrint,
			}
			return modlist.GenerateModlist(opts)
		}
}

func generateModlist(opts modformat.ModlistOptions) error {
	clientMods, sharedMods, serverMods, err := getCategorizedMods()
	if err != nil {
		return err
	}

	// Pre-fetch all author information if needed (batch API calls)
	var authorCache map[int]modformat.ModAuthorInfo
	if opts.ShowAuthors {
		fmt.Println("Fetching mod author information from APIs...")
		allMods := append(append(clientMods, sharedMods...), serverMods...)
		authorCache = modformat.FetchModAuthors(allMods)
		fmt.Printf("Fetched author info for %d mods\n", len(authorCache))
	}

	// Prepare output file (only if not raw output or print only)
	var outputFile *os.File
	packDir, _ := os.Getwd()
	outputPath := filepath.Join(packDir, "modlist.md")

	if (!opts.RawOutput) && (!opts.OnlyPrint) {
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

	// Write sections in order: Client, Shared, Server
	if opts.RawOutput {
		// For raw output, just print all mods with their URLs
		allMods := append(append(clientMods, sharedMods...), serverMods...)
		for i, mod := range allMods {
			modURL := modformat.GetModURL(mod, opts.ShowVersions)

			// Determine platform from TOML structure
			platform := modformat.GetPlatform(mod)

			if opts.ShowAuthors || opts.ShowPlatform {
				// Build output based on flags
				output := mod.Name

				if opts.ShowAuthors {
					// Use cached author info if available
					if authorInfo, ok := authorCache[i]; ok {
						output += " - by " + authorInfo.Author
					}
				}

				if opts.ShowPlatform && platform != "" {
					output += " [" + platform + "]"
				}

				fmt.Printf("%s\n%s\n\n", output, modURL)
			} else {
				fmt.Printf("%s\n%s\n\n", mod.Name, modURL)
			}
		}
	} else {
		totalMods := len(clientMods) + len(serverMods) + len(sharedMods)
		fmt.Printf("Found %d mods (%d client, %d shared, %d server)\n",
			totalMods, len(clientMods), len(sharedMods), len(serverMods))

		writeSection("## Client Mods\n\n", clientMods, outputFile, opts, authorCache, 0)
		writeSection("## Shared Mods\n\n", sharedMods, outputFile, opts, authorCache, len(clientMods))
		writeSection("## Server Mods\n\n", serverMods, outputFile, opts, authorCache, len(clientMods)+len(sharedMods))

		if outputFile != nil {
			fmt.Printf("Modlist written to modlist.md\n")
		}
	}

	return nil
}

func writeSection(header string, mods []packwiz.ModToml, f *os.File, opts modformat.ModlistOptions, authorCache map[int]modformat.ModAuthorInfo, offset int) {
	if len(mods) == 0 {
		return
	}

	// Write header to console and file
	fmt.Print(header)
	if f != nil {
		f.WriteString(header)
	}

	// Write each mod
	for i, mod := range mods {
		// Get author from cache if available, otherwise empty string
		author := ""
		if opts.ShowAuthors && authorCache != nil {
			if authorInfo, ok := authorCache[offset+i]; ok {
				author = authorInfo.Author
			}
		}

		line := modformat.FormatModLine(mod, opts, author)

		fmt.Print(line)
		if f != nil {
			f.WriteString(line)
		}
	}

	// Write newline separator
	fmt.Println()
	if f != nil {
		f.WriteString("\n")
	}
}

// generateModlistContent generates modlist content as a string without console output or file writing
func generateModlistContent(opts modformat.ModlistOptions) (string, error) {
	clientMods, sharedMods, serverMods, err := getCategorizedMods()
	if err != nil {
		return "", err
	}

	// Generate markdown content
	var content strings.Builder

	// Write sections in order: Client, Shared, Server
	content.WriteString("## Client Mods\n\n")
	writeModSectionToBuffer(&content, clientMods, opts)

	content.WriteString("## Shared Mods\n\n")
	writeModSectionToBuffer(&content, sharedMods, opts)

	content.WriteString("## Server Mods\n\n")
	writeModSectionToBuffer(&content, serverMods, opts)

	return content.String(), nil
}

// getCategorizedMods processes mod files and returns them categorized by side
func getCategorizedMods() ([]packwiz.ModToml, []packwiz.ModToml, []packwiz.ModToml, error) {
	packDir, _ := os.Getwd()

	// Find pack directory
	packLocation := utils.FindPackToml(packDir)
	if packLocation == "" {
		return nil, nil, nil, fmt.Errorf("pack.toml not found")
	}

	// Read index.toml
	indexFile := filepath.Join(packLocation, "index.toml")
	indexFileHandler, err := os.Open(indexFile)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open index.toml: %w", err)
	}
	defer indexFileHandler.Close()

	var index packwiz.IndexToml
	if err := toml.NewDecoder(indexFileHandler).Decode(&index); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode index.toml: %w", err)
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
			continue
		}

		var mod packwiz.ModToml
		if err := toml.NewDecoder(modFile).Decode(&mod); err != nil {
			modFile.Close()
			continue
		}
		modFile.Close()

		// Set mod.Parse.ModID to the last part of the path without the .pw.toml extension
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

	return clientMods, sharedMods, serverMods, nil
}

// writeModSectionToBuffer writes a mod section to a string buffer
func writeModSectionToBuffer(buf *strings.Builder, mods []packwiz.ModToml, opts modformat.ModlistOptions) {
	if len(mods) == 0 {
		return
	}

	// Write each mod
	for _, mod := range mods {
		// Get author if needed (will be fetched on demand)
		author := ""
		if opts.ShowAuthors {
			platform := modformat.GetPlatform(mod)
			author = modformat.GetModAuthor(mod, platform)
		}

		line := modformat.FormatModLine(mod, opts, author)
		buf.WriteString(line)
	}

	// Write newline separator
	buf.WriteString("\n")
}
