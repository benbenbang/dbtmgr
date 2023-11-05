package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"dbtmgr/cmd/lock"
	"dbtmgr/cmd/manifest"
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

	groups := templates.CommandGroups{
		{
			Message:  "Lock & Manifest Management Commands",
			Commands: []*cobra.Command{lock.LockCmd, manifest.ManifestCmd},
		},
		{
			Message:  "Lock Management Subommands",
			Commands: lockCmds,
		},
		{
			Message:  "Manifest Management Subcommands",
			Commands: manifestCmds,
		},
		{
			Message:  "Settings Commands",
			Commands: mngCmds,
		},
	}

	templates.ActsAsRootCommand(rootCmd, []string{"options"}, groups...)
}

var DefaultCmd = rootCmd

var rootCmd = &cobra.Command{
	Use:   "dbtmgr",
	Short: "DBT state management and synchronization tool",
	Long: `dbtmgr is a command-line utility designed to manage, synchronize, and
lock the state files for DBT (Data Build Tool) manifests. It facilitates
development workflows by ensuring consistent state across multiple environments
and preventing concurrent operations that could lead to conflicts.

With dbtmgr, developers or CI can acquire and release locks on the DBT state file
residing within an S3 bucket, pull the latest state for local comparison, and
push updates to the remote state safely. It is built to handle the state as a
source of truth for all schema changes and to help DBT in identifying and running
tests on modified columns.

The tool uses AWS services to manage state files and employs an S3-based locking
mechanism to prevent concurrent updates, ensuring a smooth and error-free
release process.

For example, to refresh your local state, run:

  dbtmgr refresh

To acquire a lock before making changes, use:

  dbtmgr lock acquire

dbtmgr integrates with CI/CD pipelines, providing a seamless interface for
managing DBT states within team development practices.`,
}
