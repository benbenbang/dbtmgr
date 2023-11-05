package manifest

import (
	"context"
	"dbtmgr/internal/aws"
	"os"

	"github.com/spf13/cobra"
)

var (
	bucket       string
	key          string
	manifestInfo string
	localPath    string
)

var UploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a manifest to the S3 bucket",
	Long: `Upload a manifest to the S3 bucket to track the state of the database schema.
This command creates a new manifest file in the specified S3 bucket, which
tracks the state of the database schema. This manifest file is used to
coordinate safe access to the state file among multiple developers or
automation tools.
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		log.Debug("Running manifest upload command")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		if err := aws.UploadManifest(ctx, bucket, key, manifestInfo); err != nil {
			log.Errorf("Failed to upload manifest to S3 bucket: %v", err)
			os.Exit(1)
		}

		cmd.Println("dbtmgr manifest uploaded")
	},
}

var SyncCmd = &cobra.Command{
	Use:   "download",
	Short: "download the local state file with the S3 bucket",
	Long: `download the local state file with the S3 bucket to update the remote state.
This command updates the state file in the specified S3 bucket with the
contents of the local state file. This command should be used after acquiring
a lock on the S3 bucket to ensure that the state file is not modified by
another user or process.
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		log.Debug("Running manifest download command")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		if err := aws.DownloadManifest(ctx, bucket, key, localPath); err != nil {
			log.Errorf("Failed to sync state file with S3 bucket: %v", err)
			os.Exit(1)
		}

		cmd.Println("dbtmgr manifest synced")
	},
}
