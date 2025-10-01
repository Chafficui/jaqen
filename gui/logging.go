package gui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	internal "jaqen/internal"
)

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

	g.logger = log.New(logFile, "", log.LstdFlags)
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
		// If we can't access user directory, just use stdout
		g.logger = log.New(os.Stdout, "", log.LstdFlags)
		return
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// If we can't create log file, just use stdout
		g.logger = log.New(os.Stdout, "", log.LstdFlags)
		return
	}

	g.logger = log.New(logFile, "", log.LstdFlags)
	g.logger.Printf("Jaqen NewGen Tool startup - %s", time.Now().Format("2006-01-02 15:04:05"))
}
