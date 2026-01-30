package fs

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/afero"
)

// RealFileSystem performs actual filesystem operations.
type RealFileSystem struct {
	afero.Fs
}

// Rename performs the rename operation.
func (r *RealFileSystem) Rename(oldname, newname string) error {
	slog.Debug("renaming", "from", oldname, "to", newname)
	return r.Fs.Rename(oldname, newname)
}

// Remove performs the remove operation.
func (r *RealFileSystem) Remove(name string) error {
	slog.Debug("removing", "path", name)
	return r.Fs.Remove(name)
}

// RemoveAll performs the recursive remove operation.
func (r *RealFileSystem) RemoveAll(path string) error {
	slog.Debug("removing all", "path", path)
	return r.Fs.RemoveAll(path)
}

// Copy copies a file or directory from src to dst.
func (r *RealFileSystem) Copy(src, dst string) error {
	slog.Debug("copying", "from", src, "to", dst)
	srcInfo, err := r.Fs.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return copyDir(r.Fs, src, dst)
	}
	return copyFile(r.Fs, src, dst, srcInfo.Mode())
}

// Trash moves a file/directory to the system trash (platform-specific).
func (r *RealFileSystem) Trash(path string) error {
	slog.Debug("trashing", "path", path)
	switch runtime.GOOS {
	case "darwin":
		return trashDarwin(path)
	case "windows":
		return trashWindows(path)
	case "linux":
		return trashLinux(r.Fs, path)
	default:
		return fmt.Errorf("trash not supported on %s", runtime.GOOS)
	}
}

// ResolveConflict handles destination file conflicts.
func (r *RealFileSystem) ResolveConflict(mode ConflictMode, srcPath, destPath string) (string, bool, error) {
	switch mode {
	case ConflictRenameWithSuffix:
		newPath := r.findAvailableSuffixedPath(destPath)
		slog.Debug("renaming to avoid conflict", "src", srcPath, "dest", newPath)
		return newPath, true, nil

	case ConflictSkip:
		slog.Debug("skipping file, destination exists", "src", srcPath, "dest", destPath)
		return destPath, false, nil

	case ConflictOverwrite:
		slog.Debug("overwriting destination file", "src", srcPath, "dest", destPath)
		if err := r.Fs.Remove(destPath); err != nil {
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
func (r *RealFileSystem) findAvailableSuffixedPath(destPath string) string {
	for i := 2; ; i++ {
		candidate := GenerateSuffixedPath(destPath, i)
		if _, err := r.Fs.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

// trashDarwin moves a file to trash on macOS using AppleScript.
func trashDarwin(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	script := fmt.Sprintf(`tell application "Finder" to delete POSIX file %q`, absPath)
	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

// trashWindows moves a file to the Recycle Bin on Windows using PowerShell.
func trashWindows(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Escape single quotes for PowerShell single-quoted string (double them)
	escaped := strings.ReplaceAll(absPath, "'", "''")

	// Use Shell.Application COM object to move to recycle bin
	script := fmt.Sprintf(`
		$shell = New-Object -ComObject Shell.Application
		$item = $shell.Namespace(0).ParseName('%s')
		$item.InvokeVerb('delete')
	`, escaped)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	return cmd.Run()
}

// trashLinux moves a file to trash following the FreeDesktop.org Trash specification.
// See: https://specifications.freedesktop.org/trash-spec/trashspec-latest.html
func trashLinux(fs afero.Fs, path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Get the trash directory
	trashDir, err := getLinuxTrashDir()
	if err != nil {
		return err
	}

	// Ensure trash directories exist
	filesDir := filepath.Join(trashDir, "files")
	infoDir := filepath.Join(trashDir, "info")
	if err := fs.MkdirAll(filesDir, 0700); err != nil {
		return err
	}
	if err := fs.MkdirAll(infoDir, 0700); err != nil {
		return err
	}

	// Generate a unique name for the trashed file
	baseName := filepath.Base(absPath)
	trashedName := baseName
	trashedPath := filepath.Join(filesDir, trashedName)

	// If file already exists in trash, add a suffix
	for i := 1; ; i++ {
		if _, err := fs.Stat(trashedPath); os.IsNotExist(err) {
			break
		}
		trashedName = fmt.Sprintf("%s.%d", baseName, i)
		trashedPath = filepath.Join(filesDir, trashedName)
	}

	// Create the .trashinfo file
	infoPath := filepath.Join(infoDir, trashedName+".trashinfo")
	infoContent := fmt.Sprintf("[Trash Info]\nPath=%s\nDeletionDate=%s\n",
		absPath,
		time.Now().Format("2006-01-02T15:04:05"),
	)
	if err := afero.WriteFile(fs, infoPath, []byte(infoContent), 0600); err != nil {
		return err
	}

	// Move the file to trash
	if err := fs.Rename(absPath, trashedPath); err != nil {
		// Clean up the info file if move fails
		fs.Remove(infoPath)
		return err
	}

	return nil
}

// getLinuxTrashDir returns the path to the user's trash directory.
func getLinuxTrashDir() (string, error) {
	// First try XDG_DATA_HOME
	if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		return filepath.Join(xdgDataHome, "Trash"), nil
	}

	// Fall back to ~/.local/share/Trash
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "Trash"), nil
}

// copyFile copies a single file.
func copyFile(afs afero.Fs, src, dst string, mode os.FileMode) error {
	srcFile, err := afs.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := afs.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return nil
}

// copyDir recursively copies a directory.
func copyDir(afs afero.Fs, src, dst string) error {
	srcInfo, err := afs.Stat(src)
	if err != nil {
		return err
	}

	if err := afs.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := afero.ReadDir(afs, src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(afs, srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(afs, srcPath, dstPath, entry.Mode()); err != nil {
				return err
			}
		}
	}

	return nil
}
