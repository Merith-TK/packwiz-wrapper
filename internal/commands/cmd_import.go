package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Merith-TK/packwiz-wrapper/internal/utils"
)

// CmdImport provides mod import functionality
func CmdImport() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"import", "load"},
		"Import mods from file or URL list",
		`Usage:
  pw import [-i <file>] [-y] [<url>...]

Options:
  -i <file>           - Import from file (default: import.txt)
  -y                  - Auto-confirm all imports
  <url>...            - Import mods from URLs directly

Examples:
  pw import           - Import from import.txt
  pw import -y        - Import with auto-confirm
  pw import <url>     - Import single mod`,
		func(args []string) error {
			autoConfirm := false
			importFile := false
			filename := "./import.txt"

			// Parse arguments
			filteredArgs := []string{}
			for i, arg := range args {
				switch arg {
				case "-y":
					autoConfirm = true
				case "-i":
					importFile = true
					if i+1 < len(args) {
						filename = args[i+1]
						// Skip the next argument (filename)
						continue
					}
				default:
					if i > 0 && args[i-1] == "-i" {
						// This is the filename, skip it
						continue
					}
					filteredArgs = append(filteredArgs, arg)
				}
			}

			if importFile || len(filteredArgs) == 0 {
				// Import from file
				return importFromFile(filename, autoConfirm)
			}

			// Import from command line arguments
			fmt.Println("[PackWrap] [NOTICE] importing from command line arguments")
			return importFromStrings(filteredArgs, autoConfirm)
		}
}

func importFromFile(filename string, autoConfirm bool) error {
	fmt.Printf("[PackWrap] Importing from file: %s\n", filename)

	// Ensure file is UTF-8 encoded (convert if needed)
	if err := utils.EnsureUTF8(filename); err != nil {
		return fmt.Errorf("failed to normalize file encoding: %w", err)
	}

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open import file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var urls []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Only accept valid URLs or shorthand identifiers
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			urls = append(urls, line)
		} else if strings.Contains(line, ":") && !strings.Contains(line, "://") {
			// Check for shorthand identifiers (mr:slug, cf:slug, etc.)
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				source := strings.ToLower(parts[0])
				if source == "mr" || source == "cf" || source == "modrinth" || source == "curseforge" {
					urls = append(urls, line)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	if len(urls) == 0 {
		fmt.Println("No valid mod identifiers found in import file")
		return nil
	}

	return importMods(urls, autoConfirm)
}

func importFromStrings(urls []string, autoConfirm bool) error {
	if len(urls) == 0 {
		return fmt.Errorf("no URLs provided")
	}

	return importMods(urls, autoConfirm)
}

func importMods(urls []string, autoConfirm bool) error {
	packDir, _ := os.Getwd()
	packLocation := utils.FindPackToml(packDir)
	if packLocation == "" {
		return fmt.Errorf("pack.toml not found")
	}

	fmt.Printf("Found %d mod(s) to import:\n", len(urls))
	for i, url := range urls {
		fmt.Printf("  %d. %s\n", i+1, url)
	}

	if !autoConfirm {
		fmt.Print("Do you want to continue? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Import cancelled")
			return nil
		}
	}

	fmt.Println("Starting import process...")
	var errors []string

	for i, url := range urls {
		fmt.Printf("\n[%d/%d] Importing: %s\n", i+1, len(urls), url)

		// Use smart mod adding (same logic as pw mod add)
		if err := AddModSmart(packLocation, url); err != nil {
			errorMsg := fmt.Sprintf("Failed to import %s: %v", url, err)
			errors = append(errors, errorMsg)
			fmt.Printf("  ERROR: %s\n", errorMsg)
		} else {
			fmt.Printf("  SUCCESS: Imported %s\n", url)
		}
	}

	if len(errors) > 0 {
		fmt.Printf("\nImport completed with %d error(s):\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		return fmt.Errorf("%d imports failed", len(errors))
	}

	fmt.Printf("\nSuccessfully imported all %d mod(s)!\n", len(urls))
	return nil
}
