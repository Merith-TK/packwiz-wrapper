package commands

import (
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strings"

	"github.com/Merith-TK/packwiz-wrapper/internal/packwiz"
)

// Command interface for all commands - minimal and simple
type Command interface {
	Execute(args []string) error // REQUIRED: Run the command
	Names() []string             // REQUIRED: Command names (first is primary, rest are aliases)
	ShortHelp() string           // REQUIRED: One-line description for main help
	LongHelp() string            // REQUIRED: Detailed help text
}

// CommandFunc is a function that returns command information and execution
type CommandFunc func() (names []string, shortHelp, longHelp string, execute func([]string) error)

// FuncCommand wraps a CommandFunc to implement the Command interface
type FuncCommand struct {
	names     []string
	shortHelp string
	longHelp  string
	execute   func([]string) error
}

// NewFuncCommand creates a Command from a CommandFunc
func NewFuncCommand(fn CommandFunc) Command {
	names, shortHelp, longHelp, execute := fn()
	return &FuncCommand{
		names:     names,
		shortHelp: shortHelp,
		longHelp:  longHelp,
		execute:   execute,
	}
}

func (f *FuncCommand) Names() []string             { return f.names }
func (f *FuncCommand) ShortHelp() string           { return f.shortHelp }
func (f *FuncCommand) LongHelp() string            { return f.longHelp }
func (f *FuncCommand) Execute(args []string) error { return f.execute(args) }

// SimpleFuncCommand wraps a simple func([]string) to implement Command interface
type SimpleFuncCommand struct {
	name string
	fn   func([]string)
}

func (s *SimpleFuncCommand) Names() []string {
	return []string{s.name}
}

func (s *SimpleFuncCommand) ShortHelp() string {
	return fmt.Sprintf("Run %s command", s.name)
}

func (s *SimpleFuncCommand) LongHelp() string {
	return fmt.Sprintf("Executes the %s command.", s.name)
}

func (s *SimpleFuncCommand) Execute(args []string) error {
	s.fn(args)
	return nil
}

// extractFunctionName extracts the command name from a function
func extractFunctionName(fn interface{}) string {
	name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	// Extract just the function name part (e.g., "commands.CmdVersion" -> "version")
	parts := strings.Split(name, ".")
	if len(parts) > 0 {
		funcName := parts[len(parts)-1]
		// Remove "Cmd" prefix if present
		if strings.HasPrefix(funcName, "Cmd") {
			return strings.ToLower(funcName[3:])
		}
		return strings.ToLower(funcName)
	}
	return ""
}

// BaseCommand provides common functionality for all commands
type BaseCommand struct {
	PackDir string
	Client  *packwiz.Client
}

// NewBaseCommand creates a new base command
func NewBaseCommand(packDir string) *BaseCommand {
	return &BaseCommand{
		PackDir: packDir,
		Client:  packwiz.NewClient(packDir),
	}
}

// CommandRegistry holds all registered commands
type CommandRegistry struct {
	commands map[string]Command
	aliases  map[string]string // alias -> command name mapping
}

// NewCommandRegistry creates a new command registry
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]Command),
		aliases:  make(map[string]string),
	}
}

// Register adds a command to the registry - accepts Command interface or CommandFunc
func (r *CommandRegistry) Register(cmd interface{}) {
	var command Command

	switch c := cmd.(type) {
	case Command:
		command = c
	case func() ([]string, string, string, func([]string) error):
		// Handle function directly
		names, shortHelp, longHelp, execute := c()
		command = &FuncCommand{
			names:     names,
			shortHelp: shortHelp,
			longHelp:  longHelp,
			execute:   execute,
		}
	case func([]string):
		// Handle simple function - create a wrapper with auto-detected command info
		name := extractFunctionName(cmd)
		if name == "" {
			return // Skip if we can't extract the name
		}
		command = &SimpleFuncCommand{
			name: name,
			fn:   c,
		}
	default:
		fmt.Printf("Warning: Unknown command type: %T\n", cmd)
		return
	}

	names := command.Names()
	if len(names) == 0 {
		return // Skip commands with no names
	}

	primaryName := names[0]
	r.commands[primaryName] = command

	// Register all names (including primary) as valid lookups
	for _, name := range names {
		r.aliases[name] = primaryName
	}
}

// RegisterAll adds multiple commands to the registry at once
func (r *CommandRegistry) RegisterAll(commands ...interface{}) {
	for _, cmd := range commands {
		r.Register(cmd)
	}
}

// Get retrieves a command by name or alias
func (r *CommandRegistry) Get(name string) (Command, bool) {
	// Try direct lookup first
	if cmd, exists := r.commands[name]; exists {
		return cmd, true
	}

	// Try alias lookup
	if actualName, exists := r.aliases[name]; exists {
		if cmd, exists := r.commands[actualName]; exists {
			return cmd, true
		}
	}

	return nil, false
}

// List returns all registered commands
func (r *CommandRegistry) List() []Command {
	var commands []Command
	for _, cmd := range r.commands {
		commands = append(commands, cmd)
	}

	// Sort commands alphabetically by primary name
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Names()[0] < commands[j].Names()[0]
	})

	return commands
}

// GenerateHelp creates a formatted help text for all commands
func (r *CommandRegistry) GenerateHelp(version string) string {
	var help strings.Builder

	help.WriteString(fmt.Sprintf("PackWiz Wrapper v%s - Enhanced PackWiz with Passthrough\n\n", version))

	help.WriteString("Enhanced Commands:\n")
	commands := r.List()

	for _, cmd := range commands {
		names := cmd.Names()
		if len(names) == 0 {
			continue
		}

		// Format: primary [aliases] - short help
		nameDisplay := names[0]
		if len(names) > 1 {
			aliases := names[1:]
			nameDisplay += fmt.Sprintf(" [%s]", strings.Join(aliases, "|"))
		}
		help.WriteString(fmt.Sprintf("  pw %-20s - %s\n", nameDisplay, cmd.ShortHelp()))
	}

	help.WriteString("\nPassthrough Examples:\n")
	help.WriteString("  pw install sodium         -> packwiz install sodium\n")
	help.WriteString("  pw remove jei             -> packwiz remove jei\n")
	help.WriteString("  pw refresh                -> packwiz refresh\n")
	help.WriteString("  pw list                   -> packwiz list\n")
	help.WriteString("\n")
	help.WriteString("Flags:\n")
	help.WriteString("  -d [dir]                  - Set pack directory\n")
	help.WriteString("  -r                        - Auto-refresh after operations\n")
	help.WriteString("  -h                        - Show this help\n")

	return help.String()
}
