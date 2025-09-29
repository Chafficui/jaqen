package cmd

import (
	"fmt"
	internal "jaqen/internal"
	mapper "jaqen/pkgs"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/spf13/cobra"
	nativeDialog "github.com/sqweek/dialog"
)

type JaqenGUI struct {
	app            fyne.App
	window         fyne.Window
	settingsWindow fyne.Window

	// Main content
	imgDirEntry     *widget.Entry
	xmlPathEntry    *widget.Entry
	rtfPathEntry    *widget.Entry
	fmVersionSelect *widget.Select
	statusLabel     *widget.Label
	progressBar     *widget.ProgressBar
	runButton       *widget.Button

	// Settings
	preserveCheck       *widget.Check
	allowDuplicateCheck *widget.Check
	mappingOverrideList *widget.List
	mappingOverrides    map[string]string

	// Image preview
	imagePreview1 *widget.Card
	imagePreview2 *widget.Card
	imagePreview3 *widget.Card

	// State tracking
	hasShownRTFPopup bool

	config internal.JaqenConfig
}

func NewJaqenGUI() *JaqenGUI {
	myApp := app.NewWithID("com.jaqen.footballmanager")
	myApp.SetIcon(theme.DocumentIcon())

	window := myApp.NewWindow("Jaqen NewGen Tool - Football Manager Face Manager")
	window.Resize(fyne.NewSize(1400, 900))
	window.CenterOnScreen()

	return &JaqenGUI{
		app:              myApp,
		window:           window,
		mappingOverrides: make(map[string]string),
	}
}

func (g *JaqenGUI) loadConfig() {
	configPath := internal.DefaultConfigPath

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

func (g *JaqenGUI) applyConfigToSettings() {
	// Apply config specifically to settings widgets
	if g.preserveCheck != nil && g.config.Preserve != nil {
		g.preserveCheck.SetChecked(*g.config.Preserve)
	}
	if g.allowDuplicateCheck != nil && g.config.AllowDuplicate != nil {
		g.allowDuplicateCheck.SetChecked(*g.config.AllowDuplicate)
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
	configPath := internal.DefaultConfigPath
	err := internal.WriteConfig(g.config, configPath)
	if err != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("error saving config: %w", err), g.window)
		})
	}
}

func (g *JaqenGUI) autoSaveConfig() {
	// Auto-save config whenever settings change
	go func() {
		g.saveConfig()
	}()
}

func (g *JaqenGUI) detectPathsFromImageFolder(imgPath string) {
	if imgPath == "" {
		return
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
		// config.xml doesn't exist, generate it
		g.autoGenerateConfigXML(imgPath)
		g.xmlPathEntry.SetText(configPath)
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

	// Auto-distribute views and filters to the FM directory
	g.autoDistributeViewsAndFilters(imgPath)
}

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
			// Log error but don't show to user - this is a background operation
			fmt.Printf("Warning: Failed to distribute views/filters: %v\n", err)
			return
		}

		// Update status to show success
		fyne.Do(func() {
			if g.statusLabel != nil {
				g.statusLabel.SetText(fmt.Sprintf("Views and filters distributed to FM directory: %s", fmDir.BasePath))
			}
		})
	}()
}

// autoGenerateConfigXML automatically generates a config.xml file if it doesn't exist
func (g *JaqenGUI) autoGenerateConfigXML(imgPath string) {
	go func() {
		err := mapper.GenerateConfigXML(imgPath)
		if err != nil {
			// Log error but don't show to user - this is a background operation
			fmt.Printf("Warning: Failed to generate config.xml: %v\n", err)
			return
		}

		// Update status to show success
		fyne.Do(func() {
			if g.statusLabel != nil {
				g.statusLabel.SetText(fmt.Sprintf("Generated config.xml in: %s", imgPath))
			}
		})
	}()
}

// showRTFInstructionsPopup displays complete Jaqen NewGen instructions
func (g *JaqenGUI) showRTFInstructionsPopup() {
	// Create a custom dialog with video link and instructions
	videoLabel := widget.NewLabel("📺 Watch the Tutorial Video:")
	videoLabel.TextStyle.Bold = true

	videoButton := widget.NewButton("🎥 Jaqen NewGen Tutorial", func() {
		// Open YouTube link in default browser
		cmd := exec.Command("xdg-open", "https://youtu.be/aHnrpfH--ic")
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/c", "start", "https://youtu.be/aHnrpfH--ic")
		} else if runtime.GOOS == "darwin" {
			cmd = exec.Command("open", "https://youtu.be/aHnrpfH--ic")
		}
		cmd.Run()
	})
	videoButton.Importance = widget.HighImportance

	instructions := `# Quick Setup Guide

## 1. Export RTF from Football Manager
- Go to Scouting → Players in Range
- Import "SCRIPT FACES player search" view
- Apply "is newgen search filter"
- Select all (Ctrl+A) → Print to text file (Ctrl+P)
- Save as "newgen.rtf" in your image folder

## 2. Configure Jaqen
- Select your image directory
- Choose settings (Preserve, Allow Duplicates, etc.)
- Click "Assign Face Mappings"

## 3. Apply in Football Manager
- Restart Football Manager
- Newgen faces will use assigned images

**Note:** Views, filters, and config.xml are auto-distributed/generated.`

	content := widget.NewRichTextFromMarkdown(instructions)
	content.Wrapping = fyne.TextWrapWord

	// Combine video button and instructions
	videoContainer := container.NewVBox(videoLabel, videoButton)
	instructionsContainer := container.NewScroll(content)
	instructionsContainer.SetMinSize(fyne.NewSize(600, 300))

	mainContainer := container.NewVBox(
		videoContainer,
		widget.NewSeparator(),
		instructionsContainer,
	)

	dialog := dialog.NewCustom("Jaqen NewGen Tool - Instructions", "Close", mainContainer, g.window)
	dialog.Resize(fyne.NewSize(650, 500))
	dialog.Show()
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

func (g *JaqenGUI) createHeaderBar() *fyne.Container {
	title := widget.NewLabel("Jaqen NewGen - Football Manager Face Manager")
	title.TextStyle.Bold = true

	settingsButton := widget.NewButtonWithIcon("", theme.SettingsIcon(), g.openSettings)
	settingsButton.Importance = widget.MediumImportance

	return container.NewBorder(nil, nil, title, settingsButton, widget.NewSeparator())
}

func (g *JaqenGUI) openSettings() {
	if g.settingsWindow != nil {
		g.settingsWindow.Show()
		return
	}

	g.settingsWindow = g.app.NewWindow("Settings")
	g.settingsWindow.Resize(fyne.NewSize(600, 500)) // Larger settings window
	g.settingsWindow.SetContent(g.createSettingsContent())

	// Apply config to settings widgets after they're created
	g.applyConfigToSettings()

	// Close settings window when main window closes
	g.settingsWindow.SetCloseIntercept(func() {
		g.settingsWindow.Hide()
	})

	g.settingsWindow.Show()
}

func (g *JaqenGUI) createSettingsContent() *fyne.Container {
	// Preserve checkbox
	g.preserveCheck = widget.NewCheck("Preserve existing mappings", nil)
	g.preserveCheck.SetChecked(true)
	g.preserveCheck.OnChanged = func(_ bool) { g.autoSaveConfig() }

	// Allow duplicate checkbox
	g.allowDuplicateCheck = widget.NewCheck("Allow duplicate mappings", nil)
	g.allowDuplicateCheck.SetChecked(true)
	g.allowDuplicateCheck.OnChanged = func(_ bool) { g.autoSaveConfig() }

	// Mapping overrides section
	mappingOverrideCard := widget.NewCard("Mapping Overrides", "", g.createMappingOverrideSection())

	return container.NewVBox(
		widget.NewCard("General Settings", "", container.NewVBox(
			g.preserveCheck,
			g.allowDuplicateCheck,
		)),
		mappingOverrideCard,
	)
}

func (g *JaqenGUI) updateMappingOverrideList() {
	if g.mappingOverrideList != nil {
		g.mappingOverrideList.Refresh()
	}
}

func (g *JaqenGUI) findRandomImages(imgDir string) []string {
	var imageFiles []string

	// Supported image extensions
	extensions := []string{".png", ".jpg", ".jpeg", ".gif", ".bmp"}

	// Walk through the directory and subdirectories
	err := filepath.Walk(imgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			for _, supportedExt := range extensions {
				if ext == supportedExt {
					imageFiles = append(imageFiles, path)
					break
				}
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("Error walking directory: %v", err)
		return []string{}
	}

	// Shuffle the images and return up to 3
	if len(imageFiles) == 0 {
		return []string{}
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(imageFiles), func(i, j int) {
		imageFiles[i], imageFiles[j] = imageFiles[j], imageFiles[i]
	})

	// Return up to 3 images
	if len(imageFiles) > 3 {
		return imageFiles[:3]
	}
	return imageFiles
}

func (g *JaqenGUI) updateImagePreviews(imgDir string) {
	if imgDir == "" {
		// Clear previews if no directory selected
		if g.imagePreview1 != nil {
			g.imagePreview1.SetContent(widget.NewLabel("No folder selected"))
		}
		if g.imagePreview2 != nil {
			g.imagePreview2.SetContent(widget.NewLabel("No folder selected"))
		}
		if g.imagePreview3 != nil {
			g.imagePreview3.SetContent(widget.NewLabel("No folder selected"))
		}
		return
	}

	// Find random images
	images := g.findRandomImages(imgDir)

	// Update preview cards
	previews := []*widget.Card{g.imagePreview1, g.imagePreview2, g.imagePreview3}

	for i, preview := range previews {
		if preview != nil {
			if i < len(images) {
				// Load and display the actual image
				imagePath := images[i]
				imageName := filepath.Base(imagePath)

				// Load the image file
				imageFile, err := os.Open(imagePath)
				if err != nil {
					preview.SetContent(widget.NewLabel(fmt.Sprintf("Error loading\n%s", imageName)))
					continue
				}

				// Read the file content
				fileInfo, err := imageFile.Stat()
				if err != nil {
					imageFile.Close()
					preview.SetContent(widget.NewLabel(fmt.Sprintf("Error loading\n%s", imageName)))
					continue
				}

				imageData := make([]byte, fileInfo.Size())
				_, err = imageFile.Read(imageData)
				imageFile.Close()

				if err != nil {
					preview.SetContent(widget.NewLabel(fmt.Sprintf("Error loading\n%s", imageName)))
					continue
				}

				// Create image resource
				imageResource := fyne.NewStaticResource(imageName, imageData)

				// Create canvas image for proper scaling
				imageWidget := canvas.NewImageFromResource(imageResource)
				imageWidget.FillMode = canvas.ImageFillContain
				imageWidget.SetMinSize(fyne.NewSize(300, 300))

				// Create a centered container
				imageContainer := container.NewCenter(imageWidget)

				// Show the image container
				preview.SetContent(imageContainer)
			} else {
				// No more images available
				preview.SetContent(widget.NewLabel("No image\navailable"))
			}
		}
	}
}

func (g *JaqenGUI) createMainContent() *fyne.Container {
	// Image folder selection (Step 1)
	imgLabel := widget.NewLabel("Image Folder:")
	imgLabel.TextStyle.Bold = true
	g.imgDirEntry = widget.NewEntry()
	g.imgDirEntry.SetPlaceHolder("Select the folder containing your face images...")
	g.imgDirEntry.OnChanged = func(text string) {
		g.detectPathsFromImageFolder(text)
		g.updateImagePreviews(text)
		g.autoSaveConfig()
	}
	imgButton := g.createFileSelector(g.imgDirEntry, "Select Image Folder", "directory")

	// Detected files section (Step 2)
	xmlLabel := widget.NewLabel("Config XML File:")
	xmlLabel.TextStyle.Bold = true
	g.xmlPathEntry = widget.NewEntry()
	g.xmlPathEntry.SetPlaceHolder("Will be auto-detected from image folder...")
	g.xmlPathEntry.OnChanged = func(_ string) { g.autoSaveConfig() }
	xmlButton := g.createFileSelector(g.xmlPathEntry, "Select XML File", "xml")

	rtfLabel := widget.NewLabel("RTF Player File:")
	rtfLabel.TextStyle.Bold = true
	g.rtfPathEntry = widget.NewEntry()
	g.rtfPathEntry.SetPlaceHolder("Will be auto-detected from image folder...")
	g.rtfPathEntry.OnChanged = func(_ string) { g.autoSaveConfig() }
	rtfButton := g.createFileSelector(g.rtfPathEntry, "Select RTF File", "rtf")

	// FM Version (Step 3)
	fmVersionLabel := widget.NewLabel("Football Manager Version:")
	fmVersionLabel.TextStyle.Bold = true
	g.fmVersionSelect = widget.NewSelect([]string{"2024", "2023", "2022", "2021", "2020"}, nil)
	g.fmVersionSelect.SetSelected("2024")
	g.fmVersionSelect.OnChanged = func(_ string) { g.autoSaveConfig() }

	// Status and action
	g.statusLabel = widget.NewLabel("Select an image folder to get started...")
	g.statusLabel.Wrapping = fyne.TextWrapWord

	g.progressBar = widget.NewProgressBar()
	g.progressBar.Hide()

	g.runButton = widget.NewButton("Assign Face Mappings", g.runProcessing)
	g.runButton.Importance = widget.HighImportance

	// Initialize checkboxes if they don't exist (for main window)
	if g.preserveCheck == nil {
		g.preserveCheck = widget.NewCheck("Preserve existing mappings", nil)
		g.preserveCheck.SetChecked(true)
		g.preserveCheck.OnChanged = func(_ bool) { g.autoSaveConfig() }
	}
	if g.allowDuplicateCheck == nil {
		g.allowDuplicateCheck = widget.NewCheck("Allow duplicate mappings", nil)
		g.allowDuplicateCheck.SetChecked(true)
		g.allowDuplicateCheck.OnChanged = func(_ bool) { g.autoSaveConfig() }
	}

	// Create image preview cards with better styling - no titles
	g.imagePreview1 = widget.NewCard("", "", widget.NewLabel("No folder selected"))
	g.imagePreview2 = widget.NewCard("", "", widget.NewLabel("No folder selected"))
	g.imagePreview3 = widget.NewCard("", "", widget.NewLabel("No folder selected"))

	previewSize := fyne.NewSize(350, 350)
	g.imagePreview1.Resize(previewSize)
	g.imagePreview2.Resize(previewSize)
	g.imagePreview3.Resize(previewSize)

	// Create a more organized layout with better spacing
	setupCard := widget.NewCard("Configuration", "", container.NewVBox(
		container.NewBorder(nil, nil, imgLabel, imgButton, g.imgDirEntry),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, xmlLabel, xmlButton, g.xmlPathEntry),
		container.NewBorder(nil, nil, rtfLabel, rtfButton, g.rtfPathEntry),
		container.NewBorder(nil, nil, fmVersionLabel, nil, g.fmVersionSelect),
	))

	// Create image preview section with better layout
	imagePreviewCard := widget.NewCard("Image Preview", "Sample images from your selected folder",
		container.NewHBox(
			g.imagePreview1,
			widget.NewSeparator(),
			g.imagePreview2,
			widget.NewSeparator(),
			g.imagePreview3,
		))

	// Create action section with better styling
	actionCard := widget.NewCard("Actions", "", container.NewVBox(
		g.statusLabel,
		g.progressBar,
		container.NewCenter(g.runButton),
	))

	return container.NewVBox(
		setupCard,
		widget.NewSeparator(),
		imagePreviewCard,
		widget.NewSeparator(),
		actionCard,
	)
}

func (g *JaqenGUI) createMappingOverrideSection() *fyne.Container {
	// Create a simple list with better layout
	g.mappingOverrideList = widget.NewList(
		func() int {
			return len(g.mappingOverrides)
		},
		func() fyne.CanvasObject {
			// Create a horizontal container with proper spacing
			return container.NewBorder(
				nil, nil,
				widget.NewLabel(""),        // Country code (left)
				widget.NewButton("✕", nil), // Delete button (right)
				widget.NewLabel(""),        // Ethnic group (center)
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			container := obj.(*fyne.Container)

			// Get sorted list of countries for consistent ordering
			countries := make([]string, 0, len(g.mappingOverrides))
			for country := range g.mappingOverrides {
				countries = append(countries, country)
			}
			sort.Strings(countries) // Sort for consistent display

			if id < len(countries) {
				country := countries[id]
				ethnic := g.mappingOverrides[country]

				// Set country code label (left)
				countryLabel := container.Objects[0].(*widget.Label)
				countryLabel.SetText(country)
				countryLabel.TextStyle.Bold = true

				// Set ethnic group label (center)
				ethnicLabel := container.Objects[1].(*widget.Label)
				ethnicLabel.SetText(ethnic)

				// Set up delete button (right)
				deleteBtn := container.Objects[2].(*widget.Button)
				deleteBtn.OnTapped = func() {
					delete(g.mappingOverrides, country)
					g.updateMappingOverrideList()
					g.autoSaveConfig()
				}
			}
		},
	)

	// Set a reasonable size for the list
	g.mappingOverrideList.Resize(fyne.NewSize(550, 300))

	addButton := widget.NewButton("Add Mapping Override", g.addMappingOverride)

	return container.NewVBox(
		widget.NewCard("Mapping Overrides", "", container.NewVBox(
			widget.NewLabel("Country Code → Ethnic Group"),
			widget.NewLabel("Add custom mappings for countries not recognized by default.\nAvailable ethnicities: African, Asian, Caucasian, Central European, EECA, Italmed, MENA, MESA, SAMed, Scandinavian, Seasian, South American, SpanMed, YugoGreek"),
			g.mappingOverrideList,
			addButton,
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

	// Get all available ethnicities from the mapper package
	availableEthnicities := []string{
		"African", "Asian", "Caucasian", "Central European", "EECA",
		"Italmed", "MENA", "MESA", "SAMed", "Scandinavian",
		"Seasian", "South American", "SpanMed", "YugoGreek",
	}

	ethnicSelect := widget.NewSelect(availableEthnicities, nil)
	ethnicSelect.SetSelected(ethnic)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Country Code", Widget: countryEntry},
			{Text: "Ethnic Group", Widget: ethnicSelect},
		},
	}

	// Use settings window if it exists, otherwise main window
	targetWindow := g.window
	if g.settingsWindow != nil {
		targetWindow = g.settingsWindow
	}

	dialog.ShowForm("Edit Mapping Override", "Save", "Cancel", form.Items, func(confirmed bool) {
		if confirmed && countryEntry.Text != "" && ethnicSelect.Selected != "" {
			// Ensure mappingOverrides map is initialized
			if g.mappingOverrides == nil {
				g.mappingOverrides = make(map[string]string)
			}
			g.mappingOverrides[countryEntry.Text] = ethnicSelect.Selected
			g.updateMappingOverrideList()
			g.autoSaveConfig() // Save the new mapping
		} else if confirmed && (countryEntry.Text == "" || ethnicSelect.Selected == "") {
			// Show error if user tries to save empty values
			dialog.ShowError(fmt.Errorf("country code and ethnic group are required"), targetWindow)
		}
	}, targetWindow)
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
		dialog.ShowError(fmt.Errorf("image directory path is required"), g.window)
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
	configPath := internal.DefaultConfigPath

	if _, err := os.Stat(configPath); err == nil {
		_, err := internal.ReadConfig(configPath)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("error reading config: %w", err), g.window)
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
		g.statusLabel.SetText("Loading image pool...")
	})

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
		g.statusLabel.SetText("Processing players...")
	})

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
		g.statusLabel.SetText("Assigning faces...")
	})

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

	fyne.Do(func() {
		g.progressBar.SetValue(1.0)
		g.statusLabel.SetText("Processing completed successfully!")
		dialog.ShowInformation("Success", "Face mapping completed successfully!", g.window)
	})
}

func (g *JaqenGUI) ShowAndRun() {
	// Initialize config
	g.loadConfig()

	// Create main layout
	content := container.NewVBox(
		g.createHeaderBar(),
		widget.NewSeparator(),
		g.createMainContent(),
	)

	// Apply config to GUI after widgets are created
	g.applyConfigToGUI()

	// Add some padding
	paddedContent := container.NewPadded(content)

	g.window.SetContent(paddedContent)

	// Close settings window when main window closes
	g.window.SetCloseIntercept(func() {
		if g.settingsWindow != nil {
			g.settingsWindow.Close()
		}
		g.window.Close()
	})

	g.window.ShowAndRun()
}

func runGUI(cmd *cobra.Command, args []string) {
	gui := NewJaqenGUI()
	gui.ShowAndRun()
}

var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Launch Jaqen NewGen Tool",
	Long:  "Launch Jaqen NewGen Tool - Football Manager Face Manager",
	Run:   runGUI,
}
