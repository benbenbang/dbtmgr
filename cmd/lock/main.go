package lock

import (
	"dbtmgr/internal/logging"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	LockCmd.AddCommand(
		AcquireCmd,
		ReleaseCmd,
		RefreshCmd,
		SyncCmd,
	)

	var verbose bool

	LockCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "set verbose output")
}

var log = logging.GetLogger()

var LockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Manage locks on the S3 bucket state file",
	Long: `The lock command group is used for managing locks on the S3 bucket state file.

With these commands, you can acquire and release locks to coordinate safe access
to the state file among multiple developers or automation tools. Use refresh to
update your local state from S3, and sync to update the remote state in S3 with
your local changes after acquiring a lock.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			log.SetLevel(logrus.DebugLevel)
		}
		log.Debugf("Running lock command group")
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			log.Errorf("Error displaying help for lock command group: %v", err)
		}
	},
}
