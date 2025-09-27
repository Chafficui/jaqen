package mapper

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FMDirectory represents a Football Manager directory structure
type FMDirectory struct {
	BasePath    string // Path to Documents/Sports Interactive/Football Manager XXXX
	GraphicsDir string // Path to graphics directory
	ViewsDir    string // Path to views directory
	FiltersDir  string // Path to filters directory
	ConfigPath  string // Path to config.xml
}

// FindFMDirectoryFromImagePath finds the FM directory based on the image path
func FindFMDirectoryFromImagePath(imagePath string) (*FMDirectory, error) {
	if imagePath == "" {
		return nil, fmt.Errorf("image path is empty")
	}

	// Normalize the path
	imagePath = filepath.Clean(imagePath)

	// Walk up the directory tree to find the FM directory
	currentPath := imagePath
	for {
		// Check if this looks like an FM graphics directory
		if isFMGraphicsDirectory(currentPath) {
			// Found the graphics directory, now find the FM base directory
			fmBasePath := findFMBaseDirectory(currentPath)
			if fmBasePath != "" {
				return &FMDirectory{
					BasePath:    fmBasePath,
					GraphicsDir: currentPath,
					ViewsDir:    filepath.Join(fmBasePath, "views"),
					FiltersDir:  filepath.Join(fmBasePath, "filters"),
					ConfigPath:  filepath.Join(currentPath, "config.xml"),
				}, nil
			}
		}

		// Move up one directory
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached root, stop searching
			break
		}
		currentPath = parentPath
	}

	return nil, fmt.Errorf("could not find FM directory from image path: %s", imagePath)
}

// isFMGraphicsDirectory checks if a path looks like an FM graphics directory
func isFMGraphicsDirectory(path string) bool {
	// Check if path contains "graphics" and is under "Sports Interactive"
	pathLower := strings.ToLower(path)
	if !strings.Contains(pathLower, "graphics") {
		return false
	}

	// Check if it's under Sports Interactive directory
	if !strings.Contains(pathLower, "sports interactive") {
		return false
	}

	// Check if it contains Football Manager
	if !strings.Contains(pathLower, "football manager") {
		return false
	}

	// Additional check: look for config.xml in this directory
	configPath := filepath.Join(path, "config.xml")
	if _, err := os.Stat(configPath); err == nil {
		return true
	}

	return false
}

// findFMBaseDirectory finds the FM base directory from a graphics directory
func findFMBaseDirectory(graphicsPath string) string {
	// The FM base directory should be the parent of the graphics directory
	// and should contain "Football Manager" in its name
	currentPath := filepath.Dir(graphicsPath)

	// Walk up to find the actual FM directory
	for {
		dirName := filepath.Base(currentPath)
		dirNameLower := strings.ToLower(dirName)

		// Check if this looks like an FM directory
		if strings.Contains(dirNameLower, "football manager") ||
			strings.Contains(dirNameLower, "fm") {
			// Verify it's under Sports Interactive
			pathLower := strings.ToLower(currentPath)
			if strings.Contains(pathLower, "sports interactive") {
				return currentPath
			}
		}

		// Move up one directory
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached root, stop searching
			break
		}
		currentPath = parentPath
	}

	return ""
}

// DistributeViewsAndFilters copies views and filters to the FM directory
func DistributeViewsAndFilters(fmDir *FMDirectory) error {
	if fmDir == nil {
		return fmt.Errorf("FM directory is nil")
	}

	// Ensure views and filters directories exist
	if err := os.MkdirAll(fmDir.ViewsDir, 0755); err != nil {
		return fmt.Errorf("failed to create views directory: %w", err)
	}

	if err := os.MkdirAll(fmDir.FiltersDir, 0755); err != nil {
		return fmt.Errorf("failed to create filters directory: %w", err)
	}

	// Copy views
	if err := copyDirectory("views", fmDir.ViewsDir); err != nil {
		return fmt.Errorf("failed to copy views: %w", err)
	}

	// Copy filters
	if err := copyDirectory("filters", fmDir.FiltersDir); err != nil {
		return fmt.Errorf("failed to copy filters: %w", err)
	}

	return nil
}

// GenerateConfigXML creates a default config.xml file in the specified directory
func GenerateConfigXML(imageDir string) error {
	if imageDir == "" {
		return fmt.Errorf("image directory is empty")
	}

	configPath := filepath.Join(imageDir, "config.xml")

	// Check if config.xml already exists
	if _, err := os.Stat(configPath); err == nil {
		// File already exists, don't overwrite
		return nil
	}

	// Create the default config.xml content
	configContent := `<?xml version="1.0" encoding="UTF-8"?>
<record>
	<boolean id="preload" value="false"/>
	<boolean id="amap" value="false"/>
	<list id="maps">
	</list>
</record>`

	// Write the config.xml file
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write config.xml: %w", err)
	}

	return nil
}

// copyDirectory copies all files from source directory to destination
func copyDirectory(srcDir, destDir string) error {
	// Check if source directory exists
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		// Source directory doesn't exist, that's okay
		return nil
	}

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		return copyFile(path, destPath)
	})
}

// copyFile copies a single file
func copyFile(src, dest string) error {
	// Ensure destination directory exists
	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

// GetFMVersionFromPath extracts FM version from the path
func GetFMVersionFromPath(fmPath string) string {
	// Look for patterns like "Football Manager 2022", "FM2022", "2022", etc.
	pathParts := strings.Split(fmPath, string(os.PathSeparator))
	for _, part := range pathParts {
		if strings.Contains(strings.ToLower(part), "football manager") ||
			strings.Contains(strings.ToLower(part), "fm") {
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
