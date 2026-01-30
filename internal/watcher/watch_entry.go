package watcher

import "time"

type LostRoot struct {
	// The path where the root should be.
	targetPath string

	// Whether the root should be recursive when found.
	isRecursive bool
}

// RecreatedRoot represents a root that was recreated after being lost.
type RecreatedRoot struct {
	Path string
	Time time.Time
}

type RootType int

const (
	None RootType = iota
	NonRecursiveRoot
	RecursiveRoot
)

type WatchEntry struct {
	// The path being watched.
	path string

	// If this path is a root, the type of root it is.
	rootType RootType

	// Whether this path is being watched recursively.
	isRecursive bool

	// Lost roots being tracked under this path because it is the closest accessible ancestor.
	lostRoots []*LostRoot

	// Debounce timer for fsnotify.Create events directly under this path.
	// Used for avoiding excessive Stats when many files/dirs are created at once.
	createDebounceTimer *time.Timer

	// Paths created since the last debounce timer expiration.
	createDebouncedPaths map[string]struct{}
}

func (we *WatchEntry) shouldRelocateWhenRemoved() bool {
	return len(we.lostRoots) > 0 || we.rootType == RecursiveRoot || we.rootType == NonRecursiveRoot
}

func (we *WatchEntry) removeLostRoot(targetPath string) {
	var newLostRoots []*LostRoot
	for _, lr := range we.lostRoots {
		if lr.targetPath != targetPath {
			newLostRoots = append(newLostRoots, lr)
		}
	}
	we.lostRoots = newLostRoots
}
