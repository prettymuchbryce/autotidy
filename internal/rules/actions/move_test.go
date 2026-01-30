package actions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	"github.com/prettymuchbryce/autotidy/internal/fs"
	"github.com/prettymuchbryce/autotidy/internal/rules"
	"github.com/prettymuchbryce/autotidy/internal/testutil"
	"github.com/prettymuchbryce/autotidy/internal/utils"
)

func TestMove_Execute(t *testing.T) {
	home, _ := os.UserHomeDir()
	src := testutil.Path("/", "src")
	dest := testutil.Path("/", "dest")
	nestedDest := testutil.Path("/", "new", "nested", "dest")

	tests := []struct {
		name            string
		dest            string
		srcPath         string
		srcContent      string
		expectedNewPath string
		wantErr         bool
	}{
		{
			name:            "moves file to destination",
			dest:            dest,
			srcPath:         testutil.Path(src, "file.txt"),
			srcContent:      "hello world",
			expectedNewPath: testutil.Path(dest, "file.txt"),
			wantErr:         false,
		},
		{
			name:            "creates destination directory",
			dest:            nestedDest,
			srcPath:         testutil.Path(src, "file.txt"),
			srcContent:      "content",
			expectedNewPath: testutil.Path(nestedDest, "file.txt"),
			wantErr:         false,
		},
		{
			name:            "same source and dest returns nil",
			dest:            src,
			srcPath:         testutil.Path(src, "file.txt"),
			srcContent:      "content",
			expectedNewPath: "",
			wantErr:         false,
		},
		{
			name:            "expands tilde in destination",
			dest:            "~/dest",
			srcPath:         testutil.Path(src, "file.txt"),
			srcContent:      "content",
			expectedNewPath: filepath.Join(home, "dest", "file.txt"),
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filesystem := fs.NewMem()

			// Create source directory and file
			srcDir := src
			if err := filesystem.MkdirAll(srcDir, 0755); err != nil {
				t.Fatalf("failed to create source dir: %v", err)
			}
			if err := afero.WriteFile(filesystem, tt.srcPath, []byte(tt.srcContent), 0644); err != nil {
				t.Fatalf("failed to create source file: %v", err)
			}

			// Create action
			m := &Move{
				Dest: utils.Template(tt.dest),
			}

			// Execute
			result, err := m.Execute(tt.srcPath, filesystem)

			// Check error
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

			// Check result
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

			// Verify file exists at destination
			exists, err := afero.Exists(filesystem, tt.expectedNewPath)
			if err != nil {
				t.Errorf("error checking dest file: %v", err)
			}
			if !exists {
				t.Errorf("destination file should exist at %s", tt.expectedNewPath)
			}

			// Verify source no longer exists
			exists, err = afero.Exists(filesystem, tt.srcPath)
			if err != nil {
				t.Errorf("error checking source file: %v", err)
			}
			if exists {
				t.Errorf("source file should not exist at %s after move", tt.srcPath)
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

func TestMove_Execute_NonExistentSource(t *testing.T) {
	dest := testutil.Path("/", "dest")
	nonexistent := testutil.Path("/", "nonexistent", "file.txt")

	filesystem := fs.NewMem()

	m := &Move{
		Dest: utils.Template(dest),
	}

	_, err := m.Execute(nonexistent, filesystem)
	if err == nil {
		t.Errorf("expected error for non-existent source, got none")
	}
}

func TestMove_Execute_DestinationIsFile(t *testing.T) {
	src := testutil.Path("/", "src")
	srcFile := testutil.Path(src, "file.txt")
	dest := testutil.Path("/", "dest")

	filesystem := fs.NewMem()
	filesystem.MkdirAll(src, 0755)

	// Create source file
	afero.WriteFile(filesystem, srcFile, []byte("content"), 0644)

	// Create a file at the destination path (not a directory)
	afero.WriteFile(filesystem, dest, []byte("I am a file"), 0644)

	m := &Move{
		Dest: utils.Template(dest),
	}

	_, err := m.Execute(srcFile, filesystem)
	if err == nil {
		t.Errorf("expected error when destination is a file, got none")
	}

	// Verify error message contains key info
	expectedMsg := "move destination must be a directory, not a file: " + dest
	if err.Error() != expectedMsg {
		t.Errorf("error = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestMove_Execute_ConflictSkip(t *testing.T) {
	src := testutil.Path("/", "src")
	dest := testutil.Path("/", "dest")
	srcFile := testutil.Path(src, "file.txt")
	destFile := testutil.Path(dest, "file.txt")

	filesystem := fs.NewMem()
	filesystem.MkdirAll(src, 0755)
	filesystem.MkdirAll(dest, 0755)

	// Create source and destination files
	afero.WriteFile(filesystem, srcFile, []byte("source content"), 0644)
	afero.WriteFile(filesystem, destFile, []byte("dest content"), 0644)

	m := &Move{
		Dest:       utils.Template(dest),
		OnConflict: fs.ConflictSkip,
	}

	result, err := m.Execute(srcFile, filesystem)
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

func TestMove_Execute_ConflictRenameWithSuffixIsDefault(t *testing.T) {
	src := testutil.Path("/", "src")
	dest := testutil.Path("/", "dest")
	srcFile := testutil.Path(src, "file.txt")
	destFile := testutil.Path(dest, "file.txt")
	expectedDestFile := testutil.Path(dest, "file_2.txt")

	filesystem := fs.NewMem()
	filesystem.MkdirAll(src, 0755)
	filesystem.MkdirAll(dest, 0755)

	// Create source and destination files
	afero.WriteFile(filesystem, srcFile, []byte("source content"), 0644)
	afero.WriteFile(filesystem, destFile, []byte("dest content"), 0644)

	m := &Move{
		Dest: utils.Template(dest),
		// OnConflict not set - should default to rename_with_suffix
	}

	result, err := m.Execute(srcFile, filesystem)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should move to file_2.txt
	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.NewPath != expectedDestFile {
		t.Errorf("NewPath = %q, want %q", result.NewPath, expectedDestFile)
	}

	// Source should be gone
	exists, _ := afero.Exists(filesystem, srcFile)
	if exists {
		t.Error("source file should not exist after move")
	}

	// Original destination should be unchanged
	content, _ := afero.ReadFile(filesystem, destFile)
	if string(content) != "dest content" {
		t.Errorf("original dest content = %q, want %q", string(content), "dest content")
	}

	// New destination should have source content
	content, _ = afero.ReadFile(filesystem, expectedDestFile)
	if string(content) != "source content" {
		t.Errorf("new dest content = %q, want %q", string(content), "source content")
	}
}

func TestMove_Execute_ConflictOverwrite(t *testing.T) {
	src := testutil.Path("/", "src")
	dest := testutil.Path("/", "dest")
	srcFile := testutil.Path(src, "file.txt")
	destFile := testutil.Path(dest, "file.txt")

	filesystem := fs.NewMem()
	filesystem.MkdirAll(src, 0755)
	filesystem.MkdirAll(dest, 0755)

	// Create source and destination files
	afero.WriteFile(filesystem, srcFile, []byte("source content"), 0644)
	afero.WriteFile(filesystem, destFile, []byte("dest content"), 0644)

	m := &Move{
		Dest:       utils.Template(dest),
		OnConflict: fs.ConflictOverwrite,
	}

	result, err := m.Execute(srcFile, filesystem)
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

func TestMove_Execute_ConflictTrashNotImplemented(t *testing.T) {
	src := testutil.Path("/", "src")
	dest := testutil.Path("/", "dest")
	srcFile := testutil.Path(src, "file.txt")
	destFile := testutil.Path(dest, "file.txt")

	filesystem := fs.NewMem()
	filesystem.MkdirAll(src, 0755)
	filesystem.MkdirAll(dest, 0755)

	afero.WriteFile(filesystem, srcFile, []byte("content"), 0644)
	afero.WriteFile(filesystem, destFile, []byte("existing"), 0644)

	m := &Move{
		Dest:       utils.Template(dest),
		OnConflict: fs.ConflictTrash,
	}

	_, err := m.Execute(srcFile, filesystem)
	if err == nil {
		t.Error("expected error for trash mode, got none")
	}

	expectedMsg := "trash conflict mode is not yet implemented"
	if err.Error() != expectedMsg {
		t.Errorf("error = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestDeserializeMove(t *testing.T) {
	tests := []struct {
		name               string
		yaml               string
		expectedDest       string
		expectedOnConflict fs.ConflictMode
		wantErr            bool
	}{
		{
			name:               "scalar string",
			yaml:               "move: /dest/path",
			expectedDest:       "/dest/path",
			expectedOnConflict: "", // defaults to rename_with_suffix at runtime
			wantErr:            false,
		},
		{
			name:               "scalar with tilde",
			yaml:               "move: ~/Documents",
			expectedDest:       "~/Documents",
			expectedOnConflict: "",
			wantErr:            false,
		},
		{
			name:               "mapping with dest",
			yaml:               "move:\n  dest: /other/path",
			expectedDest:       "/other/path",
			expectedOnConflict: "",
			wantErr:            false,
		},
		{
			name:               "mapping with on_conflict skip",
			yaml:               "move:\n  dest: /path\n  on_conflict: skip",
			expectedDest:       "/path",
			expectedOnConflict: fs.ConflictSkip,
			wantErr:            false,
		},
		{
			name:               "mapping with on_conflict overwrite",
			yaml:               "move:\n  dest: /path\n  on_conflict: overwrite",
			expectedDest:       "/path",
			expectedOnConflict: fs.ConflictOverwrite,
			wantErr:            false,
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

			if a.Name != "move" {
				t.Errorf("action name = %q, want %q", a.Name, "move")
				return
			}

			// The inner should be a *Move
			moveAction, ok := a.Inner.(*Move)
			if !ok {
				t.Errorf("inner is not *Move, got %T", a.Inner)
				return
			}

			if moveAction.Dest.String() != tt.expectedDest {
				t.Errorf("Dest = %q, want %q", moveAction.Dest.String(), tt.expectedDest)
			}

			if moveAction.OnConflict != tt.expectedOnConflict {
				t.Errorf("OnConflict = %q, want %q", moveAction.OnConflict, tt.expectedOnConflict)
			}
		})
	}
}
