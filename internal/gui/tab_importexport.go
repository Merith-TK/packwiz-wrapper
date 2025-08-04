package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/Merith-TK/packwiz-wrapper/internal/core"
	"github.com/Merith-TK/packwiz-wrapper/pkg/packwrap"
)

// CreateImportExportTab creates the import/export tab
func CreateImportExportTab() fyne.CanvasObject {
	// Pack directory input that syncs with global state
	packDirEntry := widget.NewEntry()
	packDirEntry.SetPlaceHolder("Pack directory (synced globally)")
	packDirEntry.SetText(GetGlobalPackDir())

	// Register callback to update entry when global pack dir changes
	RegisterPackDirCallback(func(dir string) {
		packDirEntry.SetText(dir)
	})

	packDirEntry.OnChanged = func(text string) {
		SetGlobalPackDir(text)
	}

	// Import section
	importFileEntry := widget.NewEntry()
	importFileEntry.SetPlaceHolder("import.txt")

	importBrowseButton := widget.NewButton("Browse", func() {
		// Create a file dialog for import files
		fileDialog := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
			if err != nil {
				if GlobalLogWidget != nil {
					GlobalLogWidget.ParseMarkdown(GlobalLogWidget.String() + "\n[ERROR] Failed to select file: " + err.Error())
				}
				return
			}
			if file != nil {
				importFileEntry.SetText(file.URI().Path())
				file.Close() // Don't forget to close the file
			}
		}, fyne.CurrentApp().Driver().AllWindows()[0])

		// Set file filter for common import files
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt", ".md"}))
		fileDialog.Show()
	})

	importButton := widget.NewButton("Import from File", func() {
		importFromFile(GetGlobalPackDir(), importFileEntry.Text)
	})

	importCommandButton := widget.NewButton("Show Import Command", func() {
		if importFileEntry.Text != "" {
			RunPwCommand("Import from File", []string{"import", importFileEntry.Text}, GetGlobalPackDir())
		}
	})

	importContainer := container.NewVBox(
		widget.NewLabel("Import Mods:"),
		container.NewBorder(nil, nil, importBrowseButton, nil, importFileEntry),
		container.NewHBox(importButton, importCommandButton),
	)

	// Export section
	exportFormatSelect := widget.NewSelect(
		[]string{"CurseForge", "Modrinth", "MultiMC", "Technic", "Server", "All"},
		func(selected string) {
			// Selection callback (optional)
		},
	)
	exportFormatSelect.SetSelected("CurseForge")

	exportButton := widget.NewButton("Export Pack", func() {
		exportPack(GetGlobalPackDir(), exportFormatSelect.Selected)
	})

	exportContainer := container.NewVBox(
		widget.NewLabel("Export Pack:"),
		exportFormatSelect,
		exportButton,
	)

	// Layout everything - compact without header
	content := container.NewVBox(
		widget.NewCard("📂 Pack Directory", "Current pack location", packDirEntry),

		// Two-column layout for import/export
		container.NewGridWithColumns(2,
			widget.NewCard("📥 Import Mods", "Import from files", importContainer),
			widget.NewCard("📤 Export Pack", "Export for sharing", exportContainer),
		),
	)

	// Wrap in scroll container
	scroll := container.NewScroll(content)

	return scroll
}

func importFromFile(packDir string, filename string) {
	if filename == "" {
		if GlobalLogWidget != nil {
			GlobalLogWidget.ParseMarkdown(GlobalLogWidget.String() + "\n[WARN] No import file specified")
		}
		return
	}

	logger := NewGUILogger(GlobalLogWidget)
	manager := core.NewManager(logger)

	logger.Info("Importing mods from file: %s", filename)

	err := manager.ImportFromFile(packDir, filename)
	if err != nil {
		logger.Error("Failed to import from file: %s", err.Error())
		return
	}

	logger.Info("Successfully imported mods from: %s", filename)
}

func exportPack(packDir string, format string) {
	logger := NewGUILogger(GlobalLogWidget)
	manager := core.NewManager(logger)

	logger.Info("Exporting pack in %s format", format)

	var exportFormat packwrap.ExportFormat
	switch format {
	case "CurseForge":
		exportFormat = packwrap.ExportCurseForge
	case "Modrinth":
		exportFormat = packwrap.ExportModrinth
	case "MultiMC":
		exportFormat = packwrap.ExportMultiMC
	case "Technic":
		exportFormat = packwrap.ExportTechnic
	case "Server":
		exportFormat = packwrap.ExportServer
	case "All":
		exportFormat = packwrap.ExportAll
	default:
		logger.Error("Unknown export format: %s", format)
		return
	}

	if exportFormat == packwrap.ExportAll {
		logger.Info("Exporting all formats...")
		formats := []packwrap.ExportFormat{packwrap.ExportCurseForge, packwrap.ExportModrinth}
		for _, fmt := range formats {
			logger.Info("Exporting %s format...", fmt)
			_, err := manager.ExportPack(packDir, fmt)
			if err != nil {
				logger.Error("Failed to export %s: %s", fmt, err.Error())
			} else {
				logger.Info("Successfully exported %s format", fmt)
			}
		}
	} else {
		_, err := manager.ExportPack(packDir, exportFormat)
		if err != nil {
			logger.Error("Failed to export pack: %s", err.Error())
			return
		}
		logger.Info("Successfully exported pack in %s format", format)
	}
}
