package gui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	internal "jaqen/internal"
	mapper "jaqen/pkgs"
)

// runProcessing starts the face mapping process
func (g *JaqenGUI) runProcessing() {
	// Validate inputs
	if g.xmlPathEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("XML file path is required"), g.window)
		return
	}
	if g.rtfPathEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("RTF file path is required"), g.window)
		return
	}
	if g.imgDirEntry.Text == "" {
		dialog.ShowError(fmt.Errorf("image directory path is required"), g.window)
		return
	}

	// Log processing start
	if g.logger != nil {
		g.logger.Printf("Starting face mapping process")
		g.logger.Printf("XML Path: %s", g.xmlPathEntry.Text)
		g.logger.Printf("RTF Path: %s", g.rtfPathEntry.Text)
		g.logger.Printf("Image Directory: %s", g.imgDirEntry.Text)
		g.logger.Printf("FM Version: %s", g.fmVersionSelect.Selected)
	}

	// Save current profile before processing
	if g.profileManager != nil && g.currentProfile != nil {
		g.saveCurrentProfile()
	}

	g.runButton.SetText("Processing...")
	g.runButton.Disable()
	g.progressBar.Show()

	if g.logger != nil {
		g.logger.Println("Starting face mapping process...")
	}

	// Run processing in a goroutine to keep UI responsive
	go g.processFiles()
}

// processFiles performs the actual file processing
func (g *JaqenGUI) processFiles() {
	defer func() {
		if g.logger != nil {
			g.logger.Printf("Face mapping process completed")
		}
		fyne.Do(func() {
			g.runButton.SetText("Assign Face Mappings")
			g.runButton.Enable()
			g.progressBar.Hide()
		})
	}()

	// Update progress
	fyne.Do(func() {
		g.progressBar.SetValue(0.1)
	})

	if g.logger != nil {
		g.logger.Println("Reading configuration...")
	}

	// Load config if exists
	configPath := internal.GetDefaultConfigPath()

	if _, err := os.Stat(configPath); err == nil {
		_, err := internal.ReadConfig(configPath)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("error reading config: %w", err), g.window)
			})
			return
		}
	}

	fyne.Do(func() {
		g.progressBar.SetValue(0.2)
	})

	if g.logger != nil {
		g.logger.Println("Creating mapping...")
	}

	// Apply mapping overrides
	if len(g.mappingOverrides) > 0 {
		err := mapper.OverrideNationEthnicMapping(g.mappingOverrides)
		if err != nil {
			fyne.Do(func() {
				// Create a detailed error message
				dialog.ShowError(fmt.Errorf("error applying mapping overrides:\n\n%v\n\nPlease check your mapping overrides in Settings and ensure all ethnic groups are valid", err), g.window)
			})
			return
		}
	}

	// Create mapping
	mapping, err := mapper.NewMapping(g.xmlPathEntry.Text, g.fmVersionSelect.Selected)
	if err != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("error creating mapping: %w", err), g.window)
		})
		return
	}

	fyne.Do(func() {
		g.progressBar.SetValue(0.3)
	})

	if g.logger != nil {
		g.logger.Println("Loading image pool...")
	}

	// Create image pool
	imagePool, err := mapper.NewImagePool(g.imgDirEntry.Text)
	if err != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("error loading image pool: %w", err), g.window)
		})
		return
	}

	fyne.Do(func() {
		g.progressBar.SetValue(0.4)
	})

	if g.logger != nil {
		g.logger.Println("Processing players...")
	}

	// Get players
	players, err := mapper.GetPlayers(g.rtfPathEntry.Text)
	if err != nil {
		fyne.Do(func() {
			// Check if this is an ethnicity-related error
			errorStr := err.Error()
			if strings.Contains(errorStr, "ethnic not found for country initials") ||
				strings.Contains(errorStr, "ethnic value not found") {
				// Create a detailed error message for ethnicity issues
				dialog.ShowError(fmt.Errorf("ethnicity detection error:\n\n%v\n\nThis means some countries in your RTF file are not recognized.\n\nYou can add custom mappings in Settings → Mapping Overrides to fix this", err), g.window)
			} else {
				dialog.ShowError(fmt.Errorf("error reading players: %w", err), g.window)
			}
		})
		return
	}

	fyne.Do(func() {
		g.progressBar.SetValue(0.5)
	})

	if g.logger != nil {
		g.logger.Println("Assigning faces to players...")
	}

	// Process each player
	totalPlayers := len(players)
	for i, player := range players {
		if g.preserveCheck != nil && g.preserveCheck.Checked && mapping.Exist(player.ID) {
			continue
		}

		allowDuplicates := true
		if g.allowDuplicateCheck != nil {
			allowDuplicates = g.allowDuplicateCheck.Checked
		}
		imgFilename, err := imagePool.GetRandomImagePath(player.Ethnic, !allowDuplicates)
		if err != nil {
			log.Printf("Error getting image for player %s: %v", player.ID, err)
			continue
		}

		// Calculate relative path
		rel := ""
		imgDirPathAbs, _ := filepath.Abs(g.imgDirEntry.Text)
		xmlFilePathAbs, _ := filepath.Abs(g.xmlPathEntry.Text)

		if imgDirPathAbs != filepath.Dir(xmlFilePathAbs) {
			rel, _ = filepath.Rel(xmlFilePathAbs, imgDirPathAbs)
		}
		rel = strings.TrimPrefix(rel, "./")

		mapping.MapToImage(player.ID, mapper.FilePath(filepath.Join(rel, string(player.Ethnic), string(imgFilename))))

		// Update progress
		progress := 0.5 + (float64(i+1)/float64(totalPlayers))*0.4

		// Log every 10% or so to avoid spam
		if i%10 == 0 || i == totalPlayers-1 {
			if g.logger != nil {
				g.logger.Printf("Processing player %d of %d...", i+1, totalPlayers)
			}
		}

		fyne.Do(func() {
			g.progressBar.SetValue(progress)
		})
	}

	fyne.Do(func() {
		g.progressBar.SetValue(0.9)
	})

	if g.logger != nil {
		g.logger.Println("Saving mapping files...")
	}

	// Save mapping
	if err := mapping.Save(); err != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("error saving mapping: %w", err), g.window)
		})
		return
	}

	if err := mapping.Write(g.xmlPathEntry.Text); err != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("error writing XML file: %w", err), g.window)
		})
		return
	}

	if g.logger != nil {
		g.logger.Println("✅ Face mapping completed successfully!")
	}

	fyne.Do(func() {
		g.progressBar.SetValue(1.0)
		dialog.ShowInformation("Success", "Face mapping completed successfully!", g.window)
	})
}
