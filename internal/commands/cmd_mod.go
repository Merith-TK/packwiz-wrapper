package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/Merith-TK/packwiz-wrapper/internal/utils"
)

// CmdMod provides enhanced mod management with smart URL parsing
func CmdMod() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"mod", "m"},
		"Enhanced mod management with smart URL parsing",
		`Smart Mod Command:
  pw mod add <url|mr:slug:version>     - Add mod with smart URL detection
  pw mod remove <name>                 - Remove mod
  pw mod update [name]                 - Update mod(s)
  pw mod list                          - List installed mods

Smart URL Support:
  - mr:cc-tweaked:Zoo9N9Dv            - Modrinth slug + version
  - Full URLs (Modrinth/CurseForge)   - Auto-detected
  - Traditional packwiz syntax        - Passed through

Examples:
  pw mod add mr:cc-tweaked:Zoo9N9Dv
  pw mod add https://modrinth.com/mod/cc-tweaked
  pw mod remove cc-tweaked
  pw m list`,
		func(args []string) error {
			packDir, _ := os.Getwd()

			// Find pack directory
			packLocation := utils.FindPackToml(packDir)
			if packLocation == "" {
				return fmt.Errorf("pack.toml not found")
			}

			if len(args) == 0 {
				// Show our own help instead of passing through
				fmt.Println(`Smart Mod Command:
  pw mod add <url|mr:slug:version>     - Add mod with smart URL detection
  pw mod remove <name>                 - Remove mod
  pw mod update [name]                 - Update mod(s)
  pw mod list                          - List installed mods

Smart URL Support:
  - mr:cc-tweaked:Zoo9N9Dv            - Modrinth slug + version
  - Full URLs (Modrinth/CurseForge)   - Auto-detected
  - Traditional packwiz syntax        - Passed through

Examples:
  pw mod add mr:cc-tweaked:Zoo9N9Dv
  pw mod add https://modrinth.com/mod/cc-tweaked
  pw mod remove cc-tweaked
  pw m list`)
				return nil
			}

			switch args[0] {
			case "add":
				if len(args) < 2 {
					return fmt.Errorf("mod add requires an identifier")
				}
				return addModSmart(packLocation, args[1])
			case "remove", "rm":
				if len(args) < 2 {
					return fmt.Errorf("mod remove requires a mod name")
				}
				return ExecuteSelfCommand([]string{"remove", args[1]}, packLocation)
			case "update":
				if len(args) > 1 {
					// Update specific mod
					return ExecuteSelfCommand([]string{"update", args[1]}, packLocation)
				} else {
					// Update all mods
					return ExecuteSelfCommand([]string{"update"}, packLocation)
				}
			case "list", "ls":
				return ExecuteSelfCommand([]string{"list"}, packLocation)
			default:
				return fmt.Errorf("unknown mod command: %s. Use 'pw mod' for help", args[0])
			}
		}
}

// Helper functions for smart mod management
func addModSmart(packLocation, identifier string) error {
	source, slug, version := parseModIdentifier(identifier)

	switch source {
	case "modrinth", "mr":
		return addModrinthMod(packLocation, slug, version)
	case "curseforge", "cf":
		return addCurseforgeMod(packLocation, slug, version)
	case "auto":
		// Try Modrinth first, then CurseForge
		if err := addModrinthMod(packLocation, slug, version); err != nil {
			fmt.Printf("Modrinth failed, trying CurseForge: %v\n", err)
			return addCurseforgeMod(packLocation, slug, version)
		}
		return nil
	case "url":
		// Detect URL type and use appropriate command
		if strings.Contains(identifier, "modrinth.com") {
			return ExecuteSelfCommand([]string{"modrinth", "add", identifier}, packLocation)
		} else if strings.Contains(identifier, "curseforge.com") {
			return ExecuteSelfCommand([]string{"curseforge", "add", identifier}, packLocation)
		} else {
			// Generic URL - try modrinth first
			if err := ExecuteSelfCommand([]string{"modrinth", "add", identifier}, packLocation); err != nil {
				fmt.Printf("Modrinth failed, trying CurseForge: %v\n", err)
				return ExecuteSelfCommand([]string{"curseforge", "add", identifier}, packLocation)
			}
			return nil
		}
	default:
		return fmt.Errorf("unknown mod source: %s", source)
	}
}

func parseModIdentifier(identifier string) (source, slug, version string) {
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

func addModrinthMod(packLocation, slug, version string) error {
	if version != "" {
		// For specific versions, we can pass the version ID directly
		return ExecuteSelfCommand([]string{"modrinth", "add", slug, "--version", version}, packLocation)
	} else {
		// Latest version - just pass the slug
		return ExecuteSelfCommand([]string{"modrinth", "add", slug}, packLocation)
	}
}

func addCurseforgeMod(packLocation, slug, version string) error {
	if version != "" {
		// For CurseForge, version would be file ID
		return ExecuteSelfCommand([]string{"curseforge", "add", slug, "--file", version}, packLocation)
	} else {
		// Latest version - just pass the slug
		return ExecuteSelfCommand([]string{"curseforge", "add", slug}, packLocation)
	}
}
