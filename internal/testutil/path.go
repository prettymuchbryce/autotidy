package testutil

import (
	"path/filepath"
	"runtime"
)

// Path creates a platform-independent absolute path by joining parts with the
// OS-specific separator. Use this in tests instead of hardcoded paths
// like "/root/file.txt" to ensure tests pass on Windows.
//
// On Unix, Path("/", "home", "user") returns "/home/user"
// On Windows, Path("/", "home", "user") returns "C:\\home\\user"
//
// The first argument should be "/" to indicate an absolute path from root.
func Path(parts ...string) string {
	if len(parts) == 0 {
		return ""
	}

	// If the first part is "/", it indicates an absolute path from root
	if parts[0] == "/" {
		if runtime.GOOS == "windows" {
			// On Windows, use C:\ as the root drive (C: alone is relative!)
			return "C:\\" + filepath.Join(parts[1:]...)
		}
		// On Unix, filepath.Join handles the leading "/"
		return filepath.Join(parts...)
	}

	// For relative paths, just join normally
	return filepath.Join(parts...)
}
