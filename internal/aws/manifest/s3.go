package manifest

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"statectl/internal/utils/fs"
	t "statectl/internal/utils/types"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ListManifests lists all manifest files within a specified folder in an S3 bucket.
func ListManifests(ctx context.Context, cli *s3.Client, bucket, prefix string) (map[string]interface{}, error) {
	const fileIndicator = "<file>"

	prefix = fs.GetTopLevelDir(prefix)

	resp, err := cli.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
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
func DownloadManifest(ctx context.Context, cli *s3.Client, bucket, keyPrefix, localFolderPath string) error {
	paginator := s3.NewListObjectsV2Paginator(cli, &s3.ListObjectsV2Input{
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
			output, err := cli.GetObject(ctx, &s3.GetObjectInput{
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
func UploadManifest(ctx context.Context, cli *s3.Client, bucket, localFolderPath string, singleFile bool) error {

	// Get the top-level directory from the localFolderPath
	if !singleFile {
		localFolderPath = fs.GetTopLevelDir(localFolderPath)
	}

	// Trim the localFolderPath to ensure it ends with a separator
	// and remove it from the path to get the correct key structure
	if isDir, err := fs.IsDir(localFolderPath); err != nil {
		return err
	} else if isDir {
		localFolderPath = strings.TrimRight(localFolderPath, string(filepath.Separator)) + string(filepath.Separator)
	} else {
		localFolderPath = strings.TrimRight(localFolderPath, string(filepath.Separator))
	}

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
		_, err = cli.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   file,
		})
		return err
	})

	return err
}

func CreateStateJSON(ctx context.Context, cli *s3.Client, bucket, key, filePath string) error {
	// Get the version ID from S3
	resp, err := cli.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to get object head: %w", err)
	}

	// Get the current commit SHA from Git
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get git commit SHA: %w", err)
	}
	commitSHA := strings.Trim(string(output), "\n")

	// Create the state structure
	state := t.State{
		VersionID: *resp.VersionId,
		CommitSHA: commitSHA,
		Bucket:    bucket,
		Key:       key,
	}

	// Marshal into JSON
	stateJSON, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to file
	err = os.WriteFile(filePath, stateJSON, 0644)
	if err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}
