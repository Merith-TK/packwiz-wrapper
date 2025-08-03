package commands

import "fmt"

// CmdHelp provides help information
func CmdHelp() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"help", "h", "--help"},
		"Show help information",
		`Help Command:
  pw help                 - Show main help
  pw help <command>       - Show command-specific help
  pw h                    - Show help (short alias)`,
		func(args []string) error {
			if len(args) == 0 {
				fmt.Println("Use 'pw' to see main help or 'pw help <command>' for specific help")
				return nil
			}
			fmt.Printf("Help for command: %s\n", args[0])
			fmt.Println("(Command-specific help would go here)")
			return nil
		}
}
