package actions

import (
	"testing"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	"github.com/prettymuchbryce/autotidy/internal/fs"
	"github.com/prettymuchbryce/autotidy/internal/rules"
	"github.com/prettymuchbryce/autotidy/internal/testutil"
)

func TestDelete_Execute(t *testing.T) {
	dir := testutil.Path("/", "dir")
	dirFile := testutil.Path(dir, "file.txt")
	dirEmpty := testutil.Path(dir, "empty")
	dirNonempty := testutil.Path(dir, "nonempty")
	dirNonemptySub := testutil.Path(dirNonempty, "sub")
	dirNonemptyFile := testutil.Path(dirNonempty, "file.txt")
	dirNonemptySubNested := testutil.Path(dirNonemptySub, "nested.txt")
	nonexistent := testutil.Path("/", "nonexistent")

	tests := []struct {
		name    string
		setup   func(filesystem fs.FileSystem)
		path    string
		wantErr bool
	}{
		{
			name: "deletes file",
			setup: func(filesystem fs.FileSystem) {
				filesystem.MkdirAll(dir, 0755)
				afero.WriteFile(filesystem, dirFile, []byte("content"), 0644)
			},
			path:    dirFile,
			wantErr: false,
		},
		{
			name: "deletes empty directory",
			setup: func(filesystem fs.FileSystem) {
				filesystem.MkdirAll(dirEmpty, 0755)
			},
			path:    dirEmpty,
			wantErr: false,
		},
		{
			name: "deletes directory with contents",
			setup: func(filesystem fs.FileSystem) {
				filesystem.MkdirAll(dirNonemptySub, 0755)
				afero.WriteFile(filesystem, dirNonemptyFile, []byte("content"), 0644)
				afero.WriteFile(filesystem, dirNonemptySubNested, []byte("nested"), 0644)
			},
			path:    dirNonempty,
			wantErr: false,
		},
		{
			name:    "errors on non-existent path",
			setup:   func(filesystem fs.FileSystem) {},
			path:    nonexistent,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filesystem := fs.NewMem()
			tt.setup(filesystem)

			d := &Delete{}

			result, err := d.Execute(tt.path, filesystem)

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

			if result == nil {
				t.Errorf("expected result, got nil")
				return
			}

			if !result.Deleted {
				t.Errorf("Deleted = false, want true")
			}

			// Verify path no longer exists
			exists, err := afero.Exists(filesystem, tt.path)
			if err != nil {
				t.Errorf("error checking path: %v", err)
			}
			if exists {
				t.Errorf("path should not exist after delete: %s", tt.path)
			}
		})
	}
}

func TestDeserializeDelete(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name:    "bare action name",
			yaml:    "delete",
			wantErr: false,
		},
		{
			name:    "null value",
			yaml:    "delete: null",
			wantErr: false,
		},
		{
			name:    "empty mapping",
			yaml:    "delete: {}",
			wantErr: false,
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

			if a.Name != "delete" {
				t.Errorf("action name = %q, want %q", a.Name, "delete")
				return
			}

			_, ok := a.Inner.(*Delete)
			if !ok {
				t.Errorf("inner is not *Delete, got %T", a.Inner)
				return
			}
		})
	}
}
