package subproc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"statectl/internal/logging"
	t "statectl/internal/utils/types"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws"
)

type LockInfo struct {
	LockID string `json:"lock_id"`
}

var KeyNotFound *types.NoSuchKey
var log = logging.GetLogger()

// FetchLocalSHA returns the current local git commit SHA.
func FetchLocalSHA() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error fetching local git SHA: %v", err)
	}
	log.Debugf("Local git SHA: %s", string(output))
	return string(bytes.TrimSpace(output)), nil
}

// FetchRemoteSHA fetches the git commit SHA from the state lock file in the S3 bucket.
func FetchRemoteSHA(ctx context.Context, cli t.S3Client, bucket, key string) (string, error) {
	resp, err := cli.GetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		if errors.As(err, &KeyNotFound) {
			return "", nil // Lock does not exist.
		}
		return "", err // Some other error occurred.
	}
	defer resp.Body.Close()

	lockInfoRaw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	lockInfo := LockInfo{}
	err = json.Unmarshal(lockInfoRaw, &lockInfo)
	if err != nil {
		return "", err
	}

	log.Debugf("Remote Lock ID (git SHA): %s", lockInfo.LockID)
	return lockInfo.LockID, nil
}

// CompareSHAs compares the local and remote git commit SHAs and returns them.
func CompareSHAs(ctx context.Context, cli t.S3Client, bucket, key string) (bool, error) {
	var localSHA, remoteSHA string
	var err error

	CI_COMMIT_SHA := os.Getenv("CI_COMMIT_SHA")
	if CI_COMMIT_SHA == "" {
		localSHA, err = FetchLocalSHA()
		if err != nil {
			return false, err
		}
	} else {
		localSHA = CI_COMMIT_SHA
	}

	remoteSHA, err = FetchRemoteSHA(ctx, cli, bucket, key)
	if err != nil {
		return false, err
	}

	return localSHA == remoteSHA, nil
}
