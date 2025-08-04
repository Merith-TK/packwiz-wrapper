package commands

import (
	"fmt"
	"os"
	"os/exec"
)

// CmdArbitrary provides arbitrary command execution
func CmdArbitrary() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"arbitrary", "arb", "exec", "run"},
		"Execute arbitrary commands in pack context",
		`Arbitrary Commands:
  pw arbitrary <command>  - Execute any command in pack directory
  pw exec <command>       - Same as arbitrary (alias)
  pw run <command>        - Same as arbitrary (alias)

Examples:
  pw arbitrary git status - Run git status in pack directory
  pw exec ls -la          - List pack directory contents
  pw run make build       - Run make build command

Note: This is primarily useful for batch operations where you want to
run the same command across multiple pack directories.`,
		func(args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("no command specified")
			}

			return executeArbitraryCommand(args)
		}
}

func executeArbitraryCommand(args []string) error {
	packDir, _ := os.Getwd()

	// Find pack directory to ensure we're in the right context
	packLocation := findPackToml(packDir)
	if packLocation == "" {
		fmt.Println("Warning: pack.toml not found, running command in current directory")
		packLocation = packDir
	}

	fmt.Printf("Executing arbitrary command in %s: %s\n", packLocation, args)

	// Create and execute command
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = packLocation
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("arbitrary command failed: %w", err)
	}

	return nil
}
