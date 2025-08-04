package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/Merith-TK/packwiz-wrapper/internal/core"
)

// CreateModsTab creates the mod management tab
func CreateModsTab() fyne.CanvasObject {
	// Header card
	headerCard := widget.NewCard("🧩 Mod Management", 
		"Add, remove, and update mods in your pack", 
		widget.NewRichText())

	// Mod list data
	var modData []*ModDisplayInfo
	var selectedModIndex = -1

	modList := widget.NewList(
		func() int { return len(modData) },
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(modData) {
				mod := modData[id]
				label := obj.(*widget.Label)
				label.SetText(fmt.Sprintf("%s (%s) - %s", mod.Name, mod.Version, mod.Platform))
			}
		},
	)

	modList.OnSelected = func(id widget.ListItemID) {
		selectedModIndex = id
	}

	modList.OnUnselected = func(id widget.ListItemID) {
		selectedModIndex = -1
	}

	// Wrap mod list in a scrollable container with fixed height
	modListScroll := container.NewScroll(modList)
	modListScroll.SetMinSize(fyne.NewSize(400, 300))

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

	// Load mods button
	loadModsButton := widget.NewButton("🔄 Load Mods", func() {
		loadMods(GetGlobalPackDir(), &modData, modList)
	})

	// Pack directory input with Load button on the left
	packDirContainer := container.NewBorder(nil, nil, loadModsButton, nil, packDirEntry)

	// Add mod section
	addModEntry := widget.NewEntry()
	addModEntry.SetPlaceHolder("Mod URL or mr:modid:version or cf:projectid")

	addModButton := widget.NewButton("➕ Add Mod", func() {
		addMod(GetGlobalPackDir(), addModEntry.Text, &modData, modList)
		addModEntry.SetText("")
	})

	addModCommandButton := widget.NewButton("Show Add Command", func() {
		if addModEntry.Text != "" {
			ShowCommandOutput("Add Mod", "./pw.exe", []string{"mod", "add", addModEntry.Text}, GetGlobalPackDir())
		}
	})

	addModContainer := container.NewVBox(
		container.NewBorder(nil, nil, addModButton, nil, addModEntry),
		addModCommandButton,
	)

	// Mod actions
	removeModButton := widget.NewButton("🗑️ Remove Selected", func() {
		removeSelectedMod(GetGlobalPackDir(), selectedModIndex, &modData, modList)
	})
	removeModButton.Importance = widget.DangerImportance

	updateModButton := widget.NewButton("⬆️ Update Selected", func() {
		updateSelectedMod(GetGlobalPackDir(), selectedModIndex, &modData, modList)
	})

	updateAllButton := widget.NewButton("🔄 Update All", func() {
		updateAllMods(GetGlobalPackDir(), &modData, modList)
	})
	updateAllButton.Importance = widget.MediumImportance

	actionContainer := container.NewGridWithColumns(3, 
		removeModButton, 
		updateModButton, 
		updateAllButton)

	actionsCard := widget.NewCard("⚡ Mod Actions", 
		"Manage your installed mods",
		actionContainer)

	// Installed mods card
	modListCard := widget.NewCard("📦 Installed Mods", 
		"Select a mod to perform actions",
		modListScroll)

	// Layout everything
	content := container.NewVBox(
		headerCard,
		widget.NewSeparator(),
		widget.NewCard("📂 Pack Directory", "Current pack location", packDirContainer),
		widget.NewSeparator(),
		widget.NewCard("➕ Add New Mod", "Search and add mods", addModContainer),
		widget.NewSeparator(),
		modListCard,
		widget.NewSeparator(),
		actionsCard,
	)

	// Wrap in scroll container
	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(600, 400))
	
	return scroll
}

func loadMods(packDir string, modData *[]*ModDisplayInfo, modList *widget.List) {
	logger := NewGUILogger(GlobalLogWidget)
	manager := core.NewManager(logger)

	logger.Info("Loading mods from: %s", packDir)

	mods, err := manager.ListMods(packDir)
	if err != nil {
		logger.Error("Failed to load mods: %s", err.Error())
		return
	}

	// Convert to display format
	*modData = make([]*ModDisplayInfo, len(mods))
	for i, mod := range mods {
		version := mod.Version
		if version == "" {
			version = "unknown"
		}
		(*modData)[i] = &ModDisplayInfo{
			ID:       mod.ID,
			Name:     mod.Name,
			Version:  version,
			Platform: mod.Platform,
		}
	}

	modList.Refresh()
	logger.Info("Loaded %d mods", len(mods))
}

func addMod(packDir string, modRef string, modData *[]*ModDisplayInfo, modList *widget.List) {
	if modRef == "" {
		return
	}

	logger := NewGUILogger(GlobalLogWidget)
	manager := core.NewManager(logger)

	logger.Info("Adding mod: %s", modRef)

	err := manager.AddMod(packDir, modRef)
	if err != nil {
		logger.Error("Failed to add mod: %s", err.Error())
		return
	}

	logger.Info("Successfully added mod: %s", modRef)

	// Reload the mod list
	loadMods(packDir, modData, modList)
}

func removeSelectedMod(packDir string, selectedIndex int, modData *[]*ModDisplayInfo, modList *widget.List) {
	if selectedIndex < 0 || selectedIndex >= len(*modData) {
		if GlobalLogWidget != nil {
			GlobalLogWidget.ParseMarkdown(GlobalLogWidget.String() + "\n[WARN] No mod selected for removal")
		}
		return
	}

	mod := (*modData)[selectedIndex]
	logger := NewGUILogger(GlobalLogWidget)
	manager := core.NewManager(logger)

	logger.Info("Removing mod: %s", mod.Name)

	err := manager.RemoveMod(packDir, mod.ID)
	if err != nil {
		logger.Error("Failed to remove mod: %s", err.Error())
		return
	}

	logger.Info("Successfully removed mod: %s", mod.Name)

	// Reload the mod list
	loadMods(packDir, modData, modList)
}

func updateSelectedMod(packDir string, selectedIndex int, modData *[]*ModDisplayInfo, modList *widget.List) {
	if selectedIndex < 0 || selectedIndex >= len(*modData) {
		if GlobalLogWidget != nil {
			GlobalLogWidget.ParseMarkdown(GlobalLogWidget.String() + "\n[WARN] No mod selected for update")
		}
		return
	}

	mod := (*modData)[selectedIndex]
	logger := NewGUILogger(GlobalLogWidget)
	manager := core.NewManager(logger)

	logger.Info("Updating mod: %s", mod.Name)

	err := manager.UpdateMod(packDir, mod.ID)
	if err != nil {
		logger.Error("Failed to update mod: %s", err.Error())
		return
	}

	logger.Info("Successfully updated mod: %s", mod.Name)

	// Reload the mod list
	loadMods(packDir, modData, modList)
}

func updateAllMods(packDir string, modData *[]*ModDisplayInfo, modList *widget.List) {
	logger := NewGUILogger(GlobalLogWidget)
	manager := core.NewManager(logger)

	logger.Info("Updating all mods in pack")

	for _, mod := range *modData {
		logger.Info("Updating: %s", mod.Name)
		err := manager.UpdateMod(packDir, mod.ID)
		if err != nil {
			logger.Error("Failed to update %s: %s", mod.Name, err.Error())
		} else {
			logger.Info("Successfully updated: %s", mod.Name)
		}
	}

	// Reload the mod list
	loadMods(packDir, modData, modList)
	logger.Info("Finished updating all mods")
}
