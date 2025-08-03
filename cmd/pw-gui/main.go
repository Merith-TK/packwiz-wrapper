package main

import (
	"fyne.io/fyne/v2/app"

	"github.com/Merith-TK/packwiz-wrapper/cmd/pw-gui/gui"
	"github.com/Merith-TK/packwiz-wrapper/internal/core"
)

func main() {
	// Create a new Fyne app
	currentApp := app.New()

	// Initialize pack manager with logger
	logger := gui.NewGUILogger(nil) // Will be set properly in the GUI
	packManager := core.NewManager(logger)

	// Initialize and run GUI
	gui.InitializeApp(currentApp, packManager)
}
