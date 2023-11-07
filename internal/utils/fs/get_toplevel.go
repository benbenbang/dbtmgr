package fs

import (
	"path/filepath"
	"strings"
)

// GetTopLevelDir extracts the top-level directory from the given path.
func GetTopLevelDir(path string) string {
	// Clean the path to ensure consistent separators and remove any trailing slash
	cleanPath := filepath.Clean(path)

	// Check if the path is absolute
	isAbs := filepath.IsAbs(cleanPath)

	// Split the path into parts using the OS-specific separator
	parts := strings.Split(cleanPath, string(filepath.Separator))

	// If the path is absolute, the first part is the root, so we return the second part
	if isAbs && len(parts) > 1 {
		return parts[1]
	}

	// If the path is relative and has multiple parts, return the first one
	if len(parts) > 0 {
		return parts[0]
	}

	// If none of the above, return the original path
	return cleanPath
}
