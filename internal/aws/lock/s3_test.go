package lock_test

import (
	"bytes"
	"context"
	"io"
	"statectl/internal/aws/lock"
	"statectl/internal/utils/subproc"
	"statectl/test"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/mock"
)

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

func TestCheckStateLock(t *testing.T) {
	// Set up
	ctx := context.Background()
	mockS3 := new(test.MockS3Client)

	bucket := "test-bucket"
	key := "test-key"
	expectedSHA := "1234567890"
	lockInfo, lockInfoRaw := test.CreateLockInfo(expectedSHA)

	// CheckStateLock: GetObject
	// Mock GetObject
	mockS3.On("GetObject", mock.AnythingOfType("backgroundCtx"), mock.AnythingOfType("*s3.GetObjectInput"), mock.AnythingOfType("[]func(*s3.Options)")).Return(
		&s3.GetObjectOutput{
			Body: io.NopCloser(bytes.NewReader(lockInfoRaw)),
		}, nil,
	)

	// Call the function under test
	exist, lockInfoOut, err := lock.CheckStateLock(ctx, mockS3, bucket, key, true)
	if err != nil {
		t.Errorf("error checking state lock: %v", err)
	}

	// Assertions
	mockS3.AssertExpectations(t)
	if !exist {
		t.Errorf("expected lock to exist")
	}
	if lockInfo.LockID != lockInfoOut.LockID {
		t.Errorf("expected lock info to match")
	}
}

func TestReleaseStateLock(t *testing.T) {
	// Set up
	ctx := context.Background()
	mockS3 := new(test.MockS3Client)

	bucket := "test-bucket"
	key := "test-key"
	expectedSHA, err := subproc.FetchLocalSHA()
	if err != nil {
		t.Errorf("error fetching local git SHA: %v", err)
	}
	_, lockInfoRaw := test.CreateLockInfo(expectedSHA)

	// CheckStateLock: GetObject
	// Mock GetObject
	mockS3.On("GetObject", mock.AnythingOfType("backgroundCtx"), mock.AnythingOfType("*s3.GetObjectInput"), mock.AnythingOfType("[]func(*s3.Options)")).Return(
		&s3.GetObjectOutput{
			Body: io.NopCloser(bytes.NewReader(lockInfoRaw)),
		}, nil,
	)

	// Mock DeleteObject
	mockS3.On("DeleteObject", mock.AnythingOfType("backgroundCtx"), mock.AnythingOfType("*s3.DeleteObjectInput"), mock.AnythingOfType("[]func(*s3.Options)")).Return(
		&s3.DeleteObjectOutput{}, nil,
	)

	// Call the function under test
	if err := lock.ReleaseStateLock(ctx, mockS3, bucket, key); err != nil {
		t.Errorf("error releasing state lock: %v", err)
	}

	// Assertions
	mockS3.AssertExpectations(t)
}

func TestReleaseStateLockForce(t *testing.T) {
	// Set up
	ctx := context.Background()
	mockS3 := new(test.MockS3Client)

	bucket := "test-bucket"
	key := "test-key"
	expectedSHA := "1234567890"
	_, lockInfoRaw := test.CreateLockInfo(expectedSHA)

	// CheckStateLock: GetObject
	// Mock GetObject
	mockS3.On("GetObject", mock.AnythingOfType("backgroundCtx"), mock.AnythingOfType("*s3.GetObjectInput"), mock.AnythingOfType("[]func(*s3.Options)")).Return(
		&s3.GetObjectOutput{
			Body: io.NopCloser(bytes.NewReader(lockInfoRaw)),
		}, nil,
	)

	// Mock DeleteObject
	mockS3.On("DeleteObject", mock.AnythingOfType("backgroundCtx"), mock.AnythingOfType("*s3.DeleteObjectInput"), mock.AnythingOfType("[]func(*s3.Options)")).Return(
		&s3.DeleteObjectOutput{}, nil,
	)

	// Call the function under test
	if err := lock.ForceReleaseLock(ctx, mockS3, bucket, key); err != nil {
		t.Errorf("error releasing state lock: %v", err)
	}

	// Assertions
	mockS3.AssertExpectations(t)
}
