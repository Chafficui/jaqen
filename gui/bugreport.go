package gui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	internal "jaqen/internal"
)

const (
	githubRepoURL  = "https://github.com/chafficui/jaqen-newgen-tool"
	maxLogLines    = 100
	maxConfigLines = 500
)

// openBrowser opens the specified URL in the default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// readFileWithLimit reads a file and returns the last N lines
func readFileWithLimit(path string, maxLines int) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")

	// If file is within limit, return all
	if len(lines) <= maxLines {
		return string(data), nil
	}

	// Otherwise, return last N lines with a note
	truncatedLines := lines[len(lines)-maxLines:]
	header := fmt.Sprintf("[... %d earlier lines truncated ...]\n\n", len(lines)-maxLines)
	return header + strings.Join(truncatedLines, "\n"), nil
}

// getSystemInfo returns basic system information
func getSystemInfo() string {
	return fmt.Sprintf("OS: %s\nArchitecture: %s\nGo Version: %s",
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version(),
	)
}

// collectBugReportData gathers all relevant information for a bug report
func (g *JaqenGUI) collectBugReportData() string {
	var report strings.Builder

	report.WriteString("## Bug Report\n\n")
	report.WriteString("### System Information\n```\n")
	report.WriteString(getSystemInfo())
	report.WriteString("\n```\n\n")

	// Show current profile and all available profiles
	report.WriteString("### Profile Information\n")
	if g.currentProfile != nil {
		report.WriteString(fmt.Sprintf("**Current Profile:** %s\n\n", g.currentProfile.Name))
	}

	// List all available profiles
	if g.profileManager != nil {
		profiles := g.profileManager.ListProfiles()
		report.WriteString(fmt.Sprintf("**Available Profiles:** %s\n\n", strings.Join(profiles, ", ")))

		// Show current profile configuration
		if g.currentProfile != nil {
			report.WriteString("**Current Profile Configuration:**\n```json\n")

			// Format current profile settings
			report.WriteString(fmt.Sprintf("Profile Name: %s\n", g.currentProfile.Name))
			report.WriteString(fmt.Sprintf("Created: %s\n", g.currentProfile.CreatedAt.Format("2006-01-02 15:04:05")))
			report.WriteString(fmt.Sprintf("Updated: %s\n\n", g.currentProfile.UpdatedAt.Format("2006-01-02 15:04:05")))

			cfg := g.currentProfile.Config
			if cfg.Preserve != nil {
				report.WriteString(fmt.Sprintf("Preserve: %v\n", *cfg.Preserve))
			}
			if cfg.AllowDuplicate != nil {
				report.WriteString(fmt.Sprintf("Allow Duplicate: %v\n", *cfg.AllowDuplicate))
			}
			if cfg.FMVersion != nil {
				report.WriteString(fmt.Sprintf("FM Version: %s\n", *cfg.FMVersion))
			}
			if cfg.XMLPath != nil {
				report.WriteString(fmt.Sprintf("XML Path: %s\n", *cfg.XMLPath))
			}
			if cfg.RTFPath != nil {
				report.WriteString(fmt.Sprintf("RTF Path: %s\n", *cfg.RTFPath))
			}
			if cfg.IMGPath != nil {
				report.WriteString(fmt.Sprintf("Image Path: %s\n", *cfg.IMGPath))
			}
			if cfg.MappingOverride != nil && len(*cfg.MappingOverride) > 0 {
				report.WriteString("\nMapping Overrides:\n")
				for country, ethnic := range *cfg.MappingOverride {
					report.WriteString(fmt.Sprintf("  %s → %s\n", country, ethnic))
				}
			}

			report.WriteString("```\n\n")
		}
	} else {
		report.WriteString("*Profile system not initialized*\n\n")
	}

	// Try to read log file
	report.WriteString("### Log File\n")

	// Try user log directory first
	logContent := ""
	if userLogPath, err := internal.GetUserLogPath(); err == nil {
		if content, err := readFileWithLimit(userLogPath, maxLogLines); err == nil {
			logContent = content
		}
	}

	// Try current directory if user log not found
	if logContent == "" {
		if g.imgDirEntry != nil && g.imgDirEntry.Text != "" {
			logPath := filepath.Join(g.imgDirEntry.Text, "jaqen.log")
			if content, err := readFileWithLimit(logPath, maxLogLines); err == nil {
				logContent = content
			}
		}
	}

	if logContent != "" {
		report.WriteString("```\n")
		report.WriteString(logContent)
		report.WriteString("\n```\n\n")
	} else {
		report.WriteString("*Log file not found or empty*\n\n")
	}

	report.WriteString("### Description\n")
	report.WriteString("**Please describe what happened and what you expected to happen:**\n\n")
	report.WriteString("[Your description here]\n\n")

	report.WriteString("### Steps to Reproduce\n")
	report.WriteString("1. \n")
	report.WriteString("2. \n")
	report.WriteString("3. \n")

	return report.String()
}

// showBugReportDialog shows a dialog with bug report information
func (g *JaqenGUI) showBugReportDialog() {
	// Show a loading message while collecting data
	progressDialog := dialog.NewCustom("Collecting Bug Report Data", "Cancel",
		container.NewVBox(
			widget.NewLabel("Collecting system information, logs, and configuration..."),
			widget.NewProgressBarInfinite(),
		), g.window)

	progressDialog.Show()

	// Collect data in background
	go func() {
		reportBody := g.collectBugReportData()

		// Update UI on main thread
		fyne.Do(func() {
			// Close progress dialog
			progressDialog.Hide()

			// Create the report display
			reportText := widget.NewMultiLineEntry()
			reportText.SetText(reportBody)
			reportText.Wrapping = fyne.TextWrapWord
			reportText.SetMinRowsVisible(20)

			// Create buttons
			openBrowserBtn := widget.NewButton("Open GitHub Issue", func() {
				// URL encode the data (truncate if too long for URL)
				issueTitle := "Bug Report: [Please add a brief description]"

				// Try to open browser with pre-filled data
				// Note: URLs have length limits, so we provide a simpler version
				issueURL := fmt.Sprintf("%s/issues/new?title=%s",
					githubRepoURL,
					strings.ReplaceAll(issueTitle, " ", "+"),
				)

				if err := openBrowser(issueURL); err != nil {
					dialog.ShowError(fmt.Errorf("could not open browser: %w", err), g.window)
				} else {
					dialog.ShowInformation("Browser Opened",
						"Please paste the bug report information from the text field below into the GitHub issue.",
						g.window)
				}
			})
			openBrowserBtn.Importance = widget.HighImportance

			copyBtn := widget.NewButton("Copy to Clipboard", func() {
				g.window.Clipboard().SetContent(reportBody)
				dialog.ShowInformation("Copied", "Bug report copied to clipboard!", g.window)
			})

			// Create dialog content
			content := container.NewBorder(
				container.NewVBox(
					widget.NewLabel("Bug Report Information"),
					widget.NewLabel("Copy this information and paste it into a new GitHub issue:"),
					widget.NewSeparator(),
				),
				container.NewHBox(
					openBrowserBtn,
					copyBtn,
				),
				nil, nil,
				container.NewScroll(reportText),
			)

			// Show the bug report dialog
			bugDialog := dialog.NewCustom("Bug Report", "Close", content, g.window)
			bugDialog.Resize(fyne.NewSize(800, 600))
			bugDialog.Show()
		})
	}()
}
