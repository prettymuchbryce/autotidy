package filters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/prettymuchbryce/autotidy/internal/rules"

	"gopkg.in/yaml.v3"
)

func TestDateAccessed_Evaluate(t *testing.T) {
	// Create a temp file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Read the file to update access time
	_, _ = os.ReadFile(testFile)

	tests := []struct {
		name    string
		filter  DateAccessed
		want    bool
		wantErr bool
	}{
		{
			name: "after 1 hour ago matches recently accessed file",
			filter: DateAccessed{
				After: &DateSpec{HoursAgo: ptr(1.0)},
			},
			want: true,
		},
		{
			name: "before 1 hour ago does not match recently accessed file",
			filter: DateAccessed{
				Before: &DateSpec{HoursAgo: ptr(1.0)},
			},
			want: false,
		},
		{
			name: "after 1 day ago matches recently accessed file",
			filter: DateAccessed{
				After: &DateSpec{DaysAgo: ptr(1.0)},
			},
			want: true,
		},
		{
			name: "before now (1 minute in future) matches",
			filter: DateAccessed{
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

func TestDeserializeDateAccessed(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		wantErr     bool
		errContains string
	}{
		{
			name: "before days_ago",
			yaml: "date_accessed:\n  before:\n    days_ago: 7",
		},
		{
			name: "after hours_ago",
			yaml: "date_accessed:\n  after:\n    hours_ago: 2",
		},
		{
			name: "before and after",
			yaml: "date_accessed:\n  before:\n    days_ago: 1\n  after:\n    days_ago: 7",
		},
		{
			name: "with unix timestamp",
			yaml: "date_accessed:\n  after:\n    unix: 1704067200",
		},
		{
			name: "with date string",
			yaml: "date_accessed:\n  before:\n    date: \"2024-01-01\"",
		},
		{
			name: "with datetime string",
			yaml: "date_accessed:\n  before:\n    date: \"2024-01-01T12:00:00\"",
		},
		{
			name: "float days_ago",
			yaml: "date_accessed:\n  after:\n    days_ago: 7.5",
		},
		{
			name:        "empty filter",
			yaml:        "date_accessed: {}",
			wantErr:     true,
			errContains: "requires at least one of",
		},
		{
			name:        "multiple time specs in before",
			yaml:        "date_accessed:\n  before:\n    days_ago: 7\n    hours_ago: 2",
			wantErr:     true,
			errContains: "only one time specification allowed",
		},
		{
			name: "date format is validated at eval time not deserialize time",
			yaml: "date_accessed:\n  before:\n    date: \"not-a-date\"",
			// Note: invalid date format is only caught when ToTime() is called during Evaluate()
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

			if f.Name != "date_accessed" {
				t.Errorf("filter name = %q, want %q", f.Name, "date_accessed")
				return
			}

			_, ok := f.Inner.(*DateAccessed)
			if !ok {
				t.Errorf("inner is not *DateAccessed, got %T", f.Inner)
			}
		})
	}
}

func TestDateSpec_ToTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		spec     DateSpec
		wantDiff time.Duration
		wantErr  bool
	}{
		{
			name:     "seconds_ago",
			spec:     DateSpec{SecondsAgo: ptr(30.0)},
			wantDiff: 30 * time.Second,
		},
		{
			name:     "minutes_ago",
			spec:     DateSpec{MinutesAgo: ptr(5.0)},
			wantDiff: 5 * time.Minute,
		},
		{
			name:     "hours_ago",
			spec:     DateSpec{HoursAgo: ptr(2.0)},
			wantDiff: 2 * time.Hour,
		},
		{
			name:     "days_ago",
			spec:     DateSpec{DaysAgo: ptr(1.0)},
			wantDiff: 24 * time.Hour,
		},
		{
			name:     "weeks_ago",
			spec:     DateSpec{WeeksAgo: ptr(1.0)},
			wantDiff: 7 * 24 * time.Hour,
		},
		{
			name:     "float days_ago",
			spec:     DateSpec{DaysAgo: ptr(1.5)},
			wantDiff: 36 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.spec.ToTime()

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

			expectedTime := now.Add(-tt.wantDiff)
			diff := got.Sub(expectedTime)
			if diff < -time.Second || diff > time.Second {
				t.Errorf("ToTime() = %v, want approximately %v (diff: %v)", got, expectedTime, diff)
			}
		})
	}
}

func TestDateSpec_ToTime_Absolute(t *testing.T) {
	tests := []struct {
		name    string
		spec    DateSpec
		want    time.Time
		wantErr bool
	}{
		{
			name: "unix timestamp",
			spec: DateSpec{Unix: ptrInt64(1704067200)}, // 2024-01-01 00:00:00 UTC
			want: time.Unix(1704067200, 0),
		},
		{
			name: "date string",
			spec: DateSpec{Date: ptrString("2024-01-15")},
			want: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "datetime string",
			spec: DateSpec{Date: ptrString("2024-01-15T14:30:00")},
			want: time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
		},
		{
			name:    "invalid date",
			spec:    DateSpec{Date: ptrString("not-a-date")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.spec.ToTime()

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

			if !got.Equal(tt.want) {
				t.Errorf("ToTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDateSpec_Validate(t *testing.T) {
	tests := []struct {
		name        string
		spec        DateSpec
		wantErr     bool
		errContains string
	}{
		{
			name: "valid days_ago",
			spec: DateSpec{DaysAgo: ptr(7.0)},
		},
		{
			name: "valid unix",
			spec: DateSpec{Unix: ptrInt64(1704067200)},
		},
		{
			name:        "empty spec",
			spec:        DateSpec{},
			wantErr:     true,
			errContains: "no time specification provided",
		},
		{
			name:        "multiple specs",
			spec:        DateSpec{DaysAgo: ptr(7.0), HoursAgo: ptr(2.0)},
			wantErr:     true,
			errContains: "only one time specification allowed",
		},
		{
			name:        "negative unix",
			spec:        DateSpec{Unix: ptrInt64(-1)},
			wantErr:     true,
			errContains: "must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()

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
			}
		})
	}
}

func ptrInt64(i int64) *int64 {
	return &i
}

func ptrString(s string) *string {
	return &s
}
