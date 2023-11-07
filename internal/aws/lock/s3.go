package lock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"statectl/internal/logging"
	"statectl/internal/utils/subproc"
	t "statectl/internal/utils/types"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Initialize a global S3 client
var KeyNotFound *types.NoSuchKey
var log = logging.GetLogger()
var LockExists = errors.New("lock exists")

// AcquireStateLock attempts to create or update the state lock file in an S3 bucket.
func AcquireStateLock(ctx context.Context, cli *s3.Client, bucket, key string, lockInfo t.LockInfo) error {
	exist, _, err := CheckStateLock(ctx, cli, bucket, key, false)
	if err != nil {
		return err
	}

	if exist {
		same_sha, err := subproc.CompareSHAs(cli, bucket, key)
		if err != nil {
			return fmt.Errorf("unable to acquire lock: lock already exists")
		}
		if !same_sha {
			return fmt.Errorf("unable to acquire lock: lock already exists and it's not owned by this process.\nthis can happen if the lock was created by another process or user.\nplease retry after the lock is released or use the force-acquire command")
		}
		return LockExists
	}

	lockInfoRaw, err := json.Marshal(lockInfo)
	if err != nil {
		return err
	}
	log.Debugf("Raw lock info: %s\n", lockInfoRaw)

	_, err = cli.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(string(lockInfoRaw)),
	})
	return err
}

// CheckStateLock reads the state lock file from an S3 bucket.
func CheckStateLock(ctx context.Context, cli *s3.Client, bucket, key string, serialize bool) (bool, t.LockInfo, error) {
	resp, err := cli.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if errors.As(err, &KeyNotFound) {
			return false, t.LockInfo{}, nil // Lock does not exist.
		}
		return false, t.LockInfo{}, err // Some other error occurred.
	}
	defer resp.Body.Close()

	lockInfoRaw, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, t.LockInfo{}, err
	}
	log.Debugf("Raw lock info: %s\n", lockInfoRaw)

	if !serialize {
		return true, t.LockInfo{}, nil
	}

	lockInfo := t.LockInfo{}
	err = json.Unmarshal(lockInfoRaw, &lockInfo)
	if err != nil {
		return false, lockInfo, err
	}

	return true, lockInfo, nil
}

// ReleaseStateLock deletes the state lock file in an S3 bucket if it exists.
func ReleaseStateLock(ctx context.Context, cli *s3.Client, bucket, key string) error {
	_, _, err := CheckStateLock(ctx, cli, bucket, key, false)
	if err != nil {
		return err
	}

	same_sha, err := subproc.CompareSHAs(cli, bucket, key)
	if err != nil {
		return fmt.Errorf("unable to release lock: %v", err)
	}

	if !same_sha {
		return fmt.Errorf("unable to release lock: local and remote SHAs do not match. If you are sure you want to release the lock, use the force-release command")
	}

	// If the lock file exists, attempt to delete it
	_, err = cli.DeleteObject(
		ctx,
		&s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		},
	)
	return err
}

// ForceReleaseLock deletes the state lock file or clears its content in an S3 bucket.
func ForceReleaseLock(ctx context.Context, cli *s3.Client, bucket, key string) error {
	_, _, err := CheckStateLock(ctx, cli, bucket, key, false)
	if err != nil {
		return err
	}

	// If the lock file exists, attempt to delete it
	_, err = cli.DeleteObject(
		ctx,
		&s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		},
	)
	return err
}
