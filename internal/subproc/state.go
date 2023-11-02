package subproc

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Assuming the S3 bucket and object key for the state lock file
const (
	stateLockFile = "path/to/state.lock"
	bucketName    = "your-bucket-name"
)

// FetchLocalSHA returns the current local git commit SHA.
func FetchLocalSHA() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error fetching local git SHA: %v", err)
	}
	return string(bytes.TrimSpace(output)), nil
}

// FetchRemoteSHA fetches the git commit SHA from the state lock file in the S3 bucket.
func FetchRemoteSHA() (string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("your-region"), // Specify your AWS region
	)
	if err != nil {
		return "", fmt.Errorf("error loading AWS configuration: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	// Get the state lock file from S3
	result, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(stateLockFile),
	})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve state lock file: %v", err)
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(result.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read state lock file: %v", err)
	}

	return string(buf.String()), nil
}

// CompareSHAs compares the local and remote git commit SHAs and returns them.
func CompareSHAs() (localSHA, remoteSHA string, err error) {
	localSHA, err = FetchLocalSHA()
	if err != nil {
		return "", "", err
	}

	remoteSHA, err = FetchRemoteSHA()
	if err != nil {
		return "", "", err
	}

	return localSHA, remoteSHA, nil
}
