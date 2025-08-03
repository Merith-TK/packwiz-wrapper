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
	// Pack directory selection
	packDirEntry := widget.NewEntry()
	packDirEntry.SetPlaceHolder("Select pack directory...")
	packDirEntry.SetText(GetGlobalPackDir()) // Use global pack dir

	// Update global pack dir when entry changes
	packDirEntry.OnChanged = func(text string) {
		SetGlobalPackDir(text)
	}

	var packInfoText *widget.RichText // Declare here so it's accessible in the closure

	packDirButton := widget.NewButton("Browse", func() {
		// Create a folder dialog
		folderDialog := dialog.NewFolderOpen(func(folder fyne.ListableURI, err error) {
			if err != nil {
				// Handle error - could show an error dialog here
				if GlobalLogWidget != nil {
					GlobalLogWidget.ParseMarkdown(GlobalLogWidget.String() + "\n[ERROR] Failed to select folder: " + err.Error())
				}
				return
			}
			if folder != nil {
				// Set the selected folder path
				packDirEntry.SetText(folder.Path())
				SetGlobalPackDir(folder.Path())
				// Automatically refresh pack info when folder is selected
				if packInfoText != nil {
					refreshPackInfo(folder.Path(), packInfoText)
				}
			}
		}, fyne.CurrentApp().Driver().AllWindows()[0])

		// Set the dialog to start in the current directory
		folderDialog.Show()
	})

	packDirContainer := container.NewBorder(nil, nil, packDirButton, nil, packDirEntry)

	// Pack info display
	packInfoText = widget.NewRichText() // Initialize here
	packInfoText.ParseMarkdown(`**No pack loaded**

Select a pack directory above by:
- Clicking "Browse" to pick a folder
- Or typing a path like "./" for current directory

The directory should contain a **pack.toml** file or have a **.minecraft** subdirectory with pack.toml.`)

	// Wrap the pack info in a scrollable container
	packInfoScroll := container.NewScroll(packInfoText)
	packInfoScroll.SetMinSize(fyne.NewSize(400, 200))

	refreshButton := widget.NewButton("Refresh Pack Info", func() {
		refreshPackInfo(GetGlobalPackDir(), packInfoText)
	})

	refreshPackButton := widget.NewButton("Refresh Pack", func() {
		refreshPack(GetGlobalPackDir(), packInfoText)
	})

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Pack Directory:"),
			packDirContainer,
			widget.NewSeparator(),
			container.NewHBox(refreshButton, refreshPackButton),
			widget.NewSeparator(),
			widget.NewLabel("Pack Information:"),
		),
		nil, nil, nil,
		packInfoScroll,
	)
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
