package cmd

import (
	"github.com/spf13/cobra"

	"statemgr/cmd/lock"
	"statemgr/cmd/manifest"
	"statemgr/pkg/template"
)

func init() {
	var verbose bool

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "set verbose output")

	rootCmd.AddCommand(
		lock.LockCmd,
		versionCmd,
		updateCmd,
		completionCmd,
	)

	lockCmds := []*cobra.Command{lock.AcquireCmd, lock.ReleaseCmd, lock.ForceReleaseCmd}
	manifestCmds := []*cobra.Command{manifest.UploadCmd, manifest.SyncCmd}
	mngCmds := []*cobra.Command{versionCmd, updateCmd, completionCmd}

	cmdGroup := template.CreatCmdGroup(
		template.CmdTemplate{
			Title:    "Lock & Management Commands",
			Commands: []*cobra.Command{lock.LockCmd, manifest.ManifestCmd},
		},
		template.CmdTemplate{
			Title:    "Lock Managment Subcommands",
			Commands: lockCmds,
		},
		template.CmdTemplate{
			Title:    "Manifest Managment Subcommands",
			Commands: manifestCmds,
		},
		template.CmdTemplate{
			Title:    "Management Commands",
			Commands: mngCmds,
		},
	)

	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		template.HelpFunc(cmd, cmdGroup)
	})

}

var DefaultCmd = rootCmd

// var DefaultCmd = rootCmd

var rootCmd = &cobra.Command{
	Use:   "statemgr",
	Short: "DBT state management and synchronization tool",
	Long: `statemgr is a command-line utility designed to manage, synchronize, and
lock the state files for DBT (Data Build Tool) manifests. It facilitates
development workflows by ensuring consistent state across multiple environments
and preventing concurrent operations that could lead to conflicts.

With statemgr, developers or CI can acquire and release locks on the DBT state file
residing within an S3 bucket, pull the latest state for local comparison, and
push updates to the remote state safely. It is built to handle the state as a
source of truth for all schema changes and to help DBT in identifying and running
tests on modified columns.

The tool uses AWS services to manage state files and employs an S3-based locking
mechanism to prevent concurrent updates, ensuring a smooth and error-free
release process.

For example, to refresh your local state, run:

  statemgr refresh

To acquire a lock before making changes, use:

  statemgr lock acquire

statemgr integrates with CI/CD pipelines, providing a seamless interface for
managing DBT states within team development practices.`,
}
