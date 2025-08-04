package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// CreateServerTab creates the server management tab
func CreateServerTab() fyne.CanvasObject {
	// Header card
	headerCard := widget.NewCard("🖥️ Development Server", 
		"Test your modpack with a local Minecraft server", 
		widget.NewRichText())

	// Pack directory section
	packDirEntry := widget.NewEntry()
	packDirEntry.SetPlaceHolder("Pack directory (synced globally)")
	packDirEntry.SetText(GetGlobalPackDir())

	RegisterPackDirCallback(func(dir string) {
		packDirEntry.SetText(dir)
	})

	packDirEntry.OnChanged = func(text string) {
		SetGlobalPackDir(text)
	}

	packDirCard := widget.NewCard("📂 Pack Directory", 
		"Server will be created based on this pack",
		packDirEntry)

	// Server setup section
	setupButton := widget.NewButton("⚙️ Setup Server", func() {
		setupServer(GetGlobalPackDir())
	})
	setupButton.Importance = widget.HighImportance

	setupInfo := widget.NewRichText()
	setupInfo.ParseMarkdown(`**Setup Process:**
- Downloads Minecraft server
- Installs Forge/Fabric loader
- Copies mods from your pack
- Configures server settings`)

	setupCard := widget.NewCard("⚙️ Server Setup", 
		"Initialize a new test server",
		container.NewVBox(
			setupButton,
			widget.NewSeparator(),
			setupInfo,
		))

	// Server control section
	startButton := widget.NewButton("▶️ Start Server", func() {
		startServer(GetGlobalPackDir())
	})
	startButton.Importance = widget.HighImportance

	stopButton := widget.NewButton("⏹️ Stop Server", func() {
		// This would need to be implemented
		dialog.ShowInformation("Stop Server", "Server stop functionality will be added in a future update", Window)
	})

	restartButton := widget.NewButton("🔄 Restart Server", func() {
		// This would need to be implemented
		dialog.ShowInformation("Restart Server", "Server restart functionality will be added in a future update", Window)
	})

	controlActions := container.NewGridWithColumns(3, startButton, stopButton, restartButton)

	controlCard := widget.NewCard("🎮 Server Control", 
		"Start, stop, and manage your test server",
		controlActions)

	// Server maintenance section
	cleanButton := widget.NewButton("🗑️ Clean Server Files", func() {
		dialog.ShowConfirm("Clean Server", 
			"This will remove all server files in the .run/ directory.\nYour pack files will not be affected.\n\nContinue?", 
			func(confirmed bool) {
				if confirmed {
					cleanServer(GetGlobalPackDir())
				}
			}, Window)
	})
	cleanButton.Importance = widget.DangerImportance

	backupButton := widget.NewButton("💾 Backup Server", func() {
		// This would need to be implemented
		dialog.ShowInformation("Backup Server", "Server backup functionality will be added in a future update", Window)
	})

	maintenanceActions := container.NewHBox(cleanButton, backupButton)

	maintenanceCard := widget.NewCard("🔧 Maintenance", 
		"Clean up and backup server data",
		maintenanceActions)

	// Server status section
	statusText := widget.NewRichText()
	statusText.ParseMarkdown(`**Server Status:** Not Running

**Server Directory:** .run/
**Server Type:** Will be determined by pack
**Port:** 25565 (default)

**💡 Quick Start:**
1. Click "Setup Server" to initialize
2. Click "Start Server" to begin testing
3. Use the Logs tab to monitor output
4. Connect with localhost:25565`)

	statusCard := widget.NewCard("📊 Server Status", 
		"Current server information",
		statusText)

	// Help section
	helpText := widget.NewRichText()
	helpText.ParseMarkdown(`## 🖥️ Server Guide

### What This Does:
- Creates a local Minecraft server for testing your pack
- Automatically installs the correct Forge/Fabric version
- Copies all mods from your pack to the server
- Provides a safe environment to test mod compatibility

### File Locations:
- **Server files:** .run/ directory in your pack folder
- **World data:** .run/world/
- **Server logs:** .run/logs/
- **Configuration:** .run/server.properties

### Testing Workflow:
1. **Setup** - Initialize server with your pack
2. **Start** - Launch the server
3. **Connect** - Join with Minecraft client (localhost:25565)
4. **Test** - Verify mods work correctly
5. **Clean** - Remove server files when done

### 🔒 Safety Notes:
- Server runs locally only (not accessible from internet)
- Your original pack files are never modified
- Server files are kept separate in .run/ directory`)

	helpCard := widget.NewCard("ℹ️ Help & Guide", 
		"Learn how to use the development server",
		helpText)

	// Layout everything
	content := container.NewVBox(
		headerCard,
		widget.NewSeparator(),
		packDirCard,
		widget.NewSeparator(),
		
		// Two-column layout for setup and control
		container.NewGridWithColumns(2,
			setupCard,
			controlCard,
		),
		
		widget.NewSeparator(),
		
		// Two-column layout for maintenance and status
		container.NewGridWithColumns(2,
			maintenanceCard,
			statusCard,
		),
		
		widget.NewSeparator(),
		helpCard,
	)

	// Wrap in scroll container
	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(600, 400))
	
	return scroll
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
