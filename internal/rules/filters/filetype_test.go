package filters

import (
	"testing"

	"github.com/prettymuchbryce/autotidy/internal/rules"
	"github.com/prettymuchbryce/autotidy/internal/testutil"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

func TestFileType_Evaluate(t *testing.T) {
	filePath := testutil.Path("/", "file.txt")
	subdir := testutil.Path("/", "subdir")
	nestedFile := testutil.Path(subdir, "nested.txt")
	nonexistent := testutil.Path("/", "nonexistent")

	fs := afero.NewMemMapFs()

	// Create a regular file
	afero.WriteFile(fs, filePath, []byte("content"), 0644)

	// Create a directory
	fs.Mkdir(subdir, 0755)

	// Create a file inside directory
	afero.WriteFile(fs, nestedFile, []byte("nested"), 0644)

	tests := []struct {
		name    string
		types   []string
		path    string
		want    bool
		wantErr bool
	}{
		{
			name:  "matches file",
			types: []string{"file"},
			path:  filePath,
			want:  true,
		},
		{
			name:  "matches directory",
			types: []string{"directory"},
			path:  subdir,
			want:  true,
		},
		{
			name:  "matches directory with dir alias",
			types: []string{"dir"},
			path:  subdir,
			want:  true,
		},
		{
			name:  "matches directory with folder alias",
			types: []string{"folder"},
			path:  subdir,
			want:  true,
		},
		{
			name:  "file does not match directory type",
			types: []string{"directory"},
			path:  filePath,
			want:  false,
		},
		{
			name:  "directory does not match file type",
			types: []string{"file"},
			path:  subdir,
			want:  false,
		},
		{
			name:  "matches with multiple types",
			types: []string{"file", "directory"},
			path:  subdir,
			want:  true,
		},
		{
			name:  "matches file in multiple types",
			types: []string{"directory", "file"},
			path:  filePath,
			want:  true,
		},
		{
			name:  "nested file matches file type",
			types: []string{"file"},
			path:  nestedFile,
			want:  true,
		},
		{
			name:    "error on non-existent path",
			types:   []string{"file"},
			path:    nonexistent,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileType{Types: tt.types, Fs: fs}
			got, err := f.Evaluate(tt.path)

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

func TestDeserializeFileType(t *testing.T) {
	tests := []struct {
		name          string
		yaml          string
		expectedTypes []string
		wantErr       bool
	}{
		{
			name:          "scalar string",
			yaml:          "file_type: file",
			expectedTypes: []string{"file"},
		},
		{
			name:          "scalar directory",
			yaml:          "file_type: directory",
			expectedTypes: []string{"directory"},
		},
		{
			name:          "scalar with alias",
			yaml:          "file_type: dir",
			expectedTypes: []string{"dir"},
		},
		{
			name:          "sequence",
			yaml:          "file_type: [file, directory]",
			expectedTypes: []string{"file", "directory"},
		},
		{
			name:          "mapping with scalar",
			yaml:          "file_type:\n  types: file",
			expectedTypes: []string{"file"},
		},
		{
			name:          "mapping with sequence",
			yaml:          "file_type:\n  types: [file, symlink]",
			expectedTypes: []string{"file", "symlink"},
		},
		{
			name:    "invalid type",
			yaml:    "file_type: invalid",
			wantErr: true,
		},
		{
			name:    "invalid type in sequence",
			yaml:    "file_type: [file, invalid]",
			wantErr: true,
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

			if f.Name != "file_type" {
				t.Errorf("filter name = %q, want %q", f.Name, "file_type")
				return
			}

			ftFilter, ok := f.Inner.(*FileType)
			if !ok {
				t.Errorf("inner is not *FileType, got %T", f.Inner)
				return
			}

			if len(ftFilter.Types) != len(tt.expectedTypes) {
				t.Errorf("Types length = %d, want %d", len(ftFilter.Types), len(tt.expectedTypes))
				return
			}

			for i, ft := range tt.expectedTypes {
				if ftFilter.Types[i] != ft {
					t.Errorf("Types[%d] = %q, want %q", i, ftFilter.Types[i], ft)
				}
			}
		})
	}
}
