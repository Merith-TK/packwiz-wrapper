package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/Merith-TK/packwiz-wrapper/cmd/pw/internal/packwiz"
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
			client := packwiz.NewClient(packDir)
			
			if len(args) == 0 {
				// Pass through to packwiz for help
				return client.Execute([]string{"mod"})
			}

			switch args[0] {
			case "add":
				if len(args) < 2 {
					return fmt.Errorf("mod add requires an identifier")
				}
				return addModSmart(client, args[1])
			default:
				// Pass through to packwiz
				return client.Execute(append([]string{"mod"}, args...))
			}
		}
}

// Helper functions for smart mod management
func addModSmart(client *packwiz.Client, identifier string) error {
	source, slug, version := parseModIdentifier(identifier)
	
	switch source {
	case "modrinth", "mr":
		return addModrinthMod(client, slug, version)
	case "curseforge", "cf":
		return addCurseforgeMod(client, slug, version)
	case "auto":
		// Try Modrinth first, then CurseForge
		if err := addModrinthMod(client, slug, version); err != nil {
			fmt.Printf("Modrinth failed, trying CurseForge: %v\n", err)
			return addCurseforgeMod(client, slug, version)
		}
		return nil
	case "url":
		// Detect URL type and use appropriate command
		if strings.Contains(identifier, "modrinth.com") {
			return client.Execute([]string{"modrinth", "add", identifier})
		} else if strings.Contains(identifier, "curseforge.com") {
			return client.Execute([]string{"curseforge", "add", identifier})
		} else {
			// Generic URL - try modrinth first
			if err := client.Execute([]string{"modrinth", "add", identifier}); err != nil {
				fmt.Printf("Modrinth failed, trying CurseForge: %v\n", err)
				return client.Execute([]string{"curseforge", "add", identifier})
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

func addModrinthMod(client *packwiz.Client, slug, version string) error {
	if version != "" {
		// For specific versions, we can pass the version ID directly
		return client.Execute([]string{"modrinth", "add", slug, "--version", version})
	} else {
		// Latest version - just pass the slug
		return client.Execute([]string{"modrinth", "add", slug})
	}
}

func addCurseforgeMod(client *packwiz.Client, slug, version string) error {
	if version != "" {
		// For CurseForge, version would be file ID
		return client.Execute([]string{"curseforge", "add", slug, "--file", version})
	} else {
		// Latest version - just pass the slug
		return client.Execute([]string{"curseforge", "add", slug})
	}
}
