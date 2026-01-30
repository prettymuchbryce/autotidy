package fs

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// ConflictMode defines what to do when the destination file already exists.
type ConflictMode string

const (
	ConflictRenameWithSuffix ConflictMode = "rename_with_suffix" // Rename with numeric suffix (file_2.txt)
	ConflictSkip             ConflictMode = "skip"               // Skip the file, don't proceed
	ConflictOverwrite        ConflictMode = "overwrite"          // Overwrite the destination file
	ConflictTrash            ConflictMode = "trash"              // Move destination file to trash
)

// FileSystem extends afero.Fs with autotidy-specific operations.
type FileSystem interface {
	afero.Fs

	// Copy copies a file or directory from src to dst.
	// For directories, copies recursively.
	Copy(src, dst string) error

	// Trash moves a file/directory to the system trash (platform-specific).
	// On macOS: uses Finder via AppleScript
	// On Windows: uses Recycle Bin via PowerShell
	// On Linux: follows FreeDesktop.org Trash specification
	Trash(path string) error

	// ResolveConflict handles destination file conflicts.
	// Returns (newDestPath, proceed, err) - if proceed is true, caller should continue
	// using newDestPath as the destination. For most modes newDestPath equals destPath,
	// but for RenameWithSuffix it may be different (e.g., file_2.txt).
	ResolveConflict(mode ConflictMode, srcPath, destPath string) (string, bool, error)
}

// NewReal creates a FileSystem that performs actual filesystem operations.
func NewReal() FileSystem {
	return &RealFileSystem{
		Fs: afero.NewOsFs(),
	}
}

// NewDryRun creates a FileSystem that logs operations without modifying the real filesystem.
// Uses CopyOnWriteFs so subsequent operations work correctly (e.g., mkdir followed by move).
func NewDryRun() FileSystem {
	base := afero.NewReadOnlyFs(afero.NewOsFs())
	layer := afero.NewMemMapFs()
	cow := afero.NewCopyOnWriteFs(base, layer)
	return &DryRunFileSystem{Fs: cow}
}

// NewMem creates an in-memory FileSystem for testing.
// Unlike DryRunFileSystem, it performs no logging.
func NewMem() FileSystem {
	return &MemFileSystem{Fs: afero.NewMemMapFs()}
}

// NewMemTest returns a MemFileSystem for testing with access to Must* helpers.
func NewMemTest() *MemFileSystem {
	return &MemFileSystem{Fs: afero.NewMemMapFs()}
}

// GenerateSuffixedPath generates a path with a numeric suffix.
// For example: file.txt with suffix 2 becomes file_2.txt
// For multi-extension files: archive.tar.gz becomes archive_2.tar.gz
func GenerateSuffixedPath(path string, suffix int) string {
	dir := filepath.Dir(path)
	filename := filepath.Base(path)

	// Find the base name and extensions
	// For "archive.tar.gz" we want base="archive", ext=".tar.gz"
	// For ".hidden.txt" we want base=".hidden", ext=".txt"
	base, ext := splitFilenameAndExtensions(filename)

	newFilename := fmt.Sprintf("%s_%d%s", base, suffix, ext)
	return filepath.Join(dir, newFilename)
}

// splitFilenameAndExtensions splits a filename into base and extensions.
// Unlike filepath.Ext, this treats compound extensions as one unit.
// Examples:
//   - "file.txt" → ("file", ".txt")
//   - "archive.tar.gz" → ("archive", ".tar.gz")
//   - "file" → ("file", "")
//   - ".hidden" → (".hidden", "")
//   - ".hidden.txt" → (".hidden", ".txt")
func splitFilenameAndExtensions(filename string) (base, ext string) {
	// Handle hidden files (starting with .)
	if strings.HasPrefix(filename, ".") {
		// Find the first dot after the leading dot
		rest := filename[1:]
		idx := strings.Index(rest, ".")
		if idx == -1 {
			// No extension, e.g., ".hidden"
			return filename, ""
		}
		// e.g., ".hidden.txt" → base=".hidden", ext=".txt"
		return filename[:idx+1], filename[idx+1:]
	}

	// Normal files - find the first dot
	idx := strings.Index(filename, ".")
	if idx == -1 {
		return filename, ""
	}
	return filename[:idx], filename[idx:]
}
