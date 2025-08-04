package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Merith-TK/packwiz-wrapper/internal/commands"
	"github.com/Merith-TK/packwiz-wrapper/internal/packwiz"
)

func main() {
	// Create command registry with minimal setup
	registry := commands.NewCommandRegistry()

	// ULTRA-SIMPLE REGISTRATION: All function-based commands!
	registry.RegisterAll(
		// Core commands
		commands.CmdVersion, // version, v, --version
		commands.CmdHelp,    // help, h, --help
		commands.CmdGUI,     // gui

		// Mod management
		commands.CmdMod,       // mod, m (smart URL parsing)
		commands.CmdModlist,   // modlist, list-mods, mods
		commands.CmdReinstall, // reinstall, refresh-mods

		// Build & export
		commands.CmdBuild, // build, export (all formats)

		// Pack management
		commands.CmdImport,  // import, load
		commands.CmdDetect,  // detect, detect-url, url
		commands.CmdRelease, // release, changelog

		// Batch operations
		commands.CmdBatch,     // batch, multi
		commands.CmdArbitrary, // arbitrary, exec, run

		// Development
		commands.CmdServer, // server, test-server, start

		// Just add more function references here - no () needed!
	)

	// Parse command line arguments
	args := os.Args[1:]
	if len(args) == 0 {
		showMainHelp(registry)
		return
	}

	commandName := args[0]
	commandArgs := args[1:]

	// Try to find and execute command
	if cmd, found := registry.Get(commandName); found {
		if err := cmd.Execute(commandArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// PASSTHROUGH: Unknown commands go to integrated packwiz
	// Save original args and replace with the command we want to execute
	originalArgs := os.Args
	os.Args = append([]string{os.Args[0]}, args...)

	// Call packwiz directly
	packwiz.PackwizExecute()

	// Restore original args (though we probably won't get here)
	os.Args = originalArgs
}

// showMainHelp displays the main help with all registered commands
func showMainHelp(registry *commands.CommandRegistry) {
	programName := filepath.Base(os.Args[0])

	fmt.Printf("%s v%s - Enhanced packwiz wrapper\n\n", programName, commands.Version)
	fmt.Println("Enhanced Commands:")

	// Show all registered commands with their short help
	for _, cmd := range registry.List() {
		names := cmd.Names()
		primary := names[0]

		// Show aliases if any
		if len(names) > 1 {
			fmt.Printf("  %-15s %s (aliases: %v)\n", primary, cmd.ShortHelp(), names[1:])
		} else {
			fmt.Printf("  %-15s %s\n", primary, cmd.ShortHelp())
		}
	}

	fmt.Println("\nAll other commands are passed through to packwiz.")
	fmt.Printf("Use '%s <command> help' for detailed help on any command.\n", programName)
}
