package gui

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"

	internal "jaqen/internal"
)

// loadConfig loads the configuration from the default config file
func (g *JaqenGUI) loadConfig() {
	configPath := internal.GetDefaultConfigPath()

	// Try to load existing config
	if _, err := os.Stat(configPath); err == nil {
		config, err := internal.ReadConfig(configPath)
		if err == nil {
			g.config = config
			// Only apply config to GUI if widgets exist
			if g.xmlPathEntry != nil {
				g.applyConfigToGUI()
			}
		} else {
			// If config file exists but can't be read, create default but don't overwrite existing values
			g.createDefaultConfig()
		}
	} else {
		// Create default config only if no config file exists
		g.createDefaultConfig()
	}
}

// createDefaultConfig creates a default configuration
func (g *JaqenGUI) createDefaultConfig() {
	// Preserve existing mapping overrides if they exist
	var mappingOverride map[string]string
	if g.config.MappingOverride != nil {
		mappingOverride = *g.config.MappingOverride
	} else {
		mappingOverride = make(map[string]string)
	}

	g.config = internal.JaqenConfig{
		Preserve:        &[]bool{true}[0], // Default to true
		XMLPath:         &[]string{internal.DefaultXMLPath}[0],
		RTFPath:         &[]string{internal.DefaultRTFPath}[0],
		IMGPath:         &[]string{internal.DefaultImagesPath}[0],
		FMVersion:       &[]string{internal.DefaultFMVersion}[0],
		AllowDuplicate:  &[]bool{true}[0], // Default to true
		MappingOverride: &mappingOverride,
	}
	g.mappingOverrides = make(map[string]string)
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
	if g.fmVersionSelect != nil && g.config.FMVersion != nil && *g.config.FMVersion != "" {
		g.fmVersionSelect.SetSelected(*g.config.FMVersion)
	}
	if g.xmlPathEntry != nil && g.config.XMLPath != nil && *g.config.XMLPath != "" {
		g.xmlPathEntry.SetText(*g.config.XMLPath)
	}
	if g.rtfPathEntry != nil && g.config.RTFPath != nil && *g.config.RTFPath != "" {
		g.rtfPathEntry.SetText(*g.config.RTFPath)
	}
	if g.imgDirEntry != nil && g.config.IMGPath != nil && *g.config.IMGPath != "" {
		g.imgDirEntry.SetText(*g.config.IMGPath)
	}
	if g.config.MappingOverride != nil {
		// Copy the mapping overrides from config to GUI map
		g.mappingOverrides = make(map[string]string)
		for k, v := range *g.config.MappingOverride {
			g.mappingOverrides[k] = v
		}
		g.updateMappingOverrideList()
	}
}

// saveConfig saves the current configuration to file
func (g *JaqenGUI) saveConfig() {
	// Update config from GUI (only if widgets exist)
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
		if fmVersion != "" {
			g.config.FMVersion = &fmVersion
		}
	}
	if g.xmlPathEntry != nil {
		xmlPath := g.xmlPathEntry.Text
		if xmlPath != "" {
			g.config.XMLPath = &xmlPath
		}
	}
	if g.rtfPathEntry != nil {
		rtfPath := g.rtfPathEntry.Text
		if rtfPath != "" {
			g.config.RTFPath = &rtfPath
		}
	}
	if g.imgDirEntry != nil {
		imgPath := g.imgDirEntry.Text
		if imgPath != "" {
			g.config.IMGPath = &imgPath
		}
	}
	g.config.MappingOverride = &g.mappingOverrides

	// Save to default config file
	configPath := internal.GetDefaultConfigPath()
	err := internal.WriteConfig(g.config, configPath)
	if err != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("error saving config: %w", err), g.window)
		})
	}
}

// autoSaveConfig automatically saves config in a goroutine
func (g *JaqenGUI) autoSaveConfig() {
	// Auto-save config whenever settings change
	go func() {
		g.saveConfig()
	}()
}
