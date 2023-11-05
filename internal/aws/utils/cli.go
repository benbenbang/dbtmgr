package utils

import (
	"fmt"
	"statectl/internal/config"

	"github.com/spf13/cobra"
)

func GetS3BucketAndKey(cmd *cobra.Command) (string, string, error) {
	bucket := cmd.Flag("bucket").Value.String()
	if bucket == "" {
		return "", "", fmt.Errorf(config.Red("bucket is required\n"))
	}
	key := cmd.Flag("key").Value.String()
	if key == "" {
		return "", "", fmt.Errorf(config.Red("key is required\n"))
	}
	return bucket, key, nil
}

func GetS3BucketAndManifest(cmd *cobra.Command) (string, string, error) {
	bucket := cmd.Flag("bucket").Value.String()
	if bucket == "" {
		return "", "", fmt.Errorf(config.Red("bucket is required\n"))
	}
	manifestPath := cmd.Flag("manifest").Value.String()
	if manifestPath == "" {
		return "", "", fmt.Errorf(config.Red("manifest is required\n"))
	}
	return bucket, manifestPath, nil
}
