package subproc_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"statectl/internal/utils/subproc"
	"statectl/internal/utils/types"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) PutObject(ctx context.Context, input *s3.PutObjectInput, opts ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

func (m *MockS3Client) GetObject(ctx context.Context, input *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

func (m *MockS3Client) DeleteObject(ctx context.Context, input *s3.DeleteObjectInput, opts ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*s3.DeleteObjectOutput), args.Error(1)
}

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

	infoLock := types.LockInfo{
		LockID:    expectedSHA,
		TimeStamp: "2021-01-01T00:00:00Z",
		Signer:    "test",
		Comments: types.Comments{
			Commit:  "ok",
			Trigger: "ok",
			Extra:   "ok",
		},
	}

	infoLockRaw, err := json.Marshal(infoLock)
	if err != nil {
		t.Fatal(err)
	}

	// Mock setup
	mockS3 := new(MockS3Client)
	// Using mock.AnythingOfType to match any context.Background type
	mockS3.On("GetObject", mock.AnythingOfType("backgroundCtx"), mock.AnythingOfType("*s3.GetObjectInput"), mock.AnythingOfType("[]func(*s3.Options)")).Return(
		&s3.GetObjectOutput{
			Body: io.NopCloser(bytes.NewReader(infoLockRaw)),
		}, nil,
	)

	// Test function call
	result, err := subproc.FetchRemoteSHA(ctx, mockS3, "bucket-name", "file-path")

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, expectedSHA, result)
}
