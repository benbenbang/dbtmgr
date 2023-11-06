package manifest

import (
	"context"

	"os"
	"statectl/internal/aws/manifest"
	"statectl/internal/aws/utils"
	"statectl/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	bucket       string
	manifestPath string
	statePath    string
	localPath    string
	evidencePath string
)

func init() {
	PushCmd.Flags().StringVarP(&bucket, "bucket", "b", viper.GetString("BUCKET_NAME"), "S3 bucket to store the manifest")
	PushCmd.Flags().StringVarP(&manifestPath, "manifest", "m", viper.GetString("MANIFEST_KEY_PATH"), "S3 key point to the bucket key that stores store the manifest")
	PushCmd.Flags().StringVarP(&statePath, "state", "s", viper.GetString("STATE_KEY_PATH"), "S3 key point to a specific file in the bucket, e.g. artifacts/manifest.json, which will be tracked in the local state.json file")
	PushCmd.Flags().StringVarP(&evidencePath, "evidence", "e", "evidence.json", "Local path to store the envidence file which is for tracking the manifest")

	PullCmd.Flags().StringVarP(&bucket, "bucket", "b", viper.GetString("BUCKET_NAME"), "S3 bucket point to the bucket name that stores the manifest")
	PullCmd.Flags().StringVarP(&manifestPath, "manifest", "m", viper.GetString("MANIFEST_KEY_PATH"), "S3 key point to the bucket key that stores store the manifest")
	PullCmd.Flags().StringVarP(&localPath, "local-path", "l", "", "Local path to store the manifest")

	ListCmd.Flags().StringVarP(&bucket, "bucket", "b", viper.GetString("BUCKET_NAME"), "S3 bucket point to the bucket name that stores the manifest")
	ListCmd.Flags().StringVarP(&manifestPath, "manifest", "m", viper.GetString("MANIFEST_KEY_PATH"), "S3 key point to the bucket key that stores store the manifest")
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

		bucket, manifestPath, err := utils.GetS3BucketAndManifest(cmd)
		if err != nil {
			cmd.PrintErrln(config.Red("❌ Failed to get S3 bucket/key: ", err))
			os.Exit(1)
		}
		log.Debug("S3 bucket/key: ", bucket, manifestPath)

		if err := manifest.UploadManifest(ctx, bucket, manifestPath); err != nil {
			cmd.PrintErrln(config.Red("❌ Failed to upload the manifest to S3 bucket: ", err))
			os.Exit(1)
		}

		if statePath := cmd.Flag("state").Value.String(); statePath != "" {
			evidencePath := cmd.Flag("evidence").Value.String()
			log.Debugf("S3 bucket/key: %s/%s. Local evidence path: %s\n", bucket, statePath, evidencePath)
			if err := manifest.CreateStateJSON(context.Background(), bucket, statePath, evidencePath); err != nil {
				cmd.PrintErrln(config.Red("❌ Failed to create the evidence json file: ", err))
				os.Exit(1)
			}
		}

		cmd.Println(config.Green("manifest has been successfully uploaded"))
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

		bucket, key, err := utils.GetS3BucketAndKey(cmd)
		if err != nil {
			cmd.PrintErrln(config.Red("❌ Failed to get S3 bucket/key: ", err))
			os.Exit(1)
		}
		log.Debug("S3 bucket/key: ", bucket, key)

		if err := manifest.DownloadManifest(ctx, bucket, key, localPath); err != nil {
			cmd.PrintErrln(config.Red("❌ Failed to download the manifest from S3 bucket: ", err))
			os.Exit(1)
		}

		cmd.Println(config.Green("manifest has been successfully downloaded"))
	},
}

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List manifest in the S3 bucket",
	Long: `List manifest in the S3 bucket to show all the files in the given key.
This command lists all the manifest files in the specified S3 bucket & key.
`,
	PreRun: func(cmd *cobra.Command, args []string) {
		log.Debug("Running manifest list command")
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		bucket, key, err := utils.GetS3BucketAndKey(cmd)
		if err != nil {
			cmd.PrintErrln(config.Red("❌ Failed to get S3 bucket/key: ", err))
			os.Exit(1)
		}
		log.Debug("S3 bucket/key: ", bucket, key)

		info, err := manifest.ListManifests(ctx, bucket, key)

		if err != nil {
			cmd.PrintErrln(config.Red("❌ Failed to list the manifest from S3 bucket: ", err))
			os.Exit(1)
		}

		utils.PrintTree(info, "")
	},
}
