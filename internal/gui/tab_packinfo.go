package gui

import (
	"fmt"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/Merith-TK/packwiz-wrapper/internal/core"
)

// CreatePackInfoTab creates the pack information tab
func CreatePackInfoTab() fyne.CanvasObject {
	// Pack directory selection at top - compact version
	packDirEntry := widget.NewEntry()
	packDirEntry.SetPlaceHolder("Pack directory path...")

	// Convert relative path to absolute path for display
	currentDir := GetGlobalPackDir()
	if currentDir == "./" || currentDir == "." {
		if wd, err := os.Getwd(); err == nil {
			currentDir = wd
		}
	}
	packDirEntry.SetText(currentDir)

	// Update global pack dir when entry changes (debounced to avoid checking on every keystroke)
	packDirEntry.OnChanged = debouncePathUpdate(func(text string) {
		SetGlobalPackDir(text)
	}, 500*time.Millisecond)

	var packInfoText *widget.RichText

	packDirButton := widget.NewButton("📁", func() {
		ShowEnhancedFolderDialog(func(selectedPath string) {
			if selectedPath != "" {
				packDirEntry.SetText(selectedPath)
				SetGlobalPackDir(selectedPath)
				if packInfoText != nil {
					refreshPackInfo(selectedPath, packInfoText)
				}
			}
		})
	})
	packDirButton.Importance = widget.LowImportance

	packDirContainer := container.NewBorder(nil, nil, nil, packDirButton, packDirEntry)

	// Compact directory selection
	dirSelectionCard := widget.NewCard("📂 Pack Directory", "", packDirContainer)

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
	actionsCard := widget.NewCard("", "", actionsContainer)

	// Pack info display
	packInfoText = widget.NewRichText()
	// Initialize pack info text widget with proper wrapping
	packInfoText = widget.NewRichText()
	packInfoText.Wrapping = fyne.TextWrapWord
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

	// Wrap the pack info in a scrollable container with proper sizing
	packInfoScroll := container.NewScroll(packInfoText)
	packInfoScroll.SetMinSize(fyne.NewSize(400, 200))

	// Register callback for pack directory changes
	RegisterPackDirCallback(func(dir string) {
		packDirEntry.SetText(dir)
		refreshPackInfo(dir, packInfoText)
	})

	// Don't auto-refresh on creation to avoid logging before GlobalLogWidget is ready
	// User can click refresh when they're ready

	// Use BorderContainer to give more space to the pack info area
	content := container.NewBorder(
		container.NewVBox(dirSelectionCard, actionsCard), // top
		nil,              // bottom
		nil,              // left
		nil,              // right
		packInfoScroll,   // center - takes remaining space
	)

	return content
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
