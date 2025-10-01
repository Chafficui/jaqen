package gui

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"

	internal "jaqen/internal"
)

// guiLogWriter is a custom writer that writes to both file and GUI
type guiLogWriter struct {
	fileWriter io.Writer
	gui        *JaqenGUI
}

func (w *guiLogWriter) Write(p []byte) (n int, err error) {
	// Write to file
	n, err = w.fileWriter.Write(p)

	// Also append to GUI on main thread
	if w.gui != nil && w.gui.logLabel != nil {
		msg := string(p)
		fyne.Do(func() {
			currentText := w.gui.logLabel.Text
			lines := strings.Split(currentText, "\n")

			// Keep only last 100 lines to prevent GUI from getting too slow
			if len(lines) > 100 {
				lines = lines[len(lines)-100:]
				currentText = strings.Join(lines, "\n")
			}

			// Remove placeholder text on first log
			if currentText == "System logs will appear here..." {
				currentText = ""
			}

			w.gui.logLabel.SetText(currentText + msg)
		})
	}

	return n, err
}

// setupLogger creates a logger that writes to jaqen.log in the user config directory
func (g *JaqenGUI) setupLogger(imageDir string) error {
	logPath, err := internal.GetUserLogPath()
	if err != nil {
		// Fallback to current directory if user directory is not accessible
		logPath = filepath.Join(imageDir, "jaqen.log")
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to create log file: %v", err)
	}

	// Close previous logger if it exists
	if g.logger != nil {
		// Note: We can't close the previous log file easily, but the new one will be used going forward
	}

	// Create custom writer that writes to both file and GUI
	writer := &guiLogWriter{
		fileWriter: logFile,
		gui:        g,
	}

	g.logger = log.New(writer, "", log.LstdFlags)
	g.logger.Printf("Jaqen NewGen Tool started - %s", time.Now().Format("2006-01-02 15:04:05"))
	return nil
}

// setupStartupLogger creates a logger for startup operations in the user config directory
func (g *JaqenGUI) setupStartupLogger() {
	if g.logger != nil {
		return // Already set up
	}

	logPath, err := internal.GetUserLogPath()
	if err != nil {
		// If we can't access user directory, just use GUI-only logger
		writer := &guiLogWriter{
			fileWriter: io.Discard,
			gui:        g,
		}
		g.logger = log.New(writer, "", log.LstdFlags)
		return
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// If we can't create log file, just use GUI-only logger
		writer := &guiLogWriter{
			fileWriter: io.Discard,
			gui:        g,
		}
		g.logger = log.New(writer, "", log.LstdFlags)
		return
	}

	// Create custom writer that writes to both file and GUI
	writer := &guiLogWriter{
		fileWriter: logFile,
		gui:        g,
	}

	g.logger = log.New(writer, "", log.LstdFlags)
	g.logger.Printf("Jaqen NewGen Tool startup - %s", time.Now().Format("2006-01-02 15:04:05"))
}
