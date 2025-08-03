package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// CreateServerTab creates the server management tab
func CreateServerTab() fyne.CanvasObject {
	// Pack directory input that syncs with global state
	packDirEntry := widget.NewEntry()
	packDirEntry.SetPlaceHolder("Pack directory (or leave empty for current)")
	packDirEntry.SetText(GetGlobalPackDir())

	// Register callback to update entry when global pack dir changes
	RegisterPackDirCallback(func(dir string) {
		packDirEntry.SetText(dir)
	})

	packDirEntry.OnChanged = func(text string) {
		SetGlobalPackDir(text)
	}

	setupButton := widget.NewButton("Setup Test Server", func() {
		setupServer(GetGlobalPackDir())
	})

	startButton := widget.NewButton("Start Server", func() {
		startServer(GetGlobalPackDir())
	})

	cleanButton := widget.NewButton("Clean Server Files", func() {
		cleanServer(GetGlobalPackDir())
	})

	return container.NewVBox(
		widget.NewLabel("Pack Directory:"),
		packDirEntry,
		widget.NewSeparator(),
		widget.NewLabel("Development Server:"),
		setupButton,
		startButton,
		cleanButton,
		widget.NewSeparator(),
		widget.NewLabel("Server will be created in .run/ directory"),
		widget.NewLabel("Use the logs tab to monitor server output"),
	)
}

// Server functions
func setupServer(packDir string) {
	ShowCommandOutput("Server Setup", "./pw.exe", []string{"server", "--setup"}, packDir)
}

func startServer(packDir string) {
	ShowCommandOutput("Start Server", "./pw.exe", []string{"server", "--start"}, packDir)
}

func cleanServer(packDir string) {
	dialog.ShowConfirm("Clean Server", "This will remove all server files in .run/ directory. Continue?", func(confirmed bool) {
		if confirmed {
			ShowCommandOutput("Clean Server", "./pw.exe", []string{"server", "--clean"}, packDir)
		}
	}, Window)
}
