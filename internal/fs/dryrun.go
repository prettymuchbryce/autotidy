package fs

import (
	"fmt"
	"os"

	"github.com/spf13/afero"
)

// DryRunFileSystem simulates operations without modifying the real filesystem.
// Uses CopyOnWriteFs so operations work correctly in memory.
type DryRunFileSystem struct {
	afero.Fs
}

// MkdirAll delegates to the CoW filesystem.
func (d *DryRunFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return d.Fs.MkdirAll(path, perm)
}

// Mkdir delegates to the CoW filesystem.
func (d *DryRunFileSystem) Mkdir(path string, perm os.FileMode) error {
	return d.Fs.Mkdir(path, perm)
}

// Remove is a no-op in dry-run mode.
// CoW doesn't support removing files that only exist in the base layer.
func (d *DryRunFileSystem) Remove(name string) error {
	return nil
}

// RemoveAll is a no-op in dry-run mode.
// CoW doesn't support removing files that only exist in the base layer.
func (d *DryRunFileSystem) RemoveAll(path string) error {
	return nil
}

// Rename copies to new location so subsequent actions work.
// CoW doesn't support renaming files that only exist in the base layer,
// so we copy instead. The original still exists but subsequent actions use the new path.
func (d *DryRunFileSystem) Rename(oldname, newname string) error {
	// Copy to new location so subsequent actions can find the file
	srcInfo, err := d.Fs.Stat(oldname)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return copyDir(d.Fs, oldname, newname)
	}
	return copyFile(d.Fs, oldname, newname, srcInfo.Mode())
}

// Copy performs the copy in memory so subsequent actions work.
func (d *DryRunFileSystem) Copy(src, dst string) error {
	// Actually copy to memory layer so subsequent actions can operate on the file
	srcInfo, err := d.Fs.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return copyDir(d.Fs, src, dst)
	}
	return copyFile(d.Fs, src, dst, srcInfo.Mode())
}

// Create delegates to the CoW filesystem.
func (d *DryRunFileSystem) Create(name string) (afero.File, error) {
	return d.Fs.Create(name)
}

// OpenFile delegates to the CoW filesystem.
func (d *DryRunFileSystem) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return d.Fs.OpenFile(name, flag, perm)
}

// Trash is a no-op in dry-run mode.
func (d *DryRunFileSystem) Trash(path string) error {
	return nil
}

// ResolveConflict handles destination file conflicts in dry-run mode.
func (d *DryRunFileSystem) ResolveConflict(mode ConflictMode, srcPath, destPath string) (string, bool, error) {
	switch mode {
	case ConflictRenameWithSuffix:
		newPath := d.findAvailableSuffixedPath(destPath)
		return newPath, true, nil

	case ConflictSkip:
		return destPath, false, nil

	case ConflictOverwrite:
		// Don't actually remove (CoW can't remove base layer files)
		return destPath, true, nil

	case ConflictTrash:
		return destPath, true, nil

	default:
		return "", false, fmt.Errorf("unknown conflict mode: %s", mode)
	}
}

// findAvailableSuffixedPath finds the next available path with a numeric suffix.
func (d *DryRunFileSystem) findAvailableSuffixedPath(destPath string) string {
	for i := 2; ; i++ {
		candidate := GenerateSuffixedPath(destPath, i)
		if _, err := d.Fs.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}
