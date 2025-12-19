package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Merith-TK/packwiz-wrapper/internal/modformat"
)

// CmdRelease provides release and changelog generation functionality
func CmdRelease() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"release", "changelog"},
		"Generate release files and changelogs",
		`Usage:
  pw release [changelog|files|<format>...]

Options:
  (none)              - Generate changelog + all formats
  changelog           - Generate changelog only
  files               - Generate all formats (no changelog)
  <format>...         - Generate changelog + specific format(s)

Formats: cf, mr, mmc, technic, server, all

Examples:
  pw release          - Full release
  pw release mmc      - Changelog + MultiMC
  pw release cf mr    - Changelog + CurseForge + Modrinth`,
		executeRelease
}

// executeRelease handles the main release command logic
func executeRelease(args []string) error {
	// No arguments = full release (changelog + all formats)
	if len(args) == 0 {
		return fullRelease()
	}

	// Handle special commands
	switch args[0] {
	case "changelog":
		return changelogOnly()
	case "files":
		return filesOnly()
	default:
		return releaseWithFormats(args)
	}
}

// fullRelease generates changelog and all export formats
func fullRelease() error {
	if err := generateChangelog(); err != nil {
		return fmt.Errorf("failed to generate changelog: %w", err)
	}
	return generateReleaseFiles("all")
}

// changelogOnly generates only the changelog
func changelogOnly() error {
	return generateChangelog()
}

// filesOnly generates all export formats without changelog
func filesOnly() error {
	return generateReleaseFiles("all")
}

// releaseWithFormats generates changelog and specified formats
func releaseWithFormats(formats []string) error {
	if err := generateChangelog(); err != nil {
		return fmt.Errorf("failed to generate changelog: %w", err)
	}

	for _, format := range formats {
		if err := generateReleaseFiles(format); err != nil {
			return fmt.Errorf("failed to build %s: %w", format, err)
		}
	}
	return nil
}

// generateChangelog creates a changelog file with git log and mod list
func generateChangelog() error {
	packDir, _ := os.Getwd()
	buildDir := filepath.Join(packDir, ".build")
	changelogPath := filepath.Join(buildDir, "CHANGELOG.md")

	// Ensure .build directory exists
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create .build directory: %w", err)
	}

	fmt.Println("Generating changelog...")

	// Create changelog file
	file, err := os.Create(changelogPath)
	if err != nil {
		return fmt.Errorf("failed to create changelog file: %w", err)
	}
	defer file.Close()

	// Write git log section
	if err := writeGitLog(file, packDir); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	// Write mod list section
	if err := writeModList(file, packDir); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	fmt.Printf("Changelog generated: %s\n", changelogPath)
	return nil
}

// writeGitLog writes the git log section to the changelog
func writeGitLog(file *os.File, packDir string) error {
	fmt.Println("Generating git log...")

	cmd := exec.Command("git", "log", "--pretty=format:%h - %s (%ci)", "--abbrev-commit")
	cmd.Dir = packDir
	output, err := cmd.Output()

	if err != nil {
		file.WriteString("# Changelog\n\nGit log not available.\n\n")
		return fmt.Errorf("failed to generate git log: %w", err)
	}

	file.WriteString("# Changelog\n\n")
	file.Write(output)
	file.WriteString("\n\n")
	return nil
}

// writeModList writes the mod list section to the changelog
func writeModList(file *os.File, packDir string) error {
	fmt.Println("Adding mod list to changelog...")

	modlistPath := filepath.Join(packDir, "modlist.md")

	// Try to read existing modlist, generate if needed
	modlistContent, err := readOrGenerateModlist(modlistPath)
	if err != nil {
		return fmt.Errorf("failed to get mod list: %w", err)
	}

	// Write mod list in collapsible section
	file.WriteString("<details><summary>Mod List</summary>\n\n")
	file.Write(modlistContent)
	file.WriteString("</details>\n")
	return nil
}

// readOrGenerateModlist reads existing modlist or generates a new one
func readOrGenerateModlist(modlistPath string) ([]byte, error) {
	// Try to read existing modlist
	if content, err := os.ReadFile(modlistPath); err == nil {
		return content, nil
	}

	// Generate markdown content (showVersions=true, showAuthors=false, showPlatform=false for changelog)
	opts := modformat.ModlistOptions{
		ShowVersions: true,
		ShowAuthors:  false,
		ShowPlatform: false,
	}
	content, err := generateModlistContent(opts)
	if err != nil {
		return nil, err
	}
	return []byte(content), nil
}

func generateReleaseFiles(formats ...string) error {
	packDir, _ := os.Getwd()
	packName := filepath.Base(packDir)

	// If no formats specified, default to "all"
	if len(formats) == 0 {
		formats = []string{"all"}
	}

	fmt.Println("Generating release files...")

	// Build each specified format
	for _, format := range formats {
		if err := executeBuildFormat(format, packDir, packName, false); err != nil {
			return fmt.Errorf("failed to build %s: %w", format, err)
		}
	}

	fmt.Println("Release files generated in .build directory")
	return nil
}
