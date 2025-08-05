package commands

import "fmt"

const Version = "0.8.0"

// Build information (will be set by main)
var BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

// CmdVersion shows version information
func CmdVersion() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"version", "v", "--version"},
		"Show PackWrap version information",
		`Version Command:
  pw version              - Show PackWrap version
  pw v                    - Show version (short alias)
  pw --version            - Show version (long flag)`,
		func(args []string) error {
			version := BuildInfo.Version
			if version == "" {
				version = Version // fallback to const
			}

			fmt.Printf("PackWrap version %s\n", version)

			if BuildInfo.Commit != "" && BuildInfo.Commit != "none" {
				fmt.Printf("Git commit: %s\n", BuildInfo.Commit)
			}

			if BuildInfo.Date != "" && BuildInfo.Date != "unknown" {
				fmt.Printf("Built: %s\n", BuildInfo.Date)
			}

			return nil
		}
}
