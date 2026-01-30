package rules

import (
	"errors"
	"testing"

	"github.com/prettymuchbryce/autotidy/internal/fs"
	"github.com/prettymuchbryce/autotidy/internal/testutil"
)

// mockExecutable implements Executable for testing
type mockExecutable struct {
	result *ExecutionResult
	err    error
}

func (m *mockExecutable) Execute(path string, _ fs.FileSystem) (*ExecutionResult, error) {
	return m.result, m.err
}

func TestAction_Execute(t *testing.T) {
	newPath := testutil.Path("/", "new", "path")
	testPath := testutil.Path("/", "test", "path")

	tests := []struct {
		name           string
		inner          Executable
		expectedResult *ExecutionResult
		wantErr        bool
	}{
		{
			name: "delegates to inner with result",
			inner: &mockExecutable{
				result: &ExecutionResult{
					NewPath: newPath,
				},
			},
			expectedResult: &ExecutionResult{
				NewPath: newPath,
			},
			wantErr: false,
		},
		{
			name:           "delegates to inner with nil result",
			inner:          &mockExecutable{result: nil},
			expectedResult: nil,
			wantErr:        false,
		},
		{
			name:    "delegates error from inner",
			inner:   &mockExecutable{err: errors.New("test error")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Action{
				Name:  "test",
				Inner: tt.inner,
			}
			result, err := a.Execute(testPath, fs.NewNoop())

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectedResult == nil {
				if result != nil {
					t.Errorf("expected nil result, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("expected result, got nil")
				return
			}

			if result.NewPath != tt.expectedResult.NewPath {
				t.Errorf("NewPath = %q, want %q", result.NewPath, tt.expectedResult.NewPath)
			}
		})
	}
}
