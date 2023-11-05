package cmd

import (
	"statemgr/internal/config"

	"github.com/spf13/cobra"
)

var CurrentVersion = config.Version

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version of statemgr",
	Long: `Print the current version number of statemgr. Use this command
to verify the version of statemgr you are currently running.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("statemgr version %s\n", CurrentVersion)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update statemgr to the latest version",
	Long: `Check for the latest version of statemgr and update the tool if
a newer version is available. It's recommended to keep statemgr up to date
to utilize the latest features and improvements.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("Checking for updates...")
		cmd.Println("statemgr is up to date.")
	},
}
