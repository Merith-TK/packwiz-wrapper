package commands

import (
	"fyne.io/fyne/v2/app"

	"github.com/Merith-TK/packwiz-wrapper/internal/core"
	"github.com/Merith-TK/packwiz-wrapper/internal/gui"
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
	// Create a new Fyne app
	currentApp := app.New()

	// Initialize pack manager with logger
	logger := gui.NewGUILogger(nil) // Will be set properly in the GUI
	packManager := core.NewManager(logger)

	// Initialize and run GUI
	gui.InitializeApp(currentApp, packManager)

	return nil
}
