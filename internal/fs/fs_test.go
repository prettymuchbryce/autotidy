package fs

import (
	"testing"

	"github.com/prettymuchbryce/autotidy/internal/testutil"
	"github.com/spf13/afero"
)

func TestSplitFilenameAndExtensions(t *testing.T) {
	tests := []struct {
		filename     string
		expectedBase string
		expectedExt  string
	}{
		{"file.txt", "file", ".txt"},
		{"archive.tar.gz", "archive", ".tar.gz"},
		{"file", "file", ""},
		{".hidden", ".hidden", ""},
		{".hidden.txt", ".hidden", ".txt"},
		{".hidden.tar.gz", ".hidden", ".tar.gz"},
		{"file.with.many.dots.txt", "file", ".with.many.dots.txt"},
		{"no_extension", "no_extension", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			base, ext := splitFilenameAndExtensions(tt.filename)
			if base != tt.expectedBase {
				t.Errorf("base = %q, want %q", base, tt.expectedBase)
			}
			if ext != tt.expectedExt {
				t.Errorf("ext = %q, want %q", ext, tt.expectedExt)
			}
		})
	}
}

func TestGenerateSuffixedPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		suffix   int
		expected string
	}{
		{
			name:     "simple file",
			path:     testutil.Path("/", "dir", "file.txt"),
			suffix:   2,
			expected: testutil.Path("/", "dir", "file_2.txt"),
		},
		{
			name:     "compound extension",
			path:     testutil.Path("/", "dir", "archive.tar.gz"),
			suffix:   2,
			expected: testutil.Path("/", "dir", "archive_2.tar.gz"),
		},
		{
			name:     "no extension",
			path:     testutil.Path("/", "dir", "file"),
			suffix:   3,
			expected: testutil.Path("/", "dir", "file_3"),
		},
		{
			name:     "hidden file no extension",
			path:     testutil.Path("/", "dir", ".hidden"),
			suffix:   2,
			expected: testutil.Path("/", "dir", ".hidden_2"),
		},
		{
			name:     "hidden file with extension",
			path:     testutil.Path("/", "dir", ".hidden.txt"),
			suffix:   2,
			expected: testutil.Path("/", "dir", ".hidden_2.txt"),
		},
		{
			name:     "higher suffix",
			path:     testutil.Path("/", "dir", "file.txt"),
			suffix:   10,
			expected: testutil.Path("/", "dir", "file_10.txt"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateSuffixedPath(tt.path, tt.suffix)
			if result != tt.expected {
				t.Errorf("GenerateSuffixedPath(%q, %d) = %q, want %q", tt.path, tt.suffix, result, tt.expected)
			}
		})
	}
}

func TestResolveConflict_RenameWithSuffix(t *testing.T) {
	dir := testutil.Path("/", "dest")
	srcFile := testutil.Path("/", "src", "file.txt")

	tests := []struct {
		name              string
		existingFiles     []string
		destPath          string
		expectedNewPath   string
	}{
		{
			name:            "first conflict uses _2",
			existingFiles:   []string{testutil.Path(dir, "file.txt")},
			destPath:        testutil.Path(dir, "file.txt"),
			expectedNewPath: testutil.Path(dir, "file_2.txt"),
		},
		{
			name:            "_2 exists uses _3",
			existingFiles:   []string{testutil.Path(dir, "file.txt"), testutil.Path(dir, "file_2.txt")},
			destPath:        testutil.Path(dir, "file.txt"),
			expectedNewPath: testutil.Path(dir, "file_3.txt"),
		},
		{
			name:            "gap in sequence uses next available",
			existingFiles:   []string{testutil.Path(dir, "file.txt"), testutil.Path(dir, "file_2.txt"), testutil.Path(dir, "file_3.txt")},
			destPath:        testutil.Path(dir, "file.txt"),
			expectedNewPath: testutil.Path(dir, "file_4.txt"),
		},
		{
			name:            "compound extension",
			existingFiles:   []string{testutil.Path(dir, "archive.tar.gz")},
			destPath:        testutil.Path(dir, "archive.tar.gz"),
			expectedNewPath: testutil.Path(dir, "archive_2.tar.gz"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filesystem := NewMem()
			filesystem.MkdirAll(dir, 0755)

			// Create existing files
			for _, f := range tt.existingFiles {
				afero.WriteFile(filesystem, f, []byte("content"), 0644)
			}

			newPath, proceed, err := filesystem.ResolveConflict(ConflictRenameWithSuffix, srcFile, tt.destPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !proceed {
				t.Error("expected proceed=true, got false")
			}

			if newPath != tt.expectedNewPath {
				t.Errorf("newPath = %q, want %q", newPath, tt.expectedNewPath)
			}
		})
	}
}

func TestResolveConflict_Skip(t *testing.T) {
	filesystem := NewMem()
	dir := testutil.Path("/", "dest")
	filesystem.MkdirAll(dir, 0755)

	srcFile := testutil.Path("/", "src", "file.txt")
	destFile := testutil.Path(dir, "file.txt")
	afero.WriteFile(filesystem, destFile, []byte("content"), 0644)

	newPath, proceed, err := filesystem.ResolveConflict(ConflictSkip, srcFile, destFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if proceed {
		t.Error("expected proceed=false for skip mode")
	}

	if newPath != destFile {
		t.Errorf("newPath = %q, want %q", newPath, destFile)
	}
}

func TestResolveConflict_Overwrite(t *testing.T) {
	filesystem := NewMem()
	dir := testutil.Path("/", "dest")
	filesystem.MkdirAll(dir, 0755)

	srcFile := testutil.Path("/", "src", "file.txt")
	destFile := testutil.Path(dir, "file.txt")
	afero.WriteFile(filesystem, destFile, []byte("old content"), 0644)

	newPath, proceed, err := filesystem.ResolveConflict(ConflictOverwrite, srcFile, destFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !proceed {
		t.Error("expected proceed=true for overwrite mode")
	}

	if newPath != destFile {
		t.Errorf("newPath = %q, want %q", newPath, destFile)
	}

	// Destination file should be removed
	exists, _ := afero.Exists(filesystem, destFile)
	if exists {
		t.Error("destination file should be removed after overwrite resolve")
	}
}

func TestResolveConflict_Trash(t *testing.T) {
	filesystem := NewMem()
	srcFile := testutil.Path("/", "src", "file.txt")
	destFile := testutil.Path("/", "dest", "file.txt")

	_, _, err := filesystem.ResolveConflict(ConflictTrash, srcFile, destFile)
	if err == nil {
		t.Error("expected error for trash mode (not implemented)")
	}
}

func TestResolveConflict_UnknownMode(t *testing.T) {
	filesystem := NewMem()
	srcFile := testutil.Path("/", "src", "file.txt")
	destFile := testutil.Path("/", "dest", "file.txt")

	_, _, err := filesystem.ResolveConflict("invalid_mode", srcFile, destFile)
	if err == nil {
		t.Error("expected error for unknown mode")
	}
}
