package manifest

import (
	"context"

	"os"
	"statectl/internal/aws/manifest"
	"statectl/internal/config"

	"github.com/spf13/cobra"
)

var (
	bucket       = config.DBT_STATE_BUCKET
	key          = config.DBT_STATE_KEY
	manifestPath string
	localPath    string
)

func init() {
	PushCmd.Flags().StringVarP(&bucket, "bucket", "b", bucket, "S3 bucket to store the manifest")
	PushCmd.Flags().StringVarP(&key, "key", "k", key, "S3 key to store the manifest")
	PushCmd.Flags().StringVarP(&manifestPath, "manifest", "m", "", "Manifest file to upload")
	PushCmd.Flags().StringVarP(&localPath, "local-path", "l", "", "Local path to store the manifest")

	PullCmd.Flags().StringVarP(&bucket, "bucket", "b", bucket, "S3 bucket point to the bucket name that stores the manifest")
	PullCmd.Flags().StringVarP(&key, "key", "k", key, "S3 key point to the bucket key that stores store the manifest")
	PullCmd.Flags().StringVarP(&localPath, "local-path", "l", "", "Local path to store the manifest")
}

var PushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push a manifest to the S3 bucket",
	Long: `Push a manifest to the S3 bucket to track the state of the database schema.
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

		if err := manifest.UploadManifest(ctx, bucket, key, manifestPath); err != nil {
			log.Errorf("Failed to push manifest to S3 bucket: %v", err)
			os.Exit(1)
		}

		cmd.Println("statectl manifest uploaded")
	},
}

var PullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull the local state file with the S3 bucket",
	Long: `Pull the local state file with the S3 bucket to update the remote state.
This command updates the state file in the specified S3 bucket with the
contents of the local state file. This command should be used after acquiring
a lock on the S3 bucket to ensure that the state file is not modified by
another user or process.
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		log.Debug("Running manifest pull command")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		if err := manifest.DownloadManifest(ctx, bucket, key, localPath); err != nil {
			log.Errorf("Failed to sync state file with S3 bucket: %v", err)
			os.Exit(1)
		}

		cmd.Println("statectl manifest downloaded")
	},
}
