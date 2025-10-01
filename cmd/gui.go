package cmd

import (
	"jaqen/gui"

	"github.com/spf13/cobra"
)

// runGUI launches the GUI application
func runGUI(cmd *cobra.Command, args []string) {
	app := gui.New()
	app.ShowAndRun()
}

// guiCmd represents the GUI command
var guiCmd = &cobra.Command{
	Use:   "gui",
	Short: "Launch Jaqen NewGen Tool",
	Long:  "Launch Jaqen NewGen Tool - Football Manager Face Manager",
	Run:   runGUI,
}
