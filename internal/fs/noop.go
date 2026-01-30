package fs

import (
	"os"
	"time"

	"github.com/spf13/afero"
)

// NoopFileSystem is a FileSystem that panics on any operation.
// Use this in tests where the filesystem should never be accessed.
type NoopFileSystem struct{}

// NewNoop creates a FileSystem that panics on any operation.
// Use this in tests to make explicit that the filesystem is not used.
func NewNoop() FileSystem {
	return &NoopFileSystem{}
}

func (n *NoopFileSystem) Create(name string) (afero.File, error) {
	panic("NoopFileSystem: Create called")
}

func (n *NoopFileSystem) Mkdir(name string, perm os.FileMode) error {
	panic("NoopFileSystem: Mkdir called")
}

func (n *NoopFileSystem) MkdirAll(path string, perm os.FileMode) error {
	panic("NoopFileSystem: MkdirAll called")
}

func (n *NoopFileSystem) Open(name string) (afero.File, error) {
	panic("NoopFileSystem: Open called")
}

func (n *NoopFileSystem) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	panic("NoopFileSystem: OpenFile called")
}

func (n *NoopFileSystem) Remove(name string) error {
	panic("NoopFileSystem: Remove called")
}

func (n *NoopFileSystem) RemoveAll(path string) error {
	panic("NoopFileSystem: RemoveAll called")
}

func (n *NoopFileSystem) Rename(oldname, newname string) error {
	panic("NoopFileSystem: Rename called")
}

func (n *NoopFileSystem) Stat(name string) (os.FileInfo, error) {
	panic("NoopFileSystem: Stat called")
}

func (n *NoopFileSystem) Name() string {
	return "NoopFileSystem"
}

func (n *NoopFileSystem) Chmod(name string, mode os.FileMode) error {
	panic("NoopFileSystem: Chmod called")
}

func (n *NoopFileSystem) Chown(name string, uid, gid int) error {
	panic("NoopFileSystem: Chown called")
}

func (n *NoopFileSystem) Chtimes(name string, atime time.Time, mtime time.Time) error {
	panic("NoopFileSystem: Chtimes called")
}

func (n *NoopFileSystem) Copy(src, dst string) error {
	panic("NoopFileSystem: Copy called")
}

func (n *NoopFileSystem) Trash(path string) error {
	panic("NoopFileSystem: Trash called")
}

func (n *NoopFileSystem) ResolveConflict(mode ConflictMode, srcPath, destPath string) (string, bool, error) {
	panic("NoopFileSystem: ResolveConflict called")
}
