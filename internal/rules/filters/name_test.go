package filters

import (
	"regexp"
	"strings"
	"testing"

	"github.com/prettymuchbryce/autotidy/internal/rules"
	"github.com/prettymuchbryce/autotidy/internal/testutil"

	"gopkg.in/yaml.v3"
)

func TestName_Evaluate(t *testing.T) {
	tests := []struct {
		name     string
		glob     string
		path     string
		expected bool
	}{
		// Exact match
		{
			name:     "exact match",
			glob:     "foo.txt",
			path:     testutil.Path("/", "some", "path", "foo.txt"),
			expected: true,
		},
		{
			name:     "exact match no match",
			glob:     "foo.txt",
			path:     testutil.Path("/", "some", "path", "bar.txt"),
			expected: false,
		},

		// Star wildcard
		{
			name:     "star matches extension",
			glob:     "*.jpg",
			path:     testutil.Path("/", "photos", "image.jpg"),
			expected: true,
		},
		{
			name:     "star matches prefix",
			glob:     "report*",
			path:     testutil.Path("/", "docs", "report2024.pdf"),
			expected: true,
		},
		{
			name:     "star no match different extension",
			glob:     "*.jpg",
			path:     testutil.Path("/", "photos", "image.png"),
			expected: false,
		},
		{
			name:     "star matches middle",
			glob:     "test*.log",
			path:     testutil.Path("/", "logs", "test_output.log"),
			expected: true,
		},

		// Question mark wildcard
		{
			name:     "question mark single char",
			glob:     "?.txt",
			path:     testutil.Path("/", "files", "a.txt"),
			expected: true,
		},
		{
			name:     "question mark no match multiple chars",
			glob:     "?.txt",
			path:     testutil.Path("/", "files", "ab.txt"),
			expected: false,
		},
		{
			name:     "multiple question marks",
			glob:     "???.log",
			path:     testutil.Path("/", "logs", "app.log"),
			expected: true,
		},

		// Character classes
		{
			name:     "character class match",
			glob:     "[abc].txt",
			path:     testutil.Path("/", "files", "a.txt"),
			expected: true,
		},
		{
			name:     "character class no match",
			glob:     "[abc].txt",
			path:     testutil.Path("/", "files", "d.txt"),
			expected: false,
		},
		{
			name:     "character range",
			glob:     "[0-9].log",
			path:     testutil.Path("/", "logs", "5.log"),
			expected: true,
		},
		{
			name:     "negated character class",
			glob:     "[!0-9].txt",
			path:     testutil.Path("/", "files", "a.txt"),
			expected: true,
		},

		// Complex patterns
		{
			name:     "multiple wildcards",
			glob:     "*.test.*",
			path:     testutil.Path("/", "src", "app.test.js"),
			expected: true,
		},
		{
			name:     "case sensitive match",
			glob:     "*.TXT",
			path:     testutil.Path("/", "files", "doc.TXT"),
			expected: true,
		},
		{
			name:     "case sensitive no match",
			glob:     "*.TXT",
			path:     testutil.Path("/", "files", "doc.txt"),
			expected: false,
		},

		// Edge cases
		{
			name:     "empty pattern matches nothing",
			glob:     "",
			path:     testutil.Path("/", "files", "test.txt"),
			expected: false,
		},
		{
			name:     "pattern with no wildcards",
			glob:     "exactfile.dat",
			path:     testutil.Path("/", "data", "exactfile.dat"),
			expected: true,
		},
		{
			name:     "hidden file",
			glob:     ".*",
			path:     testutil.Path("/", "home", ".bashrc"),
			expected: true,
		},
		{
			name:     "filename only no path match",
			glob:     "*.md",
			path:     "README.md",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Name{Glob: tt.glob}
			got, err := n.Evaluate(tt.path)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got != tt.expected {
				t.Errorf("Evaluate(%q) with pattern %q = %v, want %v",
					tt.path, tt.glob, got, tt.expected)
			}
		})
	}
}

func TestName_Evaluate_InvalidPattern(t *testing.T) {
	// Invalid glob patterns should return errors
	tests := []struct {
		name string
		glob string
	}{
		{
			name: "unclosed bracket",
			glob: "[abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &Name{Glob: tt.glob}
			_, err := n.Evaluate(testutil.Path("/", "test", "file.txt"))
			if err == nil {
				t.Errorf("expected error for invalid pattern %q but got none", tt.glob)
			}
		})
	}
}

func TestDeserializeName(t *testing.T) {
	tests := []struct {
		name         string
		yaml         string
		expectedGlob string
		wantErr      bool
	}{
		{
			name:         "scalar string",
			yaml:         "name: \"*.jpg\"",
			expectedGlob: "*.jpg",
			wantErr:      false,
		},
		{
			name:         "mapping with glob",
			yaml:         "name:\n  glob: \"*.png\"",
			expectedGlob: "*.png",
			wantErr:      false,
		},
		{
			name:         "simple filename",
			yaml:         "name: readme.txt",
			expectedGlob: "readme.txt",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test through rules.Filter which handles the registry lookup
			var f rules.Filter
			err := yaml.Unmarshal([]byte(tt.yaml), &f)

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

			if f.Name != "name" {
				t.Errorf("filter name = %q, want %q", f.Name, "name")
				return
			}

			// The inner should be a *Name
			nameFilter, ok := f.Inner.(*Name)
			if !ok {
				t.Errorf("inner is not *Name, got %T", f.Inner)
				return
			}

			if nameFilter.Glob != tt.expectedGlob {
				t.Errorf("Glob = %q, want %q", nameFilter.Glob, tt.expectedGlob)
			}
		})
	}
}

func TestName_EvaluateErrorMessage(t *testing.T) {
	n := &Name{Glob: "[invalid"}
	_, err := n.Evaluate(testutil.Path("/", "test", "file.txt"))
	if err == nil {
		t.Fatal("expected error but got none")
	}

	// Verify error message contains useful information
	if !strings.Contains(err.Error(), "[invalid") {
		t.Errorf("error should contain the pattern, got: %v", err)
	}
	if !strings.Contains(err.Error(), "invalid glob pattern") {
		t.Errorf("error should mention 'invalid glob pattern', got: %v", err)
	}
}

func TestName_Evaluate_Regex(t *testing.T) {
	tests := []struct {
		name     string
		regex    string
		path     string
		expected bool
	}{
		{
			name:     "simple regex match",
			regex:    `^foo\.txt$`,
			path:     testutil.Path("/", "some", "path", "foo.txt"),
			expected: true,
		},
		{
			name:     "simple regex no match",
			regex:    `^foo\.txt$`,
			path:     testutil.Path("/", "some", "path", "bar.txt"),
			expected: false,
		},
		{
			name:     "regex with digit pattern",
			regex:    `^report\d{4}\.pdf$`,
			path:     testutil.Path("/", "docs", "report2024.pdf"),
			expected: true,
		},
		{
			name:     "regex with digit pattern no match",
			regex:    `^report\d{4}\.pdf$`,
			path:     testutil.Path("/", "docs", "report24.pdf"),
			expected: false,
		},
		{
			name:     "regex with alternation",
			regex:    `\.(jpg|jpeg|png)$`,
			path:     testutil.Path("/", "photos", "image.jpeg"),
			expected: true,
		},
		{
			name:     "regex partial match",
			regex:    `test`,
			path:     testutil.Path("/", "files", "my_test_file.txt"),
			expected: true,
		},
		{
			name:     "regex case insensitive flag",
			regex:    `(?i)readme`,
			path:     testutil.Path("/", "docs", "README.md"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := regexp.MustCompile(tt.regex)
			n := &Name{Regex: re}
			got, err := n.Evaluate(tt.path)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got != tt.expected {
				t.Errorf("Evaluate(%q) with regex %q = %v, want %v",
					tt.path, tt.regex, got, tt.expected)
			}
		})
	}
}

func TestDeserializeName_Regex(t *testing.T) {
	tests := []struct {
		name          string
		yaml          string
		expectedRegex string
		wantErr       bool
		errContains   string
	}{
		{
			name:          "mapping with regex",
			yaml:          "name:\n  regex: '^foo\\d+\\.txt$'",
			expectedRegex: `^foo\d+\.txt$`,
		},
		{
			name:        "invalid regex",
			yaml:        "name:\n  regex: '[invalid'",
			wantErr:     true,
			errContains: "invalid regex pattern",
		},
		{
			name:        "both glob and regex",
			yaml:        "name:\n  glob: '*.txt'\n  regex: '^foo'",
			wantErr:     true,
			errContains: "cannot have both glob and regex",
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

			nameFilter, ok := f.Inner.(*Name)
			if !ok {
				t.Errorf("inner is not *Name, got %T", f.Inner)
				return
			}

			if nameFilter.Regex == nil {
				t.Errorf("Regex is nil, expected pattern")
				return
			}

			if nameFilter.Regex.String() != tt.expectedRegex {
				t.Errorf("Regex = %q, want %q", nameFilter.Regex.String(), tt.expectedRegex)
			}
		})
	}
}
