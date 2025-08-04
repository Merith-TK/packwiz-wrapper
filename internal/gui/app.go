package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"github.com/Merith-TK/packwiz-wrapper/internal/core"
)

var (
	App         fyne.App
	Window      fyne.Window
	PackManager *core.Manager
)

// InitializeApp sets up and runs the GUI application
func InitializeApp(app fyne.App, packManager *core.Manager) {
	App = app
	PackManager = packManager

	// Set app icon (optional)
	App.SetIcon(nil)

	// Create the main window
	Window = App.NewWindow("PackWiz Wrapper GUI")
	Window.Resize(fyne.NewSize(800, 600))
	Window.SetFixedSize(true)

	// Initialize global shared state
	InitializeSharedState()

	// Create the main UI with tabs
	tabs := container.NewAppTabs()

	// Add all tabs - create Logs tab first to initialize GlobalLogWidget
	tabs.Append(container.NewTabItem("🏠 Welcome", CreateWelcomeTab()))
	tabs.Append(container.NewTabItem("📋 Logs", CreateLogsTab()))
	tabs.Append(container.NewTabItem("📦 Pack Info", CreatePackInfoTab()))
	tabs.Append(container.NewTabItem("🧩 Mods", CreateModsTab()))
	tabs.Append(container.NewTabItem("📥 Import/Export", CreateImportExportTab()))
	tabs.Append(container.NewTabItem("🖥️ Server", CreateServerTab()))

	// Set tab location
	tabs.SetTabLocation(container.TabLocationTop)

	// Set window content and show
	Window.SetContent(tabs)
	Window.ShowAndRun()
}
