package lock_test

import (
	"context"
	"statectl/internal/aws/lock"
	"statectl/test"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/mock"
)

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) PutObject(ctx context.Context, input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

func TestAcquireStateLock(t *testing.T) {
	// Set up
	ctx := context.Background()
	mockS3 := new(test.MockS3Client)

	bucket := "test-bucket"
	key := "test-key"
	expectedSHA := "1234567890"
	lockInfo, _ := test.CreateLockInfo(expectedSHA)

	// AcquireStateLock: GetObject -> PutObject
	// Mock GetObject
	mockS3.On("GetObject", mock.AnythingOfType("backgroundCtx"), mock.AnythingOfType("*s3.GetObjectInput"), mock.AnythingOfType("[]func(*s3.Options)")).Return(
		&s3.GetObjectOutput{}, &types.NoSuchKey{},
	)
	// Mock PutObject
	mockS3.On("PutObject", mock.AnythingOfType("backgroundCtx"), mock.AnythingOfType("*s3.PutObjectInput"), mock.AnythingOfType("[]func(*s3.Options)")).Return(
		&s3.PutObjectOutput{}, nil,
	)

	// Call the function under test
	if err := lock.AcquireStateLock(ctx, mockS3, bucket, key, lockInfo); err != nil {
		t.Errorf("error acquiring state lock: %v", err)
	}

	// Assertions
	mockS3.AssertExpectations(t)
}
