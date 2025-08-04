//go:build gui

package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/Merith-TK/packwiz-wrapper/internal/core"
)

// CreateModsTab creates the mod management tab with dual-pane layout
func CreateModsTab() fyne.CanvasObject {
	// Mod list data
	var modData []*ModDisplayInfo
	var selectedModIndex = -1
	var selectedMod *ModDisplayInfo

	// LEFT PANE: Mod list
	modList := widget.NewList(
		func() int { return len(modData) },
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(modData) {
				mod := modData[id]
				label := obj.(*widget.Label)
				label.SetText(fmt.Sprintf("%s (%s)", mod.Name, mod.Version))
			}
		},
	)

	modList.OnSelected = func(id widget.ListItemID) {
		selectedModIndex = id
		if id < len(modData) {
			selectedMod = modData[id]
		}
	}

	modList.OnUnselected = func(id widget.ListItemID) {
		selectedModIndex = -1
		selectedMod = nil
	}

	// Load mods button
	loadModsButton := widget.NewButton("🔄 Load", func() {
		loadMods(GetGlobalPackDir(), &modData, modList)
	})

	// Add mod section
	addModEntry := widget.NewEntry()
	addModEntry.SetPlaceHolder("mr:sodium or cf:394468")

	addModButton := widget.NewButton("➕", func() {
		if addModEntry.Text != "" {
			addMod(GetGlobalPackDir(), addModEntry.Text, &modData, modList)
			addModEntry.SetText("")
		}
	})

	addContainer := container.NewBorder(nil, nil, nil, addModButton, addModEntry)

	// Create scrollable mod list with proper sizing
	modListScroll := container.NewScroll(modList)
	modListScroll.SetMinSize(fyne.NewSize(300, 200))

	leftPane := container.NewBorder(
		container.NewVBox(loadModsButton, widget.NewSeparator()), // top
		addContainer, // bottom
		nil,          // left
		nil,          // right
		modListScroll, // center - takes remaining space
	)

	// RIGHT PANE: Actions and help

	updateButton := widget.NewButton("⬆️ Update", func() {
		if selectedMod != nil {
			updateSelectedMod(GetGlobalPackDir(), selectedModIndex, &modData, modList)
		}
	})

	removeButton := widget.NewButton("🗑️ Delete", func() {
		if selectedMod != nil {
			dialog.ShowConfirm("Delete Mod",
				fmt.Sprintf("Remove %s?", selectedMod.Name),
				func(confirmed bool) {
					if confirmed {
						removeSelectedMod(GetGlobalPackDir(), selectedModIndex, &modData, modList)
					}
				}, Window)
		}
	})

	editButton := widget.NewButton("📝 Edit", func() {
		if selectedMod != nil {
			dialog.ShowInformation("Edit Mod",
				fmt.Sprintf("Would open: %s.pw.toml", selectedMod.Filename),
				Window)
		}
	})

	settingsButton := widget.NewButton("⚙️ Settings", func() {
		if selectedMod != nil {
			dialog.ShowInformation("Mod Settings",
				fmt.Sprintf("Settings for: %s\nClient/Server/Optional toggles", selectedMod.Name),
				Window)
		}
	})

	actionsGrid := container.NewGridWithColumns(2, updateButton, removeButton)
	toolsGrid := container.NewGridWithColumns(2, editButton, settingsButton)

	// Create help text with proper wrapping
	helpText := widget.NewLabel(`Quick Help:

⬆️ Update - Updates selected mod
🗑️ Delete - Removes mod from pack
📝 Edit - Opens .pw.toml file
⚙️ Settings - Client/server toggles

Add Formats:
- mr:sodium (Modrinth)
- cf:394468 (CurseForge)
- Direct URLs supported`)
	helpText.Wrapping = fyne.TextWrapWord

	// Create scrollable help content
	helpScroll := container.NewScroll(helpText)
	helpScroll.SetMinSize(fyne.NewSize(200, 150))

	rightPane := container.NewVBox(
		widget.NewCard("🔧 Actions", "", container.NewVBox(
			actionsGrid,
			toolsGrid,
		)),
		widget.NewCard("💡 Help", "", helpScroll),
	)

	// Main dual-pane layout - give more space to left pane for mod list
	mainContent := container.NewHSplit(leftPane, rightPane)
	mainContent.SetOffset(0.7) // 70% left, 30% right

	// Register callback for pack directory changes
	RegisterPackDirCallback(func(dir string) {
		loadMods(dir, &modData, modList)
	})

	return mainContent
}

func loadMods(packDir string, modData *[]*ModDisplayInfo, modList *widget.List) {
	if packDir == "" {
		packDir = "./"
	}

	logger := NewGUILogger(GlobalLogWidget)
	manager := core.NewManager(logger)

	logger.Info("Loading mods from: %s", packDir)

	mods, err := manager.ListMods(packDir)
	if err != nil {
		logger.Error("Failed to load mods: %s", err.Error())
		*modData = []*ModDisplayInfo{}
		modList.Refresh()
		return
	}

	*modData = make([]*ModDisplayInfo, len(mods))
	for i, mod := range mods {
		(*modData)[i] = &ModDisplayInfo{
			Name:     mod.Name,
			Version:  mod.Version,
			Platform: mod.Platform,
			Filename: mod.Filename,
			ID:       mod.ID,
		}
	}

	modList.Refresh()
	logger.Info("Loaded %d mods", len(mods))
}

func addMod(packDir, modIdentifier string, modData *[]*ModDisplayInfo, modList *widget.List) {
	if packDir == "" || modIdentifier == "" {
		return
	}

	RunPwCommand("Add Mod", []string{"mod", "add", modIdentifier}, packDir)
	loadMods(packDir, modData, modList)
}

func removeSelectedMod(packDir string, selectedIndex int, modData *[]*ModDisplayInfo, modList *widget.List) {
	if selectedIndex < 0 || selectedIndex >= len(*modData) {
		return
	}

	mod := (*modData)[selectedIndex]
	RunPwCommand("Remove Mod", []string{"mod", "remove", mod.ID}, packDir)
	loadMods(packDir, modData, modList)
}

func updateSelectedMod(packDir string, selectedIndex int, modData *[]*ModDisplayInfo, modList *widget.List) {
	if selectedIndex < 0 || selectedIndex >= len(*modData) {
		return
	}

	mod := (*modData)[selectedIndex]
	RunPwCommand("Update Mod", []string{"mod", "update", mod.ID}, packDir)
	loadMods(packDir, modData, modList)
}
