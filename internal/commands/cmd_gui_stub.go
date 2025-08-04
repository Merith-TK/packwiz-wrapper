//go:build !gui

package commands

import (
	"fmt"
)

// CmdGUI provides a stub that explains GUI is not available
func CmdGUI() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"gui"},
		"GUI not available in this build",
		`GUI Command (Not Available):
  pw gui                  - Launch the graphical user interface

This build was compiled without GUI support to reduce binary size
and dependencies for headless/embedded systems.

To use the GUI, rebuild with:
  go build -tags gui

Available commands in headless mode:
  pw help                 - Show available commands
  pw mod add <mod>        - Add mods
  pw build                - Build/export packs
  pw list                 - List installed mods`,
		func(args []string) error {
			return fmt.Errorf("GUI not available in this build - compile with 'go build -tags gui' to enable GUI functionality")
		}
}
