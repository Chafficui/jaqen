package cmd

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	internal "jaqen/internal"
	mapper "jaqen/pkgs"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	nativeDialog "github.com/sqweek/dialog"
)

type JaqenGUI struct {
	app                fyne.App
	window             fyne.Window
	xmlPathEntry       *widget.Entry
	rtfPathEntry       *widget.Entry
	imgDirEntry        *widget.Entry
	preserveCheck      *widget.Check
	allowDuplicateCheck *widget.Check
	fmVersionSelect    *widget.Select
	statusLabel        *widget.Label
	progressBar        *widget.ProgressBar
	runButton          *widget.Button
	mappingOverrideList *widget.List
	mappingOverrides   map[string]string
	config             internal.JaqenConfig
	currentStep        int
	stepContainer      *fyne.Container
	nextButton         *widget.Button
	prevButton         *widget.Button
	advancedMode       bool
}

func NewJaqenGUI() *JaqenGUI {
	myApp := app.NewWithID("com.jaqen.footballmanager")
	myApp.SetIcon(theme.DocumentIcon())
	
	window := myApp.NewWindow("Jaqen - Football Manager Face Manager")
	window.Resize(fyne.NewSize(900, 700))
	window.CenterOnScreen()

	return &JaqenGUI{
		app:              myApp,
		window:           window,
		mappingOverrides: make(map[string]string),
	}
}

func (g *JaqenGUI) loadConfig() {
	configPath := g.configPathEntry.Text
	if configPath == "" {
		configPath = internal.DefaultConfigPath
	}

	// Try to load existing config
	if _, err := os.Stat(configPath); err == nil {
		config, err := internal.ReadConfig(configPath)
		if err == nil {
			g.config = config
			g.applyConfigToGUI()
		}
	} else {
		// Create default config
		g.createDefaultConfig()
	}
}

func (g *JaqenGUI) createDefaultConfig() {
	g.config = internal.JaqenConfig{
		Preserve:        &[]bool{internal.DefaultPreserve}[0],
		XMLPath:         &[]string{internal.DefaultXMLPath}[0],
		RTFPath:         &[]string{internal.DefaultRTFPath}[0],
		IMGPath:         &[]string{internal.DefaultImagesPath}[0],
		FMVersion:       &[]string{internal.DefaultFMVersion}[0],
		AllowDuplicate:  &[]bool{internal.DefaultAllowDuplicate}[0],
		MappingOverride: &map[string]string{},
	}
	g.mappingOverrides = make(map[string]string)
}

func (g *JaqenGUI) applyConfigToGUI() {
	if g.config.Preserve != nil {
		g.preserveCheck.SetChecked(*g.config.Preserve)
	}
	if g.config.AllowDuplicate != nil {
		g.allowDuplicateCheck.SetChecked(*g.config.AllowDuplicate)
	}
	if g.config.FMVersion != nil {
		g.fmVersionSelect.SetSelected(*g.config.FMVersion)
	}
	if g.config.XMLPath != nil {
		g.xmlPathEntry.SetText(*g.config.XMLPath)
	}
	if g.config.RTFPath != nil {
		g.rtfPathEntry.SetText(*g.config.RTFPath)
	}
	if g.config.IMGPath != nil {
		g.imgDirEntry.SetText(*g.config.IMGPath)
	}
	if g.config.MappingOverride != nil {
		g.mappingOverrides = *g.config.MappingOverride
		g.updateMappingOverrideList()
	}
}

func (g *JaqenGUI) saveConfig() {
	// Update config from GUI
	preserve := g.preserveCheck.Checked
	allowDuplicate := g.allowDuplicateCheck.Checked
	fmVersion := g.fmVersionSelect.Selected
	xmlPath := g.xmlPathEntry.Text
	rtfPath := g.rtfPathEntry.Text
	imgPath := g.imgDirEntry.Text

	g.config.Preserve = &preserve
	g.config.AllowDuplicate = &allowDuplicate
	g.config.FMVersion = &fmVersion
	g.config.XMLPath = &xmlPath
	g.config.RTFPath = &rtfPath
	g.config.IMGPath = &imgPath
	g.config.MappingOverride = &g.mappingOverrides

	// Save to file
	configPath := g.configPathEntry.Text
	if configPath == "" {
		configPath = internal.DefaultConfigPath
	}

	err := internal.WriteConfig(g.config, configPath)
	if err != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("Error saving config: %w", err), g.window)
		})
	}
}

func (g *JaqenGUI) createFileSelector(entry *widget.Entry, title string, fileType string) *widget.Button {
	return widget.NewButton("Browse...", func() {
		if fileType == "directory" {
			// Use native folder dialog
			path, err := nativeDialog.Directory().Title("Select Directory").Browse()
			if err == nil && path != "" {
				entry.SetText(path)
			}
		} else {
			// Use native file dialog with proper filters
			var filter string
			switch fileType {
			case "xml":
				filter = "XML Files (*.xml)|*.xml"
			case "rtf":
				filter = "RTF Files (*.rtf)|*.rtf"
			case "toml":
				filter = "TOML Files (*.toml)|*.toml"
			default:
				filter = "All Files (*.*)|*.*"
			}
			
			path, err := nativeDialog.File().Title(title).Filter(filter).Load()
			if err == nil && path != "" {
				entry.SetText(path)
			}
		}
	})
}

func (g *JaqenGUI) updateMappingOverrideList() {
	if g.mappingOverrideList != nil {
		g.mappingOverrideList.Refresh()
	}
}

func (g *JaqenGUI) createMappingOverrideSection() *fyne.Container {
	g.mappingOverrideList = widget.NewList(
		func() int {
			return len(g.mappingOverrides)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewLabel(""),
				widget.NewLabel(""),
				widget.NewButton("Edit", nil),
				widget.NewButton("Delete", nil),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			container := obj.(*fyne.Container)
			countries := make([]string, 0, len(g.mappingOverrides))
			for country := range g.mappingOverrides {
				countries = append(countries, country)
			}
			if id < len(countries) {
				country := countries[id]
				ethnic := g.mappingOverrides[country]
				container.Objects[0].(*widget.Label).SetText(country)
				container.Objects[1].(*widget.Label).SetText(ethnic)
				
				// Edit button
				editBtn := container.Objects[2].(*widget.Button)
				editBtn.OnTapped = func() {
					g.editMappingOverride(country, ethnic)
				}
				
				// Delete button
				deleteBtn := container.Objects[3].(*widget.Button)
				deleteBtn.OnTapped = func() {
					delete(g.mappingOverrides, country)
					g.updateMappingOverrideList()
				}
			}
		},
	)

	addButton := widget.NewButton("Add Override", g.addMappingOverride)
	loadDefaultsButton := widget.NewButton("Load Defaults", g.loadDefaultMappings)

	return container.NewVBox(
		widget.NewCard("Mapping Overrides", "", container.NewVBox(
			g.mappingOverrideList,
			container.NewHBox(addButton, loadDefaultsButton),
		)),
	)
}

func (g *JaqenGUI) addMappingOverride() {
	g.editMappingOverride("", "")
}

func (g *JaqenGUI) editMappingOverride(country, ethnic string) {
	countryEntry := widget.NewEntry()
	countryEntry.SetText(country)
	countryEntry.SetPlaceHolder("Country Code (e.g., ENG)")

	ethnicSelect := widget.NewSelect([]string{
		"African", "Asian", "Caucasian", "Central European", "EECA",
		"Italmed", "MENA", "MESA", "SAMed", "Scandinavian",
		"Seasian", "South American", "SpanMed", "YugoGreek",
	}, nil)
	ethnicSelect.SetSelected(ethnic)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Country Code", Widget: countryEntry},
			{Text: "Ethnic Group", Widget: ethnicSelect},
		},
	}

	dialog.ShowForm("Edit Mapping Override", "Save", "Cancel", form.Items, func(confirmed bool) {
		if confirmed && countryEntry.Text != "" && ethnicSelect.Selected != "" {
			g.mappingOverrides[countryEntry.Text] = ethnicSelect.Selected
			g.updateMappingOverrideList()
		}
	}, g.window)
}

func (g *JaqenGUI) loadDefaultMappings() {
	// Load some common mapping overrides as examples
	g.mappingOverrides["ENG"] = "Caucasian"
	g.mappingOverrides["GER"] = "Central European"
	g.mappingOverrides["FRA"] = "Central European"
	g.mappingOverrides["ESP"] = "SpanMed"
	g.mappingOverrides["ITA"] = "Italmed"
	g.mappingOverrides["BRA"] = "South American"
	g.mappingOverrides["ARG"] = "SAMed"
	g.mappingOverrides["USA"] = "Caucasian"
	g.mappingOverrides["JPN"] = "Asian"
	g.mappingOverrides["CHN"] = "Asian"
	g.updateMappingOverrideList()
}

func (g *JaqenGUI) createInputSection() *fyne.Container {
	// XML File Selection
	xmlLabel := widget.NewLabel("XML Config File:")
	g.xmlPathEntry = widget.NewEntry()
	g.xmlPathEntry.SetText(internal.DefaultXMLPath)
	xmlButton := g.createFileSelector(g.xmlPathEntry, "Select XML File", "xml")

	// RTF File Selection
	rtfLabel := widget.NewLabel("RTF Player File:")
	g.rtfPathEntry = widget.NewEntry()
	g.rtfPathEntry.SetText(internal.DefaultRTFPath)
	rtfButton := g.createFileSelector(g.rtfPathEntry, "Select RTF File", "rtf")

	// Image Directory Selection
	imgLabel := widget.NewLabel("Image Directory:")
	g.imgDirEntry = widget.NewEntry()
	g.imgDirEntry.SetText(internal.DefaultImagesPath)
	imgButton := g.createFileSelector(g.imgDirEntry, "Select Image Directory", "directory")

	// Config File Selection
	configLabel := widget.NewLabel("Config File (Optional):")
	g.configPathEntry = widget.NewEntry()
	g.configPathEntry.SetText(internal.DefaultConfigPath)
	configButton := g.createFileSelector(g.configPathEntry, "Select Config File", "toml")

	return container.NewVBox(
		widget.NewCard("File Selection", "", container.NewVBox(
			container.NewBorder(nil, nil, xmlLabel, xmlButton, g.xmlPathEntry),
			container.NewBorder(nil, nil, rtfLabel, rtfButton, g.rtfPathEntry),
			container.NewBorder(nil, nil, imgLabel, imgButton, g.imgDirEntry),
			container.NewBorder(nil, nil, configLabel, configButton, g.configPathEntry),
		)),
	)
}

func (g *JaqenGUI) createOptionsSection() *fyne.Container {
	// Checkboxes
	g.preserveCheck = widget.NewCheck("Preserve existing mappings", nil)
	g.preserveCheck.SetChecked(internal.DefaultPreserve)
	
	g.allowDuplicateCheck = widget.NewCheck("Allow duplicate images", nil)
	g.allowDuplicateCheck.SetChecked(internal.DefaultAllowDuplicate)

	// FM Version Selection
	fmVersionLabel := widget.NewLabel("Football Manager Version:")
	g.fmVersionSelect = widget.NewSelect([]string{"2024", "2023", "2022", "2021", "2020"}, nil)
	g.fmVersionSelect.SetSelected(internal.DefaultFMVersion)

	return container.NewVBox(
		widget.NewCard("Options", "", container.NewVBox(
			g.preserveCheck,
			g.allowDuplicateCheck,
			container.NewBorder(nil, nil, fmVersionLabel, nil, g.fmVersionSelect),
		)),
	)
}

func (g *JaqenGUI) createActionSection() *fyne.Container {
	g.statusLabel = widget.NewLabel("Ready to process files...")
	g.statusLabel.Wrapping = fyne.TextWrapWord
	
	g.progressBar = widget.NewProgressBar()
	g.progressBar.Hide()

	g.runButton = widget.NewButton("Process Files", g.runProcessing)
	g.runButton.Importance = widget.HighImportance

	saveButton := widget.NewButton("Save Config", g.saveConfig)
	loadButton := widget.NewButton("Load Config", g.loadConfig)
	formatButton := widget.NewButton("Format Config", g.formatConfig)

	return container.NewVBox(
		widget.NewCard("Actions", "", container.NewVBox(
			g.statusLabel,
			g.progressBar,
			container.NewHBox(g.runButton, saveButton, loadButton, formatButton),
		)),
	)
}

func (g *JaqenGUI) formatConfig() {
	configPath := g.configPathEntry.Text
	if configPath == "" {
		configPath = internal.DefaultConfigPath
	}

	g.statusLabel.SetText("Formatting config file...")
	g.progressBar.Show()
	g.progressBar.SetValue(0.5)

	// Import the format functionality from cmd/format.go
	// For now, we'll show a simple message
	dialog.ShowInformation("Config Format", "Config file formatting would be implemented here", g.window)
	
	g.progressBar.Hide()
	g.statusLabel.SetText("Config formatting completed.")
}

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
		dialog.ShowError(fmt.Errorf("Image directory path is required"), g.window)
		return
	}

	// Save config before processing
	g.saveConfig()

	g.runButton.SetText("Processing...")
	g.runButton.Disable()
	g.progressBar.Show()
	g.statusLabel.SetText("Starting file processing...")

	// Run processing in a goroutine to keep UI responsive
	go g.processFiles()
}

func (g *JaqenGUI) processFiles() {
	defer func() {
		fyne.Do(func() {
			g.runButton.SetText("Process Files")
			g.runButton.Enable()
			g.progressBar.Hide()
		})
	}()

	// Update progress
	fyne.Do(func() {
		g.progressBar.SetValue(0.1)
		g.statusLabel.SetText("Reading configuration...")
	})

	// Load config if exists
	configPath := g.configPathEntry.Text
	if configPath == "" {
		configPath = internal.DefaultConfigPath
	}

	if _, err := os.Stat(configPath); err == nil {
		_, err := internal.ReadConfig(configPath)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("Error reading config: %w", err), g.window)
			})
			return
		}
		// TODO: Apply config overrides to GUI fields
	}

	fyne.Do(func() {
		g.progressBar.SetValue(0.2)
		g.statusLabel.SetText("Creating mapping...")
	})

	// Apply mapping overrides
	if len(g.mappingOverrides) > 0 {
		err := mapper.OverrideNationEthnicMapping(g.mappingOverrides)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("Error applying mapping overrides: %w", err), g.window)
			})
			return
		}
	}

	// Create mapping
	mapping, err := mapper.NewMapping(g.xmlPathEntry.Text, g.fmVersionSelect.Selected)
	if err != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("Error creating mapping: %w", err), g.window)
		})
		return
	}

	fyne.Do(func() {
		g.progressBar.SetValue(0.3)
		g.statusLabel.SetText("Loading image pool...")
	})

	// Create image pool
	imagePool, err := mapper.NewImagePool(g.imgDirEntry.Text)
	if err != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("Error loading image pool: %w", err), g.window)
		})
		return
	}

	fyne.Do(func() {
		g.progressBar.SetValue(0.4)
		g.statusLabel.SetText("Processing players...")
	})

	// Get players
	players, err := mapper.GetPlayers(g.rtfPathEntry.Text)
	if err != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("Error reading players: %w", err), g.window)
		})
		return
	}

	fyne.Do(func() {
		g.progressBar.SetValue(0.5)
		g.statusLabel.SetText("Assigning faces...")
	})

	// Process each player
	totalPlayers := len(players)
	for i, player := range players {
		if g.preserveCheck.Checked && mapping.Exist(player.ID) {
			continue
		}

		imgFilename, err := imagePool.GetRandomImagePath(player.Ethnic, !g.allowDuplicateCheck.Checked)
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
		fyne.Do(func() {
			g.progressBar.SetValue(progress)
			g.statusLabel.SetText(fmt.Sprintf("Processing player %d of %d...", i+1, totalPlayers))
		})
	}

	fyne.Do(func() {
		g.progressBar.SetValue(0.9)
		g.statusLabel.SetText("Saving files...")
	})

	// Save mapping
	if err := mapping.Save(); err != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("Error saving mapping: %w", err), g.window)
		})
		return
	}

	if err := mapping.Write(g.xmlPathEntry.Text); err != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("Error writing XML file: %w", err), g.window)
		})
		return
	}

	fyne.Do(func() {
		g.progressBar.SetValue(1.0)
		g.statusLabel.SetText("Processing completed successfully!")
		dialog.ShowInformation("Success", "Face mapping completed successfully!", g.window)
	})
}

func (g *JaqenGUI) ShowAndRun() {
	// Initialize config
	g.createDefaultConfig()
	
	// Create main layout
	content := container.NewVBox(
		g.createInputSection(),
		g.createOptionsSection(),
		g.createMappingOverrideSection(),
		g.createActionSection(),
	)

	// Add some padding
	paddedContent := container.NewPadded(content)

	g.window.SetContent(paddedContent)
	g.window.ShowAndRun()
}

func runGUI(cmd *cobra.Command, args []string) {
	gui := NewJaqenGUI()
	gui.ShowAndRun()
}

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Launch the graphical user interface",
	Long:  "Launch the modern GUI for Jaqen Football Manager face management",
	Run:   runGUI,
}