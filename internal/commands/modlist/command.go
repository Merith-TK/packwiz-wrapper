// Package modlist provides modlist generation functionality
package modlist

import (
	"fmt"

	"github.com/Merith-TK/packwiz-wrapper/internal/modformat"
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
			return GenerateModlist(opts)
		}
}
