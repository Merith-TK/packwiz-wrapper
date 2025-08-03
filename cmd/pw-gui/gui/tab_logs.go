package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// CreateLogsTab creates the logs tab
func CreateLogsTab() fyne.CanvasObject {
	GlobalLogWidget = widget.NewRichText()
	GlobalLogWidget.ParseMarkdown("PackWiz Wrapper GUI started\n")

	// Wrap logs in a scrollable container
	logScroll := container.NewScroll(GlobalLogWidget)
	logScroll.SetMinSize(fyne.NewSize(400, 400))

	clearButton := widget.NewButton("Clear Logs", func() {
		GlobalLogWidget.ParseMarkdown("")
	})

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Logs:"),
		),
		clearButton,
		nil, nil,
		logScroll,
	)
}
