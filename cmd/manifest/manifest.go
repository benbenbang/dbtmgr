package manifest

import (
	"context"

	"os"
	"statectl/internal/aws/manifest"
	"statectl/internal/aws/utils"
	"statectl/internal/config"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	bucket       string
	manifestPath string
	statePath    string
	localPath    string
	singleStore  bool
)

func init() {
	PushCmd.Flags().StringVarP(&bucket, "bucket", "b", viper.GetString("BUCKET_NAME"), "S3 bucket to store the manifest")
	PushCmd.Flags().StringVarP(&manifestPath, "manifest", "m", viper.GetString("MANIFEST_KEY_PATH"), "S3 key point to the bucket key that stores store the manifest")
	PushCmd.Flags().StringVarP(&statePath, "state", "s", "state.json", "Local path to store the state file which is for tracking the manifest")
	PushCmd.PersistentFlags().BoolVar(&singleStore, "disable-full-tree", false, "push from the root directory. e.g. manifestPath=artifacts/manifest.json, then push entire artifacts folder")

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
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			log.SetLevel(logrus.DebugLevel)
		}
		log.Debug("Running manifest upload command")
	},
	Run: func(cmd *cobra.Command, args []string) {
		cli := utils.GetS3Client()

		if disableFT, _ := cmd.Flags().GetBool("disable-full-tree"); disableFT {
			singleStore = true
		}

		bucket, manifestPath, err := utils.GetS3BucketAndManifest(cmd)
		if err != nil {
			cmd.PrintErrln(config.Red("❌ Failed to get S3 bucket/key: ", err))
			os.Exit(1)
		}
		log.Debug("S3 bucket/key: ", bucket, manifestPath)

		log.Debugf("storing single file: %t\n", singleStore)
		if err := manifest.UploadManifest(context.Background(), cli, bucket, manifestPath, singleStore); err != nil {
			cmd.PrintErrln(config.Red("❌ Failed to upload the manifest to S3 bucket: ", err))
			os.Exit(1)
		}

		if statePath := cmd.Flag("state").Value.String(); statePath != "" {
			log.Debugf("S3 bucket/key: %s/%s. Local evidence path: %s\n", bucket, manifestPath, statePath)
			if err := manifest.CreateStateJSON(context.Background(), cli, bucket, manifestPath, statePath); err != nil {
				cmd.PrintErrln(config.Red("❌ Failed to create the state json file: ", err))
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
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			log.SetLevel(logrus.DebugLevel)
		}
		log.Debug("Running manifest pull command")
	},
	Run: func(cmd *cobra.Command, args []string) {
		cli := utils.GetS3Client()

		bucket, key, err := utils.GetS3BucketAndKey(cmd)
		if err != nil {
			cmd.PrintErrln(config.Red("❌ Failed to get S3 bucket/key: ", err))
			os.Exit(1)
		}
		log.Debug("S3 bucket/key: ", bucket, key)

		if err := manifest.DownloadManifest(context.Background(), cli, bucket, key, localPath); err != nil {
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
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			log.SetLevel(logrus.DebugLevel)
		}
		log.Debug("Running manifest list command")
	},
	Run: func(cmd *cobra.Command, args []string) {
		cli := utils.GetS3Client()

		bucket, key, err := utils.GetS3BucketAndManifest(cmd)
		if err != nil {
			cmd.PrintErrln(config.Red("❌ Failed to get S3 bucket/key: ", err))
			os.Exit(1)
		}
		log.Debug("S3 bucket/key: ", bucket, key)

		info, err := manifest.ListManifests(context.Background(), cli, bucket, key)

		if err != nil {
			cmd.PrintErrln(config.Red("❌ Failed to list the manifest from S3 bucket: ", err))
			os.Exit(1)
		}

		utils.PrintTree(info, "")
	},
}
