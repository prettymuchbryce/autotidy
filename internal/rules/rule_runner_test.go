package rules

import (
	"os"
	"testing"

	"github.com/prettymuchbryce/autotidy/internal/fs"
	"github.com/prettymuchbryce/autotidy/internal/testutil"

	"github.com/spf13/afero"
)

func TestRuleRunner_Execute_DisabledRule(t *testing.T) {
	r := &Rule{
		Name:    "disabled-rule",
		Enabled: boolPtr(false),
	}
	runner := NewRuleRunner(r, fs.NewMem(), nil)

	_, err := runner.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRuleRunner_Execute_NonExistentLocation(t *testing.T) {
	filesystem := fs.NewMem()

	r := &Rule{
		Name:      "test-rule",
		Locations: StringList{"/nonexistent"},
	}
	runner := NewRuleRunner(r, filesystem, nil)

	_, err := runner.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRuleRunner_Execute_Directory(t *testing.T) {
	filesystem := fs.NewMem()
	filesystem.MkdirAll("/root", 0755)
	afero.WriteFile(filesystem, "/root/a.txt", []byte("a"), 0644)
	afero.WriteFile(filesystem, "/root/b.txt", []byte("b"), 0644)

	var processedPaths []string
	mockAction := &Action{
		Name: "mock",
		Inner: &testExecutable{
			onExecute: func(path string) {
				processedPaths = append(processedPaths, path)
			},
		},
	}

	r := &Rule{
		Name:      "test-rule",
		Locations: StringList{"/root"},
		Actions:   []Action{*mockAction},
	}
	runner := NewRuleRunner(r, filesystem, nil)

	_, err := runner.Execute()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(processedPaths) != 2 {
		t.Errorf("expected 2 paths processed, got %d: %v", len(processedPaths), processedPaths)
	}
}

func TestRuleRunner_ExecuteOnItem_ActionChaining(t *testing.T) {
	// When an action returns a new path, subsequent actions should use that path
	originalPath := testutil.Path("/", "original", "file.txt")
	movedPath := testutil.Path("/", "moved", "file.txt")
	var paths []string

	action1 := &Action{
		Name: "action1",
		Inner: &testExecutable{
			result: &ExecutionResult{NewPath: movedPath},
			onExecute: func(path string) {
				paths = append(paths, path)
			},
		},
	}

	action2 := &Action{
		Name: "action2",
		Inner: &testExecutable{
			onExecute: func(path string) {
				paths = append(paths, path)
			},
		},
	}

	r := &Rule{
		Name:    "test-rule",
		Actions: []Action{*action1, *action2},
	}
	runner := NewRuleRunner(r, fs.NewNoop(), nil)

	result, _, err := runner.executeOnItem(originalPath)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(paths) != 2 {
		t.Fatalf("expected 2 action calls, got %d", len(paths))
	}

	if paths[0] != originalPath {
		t.Errorf("first action got path %q, want %q", paths[0], originalPath)
	}

	if paths[1] != movedPath {
		t.Errorf("second action got path %q, want %q", paths[1], movedPath)
	}

	if result == nil || result.NewPath != movedPath {
		t.Errorf("final path = %v, want %q", result, movedPath)
	}
}

// testExecutable with callback for testing
type testExecutable struct {
	result    *ExecutionResult
	err       error
	onExecute func(path string)
}

func (m *testExecutable) Execute(path string, _ fs.FileSystem) (*ExecutionResult, error) {
	if m.onExecute != nil {
		m.onExecute(path)
	}
	return m.result, m.err
}

func Test_isFilesystemError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "PathError is filesystem error",
			err:      &os.PathError{Op: "open", Path: "/test", Err: os.ErrNotExist},
			expected: true,
		},
		{
			name:     "LinkError is filesystem error",
			err:      &os.LinkError{Op: "rename", Old: "/old", New: "/new", Err: os.ErrPermission},
			expected: true,
		},
		{
			name:     "generic error is not filesystem error",
			err:      os.ErrInvalid,
			expected: false,
		},
		{
			name:     "nil is not filesystem error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isFilesystemError(tt.err); got != tt.expected {
				t.Errorf("isFilesystemError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRuleRunner_ExecuteOnItem_FilesystemErrorSkipsItem(t *testing.T) {
	// When an action returns a filesystem error, the entire item should be skipped
	// (no subsequent actions run, nil result returned)
	var executedActions []string

	action1 := &Action{
		Name: "action1",
		Inner: &testExecutable{
			err: &os.PathError{Op: "rename", Path: "/test", Err: os.ErrPermission},
			onExecute: func(path string) {
				executedActions = append(executedActions, "action1")
			},
		},
	}

	action2 := &Action{
		Name: "action2",
		Inner: &testExecutable{
			result: &ExecutionResult{NewPath: "/new/path"},
			onExecute: func(path string) {
				executedActions = append(executedActions, "action2")
			},
		},
	}

	r := &Rule{
		Name:    "test-rule",
		Actions: []Action{*action1, *action2},
	}
	runner := NewRuleRunner(r, fs.NewNoop(), nil)

	result, hadError, err := runner.executeOnItem("/original/file.txt")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Only action1 should have been executed before the fs error caused skip
	if len(executedActions) != 1 {
		t.Errorf("expected 1 action executed, got %d: %v", len(executedActions), executedActions)
	}

	// Should indicate an error occurred
	if !hadError {
		t.Error("expected hadError to be true")
	}

	// Result should be nil (item skipped)
	if result != nil {
		t.Errorf("expected nil result for skipped item, got %v", result)
	}
}

func TestRuleRunner_ExecuteOnItem_NonFilesystemErrorStops(t *testing.T) {
	// Non-filesystem errors should stop execution and return the error
	var executedActions []string

	action1 := &Action{
		Name: "action1",
		Inner: &testExecutable{
			err: os.ErrInvalid, // Not a PathError or LinkError
			onExecute: func(path string) {
				executedActions = append(executedActions, "action1")
			},
		},
	}

	action2 := &Action{
		Name: "action2",
		Inner: &testExecutable{
			onExecute: func(path string) {
				executedActions = append(executedActions, "action2")
			},
		},
	}

	r := &Rule{
		Name:    "test-rule",
		Actions: []Action{*action1, *action2},
	}
	runner := NewRuleRunner(r, fs.NewNoop(), nil)

	_, _, err := runner.executeOnItem("/original/file.txt")
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Only action1 should have been executed
	if len(executedActions) != 1 {
		t.Errorf("expected 1 action executed, got %d: %v", len(executedActions), executedActions)
	}
}

// fsErrorEvaluable returns a filesystem error for filter testing
type fsErrorEvaluable struct{}

func (e *fsErrorEvaluable) Evaluate(path string) (bool, error) {
	return false, &os.PathError{Op: "open", Path: path, Err: os.ErrPermission}
}

func TestRuleRunner_ExecuteOnItem_FilterFilesystemErrorSkipsItem(t *testing.T) {
	// When a filter returns a filesystem error, the item should be skipped
	actionCalled := false

	action := &Action{
		Name: "action",
		Inner: &testExecutable{
			onExecute: func(path string) {
				actionCalled = true
			},
		},
	}

	r := &Rule{
		Name:    "test-rule",
		Actions: []Action{*action},
		Filters: &FilterGroups{
			Exprs: []*FilterExpr{
				{Filters: []Filter{{Name: "fs-error", Inner: &fsErrorEvaluable{}}}},
			},
		},
	}
	runner := NewRuleRunner(r, fs.NewNoop(), nil)

	result, hadError, err := runner.executeOnItem("/test/file.txt")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if actionCalled {
		t.Error("action should not be called when filter has filesystem error")
	}

	if !hadError {
		t.Error("expected hadError to be true")
	}

	if result != nil {
		t.Errorf("expected nil result for skipped item, got %v", result)
	}
}

func TestRuleRunner_ExecuteOnItem_ConflictAlreadyExistsStopsSubsequentActions(t *testing.T) {
	// Regression test: When an action returns ConflictAlreadyExists: true,
	// subsequent actions should NOT run.
	var executedActions []string

	// First action returns ConflictAlreadyExists: true (simulating a move/copy/rename that skipped due to conflict)
	skipAction := &Action{
		Name: "skip-action",
		Inner: &testExecutable{
			result: &ExecutionResult{ConflictAlreadyExists: true},
			onExecute: func(path string) {
				executedActions = append(executedActions, "skip-action")
			},
		},
	}

	// Second action should NOT run
	dangerousAction := &Action{
		Name: "dangerous-action",
		Inner: &testExecutable{
			result: &ExecutionResult{Deleted: true},
			onExecute: func(path string) {
				executedActions = append(executedActions, "dangerous-action")
			},
		},
	}

	r := &Rule{
		Name:    "test-rule",
		Actions: []Action{*skipAction, *dangerousAction},
	}
	runner := NewRuleRunner(r, fs.NewNoop(), nil)

	result, _, err := runner.executeOnItem("/original/file.txt")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Only the skip action should have been executed
	if len(executedActions) != 1 {
		t.Errorf("expected 1 action executed, got %d: %v", len(executedActions), executedActions)
	}

	if executedActions[0] != "skip-action" {
		t.Errorf("expected skip-action to be executed, got %q", executedActions[0])
	}

	// Result should indicate no changes (file unchanged, not deleted)
	if result != nil {
		t.Errorf("expected nil result (no changes), got %+v", result)
	}
}
