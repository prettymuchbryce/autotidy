package filters

import (
	"testing"

	"github.com/prettymuchbryce/autotidy/internal/rules"
	"github.com/prettymuchbryce/autotidy/internal/testutil"

	"gopkg.in/yaml.v3"
)

func TestExtension_Evaluate(t *testing.T) {
	tests := []struct {
		name       string
		extensions []string
		path       string
		want       bool
		wantErr    bool
	}{
		{
			name:       "matches exact extension",
			extensions: []string{"txt"},
			path:       testutil.Path("/", "test.txt"),
			want:       true,
		},
		{
			name:       "matches with leading dot in pattern",
			extensions: []string{".txt"},
			path:       testutil.Path("/", "test.txt"),
			want:       true,
		},
		{
			name:       "no match for different extension",
			extensions: []string{"md"},
			path:       testutil.Path("/", "test.txt"),
			want:       false,
		},
		{
			name:       "matches with glob pattern",
			extensions: []string{"j*"},
			path:       testutil.Path("/", "test.json"),
			want:       true,
		},
		{
			name:       "glob matches jpg",
			extensions: []string{"j*"},
			path:       testutil.Path("/", "photo.jpg"),
			want:       true,
		},
		{
			name:       "matches with multiple patterns",
			extensions: []string{"md", "txt"},
			path:       testutil.Path("/", "readme.txt"),
			want:       true,
		},
		{
			name:       "no match with multiple patterns",
			extensions: []string{"md", "json"},
			path:       testutil.Path("/", "test.txt"),
			want:       false,
		},
		{
			name:       "handles file without extension",
			extensions: []string{"txt"},
			path:       testutil.Path("/", "Makefile"),
			want:       false,
		},
		{
			name:       "empty extension matches file without extension",
			extensions: []string{""},
			path:       testutil.Path("/", "Makefile"),
			want:       true,
		},
		{
			name:       "handles multiple dots in filename",
			extensions: []string{"gz"},
			path:       testutil.Path("/", "archive.tar.gz"),
			want:       true,
		},
		{
			name:       "case sensitive match",
			extensions: []string{"TXT"},
			path:       testutil.Path("/", "test.txt"),
			want:       false,
		},
		{
			name:       "case sensitive match uppercase",
			extensions: []string{"TXT"},
			path:       testutil.Path("/", "test.TXT"),
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Extension{Extensions: tt.extensions}
			got, err := e.Evaluate(tt.path)

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

func TestDeserializeExtension(t *testing.T) {
	tests := []struct {
		name               string
		yaml               string
		expectedExtensions []string
		wantErr            bool
	}{
		{
			name:               "scalar string",
			yaml:               "extension: txt",
			expectedExtensions: []string{"txt"},
		},
		{
			name:               "scalar string with dot",
			yaml:               "extension: .txt",
			expectedExtensions: []string{".txt"},
		},
		{
			name:               "sequence",
			yaml:               "extension: [txt, md, json]",
			expectedExtensions: []string{"txt", "md", "json"},
		},
		{
			name:               "mapping with scalar",
			yaml:               "extension:\n  extensions: txt",
			expectedExtensions: []string{"txt"},
		},
		{
			name:               "mapping with sequence",
			yaml:               "extension:\n  extensions: [txt, md]",
			expectedExtensions: []string{"txt", "md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			if f.Name != "extension" {
				t.Errorf("filter name = %q, want %q", f.Name, "extension")
				return
			}

			extFilter, ok := f.Inner.(*Extension)
			if !ok {
				t.Errorf("inner is not *Extension, got %T", f.Inner)
				return
			}

			if len(extFilter.Extensions) != len(tt.expectedExtensions) {
				t.Errorf("Extensions length = %d, want %d", len(extFilter.Extensions), len(tt.expectedExtensions))
				return
			}

			for i, ext := range tt.expectedExtensions {
				if extFilter.Extensions[i] != ext {
					t.Errorf("Extensions[%d] = %q, want %q", i, extFilter.Extensions[i], ext)
				}
			}
		})
	}
}
