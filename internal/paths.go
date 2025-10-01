package internal

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetUserConfigDir returns the appropriate user configuration directory for the current platform
// Returns ~/.jaqen on Unix-like systems (Linux, macOS) and %APPDATA%\jaqen on Windows
func GetUserConfigDir() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\jaqen
		appData := os.Getenv("APPDATA")
		if appData == "" {
			// Fallback to user profile if APPDATA is not set
			userProfile := os.Getenv("USERPROFILE")
			if userProfile == "" {
				return "", os.ErrNotExist
			}
			configDir = filepath.Join(userProfile, "AppData", "Roaming", "jaqen")
		} else {
			configDir = filepath.Join(appData, "jaqen")
		}
	case "darwin":
		// macOS: ~/Library/Application Support/jaqen
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, "Library", "Application Support", "jaqen")
	default:
		// Linux and other Unix-like systems: ~/.jaqen
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, ".jaqen")
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return configDir, nil
}

// GetUserConfigPath returns the full path to the user's jaqen.toml config file
func GetUserConfigPath() (string, error) {
	configDir, err := GetUserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "jaqen.toml"), nil
}

// GetUserLogPath returns the full path to the user's jaqen.log file
func GetUserLogPath() (string, error) {
	configDir, err := GetUserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "jaqen.log"), nil
}
