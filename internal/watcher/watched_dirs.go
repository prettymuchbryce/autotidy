package watcher

import (
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/afero"

	"github.com/prettymuchbryce/autotidy/internal/fs"
)

var statThreshold = 3

// fsnotifyWatcher is the interface for fsnotify operations, allowing mocking in tests.
type fsnotifyWatcher interface {
	Add(name string) error
	Remove(name string) error
}

// WatchedDirs manages filesystem watches and tracks their state.
// It handles the complexity of watching directories that may not exist yet (lost roots)
// and recursively watching subdirectories.
type WatchedDirs struct {
	// fs is the filesystem abstraction for stat and readdir operations.
	fs fs.FileSystem

	// fsWatcher is the underlying fsnotify watcher.
	// Note: fsnotify auto-removes watches on delete (all platforms), but not on rename for Windows.
	// We explicitly remove watches on delete to keep our state consistent.
	fsWatcher fsnotifyWatcher

	// entries stores WatchEntry objects indexed by path.
	entries map[string]*WatchEntry

	// debounceDelay is how long to wait after a create event before processing.
	debounceDelay time.Duration

	// debounceChan receives paths when their debounce timer fires.
	debounceChan chan string

	// rootsRecreatedChan signals when lost roots have been recreated.
	// A single true value is sent; the actual roots are retrieved via GetRecreatedRoots.
	rootsRecreatedChan chan bool

	// rootsRecreated accumulates recreated roots until retrieved.
	rootsRecreated []RecreatedRoot

	// done is closed when the watcher is stopping, to unblock timer goroutines.
	done chan struct{}
}

// NewWatchedDirs creates a new WatchedDirs manager.
func NewWatchedDirs(filesystem fs.FileSystem, fsWatcher fsnotifyWatcher, debounceDelay time.Duration, debounceChan chan string, rootsRecreatedChan chan bool, done chan struct{}) *WatchedDirs {
	return &WatchedDirs{
		fs:                 filesystem,
		fsWatcher:          fsWatcher,
		entries:            make(map[string]*WatchEntry),
		debounceDelay:      debounceDelay,
		debounceChan:       debounceChan,
		rootsRecreatedChan: rootsRecreatedChan,
		done:               done,
	}
}

// Destroy stops all debounce timers. Does not close the fsWatcher since that
// is owned by the parent Watcher.
func (w *WatchedDirs) Destroy() {
	for _, we := range w.entries {
		we.createDebounceTimer.Stop()
	}
}

// GetRecreatedRoots returns and clears the list of roots that were recreated.
func (w *WatchedDirs) GetRecreatedRoots() []RecreatedRoot {
	recreated := w.rootsRecreated
	w.rootsRecreated = nil
	return recreated
}

// WatchCount returns the number of directories currently being watched.
func (w *WatchedDirs) WatchCount() int {
	return len(w.entries)
}

// ProcessEvent handles an fsnotify event, updating watch state as needed.
// For watched directories: handles removal, rename, and unexpected create events.
// For children of watched directories: queues create events for debounced processing.
func (w *WatchedDirs) ProcessEvent(te TimestampedEvent) {
	event := te.Event
	path := event.Name
	pathParent := filepath.Dir(path)

	slog.Debug("ProcessEvent", "path", path, "op", event.Op)

	wePath, wePathExists := w.entries[path]
	wePathParent, wePathParentExists := w.entries[pathParent]

	if wePathExists {
		if event.Op&fsnotify.Chmod != 0 {
			// If this is a watched directory,
			// make sure it is still reachable.
			w.removeWatchEntryIfUnreachable(wePath)
		} else if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
			w.remove(path, nil)
		} else if event.Op&fsnotify.Create != 0 {
			// The directory is already being watched so this shouldn't happen.
			// We should re-add the fsnotify path just to be safe.
			w.readdFsnotify(wePath)
		}
	}

	if wePathParentExists {
		if event.Op&fsnotify.Create != 0 {
			wePathParent.createDebouncedPaths[path] = struct{}{}
			wePathParent.createDebounceTimer.Reset(w.debounceDelay)
		}
	}
}

// removeWatchEntryIfUnreachable removes the watch entry if the path no longer exists or is not a directory.
// Returns true if the entry was removed.
func (w *WatchedDirs) removeWatchEntryIfUnreachable(we *WatchEntry) bool {
	info, err := w.fs.Stat(we.path)
	if err != nil || !info.IsDir() {
		w.remove(we.path, nil)
		return true
	}
	return false
}

// readdFsnotify re-registers the fsnotify watch for a path.
// Used when we receive a create event for an already-watched directory, which
// shouldn't happen but we handle defensively.
func (w *WatchedDirs) readdFsnotify(we *WatchEntry) {
	removed := w.removeWatchEntryIfUnreachable(we)
	if !removed {
		w.fsWatcher.Remove(we.path)
		err := w.fsWatcher.Add(we.path)
		if err != nil {
			slog.Warn("fswatcher failed to re-add watch", "path", we.path, "error", err)
			w.remove(we.path, nil)
		}
	}
}

// EvaluateDebounced processes debounced create events for a watched directory.
// For recursive watches: adds watches to new subdirectories.
// For lost roots: checks if any created directory moves us closer to the target,
// potentially restoring the original watch.
func (w *WatchedDirs) EvaluateDebounced(path string) {
	slog.Debug("EvaluateDebounced called", "path", path)
	// It's possible the watch entry was removed while waiting for the debounce timer.
	we, exists := w.entries[path]
	if !exists {
		slog.Debug("EvaluateDebounced: watch entry no longer exists", "path", path)
		return
	}

	var newDirs []string

	if !we.isRecursive && len(we.lostRoots) == 0 {
		// Nothing to do.
		slog.Debug("EvaluateDebounced: nothing to do (not recursive, no lost roots)", "path", path)
		we.createDebouncedPaths = make(map[string]struct{})
		return
	}

	numCreates := len(we.createDebouncedPaths)
	if numCreates < statThreshold {
		// Stat each path individually.
		for createdPath := range we.createDebouncedPaths {
			info, err := w.fs.Stat(createdPath)
			if err != nil {
				// Remove just in case
				w.remove(createdPath, nil)
			} else if info.IsDir() {
				newDirs = append(newDirs, createdPath)
			}
		}
	} else {
		// Read the directory once.
		entries, err := afero.ReadDir(w.fs, we.path)
		if err != nil {
			slog.Warn("failed to read directory during debounce evaluation", "path", we.path, "error", err)
			// Remove just in case
			w.remove(we.path, nil)
			return
		}
		for _, entry := range entries {
			if entry.IsDir() {
				subdirPath := filepath.Join(we.path, entry.Name())
				for createdPath := range we.createDebouncedPaths {
					if subdirPath == createdPath {
						newDirs = append(newDirs, createdPath)
					}
				}
			}
		}
	}

	slog.Debug("EvaluateDebounced: found new dirs", "path", we.path, "newDirs", newDirs)
	for _, dir := range newDirs {
		if we.isRecursive {
			w.recursivelyAddSubdirectories(dir)
		}

		for _, lr := range we.lostRoots {
			slog.Debug("EvaluateDebounced: checking lost root", "dir", dir, "lrTargetPath", lr.targetPath, "isCloser", isPathCloserToTarget(we.path, lr.targetPath, dir))
			if isPathCloserToTarget(we.path, lr.targetPath, dir) {
				we.removeLostRoot(lr.targetPath)
				if dir == lr.targetPath {
					slog.Debug("EvaluateDebounced: target reached, calling addRoot", "dir", dir)
					success, _ := w.addRoot(dir, lr.isRecursive)
					if success {
						w.rootsRecreated = append(w.rootsRecreated, RecreatedRoot{Path: dir, Time: time.Now()})
						if len(w.rootsRecreatedChan) == 0 {
							w.rootsRecreatedChan <- true
						}
					}
				} else {
					slog.Debug("EvaluateDebounced: moving closer, calling addLostRoot", "dir", dir, "targetPath", lr.targetPath)
					w.addLostRoot(dir, lr.targetPath, lr.isRecursive, nil)
				}
			}
		}
	}

	we.createDebouncedPaths = make(map[string]struct{})

	if len(we.lostRoots) == 0 && we.rootType == None && !we.isRecursive {
		// No longer need to watch this path.
		w.remove(we.path, nil)
	}
}

// addLostRoot watches an ancestor path while waiting for targetPath to be created.
// The attempted map tracks paths we've already tried to prevent infinite loops
// when directories are rapidly created/deleted.
func (w *WatchedDirs) addLostRoot(path string, targetPath string, isRecursive bool, attempted map[string]bool) {
	slog.Debug("addLostRoot called", "path", path, "targetPath", targetPath, "isRecursive", isRecursive)
	if attempted == nil {
		attempted = make(map[string]bool)
	}
	attempted[path] = true

	we, exists := w.entries[path]
	slog.Debug("addLostRoot: entry lookup", "path", path, "exists", exists)
	if !exists {
		we = w.newWatchEntry(path)
	}

	// invariant: There should never be multiple LostRoots for the same target path under the same ancestor.
	for _, lr := range we.lostRoots {
		if lr.targetPath == targetPath {
			// Already exists as lost root. This should not happen.
			// slog.Error("lost root already exists", "path", path)
			panic("lost root already exists at path: " + path)
		}
	}

	we.lostRoots = append(we.lostRoots, &LostRoot{
		targetPath:  targetPath,
		isRecursive: isRecursive,
	})

	if !exists {
		err := w.fsWatcher.Add(path)
		if err != nil {
			w.remove(path, attempted)
			return
		}
		// Race condition: For lost roots, check if we can get any closer to the target path
		// since adding the watch. It may have been created before we started watching.
		ancestorPath, found := w.findSuitableAncestor(targetPath, make(map[string]bool))
		slog.Debug("addLostRoot race check", "path", path, "targetPath", targetPath, "ancestorPath", ancestorPath, "found", found)
		if !found {
			// There's nothing we can do here.
			slog.Error("no suitable ancestor found for missing lost root", "path", path)
			w.remove(path, attempted)
			return
		}
		if path != ancestorPath {
			// We found a closer ancestor.
			// Remove this entry and add the lost root to the closer ancestor.
			slog.Debug("race condition: moving closer to target", "from", path, "to", ancestorPath, "targetPath", targetPath)
			err := w.fsWatcher.Remove(path)
			if err != nil {
				slog.Warn("fswatcher failed to remove watch", "path", path, "error", err)
			}
			we.createDebounceTimer.Stop()
			delete(w.entries, path)
			if ancestorPath != targetPath {
				w.addLostRoot(ancestorPath, targetPath, isRecursive, attempted)
			} else {
				slog.Debug("race condition: target exists, calling addRoot", "targetPath", targetPath)
				success, _ := w.addRoot(targetPath, isRecursive)
				if success {
					w.rootsRecreated = append(w.rootsRecreated, RecreatedRoot{Path: targetPath, Time: time.Now()})
					if len(w.rootsRecreatedChan) == 0 {
						w.rootsRecreatedChan <- true
					}
				}
			}
		}
	}
}

// AddRoot adds a root watch for a path. If the path doesn't exist, watches the
// nearest existing ancestor and waits for the path to be created.
func (w *WatchedDirs) AddRoot(path string, isRecursive bool) {
	we, exists := w.entries[path]
	if exists && ((isRecursive && we.rootType == RecursiveRoot) || (!isRecursive && we.rootType != None)) {
		// Already exists, no-op
		return
	}

	// Stat the directory to make sure it exists and is a directory
	info, err := w.fs.Stat(path)
	if err != nil || !info.IsDir() {
		attempted := make(map[string]bool)
		attempted[path] = true
		ancestorPath, found := w.findSuitableAncestor(path, attempted)
		if !found {
			// There's nothing we can do here.
			slog.Error("no suitable ancestor found for missing root", "path", path)
			return
		}

		w.addLostRoot(ancestorPath, path, isRecursive, attempted)
		return
	}

	_, weCreated := w.addRoot(path, isRecursive)

	// race condition: Stat again to make sure it still exists. It may have been deleted
	// before we added the watch.
	if weCreated {
		info, err = w.fs.Stat(path)
		if err != nil || !info.IsDir() {
			attempted := make(map[string]bool)
			attempted[path] = true
			ancestorPath, found := w.findSuitableAncestor(path, attempted)
			if !found {
				// There's nothing we can do here.
				slog.Error("no suitable ancestor found for missing root", "path", path)
				return
			}

			w.addLostRoot(ancestorPath, path, isRecursive, attempted)
			return
		}
	}
}

// addRoot is the internal implementation that adds a root watch.
// Returns success=true if the watch was added, weCreated=true if a new entry was created.
func (w *WatchedDirs) addRoot(path string, isRecursive bool) (success bool, weCreated bool) {
	slog.Debug("addRoot called", "path", path, "isRecursive", isRecursive)
	we, exists := w.entries[path]
	if !exists {
		we = w.newWatchEntry(path)
	}

	wasRecursive := we.isRecursive

	if isRecursive {
		if exists && we.rootType == RecursiveRoot {
			// Already exists as recursive root, no-op
			return false, false
		}
		we.rootType = RecursiveRoot
		we.isRecursive = true
	} else {
		if exists && we.rootType == NonRecursiveRoot {
			// Already exists as non-recursive root, no-op
			return false, false
		}
		we.rootType = NonRecursiveRoot
		we.isRecursive = false
	}

	if !exists {
		err := w.fsWatcher.Add(path)
		if err != nil {
			slog.Warn("fswatcher failed to add root watch", "path", path, "error", err)
			w.remove(path, nil)
			return false, false
		}

		// race condition: Stat again to make sure it still exists. It may have been deleted
		// before we added the watch.
		info, err := w.fs.Stat(path)
		if err != nil || !info.IsDir() {
			slog.Warn("root path no longer exists after adding watch", "path", path)
			w.remove(path, nil)
			return false, false
		}
	}

	if isRecursive && !wasRecursive {
		w.recursivelyAddSubdirectories(path)
	}

	return true, !exists
}

// recursivelyAddSubdirectories adds watches to all subdirectories under path.
// Used when a recursive root is added or when new directories are created under one.
func (w *WatchedDirs) recursivelyAddSubdirectories(path string) {
	dirs, err := afero.ReadDir(w.fs, path)
	if err != nil {
		slog.Warn("failed to read directory when adding subdirectory", "path", path, "error", err)
		// Remove just in case
		w.remove(path, nil)
		return
	}

	we, exists := w.entries[path]
	if !exists {
		we = w.newWatchEntry(path)
	}

	we.isRecursive = true

	if !exists {
		err = w.fsWatcher.Add(path)
		if err != nil {
			slog.Warn("fswatcher failed to add subdirectory watch", "path", path, "error", err)
			// Remove just in case
			w.remove(path, nil)
			return
		}

		// note: No need to stat again here since the parent entry
		// must exist and be receiving events for us to be recursing into it.
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			subdirPath := filepath.Join(path, dir.Name())
			w.recursivelyAddSubdirectories(subdirPath)
		}
	}
}

// Remove stops watching a path and relocates any roots or lost roots to an ancestor.
// The attempted map prevents infinite loops during relocation.
// Note: multiple Remove events for the same path are possible (e.g., deleting /a/b
// when both /a and /a/b are watched).
func (w *WatchedDirs) remove(path string, attempted map[string]bool) {
	we, exists := w.entries[path]
	// Removes of a non-existent path is a no-op.
	if !exists {
		return
	}

	if attempted == nil {
		attempted = make(map[string]bool)
		attempted[path] = true
	}

	// Remove from fsWatcher
	err := w.fsWatcher.Remove(path)
	if err != nil {
		slog.Warn("fswatcher failed to remove watch", "path", path, "error", err)
	}

	we.createDebounceTimer.Stop()

	delete(w.entries, path)

	if we.shouldRelocateWhenRemoved() {
		ancestorPath, found := w.findSuitableAncestor(path, attempted)
		if found == false {
			// There's nothing we can do here.
			slog.Error("no suitable ancestor found for relocated watch entries", "path", path)
			return
		}

		if we.rootType != None {
			w.addLostRoot(ancestorPath, path, we.isRecursive, attempted)
		}

		for _, lr := range we.lostRoots {
			w.addLostRoot(ancestorPath, lr.targetPath, lr.isRecursive, attempted)
		}
	}
}

// newWatchEntry creates a new WatchEntry for a path and adds it to the entries map.
// Panics if an entry already exists for the path.
func (w *WatchedDirs) newWatchEntry(path string) *WatchEntry {
	slog.Debug("newWatchEntry called", "path", path)
	we, exists := w.entries[path]
	if exists {
		panic("watch entry already exists for path: " + path)
	}

	we = &WatchEntry{
		path:                 path,
		rootType:             None,
		isRecursive:          false,
		lostRoots:            []*LostRoot{},
		createDebouncedPaths: make(map[string]struct{}),
	}

	t := time.AfterFunc(time.Hour, func() {
		select {
		case w.debounceChan <- path:
		case <-w.done:
		}
	})
	t.Stop()

	we.createDebounceTimer = t
	w.entries[path] = we

	return we
}

// isPathCloserToTarget returns true if path is a child of cur and is on the path to target.
// Used to determine if a newly created directory moves us closer to a lost root's target.
func isPathCloserToTarget(cur, target, path string) bool {
	curWithSep := cur + string(filepath.Separator)
	if !strings.HasPrefix(path, curWithSep) {
		return false
	}

	if path == target {
		return true
	}

	pathWithSep := path + string(filepath.Separator)
	return strings.HasPrefix(target, pathWithSep)
}

// findSuitableAncestor walks up from targetPath to find the nearest existing directory.
// Skips paths in the attempted map. Returns the path and true if found, or "", false if
// we reach the filesystem root without finding an existing directory.
func (w *WatchedDirs) findSuitableAncestor(targetPath string, attempted map[string]bool) (string, bool) {
	if attempted == nil {
		attempted = make(map[string]bool)
		attempted[targetPath] = true
	}

	current := targetPath
	for {
		if !attempted[current] {
			info, err := w.fs.Stat(current)
			if err == nil && info.IsDir() {
				return current, true
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached root without finding existing directory.
			// There are no suitable ancestors.
			return "", false
		}
		current = parent
	}
}
