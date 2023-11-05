package manifest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Initialize a global S3 client
var s3Client *s3.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}

	s3Client = s3.NewFromConfig(cfg)
}

// ListManifests lists all manifest files within a specified folder in an S3 bucket.
func ListManifests(ctx context.Context, bucket, prefix string) ([]string, error) {
	var manifests []string

	resp, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, err
	}

	for _, item := range resp.Contents {
		manifests = append(manifests, *item.Key)
	}

	return manifests, nil
}

// DownloadManifest downloads a specific manifest file from an S3 bucket.
func DownloadManifest(ctx context.Context, bucket, keyPrefix, localFolderPath string) error {
	paginator := s3.NewListObjectsV2Paginator(s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(keyPrefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}

		for _, object := range page.Contents {
			outputPath := filepath.Join(localFolderPath, *object.Key)

			// Create any directories as needed
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return err
			}

			// Create the local file
			file, err := os.Create(outputPath)
			if err != nil {
				return err
			}

			// Get the object from S3
			output, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
				Bucket: aws.String(bucket),
				Key:    object.Key,
			})
			if err != nil {
				file.Close() // Close the file if we can't download
				return err
			}

			// Write to the local file
			if _, err := io.Copy(file, output.Body); err != nil {
				file.Close() // Close the file if we can't copy
				return err
			}

			// Close the file
			if err := file.Close(); err != nil {
				return err
			}
		}
	}

	return nil
}

// UploadManifest uploads a manifest file to an S3 bucket.
func UploadManifest(ctx context.Context, bucket, keyPrefix, localFolderPath string) error {
	// Walk the directory tree
	err := filepath.Walk(localFolderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Open the file
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Define the key for the S3 object
		key := keyPrefix + strings.TrimPrefix(path, localFolderPath)
		key = strings.ReplaceAll(key, string(filepath.Separator), "/") // Ensure key uses '/' for S3

		// Upload to S3
		_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   file,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

// AcquireStateLock attempts to create or update the state lock file in an S3 bucket.
func AcquireStateLock(ctx context.Context, bucket, key, lockInfo string) error {
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(lockInfo),
	})
	return err
}

// CheckStateLock reads the state lock file from an S3 bucket.
func CheckStateLock(ctx context.Context, bucket, key string) (bool, string, error) {
	resp, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFound *types.NoSuchKey
		if errors.As(err, &notFound) {
			return false, "", nil // Lock does not exist.
		}
		return false, "", err // Some other error occurred.
	}
	defer resp.Body.Close()

	lockInfo, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", err
	}

	return true, string(lockInfo), nil
}

// ReleaseStateLock deletes the state lock file or clears its content in an S3 bucket.
func ReleaseStateLock(ctx context.Context, bucket, key string) error {
	_, err := s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

// RefreshState downloads the state file from S3 and writes it to the local file system.
func RefreshState(localStatePath, bucket, stateFile string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("your-region"),
	)
	if err != nil {
		return fmt.Errorf("error loading AWS configuration: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	resp, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(stateFile),
	})
	if err != nil {
		return fmt.Errorf("failed to retrieve state file: %v", err)
	}
	defer resp.Body.Close()

	// Write the S3 object to the local file system
	localFile, err := os.Create(localStatePath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %v", err)
	}
	defer localFile.Close()

	_, err = io.Copy(localFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write state to local file: %v", err)
	}

	return nil
}

// SyncState uploads the local state file to the S3 bucket.
func SyncState(localStatePath, bucket, stateFile string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("your-region"),
	)
	if err != nil {
		return fmt.Errorf("error loading AWS configuration: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	localFile, err := os.Open(localStatePath)
	if err != nil {
		return fmt.Errorf("failed to open local state file: %v", err)
	}
	defer localFile.Close()

	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get local file stats: %v", err)
	}

	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(stateFile),
		Body:          localFile,
		ContentLength: *aws.Int64(fileInfo.Size()),
	})
	if err != nil {
		return fmt.Errorf("failed to upload state file to S3: %v", err)
	}

	return nil
}
