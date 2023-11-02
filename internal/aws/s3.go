package aws

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	awsV2 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws"
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
		Bucket: awsV2.String(bucket),
		Prefix: awsV2.String(prefix),
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
func DownloadManifest(ctx context.Context, bucket, s3FolderPrefix, localFolderPath string) error {
	paginator := s3.NewListObjectsV2Paginator(s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(s3FolderPrefix),
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
func UploadManifest(ctx context.Context, bucket, localFolderPath, s3KeyPrefix string) error {
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
		key := s3KeyPrefix + strings.TrimPrefix(path, localFolderPath)
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
		Bucket: awsV2.String(bucket),
		Key:    awsV2.String(key),
		Body:   strings.NewReader(lockInfo),
	})
	return err
}

// CheckStateLock reads the state lock file from an S3 bucket.
func CheckStateLock(ctx context.Context, bucket, key string) (bool, string, error) {
	resp, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: awsV2.String(bucket),
		Key:    awsV2.String(key),
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
		Bucket: awsV2.String(bucket),
		Key:    awsV2.String(key),
	})
	return err
}
