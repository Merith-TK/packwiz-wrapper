package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/Merith-TK/packwiz-wrapper/internal/core"
)

// CreatePackInfoTab creates the pack information tab
func CreatePackInfoTab() fyne.CanvasObject {
	// Header card
	headerCard := widget.NewCard("📦 Pack Information", 
		"View and manage your modpack details", 
		widget.NewRichText())

	// Pack directory selection section
	packDirEntry := widget.NewEntry()
	packDirEntry.SetPlaceHolder("Pack directory path...")
	packDirEntry.SetText(GetGlobalPackDir())

	// Update global pack dir when entry changes
	packDirEntry.OnChanged = func(text string) {
		SetGlobalPackDir(text)
	}

	var packInfoText *widget.RichText

	packDirButton := widget.NewButton("📁 Browse", func() {
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
				if packInfoText != nil {
					refreshPackInfo(folder.Path(), packInfoText)
				}
			}
		}, Window)
		folderDialog.Show()
	})

	packDirContainer := container.NewBorder(nil, nil, packDirButton, nil, packDirEntry)

	// Directory selection card
	dirSelectionCard := widget.NewCard("📂 Pack Directory", 
		"Select your modpack folder",
		packDirContainer)

	// Action buttons
	refreshButton := widget.NewButton("🔄 Refresh Info", func() {
		refreshPackInfo(GetGlobalPackDir(), packInfoText)
	})
	refreshButton.Importance = widget.MediumImportance

	refreshPackButton := widget.NewButton("🔧 Refresh Pack", func() {
		refreshPack(GetGlobalPackDir(), packInfoText)
	})
	refreshPackButton.Importance = widget.HighImportance

	actionsContainer := container.NewHBox(refreshButton, refreshPackButton)
	actionsCard := widget.NewCard("⚡ Quick Actions", 
		"Manage your pack data",
		actionsContainer)

	// Pack info display
	packInfoText = widget.NewRichText()
	packInfoText.ParseMarkdown(`**No pack loaded**

📁 **Getting Started:**
- Select a pack directory above by clicking "Browse"
- Or type a path like "./" for current directory
- The directory should contain a **pack.toml** file

📋 **What you'll see here:**
- Pack name and description
- Author information  
- Minecraft version
- Mod count and list
- Pack format details

💡 **Tip:** Use "Refresh Pack" to update mod information from remote sources`)

	// Wrap the pack info in a scrollable container
	packInfoScroll := container.NewScroll(packInfoText)
	packInfoScroll.SetMinSize(fyne.NewSize(500, 300))

	// Pack details card
	detailsCard := widget.NewCard("📋 Pack Details", 
		"Current pack information",
		packInfoScroll)

	// Register callback for pack directory changes
	RegisterPackDirCallback(func(dir string) {
		packDirEntry.SetText(dir)
		refreshPackInfo(dir, packInfoText)
	})

	// Don't auto-refresh on creation to avoid logging before GlobalLogWidget is ready
	// User can click refresh when they're ready

	// Layout everything
	content := container.NewVBox(
		headerCard,
		widget.NewSeparator(),
		dirSelectionCard,
		widget.NewSeparator(),
		actionsCard,
		widget.NewSeparator(),
		detailsCard,
	)

	// Wrap in scroll container
	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(600, 400))
	
	return scroll
}

func refreshPackInfo(packDir string, infoWidget *widget.RichText) {
	if packDir == "" {
		packDir = "./"
	}

	logger := NewGUILogger(GlobalLogWidget)
	manager := core.NewManager(logger)

	logger.Info("Loading pack info from: %s", packDir)

	packInfo, err := manager.GetPackInfo(packDir)
	if err != nil {
		errorMsg := fmt.Sprintf(`**Error loading pack**

**Directory:** %s
**Error:** %s

**Tips:**
- Make sure the directory contains a pack.toml file
- Or check if there's a .minecraft subdirectory with pack.toml
- Verify the path is correct and accessible`, packDir, err.Error())

		infoWidget.ParseMarkdown(errorMsg)
		logger.Error("Failed to load pack info: %s", err.Error())
		return
	}

	logger.Info("Successfully loaded pack: %s", packInfo.Name)

	info := fmt.Sprintf(`**Pack Name:** %s
**Author:** %s
**MC Version:** %s
**Pack Format:** %s
**Description:** %s
**Mod Count:** %d
**Pack Directory:** %s

*Pack loaded successfully! You can now use other tabs to manage mods, import/export, or start a server.*`,
		packInfo.Name,
		packInfo.Author,
		packInfo.McVersion,
		packInfo.PackFormat,
		packInfo.Description,
		packInfo.ModCount,
		packInfo.PackDir,
	)

	infoWidget.ParseMarkdown(info)
}

func refreshPack(packDir string, infoWidget *widget.RichText) {
	if packDir == "" {
		packDir = "./"
	}

	logger := NewGUILogger(GlobalLogWidget)
	manager := core.NewManager(logger)

	err := manager.RefreshPack(packDir)
	if err != nil {
		infoWidget.ParseMarkdown(fmt.Sprintf("**Refresh Error:** %s", err.Error()))
		return
	}

	// Refresh the pack info after successful refresh
	refreshPackInfo(packDir, infoWidget)
}
