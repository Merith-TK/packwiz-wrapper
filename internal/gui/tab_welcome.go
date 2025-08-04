package gui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/Merith-TK/packwiz-wrapper/internal/core"
)

// CreateWelcomeTab creates a compact welcome/getting started tab
func CreateWelcomeTab() fyne.CanvasObject {
	// Pack directory section
	packDirEntry := widget.NewEntry()
	packDirEntry.SetPlaceHolder("Select your modpack directory...")
	packDirEntry.SetText(GetGlobalPackDir())

	browseButton := widget.NewButton("📁 Browse", func() {
		ShowEnhancedFolderDialog(func(selectedPath string) {
			if selectedPath != "" {
				packDirEntry.SetText(selectedPath)
				SetGlobalPackDir(selectedPath)
			}
		})
	})
	browseButton.Importance = widget.HighImportance

	packDirSection := container.NewBorder(nil, nil, nil, browseButton, packDirEntry)
	packDirCard := widget.NewCard("📂 Pack Directory",
		"Choose your modpack folder to get started",
		packDirSection)

	// Status card that will be updated dynamically
	statusCard := widget.NewCard("📊 Pack Status", "No pack loaded",
		widget.NewLabel("Select a pack directory to get started"))

	// Update global pack dir when entry changes (debounced to avoid checking on every keystroke)
	packDirEntry.OnChanged = debouncePathUpdate(func(text string) {
		SetGlobalPackDir(text)
		updateWelcomeStatus(text, statusCard)
	}, 500*time.Millisecond)

	// Quick Actions
	createButton := widget.NewButton("✨ Create New Pack", func() {
		showCreatePackDialog()
	})
	createButton.Importance = widget.HighImportance

	importButton := widget.NewButton("📥 Import Pack", func() {
		fileDialog := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
			if err != nil {
				if GlobalLogWidget != nil {
					GlobalLogWidget.ParseMarkdown(GlobalLogWidget.String() + "\n[ERROR] Failed to select file: " + err.Error())
				}
				return
			}
			if file != nil {
				if GlobalLogWidget != nil {
					GlobalLogWidget.ParseMarkdown(GlobalLogWidget.String() + "\n[INFO] Selected import file: " + file.URI().Path())
				}
				file.Close()
			}
		}, Window)
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt", ".md", ".json"}))
		fileDialog.Show()
	})

	settingsButton := widget.NewButton("⚙️ Settings", func() {
		dialog.ShowInformation("Settings", "Settings panel coming soon!", Window)
	})

	quickActionsGrid := container.NewGridWithColumns(3, createButton, importButton, settingsButton)

	// Getting Started Guide
	guideContent := widget.NewRichTextFromMarkdown(`**Quick Start Guide:**

1. **📂 Choose Directory** - Select or create a modpack folder
2. **🎯 Setup Pack** - Create new or load existing pack
3. **🔧 Manage Mods** - Use Mods tab to add/remove mods
4. **🚀 Test & Share** - Use Server tab to test locally`)

	// Two-column layout: Status on left, guide on right
	leftColumn := container.NewVBox(statusCard)
	rightColumn := container.NewVBox(
		widget.NewCard("📋 Getting Started", "", guideContent),
	)

	mainContent := container.NewGridWithColumns(2, leftColumn, rightColumn)

	// Register callback for pack directory changes from other tabs
	RegisterPackDirCallback(func(dir string) {
		packDirEntry.SetText(dir)
		updateWelcomeStatus(dir, statusCard)
	})

	// Update status immediately with current directory
	updateWelcomeStatus(GetGlobalPackDir(), statusCard)

	// Main layout - compact without scroll
	content := container.NewVBox(
		packDirCard,
		widget.NewCard("🚀 Quick Actions", "", quickActionsGrid),
		mainContent,
	)

	return content
}

// updateWelcomeStatus updates the status card based on the current pack directory
func updateWelcomeStatus(packDir string, statusCard *widget.Card) {
	if statusCard == nil {
		return
	}

	if packDir == "" || packDir == "./" {
		statusCard.SetTitle("📊 Pack Status")
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
		statusCard.SetTitle("📊 Pack Status")
		statusCard.SetSubTitle("Directory selected")
		statusCard.SetContent(widget.NewLabel("No pack found. Use 'Create New Pack' to set up here."))
		return
	}

	statusCard.SetTitle("✅ Pack Loaded")
	statusCard.SetSubTitle(packInfo.Name)
	statusCard.SetContent(widget.NewRichTextFromMarkdown(fmt.Sprintf(`**Author:** %s  
**MC Version:** %s  
**Mod Count:** %d  

Ready to manage!`,
		packInfo.Author,
		packInfo.McVersion,
		packInfo.ModCount)))
}

// showCreatePackDialog shows a dialog for creating a new pack
func showCreatePackDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("My Awesome Modpack")

	authorEntry := widget.NewEntry()
	authorEntry.SetPlaceHolder("Your Name")

	// TODO: Autopopulate with common versions from api
	mcVersionSelect := widget.NewSelect(
		[]string{"1.21.4", "1.21.3", "1.21.1", "1.21", "1.20.6", "1.20.4", "1.20.1", "1.19.4", "1.19.2", "1.18.2"},
		nil,
	)
	mcVersionSelect.SetSelected("1.21.4")

	descEntry := widget.NewMultiLineEntry()
	descEntry.SetPlaceHolder("A brief description of your modpack...")
	descEntry.Resize(fyne.NewSize(300, 60))

	content := container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Pack Name", nameEntry),
			widget.NewFormItem("Author", authorEntry),
			widget.NewFormItem("MC Version", mcVersionSelect),
			widget.NewFormItem("Description", descEntry),
		),
	)

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

	createDialog.Resize(fyne.NewSize(400, 350))
	createDialog.Show()
}

// createNewPack creates a new pack with the given parameters
func createNewPack(name, author, mcVersion, description string) {
	packDir := GetGlobalPackDir()
	if packDir == "" || packDir == "./" {
		dialog.ShowError(fmt.Errorf("please select a directory first"), Window)
		return
	}

	RunPwCommand("Create New Pack", []string{
		"init",
		"--name", name,
		"--author", author,
		"--mc-version", mcVersion,
		"--description", description,
	}, packDir)
}
