package filters

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/prettymuchbryce/autotidy/internal/rules"

	"github.com/djherbis/times"
	"gopkg.in/yaml.v3"
)

func TestDateCreated_Evaluate(t *testing.T) {
	// Create a temp file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Check if birth time is available on this platform
	ts, err := times.Stat(testFile)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	if !ts.HasBirthTime() {
		t.Skip("birth time not available on this platform")
	}

	tests := []struct {
		name    string
		filter  DateCreated
		want    bool
		wantErr bool
	}{
		{
			name: "after 1 hour ago matches recently created file",
			filter: DateCreated{
				After: &DateSpec{HoursAgo: ptr(1.0)},
			},
			want: true,
		},
		{
			name: "before 1 hour ago does not match recently created file",
			filter: DateCreated{
				Before: &DateSpec{HoursAgo: ptr(1.0)},
			},
			want: false,
		},
		{
			name: "after 1 day ago matches recently created file",
			filter: DateCreated{
				After: &DateSpec{DaysAgo: ptr(1.0)},
			},
			want: true,
		},
		{
			name: "before now (1 minute in future) matches",
			filter: DateCreated{
				Before: &DateSpec{MinutesAgo: ptr(-1.0)}, // 1 minute in the future
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.filter.Evaluate(testFile)

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

			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDateCreated_HasBirthTime(t *testing.T) {
	// This test documents the expected behavior on different platforms
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	ts, err := times.Stat(testFile)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	t.Logf("HasBirthTime: %v (platform: %s)", ts.HasBirthTime(), runtime.GOOS)

	// On macOS and Windows, we expect birth time to be available
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		if !ts.HasBirthTime() {
			t.Errorf("expected HasBirthTime() to be true on %s", runtime.GOOS)
		}
	}
}

func TestDateCreated_NoBirthTime(t *testing.T) {
	// Skip on platforms that have birth time
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	ts, err := times.Stat(testFile)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	if ts.HasBirthTime() {
		t.Skip("birth time is available on this platform")
	}

	filter := DateCreated{After: &DateSpec{DaysAgo: ptr(1.0)}}
	_, err = filter.Evaluate(testFile)
	if err == nil {
		t.Errorf("expected error when birth time is not available")
	}
	if !strings.Contains(err.Error(), "creation time is not available") {
		t.Errorf("error should mention creation time not available, got: %v", err)
	}
}

func TestDeserializeDateCreated(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		wantErr     bool
		errContains string
	}{
		{
			name: "before days_ago",
			yaml: "date_created:\n  before:\n    days_ago: 7",
		},
		{
			name: "after hours_ago",
			yaml: "date_created:\n  after:\n    hours_ago: 2",
		},
		{
			name: "before and after",
			yaml: "date_created:\n  before:\n    days_ago: 1\n  after:\n    days_ago: 7",
		},
		{
			name: "with years_ago",
			yaml: "date_created:\n  before:\n    years_ago: 1",
		},
		{
			name:        "empty filter",
			yaml:        "date_created: {}",
			wantErr:     true,
			errContains: "requires at least one of",
		},
		{
			name:        "multiple time specs in before",
			yaml:        "date_created:\n  before:\n    days_ago: 7\n    hours_ago: 2",
			wantErr:     true,
			errContains: "only one time specification allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f rules.Filter
			err := yaml.Unmarshal([]byte(tt.yaml), &f)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error should contain %q, got: %v", tt.errContains, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if f.Name != "date_created" {
				t.Errorf("filter name = %q, want %q", f.Name, "date_created")
				return
			}

			_, ok := f.Inner.(*DateCreated)
			if !ok {
				t.Errorf("inner is not *DateCreated, got %T", f.Inner)
			}
		})
	}
}
