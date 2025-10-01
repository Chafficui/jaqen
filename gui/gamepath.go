package gui

import (
	"path/filepath"
	"strings"

	mapper "jaqen/pkgs"
)

// updateGamePathFromImagePath updates the current profile's game path based on the selected image path
func (g *JaqenGUI) updateGamePathFromImagePath(imgPath string) {
	if g.currentProfile == nil || g.profileManager == nil {
		return
	}

	// Try to find the FM directory from the image path
	fmDir, err := mapper.FindFMDirectoryFromImagePath(imgPath)
	if err != nil {
		// Can't find FM directory, user might be using custom setup
		return
	}

	// Check if this is a different game path than the current profile
	if fmDir.BasePath != g.currentProfile.GamePath {
		oldGamePath := g.currentProfile.GamePath
		g.currentProfile.GamePath = fmDir.BasePath

		// Save the updated profile
		if err := g.profileManager.SaveProfile(g.currentProfile); err != nil {
			if g.logger != nil {
				g.logger.Printf("Warning: Failed to update game path: %v", err)
			}
			return
		}

		if g.logger != nil {
			if oldGamePath == "" {
				g.logger.Printf("Set game path for profile '%s': %s", g.currentProfile.Name, fmDir.BasePath)
			} else {
				g.logger.Printf("Updated game path for profile '%s': %s → %s",
					g.currentProfile.Name, oldGamePath, fmDir.BasePath)
			}
		}

		// Auto-detect version from new game path
		version := g.extractFMVersion(fmDir.BasePath)
		if version != "" && g.fmVersionSelect != nil {
			g.fmVersionSelect.SetSelected(version)
		}
	}
}

// getGraphicsPathForCurrentProfile returns the graphics path for the current profile's game
func (g *JaqenGUI) getGraphicsPathForCurrentProfile() string {
	if g.currentProfile == nil || g.currentProfile.GamePath == "" {
		return ""
	}

	return filepath.Join(g.currentProfile.GamePath, "graphics")
}

// isPathInGame checks if a path is within a specific game installation
func isPathInGame(path string, gamePath string) bool {
	if gamePath == "" {
		return false
	}

	// Normalize paths
	absPath, _ := filepath.Abs(path)
	absGamePath, _ := filepath.Abs(gamePath)

	return strings.HasPrefix(absPath, absGamePath)
}
