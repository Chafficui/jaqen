package gui

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	nativeDialog "github.com/sqweek/dialog"
)

// createHeaderBar creates the application header with title and action buttons
func (g *JaqenGUI) createHeaderBar() *fyne.Container {
	title := widget.NewLabel("Jaqen NewGen - Football Manager Face Manager")
	title.TextStyle.Bold = true

	// Add help button
	helpBtn := widget.NewButton("❓ Help", func() {
		g.showRTFInstructionsPopup()
	})

	// Add bug report button
	bugReportBtn := widget.NewButton("🐛 Report Bug", func() {
		g.showBugReportDialog()
	})

	// Create button container
	buttonContainer := container.NewHBox(helpBtn, bugReportBtn)

	return container.NewBorder(nil, nil, title, buttonContainer, widget.NewSeparator())
}

// createMainContent creates the main content area with all sections
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

	// Initialize settings widgets
	g.preserveCheck = widget.NewCheck("Preserve existing mappings", nil)
	g.preserveCheck.SetChecked(true)
	g.preserveCheck.OnChanged = func(_ bool) { g.autoSaveConfig() }

	g.allowDuplicateCheck = widget.NewCheck("Allow duplicate mappings", nil)
	g.allowDuplicateCheck.SetChecked(true)
	g.allowDuplicateCheck.OnChanged = func(_ bool) { g.autoSaveConfig() }

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

	// Create settings section
	settingsCard := widget.NewCard("Settings", "", container.NewVBox(
		g.preserveCheck,
		g.allowDuplicateCheck,
		widget.NewSeparator(),
		g.createMappingOverrideSection(),
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
		settingsCard,
		widget.NewSeparator(),
		actionCard,
	)
}

// createMappingOverrideSection creates the mapping override UI section
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

// updateMappingOverrideList refreshes the mapping override list
func (g *JaqenGUI) updateMappingOverrideList() {
	if g.mappingOverrideList != nil {
		g.mappingOverrideList.Refresh()
	}
}

// addMappingOverride shows the dialog to add a new mapping override
func (g *JaqenGUI) addMappingOverride() {
	g.editMappingOverride("", "")
}

// editMappingOverride shows the dialog to edit a mapping override
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

	targetWindow := g.window

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

// createFileSelector creates a file/folder selector button
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

// findRandomImages finds random images in the specified directory
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

// updateImagePreviews updates the image preview cards
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
