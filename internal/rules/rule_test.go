package rules

import (
	"testing"

	"github.com/prettymuchbryce/autotidy/internal/testutil"
)

func boolPtr(b bool) *bool {
	return &b
}

func TestRule_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		enabled  *bool
		expected bool
	}{
		{"nil defaults to true", nil, true},
		{"explicit true", boolPtr(true), true},
		{"explicit false", boolPtr(false), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Rule{Enabled: tt.enabled}
			if got := r.IsEnabled(); got != tt.expected {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRule_IsRecursive(t *testing.T) {
	tests := []struct {
		name      string
		recursive *bool
		expected  bool
	}{
		{"nil defaults to false", nil, false},
		{"explicit true", boolPtr(true), true},
		{"explicit false", boolPtr(false), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Rule{Recursive: tt.recursive}
			if got := r.IsRecursive(); got != tt.expected {
				t.Errorf("IsRecursive() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRule_GetTraversalMode(t *testing.T) {
	tests := []struct {
		name      string
		traversal TraversalMode
		expected  TraversalMode
	}{
		{"empty defaults to depth-first", "", TraversalDepthFirst},
		{"explicit depth-first", TraversalDepthFirst, TraversalDepthFirst},
		{"explicit breadth-first", TraversalBreadthFirst, TraversalBreadthFirst},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Rule{Traversal: tt.traversal}
			if got := r.GetTraversalMode(); got != tt.expected {
				t.Errorf("GetTraversalMode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRule_CoversPath(t *testing.T) {
	docs := testutil.Path("/", "home", "user", "docs")
	downloads := testutil.Path("/", "home", "user", "downloads")
	tests := []struct {
		name      string
		locations StringList
		recursive bool
		path      string
		expected  bool
	}{
		{
			name:      "exact match",
			locations: StringList{docs},
			recursive: false,
			path:      docs,
			expected:  true,
		},
		{
			name:      "direct child non-recursive",
			locations: StringList{docs},
			recursive: false,
			path:      testutil.Path(docs, "file.txt"),
			expected:  true,
		},
		{
			name:      "nested child non-recursive",
			locations: StringList{docs},
			recursive: false,
			path:      testutil.Path(docs, "sub", "file.txt"),
			expected:  false,
		},
		{
			name:      "nested child recursive",
			locations: StringList{docs},
			recursive: true,
			path:      testutil.Path(docs, "sub", "deep", "file.txt"),
			expected:  true,
		},
		{
			name:      "unrelated path",
			locations: StringList{docs},
			recursive: true,
			path:      testutil.Path("/", "home", "user", "other", "file.txt"),
			expected:  false,
		},
		{
			name:      "multiple locations",
			locations: StringList{docs, downloads},
			recursive: false,
			path:      testutil.Path(downloads, "file.txt"),
			expected:  true,
		},
		{
			name:      "similar prefix not matched",
			locations: StringList{testutil.Path("/", "home", "user", "doc")},
			recursive: true,
			path:      testutil.Path(docs, "file.txt"),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Rule{
				Locations: tt.locations,
				Recursive: boolPtr(tt.recursive),
			}
			if got := r.CoversPath(tt.path); got != tt.expected {
				t.Errorf("CoversPath(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
