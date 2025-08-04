package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// CreateLogsTab creates the logs display tab
func CreateLogsTab() fyne.CanvasObject {
	// Initialize global log widget
	GlobalLogWidget = widget.NewRichText()
	GlobalLogWidget.ParseMarkdown("📋 **PackWrap2 GUI Started**\n\nWelcome! Application logs will appear here as you use the GUI.\n\n")
	GlobalLogWidget.Wrapping = fyne.TextWrapWord

	// Wrap logs in a scrollable container with proper sizing
	logScroll := container.NewScroll(GlobalLogWidget)
	logScroll.SetMinSize(fyne.NewSize(400, 300))

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

	// Use BorderContainer to give more space to the log area
	content := container.NewBorder(
		controlActions, // top
		nil,           // bottom
		nil,           // left
		nil,           // right
		logScroll,     // center - takes remaining space
	)

	return content
}
