package commands

import "fmt"

// GlobalRegistry is set by main.go so help command can access other commands
var GlobalRegistry *CommandRegistry

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

			// Look up the command in the registry
			if GlobalRegistry == nil {
				fmt.Printf("Help for command: %s\n", args[0])
				fmt.Println("(Registry not available)")
				return nil
			}

			cmd, found := GlobalRegistry.Get(args[0])
			if !found {
				fmt.Printf("Unknown command: %s\n", args[0])
				fmt.Println("Use 'pw help' to see available commands")
				return nil
			}

			// Display the command's long help
			fmt.Println(cmd.LongHelp())
			return nil
		}
}
