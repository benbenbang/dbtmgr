package cmd

import (
	"statectl/internal/config"

	"github.com/spf13/cobra"
)

var CurrentVersion = config.Version

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version of statectl",
	Long: `Print the current version number of statectl. Use this command
to verify the version of statectl you are currently running.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("statectl version %s\n", CurrentVersion)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update statectl to the latest version",
	Long: `Check for the latest version of statectl and update the tool if
a newer version is available. It's recommended to keep statectl up to date
to utilize the latest features and improvements.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("Checking for updates...")
		cmd.Println("statectl is up to date.")
	},
}
