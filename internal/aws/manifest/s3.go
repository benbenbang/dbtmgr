package manifest

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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
func ListManifests(ctx context.Context, bucket, prefix string) (map[string]interface{}, error) {
	const fileIndicator = "<file>"

	resp, err := s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, err
	}

	temp := make(map[string]interface{})

	for _, item := range resp.Contents {
		key := *item.Key
		if key == prefix {
			// Skip the prefix itself
			continue
		}
		// Remove the prefix from the key
		key = strings.TrimPrefix(key, prefix)
		parts := strings.Split(key, "/")

		// Navigate/create the map structure according to the key parts
		currentMap := temp
		for i, part := range parts {
			if i == len(parts)-1 {
				// It's a file, denote it with a special value
				currentMap[part] = fileIndicator
			} else {
				// It's a directory, ensure there is a map for it
				if currentMap[part] == nil {
					currentMap[part] = make(map[string]interface{})
				}
				if subMap, ok := currentMap[part].(map[string]interface{}); ok {
					currentMap = subMap
				}
			}
		}
	}

	tree := make(map[string]interface{})
	tree[prefix] = temp[""]

	return tree, nil
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
			// Strip the keyPrefix from the S3 object key
			relativePath := strings.TrimPrefix(*object.Key, keyPrefix)

			// Ensure the relative path does not start with a slash
			relativePath = strings.TrimLeft(relativePath, "/")

			// Join the localFolderPath with the relativePath
			outputPath := filepath.Join(localFolderPath, keyPrefix, relativePath)

			// Create any directories as needed
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return err
			}
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

func ignoreFile(filename string) bool {
	// Define patterns or specific filenames to ignore
	ignorePatterns := []string{".DS_Store", "*.tmp", "*/temp/*"}

	for _, pattern := range ignorePatterns {
		matched, err := filepath.Match(pattern, filepath.Base(filename))
		if err != nil {
			// handle error
			return false
		}
		if matched {
			return true
		}
	}
	return false
}

// UploadManifest uploads a manifest file to an S3 bucket.
func UploadManifest(ctx context.Context, bucket, localFolderPath string) error {
	// Trim the localFolderPath to ensure it ends with a separator
	// and remove it from the path to get the correct key structure
	localFolderPath = strings.TrimRight(localFolderPath, string(filepath.Separator)) + string(filepath.Separator)

	err := filepath.Walk(localFolderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and ignored files
		if info.IsDir() || ignoreFile(path) {
			return nil
		}

		relativePath := strings.TrimPrefix(path, localFolderPath)
		// If you want to keep the 'target' as the root directory in the S3 key, prepend it here
		key := localFolderPath + relativePath
		// Replace OS-specific path separators with '/'
		key = strings.ReplaceAll(key, string(filepath.Separator), "/")

		// Open the file
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Upload the file to S3
		_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   file,
		})
		return err
	})

	return err
}
