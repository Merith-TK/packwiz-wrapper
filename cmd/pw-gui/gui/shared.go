package gui

import (
	"fmt"

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
	currentText := l.logWidget.String()
	l.logWidget.ParseMarkdown(currentText + entry)
}

// Pack directory management functions
func SetGlobalPackDir(dir string) {
	if dir == "" {
		dir = "./"
	}
	globalPackDir = dir
	// Notify all registered callbacks
	for _, callback := range packDirCallbacks {
		callback(dir)
	}
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
}
