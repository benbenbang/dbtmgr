package lock

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"statemgr/internal/aws/lock"
	"statemgr/internal/config"
	"statemgr/internal/subproc"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var (
	bucket = config.DBT_STATE_BUCKET
	key    = config.DBT_LOCK_KEY
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
  statemgr lock acquire

Example:
  # Acquire a lock on the S3 state file
  statemgr lock acquire`,
	PreRun: func(cmd *cobra.Command, args []string) {
		log.Debug("Running lock acquire command")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		extra_comment := ""
		var err error

		commit_sha := os.Getenv("CI_COMMIT_SHA")
		cs_comment := "ok"
		if commit_sha == "" {
			commit_sha, err = subproc.FetchLocalSHA()
			if err != nil {
				commit_sha = uuid.New().String()
			}
			cs_comment = "No commit SHA available, using random UUID"
		}

		trigger_iid := os.Getenv("CI_PIPELINE_IID")
		ti_comment := "ok"
		if trigger_iid == "" {
			trigger_iid = uuid.New().String()
			ti_comment = "No pipeline ID available, using random UUID"
		}

		if cs_comment != "ok" || ti_comment != "ok" {
			extra_comment = "WARNING: one or more environment variables were not found. Use timestamp as reference to check the exact commit and pipeline ID."
		}

		lockInfo := lock.LockInfo{
			LockID:    commit_sha,
			TimeStamp: time.Now().Format(time.RFC3339),
			Signer:    trigger_iid,
			Comments: lock.Comments{
				Commit:  cs_comment,
				Trigger: ti_comment,
				Extra:   extra_comment,
			},
		}

		if err := lock.AcquireStateLock(ctx, bucket, key, lockInfo); err != nil {
			if errors.Is(err, lock.LockExists) {
				cmd.Println(config.Yellow("Lock already acquired, exiting..."))
				os.Exit(0)
			}
			cmd.PrintErrf(config.Red("Failed to acquire lock: %v\n", err))
			os.Exit(1)
		}

		cmd.Println(config.Green("Lock acquired successfully."))
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
  statemgr lock release

Example:
  # Release the lock on the S3 state file
  statemgr lock release`,
	PreRun: func(cmd *cobra.Command, args []string) {
		log.Debug("Running lock release command")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		if err := lock.ReleaseStateLock(ctx, bucket, key); err != nil {
			cmd.PrintErrf("Failed to release lock: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Lock released successfully.")
	},
}

var ForceReleaseCmd = &cobra.Command{
	Use:   "force-release",
	Short: "Force release the S3 lock with confirmation",
	Long: `Forcefully releases the lock on the S3 state file after user confirmation.
This command should be used with caution as it can disrupt ongoing operations.
It synchronizes the local state with the latest state from the S3 bucket.

Usage:
  statemgr lock force-release

Example:
  # Prompt for confirmation and then force release the S3 lock
  statemgr lock force-release`,
	PreRun: func(cmd *cobra.Command, args []string) {
		log.Debug("Preparing to prompt for lock force-release")
	},
	Run: func(cmd *cobra.Command, args []string) {
		if exist, _, err := lock.CheckStateLock(context.Background(), bucket, key, false); err != nil {
			cmd.PrintErrf(config.Red("Failed to check lock status: %v\n", err))
			os.Exit(1)
		} else if !exist {
			cmd.PrintErrf(config.Red("Lock does not exist. Nothing to release.\n"))
			os.Exit(1)
		}

		ctx := context.Background()
		reader := bufio.NewReader(os.Stdin)

		cmd.Println(config.Yellow("WARNING: You are about to forcefully remove the remote lock file. This may disrupt ongoing operations."))
		cmd.Print("Are you sure you want to proceed? (type 'yes' to confirm): ")

		confirmation, _ := reader.ReadString('\n')
		if strings.TrimSpace(confirmation) != "yes" {
			fmt.Println("Force release cancelled.")
			return
		}

		// User confirmed, proceed with force release
		err := lock.ForceReleaseLock(ctx, bucket, key)
		if err != nil {
			cmd.PrintErrf(config.Red("Failed to forcefully release lock on S3 state file: %v\n", err))
			os.Exit(1)
		}
		fmt.Println(config.Green("Lock forcefully released successfully."))
	},
}
