package rules

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestStringList_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected StringList
		wantErr  bool
	}{
		{
			name:     "single string",
			yaml:     "locations: single",
			expected: StringList{"single"},
			wantErr:  false,
		},
		{
			name:     "single string with path",
			yaml:     "locations: /path/to/dir",
			expected: StringList{"/path/to/dir"},
			wantErr:  false,
		},
		{
			name:     "list of strings",
			yaml:     "locations:\n  - first\n  - second\n  - third",
			expected: StringList{"first", "second", "third"},
			wantErr:  false,
		},
		{
			name:     "list with single item",
			yaml:     "locations:\n  - only",
			expected: StringList{"only"},
			wantErr:  false,
		},
		{
			name:     "empty list",
			yaml:     "locations: []",
			expected: StringList{},
			wantErr:  false,
		},
		{
			name:     "list with paths",
			yaml:     "locations:\n  - ~/Desktop\n  - /tmp/test",
			expected: StringList{"~/Desktop", "/tmp/test"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wrapper struct {
				Locations StringList `yaml:"locations"`
			}

			err := yaml.Unmarshal([]byte(tt.yaml), &wrapper)

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

			if len(wrapper.Locations) != len(tt.expected) {
				t.Errorf("length mismatch: got %d, want %d", len(wrapper.Locations), len(tt.expected))
				return
			}

			for i, v := range wrapper.Locations {
				if v != tt.expected[i] {
					t.Errorf("item %d: got %q, want %q", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestStringList_UnmarshalYAML_InvalidTypes(t *testing.T) {
	// Test cases that should error
	tests := []struct {
		name string
		yaml string
	}{
		{
			name: "mapping instead of string or list",
			yaml: "locations:\n  key: value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wrapper struct {
				Locations StringList `yaml:"locations"`
			}

			err := yaml.Unmarshal([]byte(tt.yaml), &wrapper)
			if err == nil {
				t.Errorf("expected error for %s but got none", tt.name)
			}
		})
	}
}

func TestStringList_UnmarshalYAML_Coercion(t *testing.T) {
	// YAML coerces some types to strings - verify this behavior
	tests := []struct {
		name     string
		yaml     string
		expected StringList
	}{
		{
			name:     "integer coerced to string",
			yaml:     "locations: 123",
			expected: StringList{"123"},
		},
		{
			name:     "list of integers coerced to strings",
			yaml:     "locations:\n  - 1\n  - 2",
			expected: StringList{"1", "2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wrapper struct {
				Locations StringList `yaml:"locations"`
			}

			err := yaml.Unmarshal([]byte(tt.yaml), &wrapper)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(wrapper.Locations) != len(tt.expected) {
				t.Errorf("length mismatch: got %d, want %d", len(wrapper.Locations), len(tt.expected))
				return
			}

			for i, v := range wrapper.Locations {
				if v != tt.expected[i] {
					t.Errorf("item %d: got %q, want %q", i, v, tt.expected[i])
				}
			}
		})
	}
}
