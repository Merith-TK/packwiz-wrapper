package utils

import "fmt"

// UI provides consistent console output formatting
type UI struct{}

// NewUI creates a new UI instance
func NewUI() *UI {
	return &UI{}
}

// Success prints a success message with checkmark
func (ui *UI) Success(message string) {
	fmt.Printf("✅ %s\n", message)
}

// Info prints an informational message with info icon
func (ui *UI) Info(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}

// Warning prints a warning message with warning icon
func (ui *UI) Warning(message string) {
	fmt.Printf("⚠️  %s\n", message)
}

// Error prints an error message with X icon
func (ui *UI) Error(message string) {
	fmt.Printf("❌ %s\n", message)
}

// Progress prints a progress message with arrow
func (ui *UI) Progress(message string) {
	fmt.Printf("⬇️  %s\n", message)
}

// Action prints an action message with appropriate icon
func (ui *UI) Action(icon, message string) {
	fmt.Printf("%s %s\n", icon, message)
}

// Status prints a status header with separator
func (ui *UI) Status(title string) {
	fmt.Printf("📊 %s\n", title)
	fmt.Println("================")
}

// Header prints a section header
func (ui *UI) Header(message string) {
	fmt.Printf("🚀 %s\n", message)
}
