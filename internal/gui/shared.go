//go:build gui

package gui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Global pack directory state
var globalPackDir = "./"
var packDirCallbacks []func(string)

// Global log widget for the logs tab
var GlobalLogWidget *widget.RichText

// InitializeSharedState initializes the global shared state
func InitializeSharedState() {
	globalPackDir = "./"
	packDirCallbacks = make([]func(string), 0)
}

// debouncePathUpdate creates a debounced function that waits for user to stop typing
func debouncePathUpdate(callback func(string), delay time.Duration) func(string) {
	var timer *time.Timer
	return func(text string) {
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(delay, func() {
			callback(text)
		})
	}
}

// ShowEnhancedFolderDialog shows an improved folder selection dialog
func ShowEnhancedFolderDialog(callback func(string)) {
	// For now, use the native Fyne dialog but with better styling
	folderDialog := dialog.NewFolderOpen(func(folder fyne.ListableURI, err error) {
		if err != nil {
			if GlobalLogWidget != nil {
				GlobalLogWidget.ParseMarkdown(GlobalLogWidget.String() + "\n[ERROR] Failed to select folder: " + err.Error())
			}
			callback("")
			return
		}
		if folder != nil {
			callback(folder.Path())
		} else {
			callback("")
		}
	}, Window)

	// Make the dialog larger for better usability
	folderDialog.Resize(fyne.NewSize(700, 500))
	folderDialog.Show()
}

// Logger interface for the GUI
type Logger interface {
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

// GUILogger implements a simple logger interface for GUI display
type GUILogger struct {
	logWidget *widget.RichText
}

// NewGUILogger creates a new GUI logger
func NewGUILogger(logWidget *widget.RichText) *GUILogger {
	return &GUILogger{logWidget: logWidget}
}

func (l *GUILogger) Info(msg string, args ...interface{}) {
	l.appendLog("INFO", fmt.Sprintf(msg, args...))
}

func (l *GUILogger) Warn(msg string, args ...interface{}) {
	l.appendLog("WARN", fmt.Sprintf(msg, args...))
}

func (l *GUILogger) Error(msg string, args ...interface{}) {
	l.appendLog("ERROR", fmt.Sprintf(msg, args...))
}

func (l *GUILogger) Debug(msg string, args ...interface{}) {
	l.appendLog("DEBUG", fmt.Sprintf(msg, args...))
}

func (l *GUILogger) appendLog(level, msg string) {
	entry := fmt.Sprintf("[%s] %s\n", level, msg)

	// Log to the logger's own widget if available
	if l.logWidget != nil {
		fyne.Do(func() {
			currentText := l.logWidget.String()
			l.logWidget.ParseMarkdown(currentText + entry)
		})
	}

	// Also log to global widget if available
	if GlobalLogWidget != nil {
		fyne.Do(func() {
			currentText := GlobalLogWidget.String()
			GlobalLogWidget.ParseMarkdown(currentText + entry)
		})
	}
}

// Pack directory management functions
func SetGlobalPackDir(dir string) {
	if dir == "" {
		dir = "./"
	}
	globalPackDir = dir
	// Notify all registered callbacks on the main thread
	fyne.Do(func() {
		for _, callback := range packDirCallbacks {
			callback(dir)
		}
	})
}

func RegisterPackDirCallback(callback func(string)) {
	packDirCallbacks = append(packDirCallbacks, callback)
}

func GetGlobalPackDir() string {
	return globalPackDir
}

// ModDisplayInfo represents mod information for display in the GUI
type ModDisplayInfo struct {
	ID       string
	Name     string
	Version  string
	Platform string
	Filename string
}
