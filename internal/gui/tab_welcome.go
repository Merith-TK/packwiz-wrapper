//go:build gui

package gui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

	// Pack information display widget
	packInfoWidget := widget.NewRichText()
	packInfoWidget.Wrapping = fyne.TextWrapWord
	packInfoWidget.ParseMarkdown(`# No Pack Loaded

## 📁 Getting Started
- Click **"📁 Browse"** above to select your modpack directory
- The directory should contain a **pack.toml** file
- Or use **"✨ Create New Pack"** to set up a new modpack

## 💡 What You'll See Here
Once a pack is loaded, this area will display:
- **Pack Name** and description
- **Author** information
- **Minecraft Version** compatibility
- **Mod Count** and pack format
- **Pack Directory** location

---
*Ready to get started? Select a pack directory above!*`)

	// Update global pack dir when entry changes (debounced to avoid checking on every keystroke)
	packDirEntry.OnChanged = debouncePathUpdate(func(text string) {
		SetGlobalPackDir(text)
		updateWelcomeStatus(text, statusCard, packInfoWidget)
	}, 500*time.Millisecond)

	// Action buttons
	refreshButton := widget.NewButton("🔄 Refresh Info", func() {
		refreshWelcomePackInfo(GetGlobalPackDir(), statusCard, packInfoWidget)
	})
	refreshButton.Importance = widget.MediumImportance

	refreshPackButton := widget.NewButton("🔧 Refresh Pack", func() {
		refreshWelcomePack(GetGlobalPackDir(), statusCard, packInfoWidget)
	})
	refreshPackButton.Importance = widget.HighImportance

	// Quick Actions
	createButton := widget.NewButton("✨ Create New Pack", func() {
		showCreatePackDialog()
	})
	createButton.Importance = widget.HighImportance

	quickActionsGrid := container.NewGridWithColumns(3, refreshButton, refreshPackButton, createButton)

	// Two-column layout: Status and detailed pack info
	leftColumn := container.NewVBox(statusCard)

	// Pack info in a scrollable container
	packInfoScroll := container.NewScroll(packInfoWidget)
	packInfoScroll.SetMinSize(fyne.NewSize(450, 350))
	rightColumn := container.NewVBox(
		widget.NewCard("📄 Pack Information", "", packInfoScroll),
	)

	mainContent := container.NewGridWithColumns(2, leftColumn, rightColumn)

	// Register callback for pack directory changes from other tabs
	RegisterPackDirCallback(func(dir string) {
		fyne.Do(func() {
			packDirEntry.SetText(dir)
			updateWelcomeStatus(dir, statusCard, packInfoWidget)
		})
	})

	// Update status immediately with current directory
	updateWelcomeStatus(GetGlobalPackDir(), statusCard, packInfoWidget)

	// Main layout - compact without scroll
	content := container.NewVBox(
		packDirCard,
		widget.NewCard("🚀 Quick Actions", "", quickActionsGrid),
		mainContent,
	)

	return content
}

// updateWelcomeStatus updates the status card based on the current pack directory (thread-safe)
func updateWelcomeStatus(packDir string, statusCard *widget.Card, packInfoWidget *widget.RichText) {
	// Ensure all UI updates happen on the main thread
	fyne.Do(func() {
		updateWelcomeStatusUI(packDir, statusCard, packInfoWidget)
	})
}

// updateWelcomeStatusUI performs the actual UI updates (main thread only)
func updateWelcomeStatusUI(packDir string, statusCard *widget.Card, packInfoWidget *widget.RichText) {
	if statusCard == nil || packInfoWidget == nil {
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

		// Update pack info widget with error message
		errorMsg := fmt.Sprintf(`# No Pack Found

## 📁 Directory Information
- **Selected Path:** %s
- **Error:** %s

## 💡 Troubleshooting Tips
- Make sure the directory contains a **pack.toml** file
- Check if there's a **.minecraft** subdirectory with pack.toml
- Verify the path is correct and accessible
- Use **"✨ Create New Pack"** to set up a new pack here

---
*Need help? Check the path or create a new pack!*`, packDir, err.Error())

		packInfoWidget.ParseMarkdown(errorMsg)
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

	// Update detailed pack info widget
	detailedInfo := fmt.Sprintf(`# %s

## 📋 Pack Details
- **Author:** %s
- **Minecraft Version:** %s
- **Pack Format:** %s
- **Mod Count:** %d
- **Pack Directory:** %s

## 📝 Description
%s

---
*Pack loaded successfully! You can now use other tabs to manage mods, import/export, or start a server.*`,
		packInfo.Name,
		packInfo.Author,
		packInfo.McVersion,
		packInfo.PackFormat,
		packInfo.ModCount,
		packInfo.PackDir,
		packInfo.Description,
	)

	packInfoWidget.ParseMarkdown(detailedInfo)
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

// refreshWelcomePackInfo refreshes the pack information display
func refreshWelcomePackInfo(packDir string, statusCard *widget.Card, packInfoWidget *widget.RichText) {
	if packDir == "" {
		packDir = "./"
	}

	logger := NewGUILogger(GlobalLogWidget)

	logger.Info("Loading pack info from: %s", packDir)

	// Use the same logic as updateWelcomeStatus but force refresh
	fyne.Do(func() {
		updateWelcomeStatusUI(packDir, statusCard, packInfoWidget)
	})
}

// refreshWelcomePack refreshes the pack using packwiz refresh command
func refreshWelcomePack(packDir string, statusCard *widget.Card, packInfoWidget *widget.RichText) {
	if packDir == "" {
		packDir = "./"
	}

	logger := NewGUILogger(GlobalLogWidget)
	manager := core.NewManager(logger)

	err := manager.RefreshPack(packDir)
	if err != nil {
		fyne.Do(func() {
			packInfoWidget.ParseMarkdown(fmt.Sprintf("**Refresh Error:** %s", err.Error()))
		})
		return
	}

	// Refresh the pack info after successful refresh
	refreshWelcomePackInfo(packDir, statusCard, packInfoWidget)
}
