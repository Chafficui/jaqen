package gui

import (
	"os/exec"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// showRTFInstructionsPopup displays complete Jaqen NewGen instructions
func (g *JaqenGUI) showRTFInstructionsPopup() {
	// Create a custom dialog with video link and instructions
	videoLabel := widget.NewLabel("📺 Watch the Tutorial Video:")
	videoLabel.TextStyle.Bold = true

	videoButton := widget.NewButton("🎥 Jaqen NewGen Tutorial", func() {
		// Open YouTube link in default browser
		cmd := exec.Command("xdg-open", "https://youtu.be/aHnrpfH--ic")
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/c", "start", "https://youtu.be/aHnrpfH--ic")
		} else if runtime.GOOS == "darwin" {
			cmd = exec.Command("open", "https://youtu.be/aHnrpfH--ic")
		}
		cmd.Run()
	})
	videoButton.Importance = widget.HighImportance

	instructions := `# Quick Setup Guide

## 1. Export RTF from Football Manager
- Go to Scouting → Players in Range
- Import "SCRIPT FACES player search" view
- Apply "is newgen search filter"
- Select all (Ctrl+A) → Print to text file (Ctrl+P)
- Save as "newgen.rtf" in your image folder

## 2. Configure Jaqen
- Select your image directory
- Choose settings (Preserve, Allow Duplicates, etc.)
- Click "Assign Face Mappings"

## 3. Apply in Football Manager
- Restart Football Manager
- Newgen faces will use assigned images

**Note:** Views, filters, and config.xml are auto-distributed/generated.`

	content := widget.NewRichTextFromMarkdown(instructions)
	content.Wrapping = fyne.TextWrapWord

	// Combine video button and instructions
	videoContainer := container.NewVBox(videoLabel, videoButton)
	instructionsContainer := container.NewScroll(content)
	instructionsContainer.SetMinSize(fyne.NewSize(600, 300))

	mainContainer := container.NewVBox(
		videoContainer,
		widget.NewSeparator(),
		instructionsContainer,
	)

	dialog := dialog.NewCustom("Jaqen NewGen Tool - Instructions", "Close", mainContainer, g.window)
	dialog.Resize(fyne.NewSize(650, 500))
	dialog.Show()
}
