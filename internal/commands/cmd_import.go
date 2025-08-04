package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// CmdImport provides mod import functionality
func CmdImport() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"import", "load"},
		"Import mods from file or URL list",
		`Import Commands:
  pw import -i <file>      - Import mods from file (default: import.txt)
  pw import -y             - Auto-confirm imports
  pw import -i file.txt -y - Import from file with auto-confirm
  pw import <url1> <url2>  - Import mods from URLs directly

Examples:
  pw import -i import.txt  - Import from import.txt file
  pw import -y             - Import from default file with auto-confirm
  pw import <mod-url>      - Import single mod from URL`,
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

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open import file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	if len(lines) == 0 {
		fmt.Println("No mods found in import file")
		return nil
	}

	return importMods(lines, autoConfirm)
}

func importFromStrings(urls []string, autoConfirm bool) error {
	if len(urls) == 0 {
		return fmt.Errorf("no URLs provided")
	}

	return importMods(urls, autoConfirm)
}

func importMods(mods []string, autoConfirm bool) error {
	packDir, _ := os.Getwd()
	
	// Find pack directory using our helper function
	packLocation := findPackToml(packDir)
	if packLocation == "" {
		return fmt.Errorf("pack.toml not found")
	}

	fmt.Printf("Found %d mod(s) to import:\n", len(mods))
	for i, mod := range mods {
		fmt.Printf("  %d. %s\n", i+1, mod)
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

	for i, mod := range mods {
		fmt.Printf("\n[%d/%d] Importing: %s\n", i+1, len(mods), mod)

		url, path, name := parseLine(mod, "")
		fmt.Printf("  URL: %s\n", url)
		if path != "" {
			fmt.Printf("  Path: %s\n", path)
		}
		if name != "" {
			fmt.Printf("  Name: %s\n", name)
		}

		// Build packwiz command arguments
		args := []string{"add", url}
		if name != "" {
			args = append(args, "--name", name)
		}

		if err := ExecuteSelfCommand(args, packLocation); err != nil {
			errorMsg := fmt.Sprintf("Failed to import %s: %v", mod, err)
			errors = append(errors, errorMsg)
			fmt.Printf("  ERROR: %s\n", errorMsg)
		} else {
			fmt.Printf("  SUCCESS: Imported %s\n", mod)
		}
	}

	if len(errors) > 0 {
		fmt.Printf("\nImport completed with %d error(s):\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		return fmt.Errorf("%d imports failed", len(errors))
	}

	fmt.Printf("\nSuccessfully imported all %d mod(s)!\n", len(mods))
	return nil
}

// parseLine parses a mod entry line to extract URL, path, and name
// Returns url, path, name
func parseLine(line string, previousLine string) (string, string, string) {
	line = strings.TrimSpace(line)

	// If line doesn't contain a space, treat the whole line as URL
	if !strings.Contains(line, " ") {
		return line, "", ""
	}

	// Split by space to get URL and additional info
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return line, "", ""
	}

	url := parts[0]

	// Check if second part looks like a path (contains / or \)
	if strings.Contains(parts[1], "/") || strings.Contains(parts[1], "\\") {
		// Format: URL PATH [NAME...]
		path := parts[1]
		name := ""
		if len(parts) > 2 {
			name = strings.Join(parts[2:], " ")
		}
		return url, path, name
	} else {
		// Format: URL NAME
		name := strings.Join(parts[1:], " ")
		return url, "", name
	}
}
