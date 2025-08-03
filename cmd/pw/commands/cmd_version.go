package commands

import "fmt"

const Version = "0.8.0"

// CmdVersion shows version information
func CmdVersion() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"version", "v", "--version"},
		"Show PackWrap version information",
		`Version Command:
  pw version              - Show PackWrap version
  pw v                    - Show version (short alias)
  pw --version            - Show version (long flag)`,
		func(args []string) error {
			fmt.Printf("PackWrap version %s\n", Version)
			return nil
		}
}
