package manifest

import (
	"statectl/internal/logging"

	"github.com/spf13/cobra"
)

func init() {
	ManifestCmd.AddCommand(PushCmd)
	ManifestCmd.AddCommand(PullCmd)
}

var log = logging.GetLogger()

var ManifestCmd = &cobra.Command{
	Use:   "manifest",
	Short: "Manage manifests in the S3 bucket",
	Long: `The manifest command group is used for managing manifests in the S3 bucket.
Manifests are used to track the state of the database schema and are used to
coordinate safe access to the state file among multiple developers or automation
tools. Use upload to create a new manifest file in the S3 bucket, and list to
view all manifests in the S3 bucket.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		log.Debugf("Running manifest command group")
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			log.Errorf("Error displaying help for manifest command group: %v", err)
		}
	},
}
