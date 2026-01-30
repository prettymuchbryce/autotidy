package filters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prettymuchbryce/autotidy/internal/rules"

	"gopkg.in/yaml.v3"
)

func TestDateModified_Evaluate(t *testing.T) {
	// Create a temp file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name    string
		filter  DateModified
		want    bool
		wantErr bool
	}{
		{
			name: "after 1 hour ago matches recently modified file",
			filter: DateModified{
				After: &DateSpec{HoursAgo: ptr(1.0)},
			},
			want: true,
		},
		{
			name: "before 1 hour ago does not match recently modified file",
			filter: DateModified{
				Before: &DateSpec{HoursAgo: ptr(1.0)},
			},
			want: false,
		},
		{
			name: "after 1 day ago matches recently modified file",
			filter: DateModified{
				After: &DateSpec{DaysAgo: ptr(1.0)},
			},
			want: true,
		},
		{
			name: "before now (1 minute in future) matches",
			filter: DateModified{
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

func TestDeserializeDateModified(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		wantErr     bool
		errContains string
	}{
		{
			name: "before days_ago",
			yaml: "date_modified:\n  before:\n    days_ago: 7",
		},
		{
			name: "after hours_ago",
			yaml: "date_modified:\n  after:\n    hours_ago: 2",
		},
		{
			name: "before and after",
			yaml: "date_modified:\n  before:\n    days_ago: 1\n  after:\n    days_ago: 7",
		},
		{
			name: "with unix timestamp",
			yaml: "date_modified:\n  after:\n    unix: 1704067200",
		},
		{
			name: "with date string",
			yaml: "date_modified:\n  before:\n    date: \"2024-01-01\"",
		},
		{
			name: "float days_ago",
			yaml: "date_modified:\n  after:\n    days_ago: 7.5",
		},
		{
			name:        "empty filter",
			yaml:        "date_modified: {}",
			wantErr:     true,
			errContains: "requires at least one of",
		},
		{
			name:        "multiple time specs in before",
			yaml:        "date_modified:\n  before:\n    days_ago: 7\n    hours_ago: 2",
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

			if f.Name != "date_modified" {
				t.Errorf("filter name = %q, want %q", f.Name, "date_modified")
				return
			}

			_, ok := f.Inner.(*DateModified)
			if !ok {
				t.Errorf("inner is not *DateModified, got %T", f.Inner)
			}
		})
	}
}
