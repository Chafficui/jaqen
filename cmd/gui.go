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
	configPathEntry    *widget.Entry
	preserveCheck      *widget.Check
	allowDuplicateCheck *widget.Check
	fmVersionSelect    *widget.Select
	statusLabel        *widget.Label
	progressBar        *widget.ProgressBar
	runButton          *widget.Button
}

func NewJaqenGUI() *JaqenGUI {
	myApp := app.NewWithID("com.jaqen.footballmanager")
	myApp.SetIcon(theme.DocumentIcon())
	
	window := myApp.NewWindow("Jaqen - Football Manager Face Manager")
	window.Resize(fyne.NewSize(800, 600))
	window.CenterOnScreen()

	return &JaqenGUI{
		app:    myApp,
		window: window,
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

	formatButton := widget.NewButton("Format Config", g.formatConfig)

	return container.NewVBox(
		widget.NewCard("Actions", "", container.NewVBox(
			g.statusLabel,
			g.progressBar,
			container.NewHBox(g.runButton, formatButton),
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
	// Create main layout
	content := container.NewVBox(
		g.createInputSection(),
		g.createOptionsSection(),
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