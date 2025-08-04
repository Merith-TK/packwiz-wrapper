package gui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// CreateServerTab creates the server management tab
func CreateServerTab() fyne.CanvasObject {
	// Pack directory section - minimal
	packDirEntry := widget.NewEntry()
	packDirEntry.SetPlaceHolder("Pack directory")
	packDirEntry.SetText(GetGlobalPackDir())

	RegisterPackDirCallback(func(dir string) {
		packDirEntry.SetText(dir)
	})

	packDirEntry.OnChanged = debouncePathUpdate(func(text string) {
		SetGlobalPackDir(text)
	}, 500*time.Millisecond)

	// Server buttons - simple layout
	setupButton := widget.NewButton("⚙️ Setup", func() {
		setupServer(GetGlobalPackDir())
	})

	startButton := widget.NewButton("▶️ Start", func() {
		startServer(GetGlobalPackDir())
	})

	stopButton := widget.NewButton("⏹️ Stop", func() {
		dialog.ShowInformation("Stop Server", "Stop functionality coming soon", Window)
	})

	cleanButton := widget.NewButton("🗑️ Clean", func() {
		dialog.ShowConfirm("Clean Server", "Remove server files?", func(confirmed bool) {
			if confirmed {
				cleanServer(GetGlobalPackDir())
			}
		}, Window)
	})

	// Simple grid layout
	buttonGrid := container.NewGridWithColumns(4, setupButton, startButton, stopButton, cleanButton)

	// Minimal status
	statusLabel := widget.NewLabel("Status: Ready • Port: 25565 • Directory: .run/")

	// Compact layout
	content := container.NewVBox(
		packDirEntry,
		buttonGrid,
		statusLabel,
	)

	return content
}

// Server functions
func setupServer(packDir string) {
	RunPwCommand("Server Setup", []string{"server", "--setup"}, packDir)
}

func startServer(packDir string) {
	RunPwCommand("Start Server", []string{"server", "--start"}, packDir)
}

func cleanServer(packDir string) {
	dialog.ShowConfirm("Clean Server", "This will remove all server files in .run/ directory. Continue?", func(confirmed bool) {
		if confirmed {
			RunPwCommand("Clean Server", []string{"server", "--clean"}, packDir)
		}
	}, Window)
}
