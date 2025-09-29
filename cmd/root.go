package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

func runDefault(cmd *cobra.Command, args []string) {
	// Launch GUI
	runGUI(cmd, args)
}

var rootCmd = &cobra.Command{
	Use:   "jaqen-newgen-tool",
	Short: "Jaqen NewGen Tool - Football Manager Face Manager",
	Long:  `Jaqen NewGen Tool creates and manages image file mappings to face profiles in Football Manager.`,
	Run:   runDefault,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalln(err)
	}
}

func init() {
	// Add GUI command
	rootCmd.AddCommand(guiCmd)
}
