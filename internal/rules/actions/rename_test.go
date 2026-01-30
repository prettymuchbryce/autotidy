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

func TestRename_Execute(t *testing.T) {
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
			name:            "renames file in same directory",
			newName:         "newfile.txt",
			srcPath:         testutil.Path(dir, "oldfile.txt"),
			srcContent:      "hello world",
			expectedNewPath: testutil.Path(dir, "newfile.txt"),
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

			r := &Rename{
				NewName: utils.Template(tt.newName),
			}

			result, err := r.Execute(tt.srcPath, filesystem)

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

			// Verify file exists at new location
			exists, err := afero.Exists(filesystem, tt.expectedNewPath)
			if err != nil {
				t.Errorf("error checking dest file: %v", err)
			}
			if !exists {
				t.Errorf("renamed file should exist at %s", tt.expectedNewPath)
			}

			// Verify source no longer exists (if different from dest)
			if tt.srcPath != tt.expectedNewPath {
				exists, err = afero.Exists(filesystem, tt.srcPath)
				if err != nil {
					t.Errorf("error checking source file: %v", err)
				}
				if exists {
					t.Errorf("source file should not exist at %s after rename", tt.srcPath)
				}
			}

			// Verify content preserved
			content, err := afero.ReadFile(filesystem, tt.expectedNewPath)
			if err != nil {
				t.Errorf("error reading dest file: %v", err)
			}
			if string(content) != tt.srcContent {
				t.Errorf("content = %q, want %q", string(content), tt.srcContent)
			}
		})
	}
}

func TestRename_Execute_PathSeparatorInvalid(t *testing.T) {
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
			r := &Rename{
				NewName: utils.Template(tt.newName),
			}

			_, err := r.Execute(srcFile, filesystem)
			if err == nil {
				t.Error("expected error for path separator in new_name")
			}
		})
	}
}

func TestRename_Execute_ConflictSkip(t *testing.T) {
	dir := testutil.Path("/", "dir")
	srcFile := testutil.Path(dir, "source.txt")
	destFile := testutil.Path(dir, "dest.txt")

	filesystem := fs.NewMem()
	filesystem.MkdirAll(dir, 0755)

	// Create source and destination files
	afero.WriteFile(filesystem, srcFile, []byte("source content"), 0644)
	afero.WriteFile(filesystem, destFile, []byte("dest content"), 0644)

	r := &Rename{
		NewName:    utils.Template("dest.txt"),
		OnConflict: fs.ConflictSkip,
	}

	result, err := r.Execute(srcFile, filesystem)
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

func TestRename_Execute_ConflictOverwrite(t *testing.T) {
	dir := testutil.Path("/", "dir")
	srcFile := testutil.Path(dir, "source.txt")
	destFile := testutil.Path(dir, "dest.txt")

	filesystem := fs.NewMem()
	filesystem.MkdirAll(dir, 0755)

	// Create source and destination files
	afero.WriteFile(filesystem, srcFile, []byte("source content"), 0644)
	afero.WriteFile(filesystem, destFile, []byte("dest content"), 0644)

	r := &Rename{
		NewName:    utils.Template("dest.txt"),
		OnConflict: fs.ConflictOverwrite,
	}

	result, err := r.Execute(srcFile, filesystem)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.NewPath != destFile {
		t.Errorf("NewPath = %q, want %q", result.NewPath, destFile)
	}

	// Source should be gone
	exists, _ := afero.Exists(filesystem, srcFile)
	if exists {
		t.Error("source file should not exist after overwrite")
	}

	// Destination should have source content
	content, _ := afero.ReadFile(filesystem, destFile)
	if string(content) != "source content" {
		t.Errorf("destination content = %q, want %q", string(content), "source content")
	}
}

func TestDeserializeRename(t *testing.T) {
	tests := []struct {
		name               string
		yaml               string
		expectedNewName    string
		expectedOnConflict fs.ConflictMode
		wantErr            bool
	}{
		{
			name:               "mapping with new_name",
			yaml:               "rename:\n  new_name: newfile.txt",
			expectedNewName:    "newfile.txt",
			expectedOnConflict: "",
			wantErr:            false,
		},
		{
			name:               "mapping with on_conflict skip",
			yaml:               "rename:\n  new_name: newfile.txt\n  on_conflict: skip",
			expectedNewName:    "newfile.txt",
			expectedOnConflict: fs.ConflictSkip,
			wantErr:            false,
		},
		{
			name:               "mapping with on_conflict overwrite",
			yaml:               "rename:\n  new_name: newfile.txt\n  on_conflict: overwrite",
			expectedNewName:    "newfile.txt",
			expectedOnConflict: fs.ConflictOverwrite,
			wantErr:            false,
		},
		{
			name:    "missing new_name",
			yaml:    "rename:\n  on_conflict: skip",
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

			if a.Name != "rename" {
				t.Errorf("action name = %q, want %q", a.Name, "rename")
				return
			}

			renameAction, ok := a.Inner.(*Rename)
			if !ok {
				t.Errorf("inner is not *Rename, got %T", a.Inner)
				return
			}

			if renameAction.NewName.String() != tt.expectedNewName {
				t.Errorf("NewName = %q, want %q", renameAction.NewName.String(), tt.expectedNewName)
			}

			if renameAction.OnConflict != tt.expectedOnConflict {
				t.Errorf("OnConflict = %q, want %q", renameAction.OnConflict, tt.expectedOnConflict)
			}
		})
	}
}
