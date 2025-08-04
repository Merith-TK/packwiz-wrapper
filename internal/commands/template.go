/*
ULTRA-SIMPLE USAGE - TWO APPROACHES:

=== APPROACH 1: FUNCTION-BASED (RECOMMENDED) ===

func CmdMyCommand() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"mycommand", "alias1", "alias2"},
		"Brief description",
		`Detailed Help:
  pw mycommand sub1       - Does something
  pw alias1 sub2          - Does something else`,
		func(args []string) error {
			// Your command logic here
			if len(args) == 0 {
				fmt.Println("Hello from my command!")
				return nil
			}
			fmt.Printf("Args: %v\n", args)
			return nil
		}
}

// Register: registry.Register(commands.CmdMyCommand)  // NO () - just the function!

=== APPROACH 2: STRUCT-BASED (EXISTING) ===

See below for the traditional struct approach...

*/

package commands

import (
	"fmt"
)

// TEMPLATE: Replace "Example" with your actual command name
// ExampleCommand handles [describe what this command does]
type ExampleCommand struct {
	*BaseCommand
}

// NewExampleCommand creates a new example command instance
func NewExampleCommand(packDir string) *ExampleCommand {
	return &ExampleCommand{
		BaseCommand: NewBaseCommand(packDir),
	}
}

// Names returns command names (first is primary, rest are aliases)
func (c *ExampleCommand) Names() []string {
	return []string{"example", "ex", "demo"} // REQUIRED: Primary name + aliases
}

// ShortHelp returns a one-line description for main help
func (c *ExampleCommand) ShortHelp() string {
	return "Example command template for developers" // REQUIRED: Brief description
}

// LongHelp returns detailed help text
func (c *ExampleCommand) LongHelp() string {
	// REQUIRED: Provide detailed help text
	return `Example Command Help:
  pw example subcommand1 [args]  - Does something useful
  pw example subcommand2 [args]  - Does something else

Examples:
  pw example subcommand1 value   - Example usage
  pw ex subcommand2              - Using alias

This is a template command for developers to copy and modify.`
}

// Execute runs the command with the provided arguments
func (c *ExampleCommand) Execute(args []string) error {
	// REQUIRED: Implement your command logic here
	
	if len(args) == 0 {
		// Show help if no arguments provided
		fmt.Println(c.LongHelp())
		return nil
	}
	
	switch args[0] {
	case "subcommand1":
		return c.handleSubcommand1(args[1:])
	case "subcommand2":
		return c.handleSubcommand2(args[1:])
	default:
		return fmt.Errorf("unknown subcommand: %s", args[0])
	}
}

// OPTIONAL: Add private helper methods for your command logic
func (c *ExampleCommand) handleSubcommand1(args []string) error {
	// Implement subcommand1 logic
	fmt.Println("Subcommand1 executed with args:", args)
	
	// Example: Use the packwiz client for operations
	// return c.Client.Execute([]string{"some", "packwiz", "command"})
	
	return nil
}

func (c *ExampleCommand) handleSubcommand2(args []string) error {
	// Implement subcommand2 logic
	fmt.Println("Subcommand2 executed with args:", args)
	return nil
}

/*
CHOOSE YOUR APPROACH:

FUNCTION-BASED (NEW & SIMPLE):
✅ No structs needed
✅ No constructors needed  
✅ Direct registration: registry.Register(CmdMyCommand)
✅ Perfect for simple commands
✅ Closure access to packDir if needed

STRUCT-BASED (EXISTING):  
✅ Good for complex commands
✅ Access to BaseCommand utilities
✅ Can store state between calls
✅ Easy testing with dependency injection

BOTH WORK TOGETHER! Mix and match as needed.
*/
