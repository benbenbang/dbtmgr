package subproc_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"statectl/internal/utils/subproc"
	"statectl/test"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFetchLocalSHA(t *testing.T) {
	// Test that the local SHA is fetched correctly.
	sha, err := subproc.FetchLocalSHA()
	if err != nil {
		t.Errorf("error fetching local git SHA: %v", err)
	}
	fmt.Printf("Local git SHA: %s", sha)
}

func TestFetchRemoteSHA(t *testing.T) {
	expectedSHA := "1234567890"
	ctx := context.Background()
	_, lockInfoRaw := test.CreateLockInfo(expectedSHA)

	// Mock setup
	mockS3 := new(test.MockS3Client)
	// Using mock.AnythingOfType to match any context.Background type
	mockS3.On("GetObject", mock.AnythingOfType("backgroundCtx"), mock.AnythingOfType("*s3.GetObjectInput"), mock.AnythingOfType("[]func(*s3.Options)")).Return(
		&s3.GetObjectOutput{
			Body: io.NopCloser(bytes.NewReader(lockInfoRaw)),
		}, nil,
	)

	// Test function call
	result, err := subproc.FetchRemoteSHA(ctx, mockS3, "bucket-name", "file-path")

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, expectedSHA, result)
}
