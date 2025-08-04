package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// CreateLogsTab creates the logs display tab
func CreateLogsTab() fyne.CanvasObject {
	// Header card
	headerCard := widget.NewCard("📋 Application Logs", 
		"Monitor PackWrap2 operations and debug information", 
		widget.NewRichText())

	// Initialize global log widget
	GlobalLogWidget = widget.NewRichText()
	GlobalLogWidget.ParseMarkdown("📋 **PackWrap2 GUI Started**\n\nWelcome! Application logs will appear here as you use the GUI.\n\n")

	// Wrap logs in a scrollable container
	logScroll := container.NewScroll(GlobalLogWidget)
	logScroll.SetMinSize(fyne.NewSize(500, 350))

	// Log controls
	clearButton := widget.NewButton("🗑️ Clear Logs", func() {
		GlobalLogWidget.ParseMarkdown("📋 **Logs Cleared**\n\n")
	})
	clearButton.Importance = widget.DangerImportance

	exportButton := widget.NewButton("💾 Export Logs", func() {
		// This would save logs to a file
		dialog.ShowInformation("Export Logs", "Log export functionality will be added in a future update", Window)
	})

	controlActions := container.NewHBox(clearButton, exportButton)

	controlCard := widget.NewCard("🔧 Log Controls", 
		"Manage log display and export",
		controlActions)

	// Main log display
	logDisplayCard := widget.NewCard("📄 Log Output", 
		"Real-time application logs and command output",
		logScroll)

	// Layout everything
	content := container.NewVBox(
		headerCard,
		widget.NewSeparator(),
		controlCard,
		widget.NewSeparator(),
		logDisplayCard,
	)

	// Wrap in scroll container
	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(600, 400))
	
	return scroll
}
