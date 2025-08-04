//go:build gui

package gui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowCommandOutput displays a terminal output dialog that shows command execution results
func ShowCommandOutput(title, command string, args []string, workingDir string) {
	// Create a dialog to show command execution
	outputText := widget.NewRichText()
	outputText.ParseMarkdown(fmt.Sprintf("**Executing:** %s %s\n\n**Working Directory:** %s\n\n---\n\n", command, strings.Join(args, " "), workingDir))

	// Create scrollable container for output
	outputScroll := container.NewScroll(outputText)
	outputScroll.SetMinSize(fyne.NewSize(600, 400))

	// Create close button
	closeButton := widget.NewButton("Close", nil)

	// Create the dialog
	outputDialog := dialog.NewCustom(title, "Close", outputScroll, Window)
	outputDialog.Resize(fyne.NewSize(700, 500))

	// Update close button to close dialog
	closeButton.OnTapped = func() {
		outputDialog.Hide()
	}

	// Show the dialog immediately
	outputDialog.Show()

	// Run command in background and update output
	go func() {
		cmd := exec.Command(command, args...)
		if workingDir != "" && workingDir != "./" {
			cmd.Dir = workingDir
		}

		// Get combined output (stdout + stderr)
		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Update the text widget with results
		var resultText strings.Builder
		resultText.WriteString(fmt.Sprintf("**Executing:** %s %s\n\n", command, strings.Join(args, " ")))
		resultText.WriteString(fmt.Sprintf("**Working Directory:** %s\n\n", workingDir))
		resultText.WriteString("---\n\n")

		if err != nil {
			resultText.WriteString(fmt.Sprintf("**Exit Code:** %v\n\n", err))
		} else {
			resultText.WriteString("**Exit Code:** 0 (Success)\n\n")
		}

		resultText.WriteString("**Output:**\n```\n")
		if outputStr != "" {
			resultText.WriteString(outputStr)
		} else {
			resultText.WriteString("(No output)")
		}
		resultText.WriteString("\n```")

		// Update UI on main thread
		outputText.ParseMarkdown(resultText.String())

		// Auto-scroll to bottom
		outputScroll.ScrollToBottom()
	}()
}

// RunPwCommand is a wrapper function that calls the current executable as CLI
func RunPwCommand(title string, args []string, packDir string) {
	// Get the path to the current executable
	execPath, err := os.Executable()
	if err != nil {
		// Fallback to a generic error display
		ShowCommandOutput(title, "error", []string{"Failed to get executable path"}, packDir)
		return
	}
	ShowCommandOutput(title, execPath, args, packDir)
}
