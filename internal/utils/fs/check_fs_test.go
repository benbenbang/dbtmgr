package fs_test

import (
	"os"
	"statectl/internal/utils/fs"
	"testing"
)

func TestIsDir(t *testing.T) {
	// Create a temporary directory
	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create a temporary file
	file, err := os.CreateTemp(dir, "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	// Check that the directory is a directory
	isDir, err := fs.IsDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !isDir {
		t.Errorf("%s is not a directory", dir)
	}

	// Check that the file is not a directory
	isDir, err = fs.IsDir(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	if isDir {
		t.Errorf("%s is a directory", file.Name())
	}
}

func TestIsFile(t *testing.T) {
	// Create a temporary directory
	dir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create a temporary file
	file, err := os.CreateTemp(dir, "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	// Check that the directory is not a file
	isFile, err := fs.IsFile(dir)
	if err != nil {
		t.Fatal(err)
	}
	if isFile {
		t.Errorf("%s is a file", dir)
	}

	// Check that the file is a file
	isFile, err = fs.IsFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !isFile {
		t.Errorf("%s is not a file", file.Name())
	}
}
