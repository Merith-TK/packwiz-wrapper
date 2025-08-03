package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// CmdGUI launches the GUI application
func CmdGUI() (names []string, shortHelp, longHelp string, execute func([]string) error) {
	return []string{"gui"},
		"Launch the PackWiz Wrapper GUI",
		`GUI Command:
  pw gui                  - Launch the graphical user interface

The GUI provides an easy-to-use interface for:
- Pack information and management
- Mod installation and removal
- Import/export operations
- Development server setup
- Real-time logs and feedback

Examples:
  pw gui                  - Start the GUI application`,
		func(args []string) error {
			return launchGUI()
		}
}

func launchGUI() error {
	// Get the directory of the current executable
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	execDir := filepath.Dir(execPath)

	// Look for pw-gui executable in the same directory
	guiExecName := "pw-gui"
	if runtime.GOOS == "windows" {
		guiExecName = "pw-gui.exe"
	}

	guiExecPath := filepath.Join(execDir, guiExecName)

	// Check if GUI executable exists
	if _, err := os.Stat(guiExecPath); os.IsNotExist(err) {
		// Try to find it in PATH
		guiExecPath, err = exec.LookPath("pw-gui")
		if err != nil {
			return fmt.Errorf("GUI executable not found. Please build pw-gui first:\n  go build -o %s ./cmd/pw-gui", guiExecName)
		}
	}

	fmt.Println("Launching PackWiz Wrapper GUI...")

	// Launch the GUI application
	cmd := exec.Command(guiExecPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the process and don't wait for it to finish
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start GUI: %w", err)
	}

	fmt.Printf("GUI started with PID: %d\n", cmd.Process.Pid)
	return nil
}
