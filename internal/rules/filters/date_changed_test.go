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

func TestDateChanged_Evaluate(t *testing.T) {
	// Create a temp file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Check if change time is available on this platform
	ts, err := times.Stat(testFile)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	if !ts.HasChangeTime() {
		t.Skip("change time not available on this platform")
	}

	tests := []struct {
		name    string
		filter  DateChanged
		want    bool
		wantErr bool
	}{
		{
			name: "after 1 hour ago matches recently changed file",
			filter: DateChanged{
				After: &DateSpec{HoursAgo: ptr(1.0)},
			},
			want: true,
		},
		{
			name: "before 1 hour ago does not match recently changed file",
			filter: DateChanged{
				Before: &DateSpec{HoursAgo: ptr(1.0)},
			},
			want: false,
		},
		{
			name: "after 1 day ago matches recently changed file",
			filter: DateChanged{
				After: &DateSpec{DaysAgo: ptr(1.0)},
			},
			want: true,
		},
		{
			name: "before now (1 minute in future) matches",
			filter: DateChanged{
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

func TestDateChanged_HasChangeTime(t *testing.T) {
	// This test documents the expected behavior on different platforms
	switch runtime.GOOS {
	case "darwin", "linux", "freebsd":
		// These platforms should have change time
	case "windows":
		// Windows Vista+ should have change time, XP does not
	}
	// We just verify that the check works without crashing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	ts, err := times.Stat(testFile)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	t.Logf("HasChangeTime: %v (platform: %s)", ts.HasChangeTime(), runtime.GOOS)
}

func TestDeserializeDateChanged(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		wantErr     bool
		errContains string
	}{
		{
			name: "before days_ago",
			yaml: "date_changed:\n  before:\n    days_ago: 7",
		},
		{
			name: "after hours_ago",
			yaml: "date_changed:\n  after:\n    hours_ago: 2",
		},
		{
			name: "before and after",
			yaml: "date_changed:\n  before:\n    days_ago: 1\n  after:\n    days_ago: 7",
		},
		{
			name:        "empty filter",
			yaml:        "date_changed: {}",
			wantErr:     true,
			errContains: "requires at least one of",
		},
		{
			name:        "multiple time specs in before",
			yaml:        "date_changed:\n  before:\n    days_ago: 7\n    hours_ago: 2",
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

			if f.Name != "date_changed" {
				t.Errorf("filter name = %q, want %q", f.Name, "date_changed")
				return
			}

			_, ok := f.Inner.(*DateChanged)
			if !ok {
				t.Errorf("inner is not *DateChanged, got %T", f.Inner)
			}
		})
	}
}
