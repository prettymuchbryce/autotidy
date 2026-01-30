package fs

import (
	"fmt"
	"os"

	"github.com/spf13/afero"
)

// MemFileSystem is an in-memory filesystem for testing.
// Unlike DryRunFileSystem, it performs no logging.
type MemFileSystem struct {
	afero.Fs
}

// Copy copies a file or directory from src to dst.
func (m *MemFileSystem) Copy(src, dst string) error {
	srcInfo, err := m.Fs.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return copyDir(m.Fs, src, dst)
	}
	return copyFile(m.Fs, src, dst, srcInfo.Mode())
}

// Trash simulates trashing by removing the file.
func (m *MemFileSystem) Trash(path string) error {
	return m.Fs.RemoveAll(path)
}

// ResolveConflict handles destination file conflicts.
func (m *MemFileSystem) ResolveConflict(mode ConflictMode, srcPath, destPath string) (string, bool, error) {
	switch mode {
	case ConflictRenameWithSuffix:
		newPath := m.findAvailableSuffixedPath(destPath)
		return newPath, true, nil

	case ConflictSkip:
		return destPath, false, nil

	case ConflictOverwrite:
		if err := m.Fs.Remove(destPath); err != nil {
			return "", false, err
		}
		return destPath, true, nil

	case ConflictTrash:
		return "", false, fmt.Errorf("trash conflict mode is not yet implemented")

	default:
		return "", false, fmt.Errorf("unknown conflict mode: %s", mode)
	}
}

// findAvailableSuffixedPath finds the next available path with a numeric suffix.
func (m *MemFileSystem) findAvailableSuffixedPath(destPath string) string {
	for i := 2; ; i++ {
		candidate := GenerateSuffixedPath(destPath, i)
		if _, err := m.Fs.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

// MustMkdirAll creates a directory and panics on error. For use in tests.
func (m *MemFileSystem) MustMkdirAll(path string) {
	if err := m.Fs.MkdirAll(path, 0755); err != nil {
		panic(fmt.Sprintf("MustMkdirAll(%q): %v", path, err))
	}
}

// MustRemoveAll removes a path and panics on error. For use in tests.
func (m *MemFileSystem) MustRemoveAll(path string) {
	if err := m.Fs.RemoveAll(path); err != nil {
		panic(fmt.Sprintf("MustRemoveAll(%q): %v", path, err))
	}
}
