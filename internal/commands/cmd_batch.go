package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CmdBatch provides batch operations across multiple pack directories
func CmdBatch() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"batch", "multi"},
		"Run commands across multiple directories",
		`Batch Commands:
  pw batch <command>           - Run command in all subdirectories with pack.toml
  pw batch --all <command>     - Run command in ALL subdirectories
  pw batch -r <command>        - Run command and refresh packs after
  pw batch --all -r <command>  - Run in all dirs and refresh packs

Examples:
  pw batch modlist             - Generate modlists for all packs
  pw batch --all arb ls        - Run 'pw arb ls' in ALL subdirectories
  pw batch build cf            - Build CurseForge exports for all packs
  pw batch -r import -i mods.txt - Import mods and refresh all packs`,
		func(args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("no command specified for batch operation")
			}

			refresh := false
			skipPackCheck := false
			commandArgs := args

			// Parse flags
			for len(commandArgs) > 0 && strings.HasPrefix(commandArgs[0], "-") {
				flag := commandArgs[0]
				commandArgs = commandArgs[1:]

				switch flag {
				case "-r":
					refresh = true
				case "--all":
					skipPackCheck = true
				default:
					return fmt.Errorf("unknown flag: %s", flag)
				}
			}

			if len(commandArgs) == 0 {
				return fmt.Errorf("no command specified after flags")
			}

			return runBatchMode(refresh, skipPackCheck, commandArgs)
		}
}

func runBatchMode(refresh bool, skipPackCheck bool, args []string) error {
	packDir, _ := os.Getwd()

	// Get all subdirectories
	files, err := ioutil.ReadDir(packDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	var targetDirs []string
	for _, file := range files {
		if file.IsDir() {
			dirPath := filepath.Join(packDir, file.Name())

			if skipPackCheck {
				// Add ALL directories when --all flag is used
				targetDirs = append(targetDirs, dirPath)
			} else {
				// Only add directories with pack.toml (original behavior)
				if hasPackToml(dirPath) {
					targetDirs = append(targetDirs, dirPath)
				}
			}
		}
	}

	if len(targetDirs) == 0 {
		if skipPackCheck {
			return fmt.Errorf("no subdirectories found")
		} else {
			return fmt.Errorf("no pack directories found")
		}
	}

	if skipPackCheck {
		fmt.Printf("Found %d director(ies):\n", len(targetDirs))
	} else {
		fmt.Printf("Found %d pack director(ies):\n", len(targetDirs))
	}
	for _, dir := range targetDirs {
		fmt.Printf("  - %s\n", filepath.Base(dir))
	}
	fmt.Println()

	// Get current executable path
	selfExec, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	// Use full path to ensure executable can be found from any directory

	var errors []string
	successCount := 0

	for i, dir := range targetDirs {
		dirName := filepath.Base(dir)
		fmt.Printf("[%d/%d] Processing: %s\n", i+1, len(targetDirs), dirName)

		var err error
		
		// Check if this is an arbitrary command that should be handled directly
		if len(args) > 0 && (args[0] == "arb" || args[0] == "arbitrary" || args[0] == "exec" || args[0] == "run") {
			// Handle arbitrary commands directly in the target directory
			if len(args) < 2 {
				fmt.Printf("  ERROR: No command specified for arbitrary execution\n")
				errorMsg := fmt.Sprintf("Failed in %s: no command specified for arbitrary execution", dirName)
				errors = append(errors, errorMsg)
				continue
			}
			
			// Execute the arbitrary command directly in the target directory
			err = executeArbitraryInDirectory(dir, args[1:])
		} else {
			// Build command arguments - prepend executable name for regular commands
			newArgs := append([]string{selfExec}, args...)
			
			// Execute command in the pack directory
			err = executeInDirectory(dir, newArgs)
		}
		
		if err != nil {
			errorMsg := fmt.Sprintf("Failed in %s: %v", dirName, err)
			errors = append(errors, errorMsg)
			fmt.Printf("  ERROR: %s\n", errorMsg)
		} else {
			successCount++
			fmt.Printf("  SUCCESS: Completed in %s\n", dirName)
		}

		// Refresh if requested  
		if refresh {
			fmt.Printf("  Refreshing %s...\n", dirName)
			if err := executePackwizRefresh(dir); err != nil {
				fmt.Printf("  WARNING: Failed to refresh %s: %v\n", dirName, err)
			}
		}

		fmt.Println()
	}	// Summary
	fmt.Printf("Batch operation completed:\n")
	fmt.Printf("  Successful: %d/%d\n", successCount, len(targetDirs))
	if len(errors) > 0 {
		fmt.Printf("  Errors: %d\n", len(errors))
		for _, err := range errors {
			fmt.Printf("    - %s\n", err)
		}
		return fmt.Errorf("batch operation completed with %d error(s)", len(errors))
	}

	return nil
}

func hasPackToml(dir string) bool {
	// Check for pack.toml in the directory
	if _, err := os.Stat(filepath.Join(dir, "pack.toml")); err == nil {
		return true
	}

	// Check for pack.toml in .minecraft subdirectory
	if _, err := os.Stat(filepath.Join(dir, ".minecraft", "pack.toml")); err == nil {
		return true
	}

	return false
}

func executeInDirectory(dir string, args []string) error {
	// Convert directory path for cross-platform compatibility
	dir = filepath.ToSlash(dir)
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}

	fmt.Printf("  [PackWrap] Arbitrary: [%s] %s\n", dir, strings.Join(args, " "))

	// Create command
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = filepath.Dir(dir)

	// Check if directory exists, create if needed
	if _, err := os.Stat(cmd.Dir); err != nil {
		fmt.Printf("  [PackWrap] [ERROR] arbitrary directory not found, creating...\n")
		if err := os.Mkdir(cmd.Dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func executePackwizRefresh(dir string) error {
	// Convert directory path for cross-platform compatibility
	dir = filepath.ToSlash(dir)
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}

	// The dir path includes trailing slash, so we need the directory part
	workingDir := filepath.Dir(dir)

	// Check if pack.toml exists in the directory
	packTomlPath := filepath.Join(workingDir, "pack.toml")
	if _, err := os.Stat(packTomlPath); err != nil {
		// Try .minecraft subdirectory
		packTomlPath = filepath.Join(workingDir, ".minecraft", "pack.toml")
		if _, err := os.Stat(packTomlPath); err != nil {
			return fmt.Errorf("pack.toml not found in %s or %s/.minecraft", workingDir, workingDir)
		}
		// If pack.toml is in .minecraft, run from there
		workingDir = filepath.Join(workingDir, ".minecraft")
	}

	// Create packwiz refresh command
	cmd := exec.Command("packwiz", "refresh")
	cmd.Dir = workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func executeArbitraryInDirectory(dir string, args []string) error {
	// Find the pack directory within the target directory
	packTomlPath := filepath.Join(dir, "pack.toml")
	packLocation := dir
	
	// Check if pack.toml exists in the directory
	if _, err := os.Stat(packTomlPath); err != nil {
		// Try .minecraft subdirectory
		packTomlPath = filepath.Join(dir, ".minecraft", "pack.toml")
		if _, err := os.Stat(packTomlPath); err == nil {
			packLocation = filepath.Join(dir, ".minecraft")
		}
		// If no pack.toml found, just use the directory as-is (like original arbitrary command)
	} else {
		// pack.toml found in root, use that directory
		packLocation = dir
	}
	
	fmt.Printf("  [PackWrap] Executing arbitrary command in %s: %s\n", packLocation, strings.Join(args, " "))
	
	// Create and execute command
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = packLocation
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	return cmd.Run()
}
