package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/Merith-TK/packwiz-wrapper/internal/core"
)

// CreateWelcomeTab creates a welcome/getting started tab
func CreateWelcomeTab() fyne.CanvasObject {
	// Welcome header
	welcomeTitle := widget.NewCard("Welcome to PackWrap2 GUI", 
		"Your Minecraft modpack management tool", 
		widget.NewRichText())
	
	// Quick start steps
	quickStartSteps := widget.NewRichText()
	quickStartSteps.ParseMarkdown(`## Getting Started

**Step 1: Choose Your Pack Directory**
- Click "Browse" below to select your modpack folder
- Or create a new pack by selecting an empty folder
- The folder should contain (or will contain) a **pack.toml** file

**Step 2: Load or Create Your Pack**
- If you have an existing pack, it will load automatically
- For new packs, use "Create New Pack" to set up the basics
- Use "Import from File" if you have a mod list to import

**Step 3: Manage Your Mods**
- Use the "Mods" tab to add, remove, or update mods
- Search for mods by name or add them via URL/ID
- All changes are automatically saved to your pack

**Step 4: Test and Share**
- Use the "Server" tab to test your pack locally
- Export your pack in various formats for sharing`)

	// Status card showing current pack state (declare early)
	statusCard := widget.NewCard("Pack Status", "No pack loaded", 
		widget.NewLabel("Select a pack directory to get started"))

	// Pack directory selection (prominent)
	packDirEntry := widget.NewEntry()
	packDirEntry.SetPlaceHolder("No pack directory selected - click Browse to get started!")
	packDirEntry.SetText(GetGlobalPackDir())
	
	// Update global pack dir when entry changes
	packDirEntry.OnChanged = func(text string) {
		SetGlobalPackDir(text)
		updateWelcomeStatus(text, statusCard)
	}

	// Large, prominent browse button
	browseButton := widget.NewButton("📁 Browse for Pack Directory", func() {
		folderDialog := dialog.NewFolderOpen(func(folder fyne.ListableURI, err error) {
			if err != nil {
				if GlobalLogWidget != nil {
					GlobalLogWidget.ParseMarkdown(GlobalLogWidget.String() + "\n[ERROR] Failed to select folder: " + err.Error())
				}
				return
			}
			if folder != nil {
				packDirEntry.SetText(folder.Path())
				SetGlobalPackDir(folder.Path())
				updateWelcomeStatus(folder.Path(), statusCard)
			}
		}, Window)
		folderDialog.Show()
	})
	browseButton.Resize(fyne.NewSize(300, 50))
	browseButton.Importance = widget.HighImportance

	// Create new pack button
	createPackButton := widget.NewButton("✨ Create New Pack", func() {
		showCreatePackDialog()
	})
	
	// Import pack button
	importPackButton := widget.NewButton("📥 Import Pack from File", func() {
		fileDialog := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
			if err != nil {
				if GlobalLogWidget != nil {
					GlobalLogWidget.ParseMarkdown(GlobalLogWidget.String() + "\n[ERROR] Failed to select file: " + err.Error())
				}
				return
			}
			if file != nil {
				// Switch to import tab and set the file
				// This would need to be implemented to communicate with the import tab
				if GlobalLogWidget != nil {
					GlobalLogWidget.ParseMarkdown(GlobalLogWidget.String() + "\n[INFO] Selected import file: " + file.URI().Path())
				}
				file.Close()
			}
		}, Window)
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt", ".md", ".json"}))
		fileDialog.Show()
	})

	// Quick action buttons
	quickActions := container.NewGridWithColumns(3, 
		createPackButton,
		importPackButton,
		widget.NewButton("🔧 Open Settings", func() {
			// Placeholder for settings
			dialog.ShowInformation("Settings", "Settings panel coming soon!", Window)
		}),
	)

	// Register callback for pack directory changes
	RegisterPackDirCallback(func(dir string) {
		packDirEntry.SetText(dir)
		updateWelcomeStatus(dir, statusCard)
	})

	// Update status immediately
	updateWelcomeStatus(GetGlobalPackDir(), statusCard)

	// Recent packs (placeholder for future feature)
	recentPacks := widget.NewCard("Recent Packs", "Quick access to your recent modpacks",
		widget.NewLabel("No recent packs found\n(This feature will be added in a future update)"))

	// Layout everything
	content := container.NewVBox(
		welcomeTitle,
		widget.NewSeparator(),
		
		// Pack directory selection section
		widget.NewCard("Select Pack Directory", "Choose your modpack folder to get started",
			container.NewVBox(
				packDirEntry,
				browseButton,
			),
		),
		
		widget.NewSeparator(),
		
		// Quick actions
		widget.NewCard("Quick Actions", "Common tasks to get you started",
			quickActions,
		),
		
		// Two column layout for status and instructions
		container.NewGridWithColumns(2,
			statusCard,
			recentPacks,
		),
		
		widget.NewSeparator(),
		quickStartSteps,
	)

	// Wrap in scroll container
	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(600, 400))
	
	return scroll
}

// updateWelcomeStatus updates the status card based on the current pack directory
func updateWelcomeStatus(packDir string, statusCard *widget.Card) {
	if statusCard == nil {
		return
	}
	
	if packDir == "" || packDir == "./" {
		statusCard.SetTitle("Pack Status")
		statusCard.SetSubTitle("No pack loaded")
		statusCard.SetContent(widget.NewLabel("Select a pack directory to get started"))
		return
	}
	
	// Try to get pack info
	logger := NewGUILogger(GlobalLogWidget)
	manager := PackManager
	if manager == nil {
		manager = core.NewManager(logger)
	}
	
	packInfo, err := manager.GetPackInfo(packDir)
	if err != nil {
		statusCard.SetTitle("Pack Status")
		statusCard.SetSubTitle("Invalid or empty pack directory")
		statusCard.SetContent(widget.NewLabel("No pack.toml found. Use 'Create New Pack' to set up a new modpack."))
		return
	}
	
	statusCard.SetTitle("Pack Loaded Successfully!")
	statusCard.SetSubTitle(packInfo.Name)
	statusCard.SetContent(widget.NewRichText())
	statusCardContent := statusCard.Content.(*widget.RichText)
	statusCardContent.ParseMarkdown(fmt.Sprintf(`**Author:** %s
**MC Version:** %s
**Mod Count:** %d

✅ Ready to manage mods!`, 
		packInfo.Author, 
		packInfo.McVersion, 
		packInfo.ModCount))
}

// showCreatePackDialog shows a dialog for creating a new pack
func showCreatePackDialog() {
	// Pack name entry
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("My Awesome Modpack")
	
	// Author entry
	authorEntry := widget.NewEntry()
	authorEntry.SetPlaceHolder("Your Name")
	
	// MC Version selection
	mcVersionSelect := widget.NewSelect(
		[]string{"1.21.4", "1.21.3", "1.21.1", "1.21", "1.20.6", "1.20.4", "1.20.1", "1.19.4", "1.19.2"},
		nil,
	)
	mcVersionSelect.SetSelected("1.21.4")
	
	// Description entry
	descEntry := widget.NewMultiLineEntry()
	descEntry.SetPlaceHolder("A brief description of your modpack...")
	descEntry.Resize(fyne.NewSize(400, 100))
	
	// Form content
	content := container.NewVBox(
		widget.NewLabel("Create a new modpack in the selected directory:"),
		widget.NewForm(
			widget.NewFormItem("Pack Name", nameEntry),
			widget.NewFormItem("Author", authorEntry),
			widget.NewFormItem("MC Version", mcVersionSelect),
			widget.NewFormItem("Description", descEntry),
		),
	)
	
	// Create dialog
	createDialog := dialog.NewCustomConfirm(
		"Create New Pack",
		"Create",
		"Cancel",
		content,
		func(create bool) {
			if create && nameEntry.Text != "" {
				createNewPack(nameEntry.Text, authorEntry.Text, mcVersionSelect.Selected, descEntry.Text)
			}
		},
		Window,
	)
	
	createDialog.Resize(fyne.NewSize(500, 400))
	createDialog.Show()
}

// createNewPack creates a new pack with the given parameters
func createNewPack(name, author, mcVersion, description string) {
	packDir := GetGlobalPackDir()
	if packDir == "" || packDir == "./" {
		dialog.ShowError(fmt.Errorf("please select a directory first"), Window)
		return
	}
	
	// This would need to be implemented in the core manager
	// For now, just show a command that the user could run
	ShowCommandOutput("Create New Pack", "./pw.exe", []string{
		"init", 
		"--name", name,
		"--author", author, 
		"--mc-version", mcVersion,
		"--description", description,
	}, packDir)
}
