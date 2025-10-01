package gui

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	internal "jaqen/internal"
)

// initializeProfiles sets up the profile manager and auto-creates profiles for FM installations
func (g *JaqenGUI) initializeProfiles() error {
	pm, err := internal.NewProfileManager()
	if err != nil {
		return fmt.Errorf("failed to initialize profile manager: %w", err)
	}
	g.profileManager = pm

	// Find all FM installations
	fmDirs := g.findAllFMInstallations()

	if g.logger != nil {
		g.logger.Printf("Found %d FM installation(s)", len(fmDirs))
	}

	// Create profiles for each FM installation if they don't exist
	for _, fmDir := range fmDirs {
		// Extract version from path for profile name
		version := g.extractFMVersion(fmDir.BasePath)
		profileName := fmt.Sprintf("FM %s", version)
		if version == "" {
			// Fallback to directory name
			profileName = filepath.Base(fmDir.BasePath)
		}

		// Check if profile already exists
		if _, err := pm.GetProfile(profileName); err != nil {
			// Profile doesn't exist, create it
			if _, err := pm.CreateProfileForGame(profileName, fmDir.BasePath); err != nil {
				if g.logger != nil {
					g.logger.Printf("Warning: Failed to create profile for %s: %v", fmDir.BasePath, err)
				}
			} else {
				if g.logger != nil {
					g.logger.Printf("Created profile '%s' for game at: %s", profileName, fmDir.BasePath)
				}
			}
		}
	}

	// Load first available profile, or create a generic one if none exist
	firstProfile, err := pm.GetFirstProfile()
	if err != nil {
		// No profiles exist, create a generic default
		if g.logger != nil {
			g.logger.Println("No profiles found, creating generic profile...")
		}
		firstProfile, err = pm.CreateProfileForGame("My Profile", "")
		if err != nil {
			return fmt.Errorf("failed to create initial profile: %w", err)
		}
	}

	// Set flag to prevent auto-save during initial load
	// This will be unset in ShowAndRun() after applyConfigToGUI() is called
	g.isLoadingProfile = true

	g.currentProfile = firstProfile
	g.config = firstProfile.Config

	if g.logger != nil {
		g.logger.Printf("Loaded profile: %s", g.currentProfile.Name)
	}

	return nil
}

// switchProfile switches to a different profile
func (g *JaqenGUI) switchProfile(profileName string) error {
	profile, err := g.profileManager.GetProfile(profileName)
	if err != nil {
		return err
	}

	// Save current profile before switching
	if g.currentProfile != nil {
		if err := g.saveCurrentProfile(); err != nil {
			if g.logger != nil {
				g.logger.Printf("Warning: Failed to save current profile: %v", err)
			}
		}
	}

	// Set flag to prevent auto-save during profile loading
	g.isLoadingProfile = true
	defer func() {
		g.isLoadingProfile = false
	}()

	// Switch to new profile
	g.currentProfile = profile
	g.config = profile.Config

	// Apply to GUI - this will update all widgets
	if g.xmlPathEntry != nil {
		g.applyConfigToGUI()
	}

	if g.logger != nil {
		g.logger.Printf("Switched to profile: %s", profileName)
	}

	return nil
}

// saveCurrentProfile saves the current profile
func (g *JaqenGUI) saveCurrentProfile() error {
	if g.currentProfile == nil {
		return fmt.Errorf("no current profile")
	}

	// Update config from GUI
	g.updateConfigFromGUI()

	// Update profile config
	g.currentProfile.Config = g.config

	// Save to disk
	if err := g.profileManager.SaveProfile(g.currentProfile); err != nil {
		return err
	}

	return nil
}

// updateConfigFromGUI updates the config struct from GUI values
func (g *JaqenGUI) updateConfigFromGUI() {
	if g.preserveCheck != nil {
		preserve := g.preserveCheck.Checked
		g.config.Preserve = &preserve
	}
	if g.allowDuplicateCheck != nil {
		allowDuplicate := g.allowDuplicateCheck.Checked
		g.config.AllowDuplicate = &allowDuplicate
	}
	if g.fmVersionSelect != nil {
		fmVersion := g.fmVersionSelect.Selected
		g.config.FMVersion = &fmVersion
	}
	if g.xmlPathEntry != nil {
		xmlPath := g.xmlPathEntry.Text
		g.config.XMLPath = &xmlPath
	}
	if g.rtfPathEntry != nil {
		rtfPath := g.rtfPathEntry.Text
		g.config.RTFPath = &rtfPath
	}
	if g.imgDirEntry != nil {
		imgPath := g.imgDirEntry.Text
		g.config.IMGPath = &imgPath
	}
	g.config.MappingOverride = &g.mappingOverrides
}

// showCreateProfileDialog shows dialog to create a new profile
func (g *JaqenGUI) showCreateProfileDialog() {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter profile name...")

	formItems := []*widget.FormItem{
		{Text: "Profile Name", Widget: nameEntry},
	}

	dialog.ShowForm("Create New Profile", "Create", "Cancel", formItems, func(confirmed bool) {
		if !confirmed || nameEntry.Text == "" {
			return
		}

		profileName := nameEntry.Text

		// Save current profile before creating new one
		if err := g.saveCurrentProfile(); err != nil {
			if g.logger != nil {
				g.logger.Printf("Warning: Failed to save current profile before creating new: %v", err)
			}
		}

		// Create new profile based on current one
		newProfile, err := g.profileManager.CreateProfile(profileName, g.currentProfile)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to create profile: %w", err), g.window)
			return
		}

		// Set flag to prevent auto-save during switching
		g.isLoadingProfile = true
		defer func() {
			g.isLoadingProfile = false
		}()

		// Switch to new profile
		g.currentProfile = newProfile
		g.config = newProfile.Config

		// Update profile selector
		g.updateProfileSelector()

		if g.logger != nil {
			g.logger.Printf("Created new profile: %s", profileName)
		}

		dialog.ShowInformation("Success", fmt.Sprintf("Profile '%s' created successfully!", profileName), g.window)
	}, g.window)
}

// showDeleteProfileDialog shows dialog to confirm profile deletion
func (g *JaqenGUI) showDeleteProfileDialog() {
	// Check if there are other profiles to switch to
	profiles := g.profileManager.ListProfiles()
	if len(profiles) <= 1 {
		dialog.ShowError(fmt.Errorf("cannot delete the only profile"), g.window)
		return
	}

	dialog.ShowConfirm("Delete Profile",
		fmt.Sprintf("Are you sure you want to delete profile '%s'?", g.currentProfile.Name),
		func(confirmed bool) {
			if !confirmed {
				return
			}

			profileToDelete := g.currentProfile.Name

			// Switch to another profile first (the first one that's not the current)
			var targetProfile string
			for _, p := range profiles {
				if p != profileToDelete {
					targetProfile = p
					break
				}
			}

			if err := g.switchProfile(targetProfile); err != nil {
				dialog.ShowError(fmt.Errorf("failed to switch profile: %w", err), g.window)
				return
			}

			// Delete the profile
			if err := g.profileManager.DeleteProfile(profileToDelete); err != nil {
				dialog.ShowError(fmt.Errorf("failed to delete profile: %w", err), g.window)
				return
			}

			// Update profile selector
			g.updateProfileSelector()

			if g.logger != nil {
				g.logger.Printf("Deleted profile: %s", profileToDelete)
			}

			dialog.ShowInformation("Success", fmt.Sprintf("Profile '%s' deleted successfully!", profileToDelete), g.window)
		}, g.window)
}

// updateProfileSelector updates the profile dropdown with current profiles
func (g *JaqenGUI) updateProfileSelector() {
	if g.profileSelect == nil {
		return
	}

	profiles := g.profileManager.ListProfiles()
	g.profileSelect.Options = profiles
	g.profileSelect.SetSelected(g.currentProfile.Name)
	g.profileSelect.Refresh()
}

// applyConfigToGUI applies the configuration to GUI widgets
func (g *JaqenGUI) applyConfigToGUI() {
	// Only apply config if widgets have been created
	if g.preserveCheck != nil && g.config.Preserve != nil {
		g.preserveCheck.SetChecked(*g.config.Preserve)
	}
	if g.allowDuplicateCheck != nil && g.config.AllowDuplicate != nil {
		g.allowDuplicateCheck.SetChecked(*g.config.AllowDuplicate)
	}
	if g.fmVersionSelect != nil && g.config.FMVersion != nil {
		g.fmVersionSelect.SetSelected(*g.config.FMVersion)
	}

	// Apply paths - always update, even if empty
	if g.xmlPathEntry != nil && g.config.XMLPath != nil {
		g.xmlPathEntry.SetText(*g.config.XMLPath)
	}
	if g.rtfPathEntry != nil && g.config.RTFPath != nil {
		g.rtfPathEntry.SetText(*g.config.RTFPath)
	}
	if g.imgDirEntry != nil && g.config.IMGPath != nil {
		g.imgDirEntry.SetText(*g.config.IMGPath)
		// Also update image previews when loading profile
		if *g.config.IMGPath != "" {
			g.updateImagePreviews(*g.config.IMGPath)
		}
	}

	// Apply mapping overrides
	if g.config.MappingOverride != nil {
		// Copy the mapping overrides from config to GUI map
		g.mappingOverrides = make(map[string]string)
		for k, v := range *g.config.MappingOverride {
			g.mappingOverrides[k] = v
		}
		g.updateMappingOverrideList()
	} else {
		// Clear mapping overrides if none in config
		g.mappingOverrides = make(map[string]string)
		g.updateMappingOverrideList()
	}

	if g.logger != nil {
		g.logger.Printf("Applied profile settings to GUI - XMLPath: %v, RTFPath: %v, IMGPath: %v",
			ptrToStr(g.config.XMLPath),
			ptrToStr(g.config.RTFPath),
			ptrToStr(g.config.IMGPath))
	}
}

// ptrToStr converts a string pointer to string for logging
func ptrToStr(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

// autoSaveConfig automatically saves the current profile
func (g *JaqenGUI) autoSaveConfig() {
	// Don't auto-save if we're currently loading a profile
	if g.isLoadingProfile {
		return
	}

	// Auto-save profile whenever settings change
	go func() {
		if g.profileManager != nil && g.currentProfile != nil {
			if err := g.saveCurrentProfile(); err != nil {
				if g.logger != nil {
					g.logger.Printf("Error auto-saving profile: %v", err)
				}
			}
		}
	}()
}
