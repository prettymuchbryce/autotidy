package actions

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	"github.com/prettymuchbryce/autotidy/internal/fs"
	"github.com/prettymuchbryce/autotidy/internal/rules"
	"github.com/prettymuchbryce/autotidy/internal/testutil"
	"github.com/prettymuchbryce/autotidy/internal/utils"
)

func TestCopy_Execute(t *testing.T) {
	dir := testutil.Path("/", "dir")

	tests := []struct {
		name            string
		newName         string
		srcPath         string
		srcContent      string
		expectedNewPath string
		wantErr         bool
	}{
		{
			name:            "copies file in same directory",
			newName:         "copied.txt",
			srcPath:         testutil.Path(dir, "original.txt"),
			srcContent:      "hello world",
			expectedNewPath: testutil.Path(dir, "copied.txt"),
			wantErr:         false,
		},
		{
			name:            "same name returns nil",
			newName:         "file.txt",
			srcPath:         testutil.Path(dir, "file.txt"),
			srcContent:      "content",
			expectedNewPath: "",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filesystem := fs.NewMem()

			// Create source directory and file
			if err := filesystem.MkdirAll(dir, 0755); err != nil {
				t.Fatalf("failed to create dir: %v", err)
			}
			if err := afero.WriteFile(filesystem, tt.srcPath, []byte(tt.srcContent), 0644); err != nil {
				t.Fatalf("failed to create source file: %v", err)
			}

			c := &Copy{
				NewName: utils.Template(tt.newName),
			}

			result, err := c.Execute(tt.srcPath, filesystem)

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

			if tt.expectedNewPath == "" {
				if result != nil {
					t.Errorf("expected nil result, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Errorf("expected result, got nil")
				return
			}

			if result.NewPath != tt.expectedNewPath {
				t.Errorf("NewPath = %q, want %q", result.NewPath, tt.expectedNewPath)
			}

			// Verify copy exists at new location
			exists, err := afero.Exists(filesystem, tt.expectedNewPath)
			if err != nil {
				t.Errorf("error checking dest file: %v", err)
			}
			if !exists {
				t.Errorf("copied file should exist at %s", tt.expectedNewPath)
			}

			// Verify source still exists (unlike rename)
			exists, err = afero.Exists(filesystem, tt.srcPath)
			if err != nil {
				t.Errorf("error checking source file: %v", err)
			}
			if !exists {
				t.Errorf("source file should still exist at %s after copy", tt.srcPath)
			}

			// Verify content preserved in both files
			srcContent, err := afero.ReadFile(filesystem, tt.srcPath)
			if err != nil {
				t.Errorf("error reading source file: %v", err)
			}
			if string(srcContent) != tt.srcContent {
				t.Errorf("source content = %q, want %q", string(srcContent), tt.srcContent)
			}

			dstContent, err := afero.ReadFile(filesystem, tt.expectedNewPath)
			if err != nil {
				t.Errorf("error reading dest file: %v", err)
			}
			if string(dstContent) != tt.srcContent {
				t.Errorf("dest content = %q, want %q", string(dstContent), tt.srcContent)
			}
		})
	}
}

func TestCopy_Execute_PathSeparatorInvalid(t *testing.T) {
	dir := testutil.Path("/", "dir")
	srcFile := testutil.Path(dir, "file.txt")

	filesystem := fs.NewMem()
	filesystem.MkdirAll(dir, 0755)
	afero.WriteFile(filesystem, srcFile, []byte("content"), 0644)

	// Use OS-specific separator for cross-platform correctness
	sep := string(filepath.Separator)
	tests := []struct {
		name    string
		newName string
	}{
		{"path separator", "sub" + sep + "file.txt"},
		{"multiple separators", strings.Join([]string{"a", "b", "c.txt"}, sep)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Copy{
				NewName: utils.Template(tt.newName),
			}

			_, err := c.Execute(srcFile, filesystem)
			if err == nil {
				t.Error("expected error for path separator in new_name")
			}
		})
	}
}

func TestCopy_Execute_ConflictSkip(t *testing.T) {
	dir := testutil.Path("/", "dir")
	srcFile := testutil.Path(dir, "source.txt")
	destFile := testutil.Path(dir, "dest.txt")

	filesystem := fs.NewMem()
	filesystem.MkdirAll(dir, 0755)

	// Create source and destination files
	afero.WriteFile(filesystem, srcFile, []byte("source content"), 0644)
	afero.WriteFile(filesystem, destFile, []byte("dest content"), 0644)

	c := &Copy{
		NewName:    utils.Template("dest.txt"),
		OnConflict: fs.ConflictSkip,
	}

	result, err := c.Execute(srcFile, filesystem)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return ConflictAlreadyExists: true
	if result == nil || !result.ConflictAlreadyExists {
		t.Errorf("expected ConflictAlreadyExists result, got %+v", result)
	}

	// Source should still exist
	exists, _ := afero.Exists(filesystem, srcFile)
	if !exists {
		t.Error("source file should still exist after skip")
	}

	// Destination should be unchanged
	content, _ := afero.ReadFile(filesystem, destFile)
	if string(content) != "dest content" {
		t.Errorf("destination content = %q, want %q", string(content), "dest content")
	}
}

func TestCopy_Execute_ConflictOverwrite(t *testing.T) {
	dir := testutil.Path("/", "dir")
	srcFile := testutil.Path(dir, "source.txt")
	destFile := testutil.Path(dir, "dest.txt")

	filesystem := fs.NewMem()
	filesystem.MkdirAll(dir, 0755)

	// Create source and destination files
	afero.WriteFile(filesystem, srcFile, []byte("source content"), 0644)
	afero.WriteFile(filesystem, destFile, []byte("dest content"), 0644)

	c := &Copy{
		NewName:    utils.Template("dest.txt"),
		OnConflict: fs.ConflictOverwrite,
	}

	result, err := c.Execute(srcFile, filesystem)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.NewPath != destFile {
		t.Errorf("NewPath = %q, want %q", result.NewPath, destFile)
	}

	// Source should still exist (it's a copy)
	exists, _ := afero.Exists(filesystem, srcFile)
	if !exists {
		t.Error("source file should still exist after copy")
	}

	// Destination should have source content
	content, _ := afero.ReadFile(filesystem, destFile)
	if string(content) != "source content" {
		t.Errorf("destination content = %q, want %q", string(content), "source content")
	}
}

func TestCopy_Execute_Directory(t *testing.T) {
	dir := testutil.Path("/", "dir")
	srcDir := testutil.Path(dir, "srcdir")
	srcSubdir := testutil.Path(srcDir, "subdir")
	destDir := testutil.Path(dir, "destdir")
	destSubdir := testutil.Path(destDir, "subdir")

	filesystem := fs.NewMem()

	// Create source directory with files
	filesystem.MkdirAll(srcSubdir, 0755)
	afero.WriteFile(filesystem, testutil.Path(srcDir, "file1.txt"), []byte("content1"), 0644)
	afero.WriteFile(filesystem, testutil.Path(srcSubdir, "file2.txt"), []byte("content2"), 0644)

	c := &Copy{
		NewName: utils.Template("destdir"),
	}

	result, err := c.Execute(srcDir, filesystem)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.NewPath != destDir {
		t.Errorf("NewPath = %q, want %q", result.NewPath, destDir)
	}

	// Verify copied directory structure
	exists, _ := afero.DirExists(filesystem, destDir)
	if !exists {
		t.Error("destination directory should exist")
	}

	exists, _ = afero.DirExists(filesystem, destSubdir)
	if !exists {
		t.Error("destination subdirectory should exist")
	}

	content, _ := afero.ReadFile(filesystem, testutil.Path(destDir, "file1.txt"))
	if string(content) != "content1" {
		t.Errorf("file1 content = %q, want %q", string(content), "content1")
	}

	content, _ = afero.ReadFile(filesystem, testutil.Path(destSubdir, "file2.txt"))
	if string(content) != "content2" {
		t.Errorf("file2 content = %q, want %q", string(content), "content2")
	}

	// Verify source still exists
	exists, _ = afero.DirExists(filesystem, srcDir)
	if !exists {
		t.Error("source directory should still exist after copy")
	}
}

func TestDeserializeCopy(t *testing.T) {
	tests := []struct {
		name               string
		yaml               string
		expectedNewName    string
		expectedOnConflict fs.ConflictMode
		wantErr            bool
	}{
		{
			name:               "mapping with new_name",
			yaml:               "copy:\n  new_name: newfile.txt",
			expectedNewName:    "newfile.txt",
			expectedOnConflict: "",
			wantErr:            false,
		},
		{
			name:               "mapping with on_conflict skip",
			yaml:               "copy:\n  new_name: newfile.txt\n  on_conflict: skip",
			expectedNewName:    "newfile.txt",
			expectedOnConflict: fs.ConflictSkip,
			wantErr:            false,
		},
		{
			name:               "mapping with on_conflict overwrite",
			yaml:               "copy:\n  new_name: newfile.txt\n  on_conflict: overwrite",
			expectedNewName:    "newfile.txt",
			expectedOnConflict: fs.ConflictOverwrite,
			wantErr:            false,
		},
		{
			name:    "missing new_name",
			yaml:    "copy:\n  on_conflict: skip",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a rules.Action
			err := yaml.Unmarshal([]byte(tt.yaml), &a)

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

			if a.Name != "copy" {
				t.Errorf("action name = %q, want %q", a.Name, "copy")
				return
			}

			copyAction, ok := a.Inner.(*Copy)
			if !ok {
				t.Errorf("inner is not *Copy, got %T", a.Inner)
				return
			}

			if copyAction.NewName.String() != tt.expectedNewName {
				t.Errorf("NewName = %q, want %q", copyAction.NewName.String(), tt.expectedNewName)
			}

			if copyAction.OnConflict != tt.expectedOnConflict {
				t.Errorf("OnConflict = %q, want %q", copyAction.OnConflict, tt.expectedOnConflict)
			}
		})
	}
}
