package filters

import (
	"strings"
	"testing"

	"github.com/prettymuchbryce/autotidy/internal/rules"
	"github.com/prettymuchbryce/autotidy/internal/testutil"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

func TestSize_Evaluate(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Define paths
	tinyPath := testutil.Path("/", "tiny.txt")
	smallPath := testutil.Path("/", "small.txt")
	oneKbPath := testutil.Path("/", "1kb.txt")
	tenKbPath := testutil.Path("/", "10kb.txt")
	oneMbPath := testutil.Path("/", "1mb.txt")
	onePointFiveMbPath := testutil.Path("/", "1.5mb.txt")
	tenMbPath := testutil.Path("/", "10mb.txt")

	// Create files of various sizes
	afero.WriteFile(fs, tinyPath, []byte("hi"), 0644)                          // 2 bytes
	afero.WriteFile(fs, smallPath, make([]byte, 500), 0644)                    // 500 bytes
	afero.WriteFile(fs, oneKbPath, make([]byte, 1024), 0644)                   // 1 KB
	afero.WriteFile(fs, tenKbPath, make([]byte, 10*1024), 0644)                // 10 KB
	afero.WriteFile(fs, oneMbPath, make([]byte, 1024*1024), 0644)              // 1 MB
	afero.WriteFile(fs, onePointFiveMbPath, make([]byte, int(1.5*1024*1024)), 0644) // 1.5 MB
	afero.WriteFile(fs, tenMbPath, make([]byte, 10*1024*1024), 0644)           // 10 MB

	tests := []struct {
		name    string
		size    Size
		path    string
		want    bool
		wantErr bool
	}{
		// GreaterThan tests
		{
			name: "greater_than 1kb matches 10kb file",
			size: Size{GreaterThan: &SizeSpec{KB: ptr(1.0)}},
			path: tenKbPath,
			want: true,
		},
		{
			name: "greater_than 1kb does not match 500b file",
			size: Size{GreaterThan: &SizeSpec{KB: ptr(1.0)}},
			path: smallPath,
			want: false,
		},
		{
			name: "greater_than 1kb does not match exactly 1kb file",
			size: Size{GreaterThan: &SizeSpec{KB: ptr(1.0)}},
			path: oneKbPath,
			want: false,
		},

		// LessThan tests
		{
			name: "less_than 1kb matches 500b file",
			size: Size{LessThan: &SizeSpec{KB: ptr(1.0)}},
			path: smallPath,
			want: true,
		},
		{
			name: "less_than 1kb does not match 10kb file",
			size: Size{LessThan: &SizeSpec{KB: ptr(1.0)}},
			path: tenKbPath,
			want: false,
		},

		// AtLeast tests
		{
			name: "at_least 1kb matches exactly 1kb file",
			size: Size{AtLeast: &SizeSpec{KB: ptr(1.0)}},
			path: oneKbPath,
			want: true,
		},
		{
			name: "at_least 1kb matches 10kb file",
			size: Size{AtLeast: &SizeSpec{KB: ptr(1.0)}},
			path: tenKbPath,
			want: true,
		},
		{
			name: "at_least 1kb does not match 500b file",
			size: Size{AtLeast: &SizeSpec{KB: ptr(1.0)}},
			path: smallPath,
			want: false,
		},

		// AtMost tests
		{
			name: "at_most 1kb matches exactly 1kb file",
			size: Size{AtMost: &SizeSpec{KB: ptr(1.0)}},
			path: oneKbPath,
			want: true,
		},
		{
			name: "at_most 1kb matches 500b file",
			size: Size{AtMost: &SizeSpec{KB: ptr(1.0)}},
			path: smallPath,
			want: true,
		},
		{
			name: "at_most 1kb does not match 10kb file",
			size: Size{AtMost: &SizeSpec{KB: ptr(1.0)}},
			path: tenKbPath,
			want: false,
		},

		// Between tests
		{
			name: "between 1mb and 2mb matches 1.5mb file",
			size: Size{Between: &SizeBetween{
				Min: SizeSpec{MB: ptr(1.0)},
				Max: SizeSpec{MB: ptr(2.0)},
			}},
			path: onePointFiveMbPath,
			want: true,
		},
		{
			name: "between 1mb and 2mb matches exactly 1mb file",
			size: Size{Between: &SizeBetween{
				Min: SizeSpec{MB: ptr(1.0)},
				Max: SizeSpec{MB: ptr(2.0)},
			}},
			path: oneMbPath,
			want: true,
		},
		{
			name: "between 1mb and 2mb does not match 10mb file",
			size: Size{Between: &SizeBetween{
				Min: SizeSpec{MB: ptr(1.0)},
				Max: SizeSpec{MB: ptr(2.0)},
			}},
			path: tenMbPath,
			want: false,
		},
		{
			name: "between 1mb and 2mb does not match 10kb file",
			size: Size{Between: &SizeBetween{
				Min: SizeSpec{MB: ptr(1.0)},
				Max: SizeSpec{MB: ptr(2.0)},
			}},
			path: tenKbPath,
			want: false,
		},

		// Different units
		{
			name: "greater_than 100b matches 500b file",
			size: Size{GreaterThan: &SizeSpec{B: ptr(100.0)}},
			path: smallPath,
			want: true,
		},

		// Float values
		{
			name: "greater_than 1.5mb does not match 1mb file",
			size: Size{GreaterThan: &SizeSpec{MB: ptr(1.5)}},
			path: oneMbPath,
			want: false,
		},
		{
			name: "less_than 1.5mb matches 1mb file",
			size: Size{LessThan: &SizeSpec{MB: ptr(1.5)}},
			path: oneMbPath,
			want: true,
		},
	}

	// Add directory test case
	dirPath := testutil.Path("/", "testdir")
	fs.MkdirAll(dirPath, 0755)
	tests = append(tests, struct {
		name    string
		size    Size
		path    string
		want    bool
		wantErr bool
	}{
		name: "directories return false",
		size: Size{GreaterThan: &SizeSpec{B: ptr(0.0)}},
		path: dirPath,
		want: false,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.size.Fs = fs
			got, err := tt.size.Evaluate(tt.path)

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

func TestParseShorthand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantOp      string
		wantValue   float64
		wantUnit    string
		wantErr     bool
		errContains string
	}{
		// Basic operators
		{name: "greater than", input: "> 10mb", wantOp: "greater_than", wantValue: 10, wantUnit: "mb"},
		{name: "greater or equal", input: ">= 10mb", wantOp: "at_least", wantValue: 10, wantUnit: "mb"},
		{name: "less than", input: "< 10mb", wantOp: "less_than", wantValue: 10, wantUnit: "mb"},
		{name: "less or equal", input: "<= 10mb", wantOp: "at_most", wantValue: 10, wantUnit: "mb"},

		// All units
		{name: "bytes", input: "> 100b", wantOp: "greater_than", wantValue: 100, wantUnit: "b"},
		{name: "kilobytes", input: "> 100kb", wantOp: "greater_than", wantValue: 100, wantUnit: "kb"},
		{name: "megabytes", input: "> 100mb", wantOp: "greater_than", wantValue: 100, wantUnit: "mb"},
		{name: "gigabytes", input: "> 100gb", wantOp: "greater_than", wantValue: 100, wantUnit: "gb"},
		{name: "terabytes", input: "> 100tb", wantOp: "greater_than", wantValue: 100, wantUnit: "tb"},

		// Float values
		{name: "float value", input: "> 2.5gb", wantOp: "greater_than", wantValue: 2.5, wantUnit: "gb"},
		{name: "float with spaces", input: ">= 1.75 mb", wantOp: "at_least", wantValue: 1.75, wantUnit: "mb"},

		// Case insensitive
		{name: "uppercase unit", input: "> 10MB", wantOp: "greater_than", wantValue: 10, wantUnit: "mb"},
		{name: "mixed case", input: "> 10Mb", wantOp: "greater_than", wantValue: 10, wantUnit: "mb"},

		// Spacing variations
		{name: "no spaces", input: ">10mb", wantOp: "greater_than", wantValue: 10, wantUnit: "mb"},
		{name: "extra spaces", input: "  >   10   mb  ", wantOp: "greater_than", wantValue: 10, wantUnit: "mb"},

		// Invalid inputs
		{name: "invalid operator", input: "== 10mb", wantErr: true, errContains: "invalid size shorthand"},
		{name: "missing unit", input: "> 10", wantErr: true, errContains: "invalid size shorthand"},
		{name: "missing value", input: "> mb", wantErr: true, errContains: "invalid size shorthand"},
		{name: "invalid unit", input: "> 10pb", wantErr: true, errContains: "invalid size shorthand"},
		{name: "empty string", input: "", wantErr: true, errContains: "invalid size shorthand"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := parseShorthand(tt.input)

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

			// Verify the correct operator was set
			var gotSpec *SizeSpec
			switch tt.wantOp {
			case "greater_than":
				gotSpec = size.GreaterThan
			case "at_least":
				gotSpec = size.AtLeast
			case "less_than":
				gotSpec = size.LessThan
			case "at_most":
				gotSpec = size.AtMost
			}

			if gotSpec == nil {
				t.Errorf("expected %s to be set", tt.wantOp)
				return
			}

			// Verify the value and unit
			var gotValue *float64
			switch tt.wantUnit {
			case "b":
				gotValue = gotSpec.B
			case "kb":
				gotValue = gotSpec.KB
			case "mb":
				gotValue = gotSpec.MB
			case "gb":
				gotValue = gotSpec.GB
			case "tb":
				gotValue = gotSpec.TB
			}

			if gotValue == nil {
				t.Errorf("expected unit %s to be set", tt.wantUnit)
				return
			}

			if *gotValue != tt.wantValue {
				t.Errorf("value = %v, want %v", *gotValue, tt.wantValue)
			}
		})
	}
}

func TestDeserializeSize(t *testing.T) {
	tests := []struct {
		name        string
		yaml        string
		wantErr     bool
		errContains string
	}{
		// Shorthand
		{
			name: "shorthand greater than",
			yaml: "file_size: \"> 10mb\"",
		},
		{
			name: "shorthand at least",
			yaml: "file_size: \">= 2.5gb\"",
		},

		// Explicit form
		{
			name: "explicit greater_than",
			yaml: "file_size:\n  greater_than:\n    mb: 10",
		},
		{
			name: "explicit less_than",
			yaml: "file_size:\n  less_than:\n    gb: 1",
		},
		{
			name: "explicit at_least",
			yaml: "file_size:\n  at_least:\n    kb: 500",
		},
		{
			name: "explicit at_most",
			yaml: "file_size:\n  at_most:\n    mb: 2.5",
		},
		{
			name: "explicit between",
			yaml: "file_size:\n  between:\n    min:\n      mb: 1\n    max:\n      mb: 10",
		},

		// Errors
		{
			name:        "multiple comparisons",
			yaml:        "file_size:\n  greater_than:\n    mb: 10\n  less_than:\n    mb: 100",
			wantErr:     true,
			errContains: "can only have one comparison",
		},
		{
			name:        "no comparison",
			yaml:        "file_size: {}",
			wantErr:     true,
			errContains: "requires a comparison",
		},
		{
			name:        "multiple units in spec",
			yaml:        "file_size:\n  greater_than:\n    mb: 10\n    gb: 1",
			wantErr:     true,
			errContains: "only one size unit allowed",
		},
		{
			name:        "invalid shorthand",
			yaml:        "file_size: \"not valid\"",
			wantErr:     true,
			errContains: "invalid size shorthand",
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

			if f.Name != "file_size" {
				t.Errorf("filter name = %q, want %q", f.Name, "file_size")
				return
			}

			_, ok := f.Inner.(*Size)
			if !ok {
				t.Errorf("inner is not *Size, got %T", f.Inner)
			}
		})
	}
}

// Helper to create float64 pointer
func ptr(f float64) *float64 {
	return &f
}
