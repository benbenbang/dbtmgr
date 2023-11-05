package lock

import (
	"context"
	"dbtmgr/internal/aws/lock"
	"dbtmgr/internal/config"
	"dbtmgr/internal/subproc"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	bucket   = config.DBT_STATE_BUCKET
	key      = config.DBT_LOCK_KEY
	lockInfo string
)

var AcquireCmd = &cobra.Command{
	Use:   "acquire",
	Short: "Acquire a lock on the S3 bucket",
	Long: `Acquire a lock on the S3 bucket to prevent concurrent state modifications.
This command attempts to create a lock file in the specified S3 bucket, which
signals to other users and processes that the state file is currently being
modified. If the lock is already present, the command will fail and indicate
that the state file is in use.

Usage:
  dbtmgr lock acquire

Example:
  # Acquire a lock on the S3 state file
  dbtmgr lock acquire`,
	PreRun: func(cmd *cobra.Command, args []string) {
		log.Debug("Running lock acquire command")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		lockInfo := lock.LockInfo{
			LockID:    "dbtmgr",
			TimeStamp: time.Now().Format(time.RFC3339),
		}

		if err := lock.AcquireStateLock(ctx, bucket, key, lockInfo); err != nil {
			log.Errorf("Failed to acquire lock on S3 state file: %v", err)
			os.Exit(1)
		}

		cmd.Println("dbtmgr lock acquired")
	},
}

var ReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release the lock on the S3 bucket",
	Long: `Release the lock on the S3 bucket to allow other modifications.
This command removes the lock file from the S3 bucket, indicating that
the state file is no longer being modified and is available for other
users and processes to modify.

Usage:
  dbtmgr lock release

Example:
  # Release the lock on the S3 state file
  dbtmgr lock release`,
	PreRun: func(cmd *cobra.Command, args []string) {
		log.Debug("Running lock release command")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		if err := lock.ReleaseStateLock(ctx, bucket, key); err != nil {
			log.Errorf("Failed to release lock on S3 state file: %v", err)
			os.Exit(1)
		}
		fmt.Println("Lock released successfully.")
	},
}

var RefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Update local manifest from the S3 bucket",
	Long: `Fetch the latest state of the DBT manifest from the S3 bucket
and update the local manifest accordingly. This ensures that the local
workspace is in sync with the shared state and reflects any changes
that have been made by others.

Usage:
  dbtmgr refresh

Example:
  # Refresh local state to match the S3 bucket state
  dbtmgr refresh`,
	PreRun: func(cmd *cobra.Command, args []string) {
		log.Debug("Running lock refresh command")
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Check if the local and remote SHAs are different
		localSHA, remoteSHA, err := subproc.CompareSHAs()
		if err != nil {
			log.Errorf("failed to compare SHAs: %v", err)
			os.Exit(1)
		}

		if localSHA != remoteSHA {
			// SHAs are different, pull remote state
			if err := lock.RefreshState(lockInfo, bucket, key); err != nil {
				log.Errorf("failed to refresh state: %v", err)
				os.Exit(1)
			}
			fmt.Println("State refreshed successfully. Local state is up-to-date with remote state.")
		} else {
			fmt.Println("Local state is already up-to-date with remote state.")
		}
	},
}

var SyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync local manifest to the S3 bucket",
	Long: `Push the local state of the DBT manifest to the S3 bucket
to synchronize it with the central state. This should typically be done
after acquiring a lock and making changes to ensure that the shared state
is updated correctly.

Usage:
  dbtmgr sync

Example:
  # Sync local state to the S3 bucket
  dbtmgr sync`,
	PreRun: func(cmd *cobra.Command, args []string) {
		log.Debug("Running lock sync command")
	},
	Run: func(cmd *cobra.Command, args []string) {
		localSHA, err := subproc.FetchLocalSHA()
		if err != nil {
			log.Errorf("failed to fetch local git SHA: %v", err)
			os.Exit(1)
		}

		// Sync the local state to the remote state
		if err := lock.SyncState(localSHA, bucket, key); err != nil {
			log.Errorf("failed to sync state: %v", err)
			os.Exit(1)
		}
		fmt.Println("State synchronized successfully. Remote state is updated with local changes.")
	},
}
