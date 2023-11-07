package fs

import (
	"os"
)

func IsDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	if fileInfo.IsDir() {
		return true, nil
	} else {
		return false, nil
	}
}

func IsFile(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	if !fileInfo.IsDir() {
		return true, nil
	} else {
		return false, nil
	}
}
