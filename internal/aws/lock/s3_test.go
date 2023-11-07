package lock_test

import (
	"context"
	"statectl/internal/aws/lock"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
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
	ctx := context.TODO()
	mockS3 := new(MockS3Client)
	bucket := "test-bucket"
	key := "test-key"

	lockInfo := lock.LockInfo{
		LockID:    "test",
		TimeStamp: "2021-07-01T00:00:00Z",
		Signer:    "test",
		Comments: lock.Comments{
			Commit:  "test",
			Trigger: "test",
			Extra:   "test",
		},
	}

	// Define what the mock should return when PutObject is called.
	mockS3.On("PutObject", ctx, mock.AnythingOfType("*s3.PutObjectInput")).Return(&s3.PutObjectOutput{}, nil)

	// Call the function under test
	err := lock.AcquireStateLock(ctx, bucket, key, lockInfo)

	// Assertions
	assert.NoError(t, err)
	mockS3.AssertExpectations(t)
}
