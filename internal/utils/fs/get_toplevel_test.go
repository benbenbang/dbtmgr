package fs_test

import (
	"statectl/internal/utils/fs"
	"testing"
)

func TestGetTopLevelDir(t *testing.T) {
	testCase := []string{
		"/home/user/test",
		"/home/user/test/",
		"home/user/test",
		"home/user/test/",
		"/home",
		"/home/",
		"home",
		"home/",
	}

	for _, path := range testCase {
		topLevel := fs.GetTopLevelDir(path)
		if topLevel != "home" {
			t.Errorf("Expected home, got %s", topLevel)
		}
	}
}
