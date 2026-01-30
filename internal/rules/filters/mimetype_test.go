package filters

import (
	"testing"

	"github.com/prettymuchbryce/autotidy/internal/rules"
	"github.com/prettymuchbryce/autotidy/internal/testutil"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

func TestMimeType_Evaluate(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Define paths
	textPath := testutil.Path("/", "test.txt")
	zipPath := testutil.Path("/", "test.zip")
	subdirPath := testutil.Path("/", "subdir")

	// Create a text file
	afero.WriteFile(fs, textPath, []byte("hello world"), 0644)

	// Create a ZIP file (minimal valid ZIP - empty archive)
	zipData := []byte{
		0x50, 0x4B, 0x05, 0x06, // End of central directory signature
		0x00, 0x00, 0x00, 0x00, // Disk numbers
		0x00, 0x00, 0x00, 0x00, // Entry counts
		0x00, 0x00, 0x00, 0x00, // Central directory size
		0x00, 0x00, 0x00, 0x00, // Central directory offset
		0x00, 0x00, // Comment length
	}
	afero.WriteFile(fs, zipPath, zipData, 0644)

	// Create a directory
	fs.Mkdir(subdirPath, 0755)

	tests := []struct {
		name      string
		mimeTypes []string
		path      string
		want      bool
		wantErr   bool
	}{
		{
			name:      "matches text file with glob",
			mimeTypes: []string{"text/*"},
			path:      textPath,
			want:      true,
		},
		{
			name:      "matches zip with application glob",
			mimeTypes: []string{"application/*"},
			path:      zipPath,
			want:      true,
		},
		{
			name:      "matches zip with exact type",
			mimeTypes: []string{"application/zip"},
			path:      zipPath,
			want:      true,
		},
		{
			name:      "no match for wrong type",
			mimeTypes: []string{"video/*"},
			path:      textPath,
			want:      false,
		},
		{
			name:      "matches with multiple patterns",
			mimeTypes: []string{"video/*", "text/*"},
			path:      textPath,
			want:      true,
		},
		{
			name:      "directory returns false",
			mimeTypes: []string{"*/*"},
			path:      subdirPath,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MimeType{MimeTypes: tt.mimeTypes, Fs: fs}
			got, err := m.Evaluate(tt.path)

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

func TestDeserializeMimeType(t *testing.T) {
	tests := []struct {
		name              string
		yaml              string
		expectedMimeTypes []string
		wantErr           bool
	}{
		{
			name:              "scalar string",
			yaml:              "mime_type: image/*",
			expectedMimeTypes: []string{"image/*"},
		},
		{
			name:              "sequence",
			yaml:              "mime_type: [image/*, video/*]",
			expectedMimeTypes: []string{"image/*", "video/*"},
		},
		{
			name:              "mapping with scalar",
			yaml:              "mime_type:\n  mime_types: text/plain",
			expectedMimeTypes: []string{"text/plain"},
		},
		{
			name:              "mapping with sequence",
			yaml:              "mime_type:\n  mime_types: [image/png, image/jpeg]",
			expectedMimeTypes: []string{"image/png", "image/jpeg"},
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

			if f.Name != "mime_type" {
				t.Errorf("filter name = %q, want %q", f.Name, "mime_type")
				return
			}

			mimeFilter, ok := f.Inner.(*MimeType)
			if !ok {
				t.Errorf("inner is not *MimeType, got %T", f.Inner)
				return
			}

			if len(mimeFilter.MimeTypes) != len(tt.expectedMimeTypes) {
				t.Errorf("MimeTypes length = %d, want %d", len(mimeFilter.MimeTypes), len(tt.expectedMimeTypes))
				return
			}

			for i, mt := range tt.expectedMimeTypes {
				if mimeFilter.MimeTypes[i] != mt {
					t.Errorf("MimeTypes[%d] = %q, want %q", i, mimeFilter.MimeTypes[i], mt)
				}
			}
		})
	}
}
