package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CmdRelease provides release and changelog generation functionality
func CmdRelease() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"release", "changelog"},
		"Generate release files and changelogs",
		`Release Commands:
  pw release              - Generate complete release package
  pw release changelog    - Generate changelog only
  pw release files        - Generate release files only

Examples:
  pw release              - Full release generation
  pw release changelog    - Just create changelog
  pw changelog            - Same as release changelog (alias)`,
		func(args []string) error {
			action := "full"
			if len(args) > 0 {
				action = args[0]
			}
			
			switch action {
			case "changelog":
				return generateChangelog()
			case "files":
				return generateReleaseFiles()
			case "full":
				fallthrough
			default:
				// Generate both changelog and release files
				if err := generateChangelog(); err != nil {
					return fmt.Errorf("failed to generate changelog: %w", err)
				}
				return generateReleaseFiles()
			}
		}
}

func generateChangelog() error {
	packDir, _ := os.Getwd()
	buildDir := filepath.Join(packDir, ".build")
	
	// Ensure .build directory exists
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create .build directory: %w", err)
	}
	
	changelogPath := filepath.Join(buildDir, "CHANGELOG.md")
	
	fmt.Println("Generating changelog...")
	
	// Create changelog file
	file, err := os.Create(changelogPath)
	if err != nil {
		return fmt.Errorf("failed to create changelog file: %w", err)
	}
	defer file.Close()
	
	// Generate git log
	fmt.Println("Generating git log...")
	cmd := exec.Command("git", "log", "--pretty=format:%h - %s (%ci)", "--abbrev-commit")
	cmd.Dir = packDir
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("Warning: failed to generate git log: %v\n", err)
		file.WriteString("# Changelog\n\nGit log not available.\n\n")
	} else {
		file.WriteString("# Changelog\n\n")
		file.Write(output)
		file.WriteString("\n\n")
	}
	
	// Add mod list if available
	modlistPath := filepath.Join(packDir, "modlist.md")
	if _, err := os.Stat(modlistPath); err == nil {
		fmt.Println("Adding mod list to changelog...")
		file.WriteString("<details><summary>Mod List</summary>\n\n")
		
		modlistContent, err := os.ReadFile(modlistPath)
		if err != nil {
			fmt.Printf("Warning: failed to read modlist.md: %v\n", err)
		} else {
			file.Write(modlistContent)
		}
		
		file.WriteString("</details>\n")
	} else {
		// Generate mod list on the fly
		fmt.Println("Generating mod list for changelog...")
		if err := generateModlist(true, true); err != nil {
			fmt.Printf("Warning: failed to generate mod list: %v\n", err)
		} else {
			// Try to read the generated modlist
			if modlistContent, err := os.ReadFile(modlistPath); err == nil {
				file.WriteString("<details><summary>Mod List</summary>\n\n")
				file.Write(modlistContent)
				file.WriteString("</details>\n")
			}
		}
	}
	
	fmt.Printf("Changelog generated: %s\n", changelogPath)
	return nil
}

func generateReleaseFiles() error {
	fmt.Println("Generating release files...")
	
	// This would typically build all export formats
	packDir, _ := os.Getwd()
	packName := filepath.Base(packDir)
	if err := executeBuildFormat("all", packDir, packName); err != nil {
		return fmt.Errorf("failed to build release files: %w", err)
	}
	
	fmt.Println("Release files generated in .build directory")
	return nil
}
