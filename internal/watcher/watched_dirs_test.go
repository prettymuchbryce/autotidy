package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/prettymuchbryce/autotidy/internal/fs"
)

// mockFsWatcher tracks Add/Remove calls for testing.
type mockFsWatcher struct {
	added     []string
	removed   []string
	addErr    error
	removeErr error
}

func newMockFsWatcher() *mockFsWatcher {
	return &mockFsWatcher{
		added:   []string{},
		removed: []string{},
	}
}

func (m *mockFsWatcher) Add(name string) error {
	if m.addErr != nil {
		return m.addErr
	}
	m.added = append(m.added, name)
	return nil
}

func (m *mockFsWatcher) Remove(name string) error {
	if m.removeErr != nil {
		return m.removeErr
	}
	m.removed = append(m.removed, name)
	return nil
}

func (m *mockFsWatcher) hasAdded(path string) bool {
	for _, p := range m.added {
		if p == path {
			return true
		}
	}
	return false
}

func (m *mockFsWatcher) hasRemoved(path string) bool {
	for _, p := range m.removed {
		if p == path {
			return true
		}
	}
	return false
}

// mockFsWatcherWithHook allows injecting a callback when Add is called.
type mockFsWatcherWithHook struct {
	*mockFsWatcher
	onAdd func(name string)
}

func newMockFsWatcherWithHook(onAdd func(name string)) *mockFsWatcherWithHook {
	return &mockFsWatcherWithHook{
		mockFsWatcher: newMockFsWatcher(),
		onAdd:         onAdd,
	}
}

func (m *mockFsWatcherWithHook) Add(name string) error {
	err := m.mockFsWatcher.Add(name)
	if err == nil && m.onAdd != nil {
		m.onAdd(name)
	}
	return err
}

// failingStatFs embeds a mem filesystem and can fail Stat calls in two modes:
// - failAtCount >= 0: fail only the Nth stat call (0-indexed), others succeed
// - failAtCount < 0 with statsUntilFail >= 0: fail all stats after N successful calls
type failingStatFs struct {
	fs.FileSystem
	statsUntilFail int // Number of Stats that will succeed before ALL fail (-1 = use failAtCount mode)
	failAtCount    int // Specific stat call to fail (-1 = use statsUntilFail mode)
	statCount      int
}

func newFailingStatFs(statsUntilFail int) *failingStatFs {
	return &failingStatFs{
		FileSystem:     fs.NewMem(),
		statsUntilFail: statsUntilFail,
		failAtCount:    -1,
	}
}

func newFailingStatFsAt(failAtCount int) *failingStatFs {
	return &failingStatFs{
		FileSystem:     fs.NewMem(),
		statsUntilFail: -1,
		failAtCount:    failAtCount,
	}
}

func (f *failingStatFs) Stat(name string) (os.FileInfo, error) {
	currentCount := f.statCount
	f.statCount++

	// Mode 1: fail only at specific count
	if f.failAtCount >= 0 && currentCount == f.failAtCount {
		return nil, os.ErrNotExist
	}

	// Mode 2: fail all after threshold
	if f.statsUntilFail >= 0 && currentCount >= f.statsUntilFail {
		return nil, os.ErrNotExist
	}

	return f.FileSystem.Stat(name)
}

func (f *failingStatFs) MustMkdirAll(path string) {
	if err := f.FileSystem.MkdirAll(path, 0755); err != nil {
		panic(fmt.Sprintf("MustMkdirAll(%q): %v", path, err))
	}
}

// testPath creates a cross-platform absolute path for testing.
// On Unix: testPath("a", "b") returns "/a/b"
// On Windows: testPath("a", "b") returns "C:\\a\\b"
func testPath(parts ...string) string {
	if runtime.GOOS == "windows" {
		// C: alone is relative, C:\ is absolute
		return "C:\\" + filepath.Join(parts...)
	}
	return filepath.Join(append([]string{"/"}, parts...)...)
}

// newTestWatchedDirs creates a WatchedDirs with mock watcher and in-memory fs.
func newTestWatchedDirs(t *testing.T) (*WatchedDirs, *mockFsWatcher, *fs.MemFileSystem) {
	t.Helper()
	memFs := fs.NewMemTest()
	mock := newMockFsWatcher()
	debounceChan := make(chan string, 10)
	rootsRecreatedChan := make(chan bool, 1)
	done := make(chan struct{})
	wd := NewWatchedDirs(memFs, mock, 50*time.Millisecond, debounceChan, rootsRecreatedChan, done)
	return wd, mock, memFs
}

// makeEvent creates a TimestampedEvent for testing.
func makeEvent(path string, op fsnotify.Op) TimestampedEvent {
	return TimestampedEvent{
		Event: fsnotify.Event{
			Name: path,
			Op:   op,
		},
		Time: time.Now(),
	}
}

// --- Basic Operations Tests ---

func TestAddRoot_ExistingDirectory(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	path := testPath("a", "b")

	memFs.MustMkdirAll(path)

	wd.AddRoot(path, false)

	if !mock.hasAdded(path) {
		t.Errorf("expected path %s to be added to fsWatcher, got: %v", path, mock.added)
	}
}

func TestAddRoot_NonExistentDirectory(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	ancestor := testPath("a")
	target := testPath("a", "b")

	memFs.MustMkdirAll(ancestor)

	wd.AddRoot(target, false)

	// Should watch ancestor since target doesn't exist
	if !mock.hasAdded(ancestor) {
		t.Errorf("expected ancestor %s to be added, got: %v", ancestor, mock.added)
	}
	if mock.hasAdded(target) {
		t.Errorf("should not have added non-existent target %s", target)
	}
}

func TestAddRoot_Recursive(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	root := testPath("a")
	sub1 := testPath("a", "b")
	sub2 := testPath("a", "b", "c")

	memFs.MustMkdirAll(sub2)

	wd.AddRoot(root, true)

	// Should watch root and all subdirs
	for _, p := range []string{root, sub1, sub2} {
		if !mock.hasAdded(p) {
			t.Errorf("expected %s to be added for recursive watch, got: %v", p, mock.added)
		}
	}
}

func TestAddRoot_Idempotent(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	wd.AddRoot(path, false)
	wd.AddRoot(path, false)

	// Should only be added once
	count := 0
	for _, p := range mock.added {
		if p == path {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected path to be added once, got %d times: %v", count, mock.added)
	}
}

func TestRemove_ExistingWatch(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	wd.AddRoot(path, false)
	wd.remove(path, nil)

	if !mock.hasRemoved(path) {
		t.Errorf("expected path %s to be removed, got: %v", path, mock.removed)
	}
}

func TestRemove_NonExistent(t *testing.T) {
	wd, mock, _ := newTestWatchedDirs(t)
	path := testPath("nonexistent")

	// Should be a no-op, not panic
	wd.remove(path, nil)

	if len(mock.removed) != 0 {
		t.Errorf("expected no removals for non-existent path, got: %v", mock.removed)
	}
}

func TestProcessEvent_Remove(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	wd.AddRoot(path, false)

	// Simulate directory removal
	wd.ProcessEvent(makeEvent(path, fsnotify.Remove))

	if !mock.hasRemoved(path) {
		t.Errorf("expected path to be removed after REMOVE event, got: %v", mock.removed)
	}
}

func TestProcessEvent_Rename(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	wd.AddRoot(path, false)

	// Simulate directory rename
	wd.ProcessEvent(makeEvent(path, fsnotify.Rename))

	if !mock.hasRemoved(path) {
		t.Errorf("expected path to be removed after RENAME event, got: %v", mock.removed)
	}
}

func TestGetRecreatedRoots_ClearsAfterRead(t *testing.T) {
	wd, _, memFs := newTestWatchedDirs(t)
	ancestor := testPath("a")
	target := testPath("a", "b")

	memFs.MustMkdirAll(ancestor)

	// Add root for non-existent target (becomes lost root)
	wd.AddRoot(target, false)

	// Now create target and simulate events
	memFs.MustMkdirAll(target)
	wd.ProcessEvent(makeEvent(target, fsnotify.Create))
	wd.EvaluateDebounced(ancestor)

	// First read should return the recreated root
	roots := wd.GetRecreatedRoots()
	if len(roots) == 0 {
		t.Errorf("expected recreated roots, got none")
	}

	// Second read should be empty
	roots2 := wd.GetRecreatedRoots()
	if len(roots2) != 0 {
		t.Errorf("expected empty after second read, got: %v", roots2)
	}
}

// --- Lost Root Handling Tests ---

func TestLostRoot_WatchesAncestor(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	ancestor := testPath("a")
	target := testPath("a", "b", "c")

	memFs.MustMkdirAll(ancestor)

	wd.AddRoot(target, false)

	if !mock.hasAdded(ancestor) {
		t.Errorf("expected ancestor %s to be watched, got: %v", ancestor, mock.added)
	}
}

func TestLostRoot_MovesCloser(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	ancestor := testPath("a")
	middle := testPath("a", "b")
	target := testPath("a", "b", "c")

	memFs.MustMkdirAll(ancestor)

	wd.AddRoot(target, false)

	// Now create middle directory
	memFs.MustMkdirAll(middle)
	wd.ProcessEvent(makeEvent(middle, fsnotify.Create))
	wd.EvaluateDebounced(ancestor)

	// Should now be watching middle, not just ancestor
	if !mock.hasAdded(middle) {
		t.Errorf("expected middle %s to be added after create, got: %v", middle, mock.added)
	}
}

func TestLostRoot_TargetCreated(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	ancestor := testPath("a")
	target := testPath("a", "b")

	memFs.MustMkdirAll(ancestor)

	wd.AddRoot(target, false)

	// Now create target
	memFs.MustMkdirAll(target)
	wd.ProcessEvent(makeEvent(target, fsnotify.Create))
	wd.EvaluateDebounced(ancestor)

	if !mock.hasAdded(target) {
		t.Errorf("expected target %s to be added, got: %v", target, mock.added)
	}
}

func TestLostRoot_SignalsRecreation(t *testing.T) {
	wd, _, memFs := newTestWatchedDirs(t)
	ancestor := testPath("a")
	target := testPath("a", "b")

	memFs.MustMkdirAll(ancestor)

	wd.AddRoot(target, false)

	// Create target
	memFs.MustMkdirAll(target)
	wd.ProcessEvent(makeEvent(target, fsnotify.Create))
	wd.EvaluateDebounced(ancestor)

	roots := wd.GetRecreatedRoots()
	if len(roots) != 1 {
		t.Errorf("expected 1 recreated root, got %d", len(roots))
	}
	if len(roots) > 0 && roots[0].Path != target {
		t.Errorf("expected recreated root path %s, got %s", target, roots[0].Path)
	}
}

func TestLostRoot_MultipleLostRoots(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	ancestor := testPath("a")
	target1 := testPath("a", "b")
	target2 := testPath("a", "c")

	memFs.MustMkdirAll(ancestor)

	wd.AddRoot(target1, false)
	wd.AddRoot(target2, false)

	// Create both targets
	memFs.MustMkdirAll(target1)
	memFs.MustMkdirAll(target2)

	wd.ProcessEvent(makeEvent(target1, fsnotify.Create))
	wd.ProcessEvent(makeEvent(target2, fsnotify.Create))
	wd.EvaluateDebounced(ancestor)

	if !mock.hasAdded(target1) {
		t.Errorf("expected target1 %s to be added, got: %v", target1, mock.added)
	}
	if !mock.hasAdded(target2) {
		t.Errorf("expected target2 %s to be added, got: %v", target2, mock.added)
	}
}

// --- Edge Cases Tests ---

func TestRemove_RelocatesLostRoots(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	root := testPath("a")
	middle := testPath("a", "b")
	target := testPath("a", "b", "c")

	memFs.MustMkdirAll(middle)

	// Add root for non-existent target (watching middle as lost root)
	wd.AddRoot(target, false)

	// Verify we're watching middle
	if !mock.hasAdded(middle) {
		t.Errorf("expected middle %s to be watched initially, got: %v", middle, mock.added)
	}

	// Simulate middle being removed
	memFs.MustRemoveAll(middle)
	wd.ProcessEvent(makeEvent(middle, fsnotify.Remove))

	// Should relocate to root
	if !mock.hasAdded(root) {
		t.Errorf("expected root %s to be added after relocation, got: %v", root, mock.added)
	}
}

func TestEvaluateDebounced_EntryRemoved(t *testing.T) {
	wd, _, _ := newTestWatchedDirs(t)

	// Call EvaluateDebounced on a path that was never added
	// Should hit the "watch entry no longer exists" case
	path := testPath("nonexistent")
	wd.EvaluateDebounced(path)

	// No panic means success
}

func TestRecursive_SubdirCreatedDuringWatch(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	root := testPath("a")
	newSubdir := testPath("a", "newdir")

	memFs.MustMkdirAll(root)

	wd.AddRoot(root, true)

	// Create new subdirectory
	memFs.MustMkdirAll(newSubdir)
	wd.ProcessEvent(makeEvent(newSubdir, fsnotify.Create))
	wd.EvaluateDebounced(root)

	if !mock.hasAdded(newSubdir) {
		t.Errorf("expected new subdir %s to be added, got: %v", newSubdir, mock.added)
	}
}

// --- Path Handling Tests ---

func TestIsPathCloserToTarget(t *testing.T) {
	tests := []struct {
		name     string
		cur      string
		target   string
		path     string
		expected bool
	}{
		{
			name:     "path equals target",
			cur:      testPath("a"),
			target:   testPath("a", "b"),
			path:     testPath("a", "b"),
			expected: true,
		},
		{
			name:     "path is on route to target",
			cur:      testPath("a"),
			target:   testPath("a", "b", "c"),
			path:     testPath("a", "b"),
			expected: true,
		},
		{
			name:     "path is sibling not on route",
			cur:      testPath("a"),
			target:   testPath("a", "b", "c"),
			path:     testPath("a", "d"),
			expected: false,
		},
		{
			name:     "path is not under cur",
			cur:      testPath("a"),
			target:   testPath("a", "b"),
			path:     testPath("x", "y"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPathCloserToTarget(tt.cur, tt.target, tt.path)
			if result != tt.expected {
				t.Errorf("isPathCloserToTarget(%q, %q, %q) = %v, want %v",
					tt.cur, tt.target, tt.path, result, tt.expected)
			}
		})
	}
}

func TestFindSuitableAncestor_FindsNearest(t *testing.T) {
	wd, _, memFs := newTestWatchedDirs(t)
	ancestor := testPath("a", "b")
	target := testPath("a", "b", "c", "d")

	memFs.MustMkdirAll(ancestor)

	result, found := wd.findSuitableAncestor(target, nil)

	if !found {
		t.Errorf("expected to find ancestor")
	}
	if result != ancestor {
		t.Errorf("expected %s, got %s", ancestor, result)
	}
}

func TestFindSuitableAncestor_SkipsAttempted(t *testing.T) {
	wd, _, memFs := newTestWatchedDirs(t)
	root := testPath("a")
	middle := testPath("a", "b")
	target := testPath("a", "b", "c")

	memFs.MustMkdirAll(middle)

	// Mark middle as attempted
	attempted := map[string]bool{
		target: true,
		middle: true,
	}

	result, found := wd.findSuitableAncestor(target, attempted)

	if !found {
		t.Errorf("expected to find ancestor")
	}
	if result != root {
		t.Errorf("expected root %s (skipping attempted middle), got %s", root, result)
	}
}

// --- Destroy Tests ---

func TestDestroy_StopsAllTimers(t *testing.T) {
	wd, _, memFs := newTestWatchedDirs(t)
	path1 := testPath("a")
	path2 := testPath("b")

	memFs.MustMkdirAll(path1)
	memFs.MustMkdirAll(path2)

	wd.AddRoot(path1, false)
	wd.AddRoot(path2, false)

	// Should not panic
	wd.Destroy()

	// Entries should still exist but timers stopped
	if len(wd.entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(wd.entries))
	}
}

// --- ProcessEvent Edge Cases ---

func TestProcessEvent_ChmodOnWatchedDir(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	wd.AddRoot(path, false)

	// Chmod event on existing directory - should remain watched
	wd.ProcessEvent(makeEvent(path, fsnotify.Chmod))

	// Should not be removed since dir still exists
	if mock.hasRemoved(path) {
		t.Errorf("should not have removed existing directory on chmod")
	}
}

func TestProcessEvent_ChmodOnDeletedDir(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	wd.AddRoot(path, false)

	// Delete the directory
	memFs.MustRemoveAll(path)

	// Chmod event on deleted directory - should be removed
	wd.ProcessEvent(makeEvent(path, fsnotify.Chmod))

	if !mock.hasRemoved(path) {
		t.Errorf("expected deleted directory to be removed on chmod, got: %v", mock.removed)
	}
}

func TestProcessEvent_CreateOnAlreadyWatchedDir(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	wd.AddRoot(path, false)
	initialAddCount := len(mock.added)

	// Create event on already-watched directory (defensive handling)
	wd.ProcessEvent(makeEvent(path, fsnotify.Create))

	// Should re-add the fsnotify watch (remove then add)
	if !mock.hasRemoved(path) {
		t.Errorf("expected path to be removed during re-add, got: %v", mock.removed)
	}
	if len(mock.added) <= initialAddCount {
		t.Errorf("expected path to be re-added, got: %v", mock.added)
	}
}

func TestProcessEvent_CreateOnAlreadyWatchedDir_DeletedMeanwhile(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	wd.AddRoot(path, false)

	// Delete the directory before processing Create event
	memFs.MustRemoveAll(path)

	// Create event on now-deleted directory
	wd.ProcessEvent(makeEvent(path, fsnotify.Create))

	// Should be removed since it doesn't exist
	if !mock.hasRemoved(path) {
		t.Errorf("expected deleted directory to be removed, got: %v", mock.removed)
	}
}

// --- EvaluateDebounced Edge Cases ---

func TestEvaluateDebounced_NonRecursiveNoLostRoots(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	wd.AddRoot(path, false)

	// Queue a create event
	subdir := testPath("a", "b")
	memFs.MustMkdirAll(subdir)
	wd.ProcessEvent(makeEvent(subdir, fsnotify.Create))

	// Evaluate - should do nothing for non-recursive with no lost roots
	wd.EvaluateDebounced(path)

	// Subdir should NOT be added (non-recursive)
	if mock.hasAdded(subdir) {
		t.Errorf("should not add subdirs for non-recursive watch")
	}
}

func TestEvaluateDebounced_StatError(t *testing.T) {
	wd, _, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	wd.AddRoot(path, true) // recursive to trigger stat

	// Queue a create event for non-existent path
	subdir := testPath("a", "b")
	wd.ProcessEvent(makeEvent(subdir, fsnotify.Create))

	// Subdir doesn't exist - should handle gracefully (no panic)
	// Remove is called but it's a no-op since subdir was never in entries
	wd.EvaluateDebounced(path)

	// Just verify no panic and path still exists
	if _, exists := wd.entries[path]; !exists {
		t.Errorf("parent path should still be watched")
	}
}

func TestEvaluateDebounced_ManyCreates_ReadDirPath(t *testing.T) {
	// Save original threshold and restore after test
	originalThreshold := statThreshold
	statThreshold = 2 // Lower threshold to trigger ReadDir path
	defer func() { statThreshold = originalThreshold }()

	wd, mock, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	wd.AddRoot(path, true)

	// Create multiple subdirs (>= threshold)
	subdir1 := testPath("a", "b")
	subdir2 := testPath("a", "c")
	subdir3 := testPath("a", "d")

	for _, sd := range []string{subdir1, subdir2, subdir3} {
		memFs.MustMkdirAll(sd)
		wd.ProcessEvent(makeEvent(sd, fsnotify.Create))
	}

	wd.EvaluateDebounced(path)

	// All subdirs should be added via ReadDir path
	for _, sd := range []string{subdir1, subdir2, subdir3} {
		if !mock.hasAdded(sd) {
			t.Errorf("expected subdir %s to be added via ReadDir path", sd)
		}
	}
}

func TestEvaluateDebounced_ReadDir_RemovesOnError(t *testing.T) {
	// Save original threshold and restore after test
	originalThreshold := statThreshold
	statThreshold = 2 // Lower threshold to trigger ReadDir path
	defer func() { statThreshold = originalThreshold }()

	wd, mock, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	wd.AddRoot(path, true)

	// Queue enough events to trigger ReadDir
	for i := 0; i < 3; i++ {
		subdir := testPath("a", string(rune('b'+i)))
		wd.ProcessEvent(makeEvent(subdir, fsnotify.Create))
	}

	// Delete the directory before EvaluateDebounced
	memFs.MustRemoveAll(path)

	// Should handle ReadDir error gracefully
	wd.EvaluateDebounced(path)

	if !mock.hasRemoved(path) {
		t.Errorf("expected path to be removed on ReadDir error")
	}
}

func TestEvaluateDebounced_CleansUpNonRecursiveAncestor(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	ancestor := testPath("a")
	target := testPath("a", "b")

	memFs.MustMkdirAll(ancestor)

	// Add non-existent target (ancestor becomes watched with lost root)
	wd.AddRoot(target, false)

	// Create target
	memFs.MustMkdirAll(target)
	wd.ProcessEvent(makeEvent(target, fsnotify.Create))
	wd.EvaluateDebounced(ancestor)

	// Ancestor should be removed since it's no longer needed
	// (not a root, not recursive, no lost roots)
	if !mock.hasRemoved(ancestor) {
		t.Errorf("expected ancestor to be removed after lost root restored")
	}
}

// --- addLostRoot Edge Cases ---

func TestAddLostRoot_FsWatcherAddError(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	ancestor := testPath("a")
	target := testPath("a", "b")

	memFs.MustMkdirAll(ancestor)

	// Set up error for Add
	mock.addErr = fsnotify.ErrNonExistentWatch

	// Should handle error gracefully
	wd.AddRoot(target, false)

	// Entry should not be created due to error
	if _, exists := wd.entries[ancestor]; exists {
		t.Errorf("entry should not exist after Add error")
	}
}

func TestAddLostRoot_RaceCondition_TargetExistsDuringAdd(t *testing.T) {
	// Test where target is created between finding ancestor and the race check.
	// We use the fsWatcher.Add hook to create the target, simulating the race.
	memFs := fs.NewMemTest()
	ancestor := testPath("a")
	target := testPath("a", "b")

	memFs.MustMkdirAll(ancestor)

	// Hook: when fsWatcher.Add is called for ancestor, create the target
	mock := newMockFsWatcherWithHook(func(name string) {
		if name == ancestor {
			memFs.MkdirAll(target, 0755)
		}
	})

	debounceChan := make(chan string, 10)
	rootsRecreatedChan := make(chan bool, 1)
	done := make(chan struct{})
	wd := NewWatchedDirs(memFs, mock, 50*time.Millisecond, debounceChan, rootsRecreatedChan, done)

	wd.AddRoot(target, false)

	// Target should be watched directly after race detection
	if !mock.hasAdded(target) {
		t.Errorf("expected target %s to be added after race detection, got: %v", target, mock.added)
	}

	// Should signal root recreation
	roots := wd.GetRecreatedRoots()
	if len(roots) != 1 || roots[0].Path != target {
		t.Errorf("expected recreated root for %s, got: %v", target, roots)
	}
}

func TestAddLostRoot_RaceCondition_CloserAncestorAppears(t *testing.T) {
	// Test where a closer ancestor appears between finding ancestor and the race check.
	memFs := fs.NewMemTest()
	root := testPath("a")
	middle := testPath("a", "b")
	target := testPath("a", "b", "c")

	memFs.MustMkdirAll(root)

	// Hook: when fsWatcher.Add is called for root, create middle (but not target)
	mock := newMockFsWatcherWithHook(func(name string) {
		if name == root {
			memFs.MkdirAll(middle, 0755)
		}
	})

	debounceChan := make(chan string, 10)
	rootsRecreatedChan := make(chan bool, 1)
	done := make(chan struct{})
	wd := NewWatchedDirs(memFs, mock, 50*time.Millisecond, debounceChan, rootsRecreatedChan, done)

	wd.AddRoot(target, false)

	// Middle should be watched (closer to target than root)
	if !mock.hasAdded(middle) {
		t.Errorf("expected middle %s to be added after race detection, got: %v", middle, mock.added)
	}
}

func TestAddLostRoot_RaceCondition_NoAncestorFound(t *testing.T) {
	// Test where findSuitableAncestor fails during the race check in addLostRoot.
	// We need stats to succeed initially but fail during the race check.
	ancestor := testPath("a")
	target := testPath("a", "b")

	// failAtCount=3: stats 0,1,2 succeed, stat 3+ fail
	// Stat 0: AddRoot initial stat(target) - fail (path doesn't exist)
	// Stat 1: findSuitableAncestor stat(ancestor) - succeed
	// Stat 2: fsWatcher.Add happens, then race check starts
	// The race check calls findSuitableAncestor with empty attempted map:
	// Stat 2: stat(target) - fail
	// Stat 3: stat(ancestor) - FAIL (triggers !found case)
	failingFs := newFailingStatFs(3)
	failingFs.MustMkdirAll(ancestor)

	mock := newMockFsWatcher()
	debounceChan := make(chan string, 10)
	rootsRecreatedChan := make(chan bool, 1)
	done := make(chan struct{})
	wd := NewWatchedDirs(failingFs, mock, 50*time.Millisecond, debounceChan, rootsRecreatedChan, done)

	// Should not panic - hits "no suitable ancestor found for missing lost root"
	wd.AddRoot(target, false)
}

// --- AddRoot Error Handling ---

func TestAddRoot_NoSuitableAncestor(t *testing.T) {
	// Use failingStatFs that fails all Stats to trigger "no suitable ancestor" error
	failingFs := newFailingStatFs(0)
	mock := newMockFsWatcher()
	debounceChan := make(chan string, 10)
	rootsRecreatedChan := make(chan bool, 1)
	done := make(chan struct{})
	wd := NewWatchedDirs(failingFs, mock, 50*time.Millisecond, debounceChan, rootsRecreatedChan, done)

	path := testPath("a", "b", "c")

	// Should not panic, just log error about no ancestor
	wd.AddRoot(path, false)

	// No entries should be created since all Stats fail
	if len(wd.entries) != 0 {
		t.Errorf("expected no entries when no ancestor found, got %d", len(wd.entries))
	}
}

func TestAddRoot_RaceInAddRoot(t *testing.T) {
	// This tests the race condition in addRoot where a path is deleted after
	// the initial stat succeeds but before the second stat (race check) in addRoot.
	// statsUntilFail=1 means first stat succeeds, second fails.
	failingFs := newFailingStatFs(1)
	mock := newMockFsWatcher()
	debounceChan := make(chan string, 10)
	rootsRecreatedChan := make(chan bool, 1)
	done := make(chan struct{})
	wd := NewWatchedDirs(failingFs, mock, 50*time.Millisecond, debounceChan, rootsRecreatedChan, done)

	path := testPath("a")
	failingFs.MustMkdirAll(path)

	wd.AddRoot(path, false)

	// Path should be removed after race condition detected
	if _, exists := wd.entries[path]; exists {
		t.Errorf("path should not be in entries after race condition removal")
	}
}

func TestAddRoot_RaceInAddRootOuter(t *testing.T) {
	// This tests the race condition in AddRoot (outer function) where a path
	// is deleted after addRoot returns but before the final stat check.
	// failAtCount=2 means only stat #2 (0-indexed) fails, others succeed.
	// Stat #0: AddRoot initial check - succeed
	// Stat #1: addRoot race check - succeed
	// Stat #2: AddRoot race check - FAIL
	// Stat #3+: findSuitableAncestor - succeed (finds "/" as ancestor)
	failingFs := newFailingStatFsAt(2)
	mock := newMockFsWatcher()
	debounceChan := make(chan string, 10)
	rootsRecreatedChan := make(chan bool, 1)
	done := make(chan struct{})
	wd := NewWatchedDirs(failingFs, mock, 50*time.Millisecond, debounceChan, rootsRecreatedChan, done)

	path := testPath("a")
	failingFs.MustMkdirAll(path)

	wd.AddRoot(path, false)

	// The race condition should trigger addLostRoot, creating an entry at the root
	// On Unix this is "/", on Windows this is "C:\"
	root := testPath()
	if !mock.hasAdded(root) {
		t.Errorf("expected root %s to be added as ancestor after race, got: %v", root, mock.added)
	}
}

func TestAddRoot_RaceInAddRootOuter_NoAncestor(t *testing.T) {
	// This tests the race condition in AddRoot where the race is detected
	// AND no suitable ancestor can be found.
	// statsUntilFail=2 means stats #0, #1 succeed, then ALL fail.
	// Stat #0: AddRoot initial check - succeed
	// Stat #1: addRoot race check - succeed
	// Stat #2: AddRoot race check - FAIL (race detected)
	// Stat #3+: findSuitableAncestor - ALL FAIL (no ancestor found)
	failingFs := newFailingStatFs(2)
	mock := newMockFsWatcher()
	debounceChan := make(chan string, 10)
	rootsRecreatedChan := make(chan bool, 1)
	done := make(chan struct{})
	wd := NewWatchedDirs(failingFs, mock, 50*time.Millisecond, debounceChan, rootsRecreatedChan, done)

	path := testPath("a")
	failingFs.MustMkdirAll(path)

	// Should not panic - hits "no suitable ancestor found" in race condition block
	wd.AddRoot(path, false)
}

// --- addRoot Idempotency ---

func TestAddRoot_Internal_AlreadyExistsAsRecursive(t *testing.T) {
	// Test the internal addRoot check by calling it directly
	wd, _, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	// First call creates the entry
	success1, created1 := wd.addRoot(path, true)
	if !success1 || !created1 {
		t.Errorf("first addRoot should succeed and create, got success=%v created=%v", success1, created1)
	}

	// Second call should hit "Already exists as recursive root, no-op"
	success2, created2 := wd.addRoot(path, true)
	if success2 || created2 {
		t.Errorf("second addRoot should return false/false, got success=%v created=%v", success2, created2)
	}
}

func TestAddRoot_Internal_AlreadyExistsAsNonRecursive(t *testing.T) {
	// Test the internal addRoot check by calling it directly
	wd, _, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	// First call creates the entry
	success1, created1 := wd.addRoot(path, false)
	if !success1 || !created1 {
		t.Errorf("first addRoot should succeed and create, got success=%v created=%v", success1, created1)
	}

	// Second call should hit "Already exists as non-recursive root, no-op"
	success2, created2 := wd.addRoot(path, false)
	if success2 || created2 {
		t.Errorf("second addRoot should return false/false, got success=%v created=%v", success2, created2)
	}
}

func TestAddRoot_AlreadyExistsAsRecursive(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	// Add as recursive
	wd.AddRoot(path, true)
	initialAddCount := len(mock.added)

	// Try to add again as recursive
	wd.AddRoot(path, true)

	// Should not add again
	if len(mock.added) != initialAddCount {
		t.Errorf("should not re-add existing recursive root")
	}
}

func TestAddRoot_AlreadyExistsAsNonRecursive(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	// Add as non-recursive
	wd.AddRoot(path, false)
	initialAddCount := len(mock.added)

	// Try to add again as non-recursive
	wd.AddRoot(path, false)

	// Should not add again
	if len(mock.added) != initialAddCount {
		t.Errorf("should not re-add existing non-recursive root")
	}
}

func TestAddRoot_NonRecursiveThenRecursive(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	path := testPath("a")
	subdir := testPath("a", "b")

	memFs.MustMkdirAll(subdir)

	// Add as non-recursive first
	wd.AddRoot(path, false)

	// Then upgrade to recursive
	wd.AddRoot(path, true)

	// Should now watch subdirectories
	if !mock.hasAdded(subdir) {
		t.Errorf("expected subdir to be added when upgrading to recursive")
	}
}

func TestAddRoot_RecursiveThenNonRecursive(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	// Add as recursive first
	wd.AddRoot(path, true)
	initialAddCount := len(mock.added)

	// Try to add as non-recursive (should be no-op since recursive > non-recursive)
	wd.AddRoot(path, false)

	// Should not add again
	if len(mock.added) != initialAddCount {
		t.Errorf("should not re-add when already recursive")
	}
}

func TestAddRoot_FsWatcherAddError(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	mock.addErr = fsnotify.ErrNonExistentWatch

	wd.AddRoot(path, false)

	// Entry should not be kept on error
	if _, exists := wd.entries[path]; exists {
		t.Errorf("entry should not exist after Add error")
	}
}

// --- recursivelyAddSubdirectories Errors ---

func TestRecursivelyAddSubdirectories_ReadDirError(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	wd.AddRoot(path, false) // Non-recursive first

	// Remove directory to cause ReadDir error
	memFs.MustRemoveAll(path)

	// Directly call recursivelyAddSubdirectories
	wd.recursivelyAddSubdirectories(path)

	// Should remove the entry on error
	if !mock.hasRemoved(path) {
		t.Errorf("expected path to be removed on ReadDir error")
	}
}

func TestRecursivelyAddSubdirectories_FsWatcherAddError(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	root := testPath("a")
	subdir := testPath("a", "b")

	memFs.MustMkdirAll(subdir)

	// Add root first without error
	wd.AddRoot(root, false)

	// Now set error for next Add
	mock.addErr = fsnotify.ErrNonExistentWatch

	// Try to recursively add (will fail for subdir)
	wd.recursivelyAddSubdirectories(subdir)

	// Should have attempted to remove
	if !mock.hasRemoved(subdir) {
		t.Errorf("expected subdir to be removed on Add error")
	}
}

// --- Remove Edge Cases ---

func TestRemove_FsWatcherRemoveError(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	path := testPath("a")

	memFs.MustMkdirAll(path)

	wd.AddRoot(path, false)

	// Set remove error
	mock.removeErr = fsnotify.ErrNonExistentWatch

	// Delete the directory so it becomes a lost root that relocates
	memFs.MustRemoveAll(path)

	// Should not panic, just warn
	wd.remove(path, nil)

	// Entry at path should be removed (even though fsWatcher.Remove fails)
	// But a new entry might be created for the ancestor (for lost root relocation)
	if _, exists := wd.entries[path]; exists {
		t.Errorf("entry at path should be removed even if fsWatcher.Remove fails")
	}
}

func TestRemove_RelocatesRoot(t *testing.T) {
	wd, mock, memFs := newTestWatchedDirs(t)
	root := testPath("a")
	subdir := testPath("a", "b")

	memFs.MustMkdirAll(subdir)

	// Add subdir as root
	wd.AddRoot(subdir, false)

	// Delete subdir
	memFs.MustRemoveAll(subdir)

	// Process remove event
	wd.ProcessEvent(makeEvent(subdir, fsnotify.Remove))

	// Should relocate to root
	if !mock.hasAdded(root) {
		t.Errorf("expected root %s to be watched after subdir removed, got: %v", root, mock.added)
	}
}
