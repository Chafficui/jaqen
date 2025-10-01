package gui

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	mapper "jaqen/pkgs"
)

// detectPathsFromImageFolder automatically detects paths based on the selected image folder
func (g *JaqenGUI) detectPathsFromImageFolder(imgPath string) {
	if imgPath == "" {
		return
	}

	// Setup logging to jaqen.log in the image directory
	if err := g.setupLogger(imgPath); err != nil {
		// Log error but don't fail the operation
		if g.logger != nil {
			g.logger.Printf("Warning: Failed to setup logging: %v", err)
		}
	}

	// Extract FM version from path
	fmVersion := g.extractFMVersion(imgPath)
	if fmVersion != "" {
		g.fmVersionSelect.SetSelected(fmVersion)
	}

	// Set config.xml path (always in image folder)
	configPath := filepath.Join(imgPath, "config.xml")
	if _, err := os.Stat(configPath); err == nil {
		g.xmlPathEntry.SetText(configPath)
	} else {
		// Only generate config.xml if this is a valid FM graphics directory
		if mapper.IsValidFMGraphicsDirectory(imgPath) {
			g.autoGenerateConfigXML(imgPath)
			g.xmlPathEntry.SetText(configPath)
		} else {
			// Not a valid FM directory, just set the path without generating
			g.xmlPathEntry.SetText(configPath)
		}
	}

	// Try to find RTF file in common locations
	rtfPath := g.findRTFFile(imgPath)
	if rtfPath != "" {
		g.rtfPathEntry.SetText(rtfPath)
	} else {
		// RTF file not found, show instructions popup only once per session
		if !g.hasShownRTFPopup {
			g.showRTFInstructionsPopup()
			g.hasShownRTFPopup = true
		}
	}
}

// extractFMVersion extracts the FM version from a path
func (g *JaqenGUI) extractFMVersion(imgPath string) string {
	// Look for patterns like "Football Manager 2022", "FM2022", "2022", etc.
	pathParts := strings.Split(imgPath, string(os.PathSeparator))
	for _, part := range pathParts {
		if strings.Contains(strings.ToLower(part), "football manager") || strings.Contains(strings.ToLower(part), "fm") {
			// Extract year from the part
			re := regexp.MustCompile(`\b(20\d{2})\b`)
			matches := re.FindStringSubmatch(part)
			if len(matches) > 0 {
				return matches[0]
			}
		}
	}
	return ""
}

// findRTFFile searches for RTF files in common locations
func (g *JaqenGUI) findRTFFile(imgPath string) string {
	// Common locations to search for RTF files
	searchPaths := []string{
		imgPath,               // Same directory as images
		filepath.Dir(imgPath), // Parent directory
		filepath.Join(filepath.Dir(imgPath), ".."), // Grandparent directory
	}

	for _, searchPath := range searchPaths {
		// Look for common RTF file names
		rtfFiles := []string{"newgen.rtf", "players.rtf", "regens.rtf"}
		for _, rtfFile := range rtfFiles {
			fullPath := filepath.Join(searchPath, rtfFile)
			if _, err := os.Stat(fullPath); err == nil {
				return fullPath
			}
		}
	}
	return ""
}

// autoDistributeOnStartup automatically distributes views and filters to all FM installations
func (g *JaqenGUI) autoDistributeOnStartup() {
	// Find all FM installations and distribute views/filters
	fmDirs := g.findAllFMInstallations()

	// Log to a temporary log file in the current directory for startup operations
	g.setupStartupLogger()

	if g.logger != nil {
		g.logger.Printf("Starting auto-distribution of views/filters to %d FM installation(s)", len(fmDirs))
	}

	for _, fmDir := range fmDirs {
		if err := mapper.DistributeViewsAndFilters(fmDir); err != nil {
			// Log error
			if g.logger != nil {
				g.logger.Printf("Error distributing to %s: %v", fmDir.BasePath, err)
			}
		} else if g.logger != nil {
			g.logger.Printf("Successfully distributed views/filters to: %s", fmDir.BasePath)
		}
	}

	if len(fmDirs) > 0 {
		if g.logger != nil {
			g.logger.Printf("Auto-distribution completed for %d FM installation(s)", len(fmDirs))
		}
	}
}

// findAllFMInstallations searches for all FM installations on the system
func (g *JaqenGUI) findAllFMInstallations() []*mapper.FMDirectory {
	var fmDirs []*mapper.FMDirectory
	seenPaths := make(map[string]bool) // Prevent duplicates

	// Common FM installation paths to check
	searchPaths := []string{
		// Linux Steam paths
		filepath.Join(os.Getenv("HOME"), ".steam/debian-installation/steamapps/compatdata"),
		filepath.Join(os.Getenv("HOME"), ".local/share/Steam/steamapps/compatdata"),
		// Linux Heroic/Epic paths
		filepath.Join(os.Getenv("HOME"), "Games/Heroic/Prefixes/default"),
		filepath.Join(os.Getenv("HOME"), ".local/share/Steam/steamapps/common"),
		// Windows paths (if running on Windows)
		filepath.Join(os.Getenv("USERPROFILE"), "Documents", "Sports Interactive"),
		// macOS paths (if running on macOS)
		filepath.Join(os.Getenv("HOME"), "Documents", "Sports Interactive"),
	}

	for _, basePath := range searchPaths {
		if _, err := os.Stat(basePath); err == nil {
			// Search for FM directories in this path
			g.searchFMInPath(basePath, &fmDirs, seenPaths)
		}
	}

	return fmDirs
}

// searchFMInPath recursively searches for FM installations in a given path
func (g *JaqenGUI) searchFMInPath(searchPath string, fmDirs *[]*mapper.FMDirectory, seenPaths map[string]bool) {
	entries, err := os.ReadDir(searchPath)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			entryPath := filepath.Join(searchPath, entry.Name())

			// Check if this looks like an FM installation
			if g.isLikelyFMInstallation(entryPath) {
				// Try to find the graphics directory
				if fmDir, err := mapper.FindFMDirectoryFromImagePath(entryPath); err == nil {
					// Check if we've already seen this path to prevent duplicates
					if !seenPaths[fmDir.BasePath] {
						*fmDirs = append(*fmDirs, fmDir)
						seenPaths[fmDir.BasePath] = true
					}
				}
			}

			// Recursively search subdirectories (but limit depth to prevent infinite recursion)
			if g.shouldContinueSearch(entryPath, searchPath) {
				g.searchFMInPath(entryPath, fmDirs, seenPaths)
			}
		}
	}
}

// isLikelyFMInstallation checks if a path looks like an FM installation
func (g *JaqenGUI) isLikelyFMInstallation(path string) bool {
	pathLower := strings.ToLower(path)
	return strings.Contains(pathLower, "football manager") ||
		strings.Contains(pathLower, "sports interactive")
}

// shouldContinueSearch determines if we should continue searching in a subdirectory
func (g *JaqenGUI) shouldContinueSearch(entryPath, searchPath string) bool {
	// Don't go too deep (max 4 levels from search path)
	relPath, err := filepath.Rel(searchPath, entryPath)
	if err != nil {
		return false
	}

	depth := strings.Count(relPath, string(os.PathSeparator))
	return depth < 4
}

// autoDistributeViewsAndFilters automatically distributes views and filters to the FM directory
func (g *JaqenGUI) autoDistributeViewsAndFilters(imgPath string) {
	go func() {
		// Find the FM directory from the image path
		fmDir, err := mapper.FindFMDirectoryFromImagePath(imgPath)
		if err != nil {
			// FM directory not found, that's okay - user might be using a custom setup
			return
		}

		// Distribute views and filters
		err = mapper.DistributeViewsAndFilters(fmDir)
		if err != nil {
			// Log error
			if g.logger != nil {
				g.logger.Printf("Warning: Failed to distribute views/filters: %v", err)
			}
			return
		}

		// Log success
		if g.logger != nil {
			g.logger.Printf("Views and filters distributed to FM directory: %s", fmDir.BasePath)
		}
	}()
}

// autoGenerateConfigXML automatically generates a config.xml file if it doesn't exist
func (g *JaqenGUI) autoGenerateConfigXML(imgPath string) {
	go func() {
		if g.logger != nil {
			g.logger.Printf("Auto-generating config.xml in: %s", imgPath)
		}
		err := mapper.GenerateConfigXML(imgPath)
		if err != nil {
			// Log error
			if g.logger != nil {
				g.logger.Printf("Error generating config.xml: %v", err)
			}
			return
		}
		if g.logger != nil {
			g.logger.Printf("Successfully generated config.xml")
		}

		// Already logged above
	}()
}
