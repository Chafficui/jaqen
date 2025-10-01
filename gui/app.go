package gui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	internal "jaqen/internal"
)

// JaqenGUI represents the main GUI application
type JaqenGUI struct {
	app    fyne.App
	window fyne.Window
	logger *log.Logger

	// Main content
	imgDirEntry     *widget.Entry
	xmlPathEntry    *widget.Entry
	rtfPathEntry    *widget.Entry
	fmVersionSelect *widget.Select
	statusLabel     *widget.Label
	progressBar     *widget.ProgressBar
	runButton       *widget.Button

	// Settings
	preserveCheck       *widget.Check
	allowDuplicateCheck *widget.Check
	mappingOverrideList *widget.List
	mappingOverrides    map[string]string

	// Image preview
	imagePreview1 *widget.Card
	imagePreview2 *widget.Card
	imagePreview3 *widget.Card

	// State tracking
	hasShownRTFPopup bool

	config internal.JaqenConfig
}

// New creates a new JaqenGUI instance
func New() *JaqenGUI {
	myApp := app.NewWithID("com.jaqen.footballmanager")
	myApp.SetIcon(theme.DocumentIcon())

	window := myApp.NewWindow("Jaqen NewGen Tool - Football Manager Face Manager")
	window.Resize(fyne.NewSize(1400, 900))
	window.CenterOnScreen()

	return &JaqenGUI{
		app:              myApp,
		window:           window,
		mappingOverrides: make(map[string]string),
	}
}

// ShowAndRun displays the window and runs the application
func (g *JaqenGUI) ShowAndRun() {
	// Initialize config
	g.loadConfig()

	// Auto-distribute views and filters to all FM installations on startup
	go g.autoDistributeOnStartup()

	// Create main layout
	content := container.NewVBox(
		g.createHeaderBar(),
		widget.NewSeparator(),
		g.createMainContent(),
	)

	// Apply config to GUI after widgets are created
	g.applyConfigToGUI()

	// Add some padding
	paddedContent := container.NewPadded(content)

	g.window.SetContent(paddedContent)

	// Close main window
	g.window.SetCloseIntercept(func() {
		g.window.Close()
	})

	g.window.ShowAndRun()
}

// GetWindow returns the main window
func (g *JaqenGUI) GetWindow() fyne.Window {
	return g.window
}

// GetLogger returns the logger instance
func (g *JaqenGUI) GetLogger() *log.Logger {
	return g.logger
}

// SetLogger sets the logger instance
func (g *JaqenGUI) SetLogger(logger *log.Logger) {
	g.logger = logger
}
