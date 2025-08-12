package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/Merith-TK/packwiz-wrapper/internal/utils"
)

func CmdChange() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"change", "modify", "adjust"},
		"Modify pack.toml fields (name, author, version)",
		`Change Commands:
  pw change name "New Name"                  - Changes the pack name
  pw change author "Someone"                 - Changes the pack author
  pw change version "1.0.0"                  - Changes the pack version
  pw change author "Someone" version "1.0.0" - Changes both the pack author and version`,
		func(args []string) error {
			if len(args) == 0 || (len(args) == 1 && args[0] == "help") {
				fmt.Println(longHelp)
				return nil
			}

			newValues := map[string]string{}
			validKeys := map[string]bool{"name": true, "author": true, "version": true}

			// Parse arguments
			for i := 0; i < len(args); i++ {
				key := args[i]

				if !validKeys[key] {
					fmt.Printf("Unknown key: %s\n", key)
					fmt.Println("Valid keys are: name, author, version")
					return nil
				}

				if i+1 >= len(args) {
					fmt.Printf("Missing value for key: %s\n", key)
					return nil
				}

				value := args[i+1]
				newValues[key] = value
				i++
			}

			return modify(newValues)
		}
}

func modify(vals map[string]string) error {
	packDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	packLocation := utils.FindPackToml(packDir)
	if packLocation == "" {
		return fmt.Errorf("pack.toml not found")
	}

	packFile := filepath.Join(packLocation, "pack.toml")
	data, err := os.ReadFile(packFile)
	if err != nil {
		return fmt.Errorf("failed to read pack.toml: %w", err)
	}

	original := string(data)
	modified := original

	for key, value := range vals {
		// Match: key = "something"
		pattern := fmt.Sprintf(`(?m)^%s\s*=\s*".*?"`, regexp.QuoteMeta(key))
		re := regexp.MustCompile(pattern)
		replacement := fmt.Sprintf(`%s = "%s"`, key, value)

		if re.MatchString(modified) {
			modified = re.ReplaceAllString(modified, replacement)
			fmt.Printf("Updated %s to \"%s\"\n", key, value)
		} else {
			fmt.Printf("Key '%s' not found — skipping\n", key)
		}
	}

	if original == modified {
		fmt.Println("No changes made.")
		return nil
	}

	if err := os.WriteFile(packFile, []byte(modified), 0644); err != nil {
		return fmt.Errorf("failed to write pack.toml: %w", err)
	}

	fmt.Println("pack.toml updated successfully.")
	return nil
}
