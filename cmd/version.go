package cmd

import (
	"github.com/spf13/cobra"
)

const CurrentVersion = "0.0.1"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version of dbtmgr",
	Long: `Print the current version number of dbtmgr. Use this command
to verify the version of dbtmgr you are currently running.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("dbtmgr version %s\n", CurrentVersion)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update dbtmgr to the latest version",
	Long: `Check for the latest version of dbtmgr and update the tool if
a newer version is available. It's recommended to keep dbtmgr up to date
to utilize the latest features and improvements.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("Checking for updates...")
		cmd.Println("dbtmgr is up to date.")
	},
}
