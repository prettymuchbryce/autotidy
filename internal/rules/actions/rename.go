package actions

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/prettymuchbryce/autotidy/internal/fs"
	"github.com/prettymuchbryce/autotidy/internal/rules"
	"github.com/prettymuchbryce/autotidy/internal/utils"

	"gopkg.in/yaml.v3"
)

func init() {
	rules.RegisterAction("rename", deserializeRename)
}

// Rename is an action that renames files in place.
type Rename struct {
	NewName    utils.Template
	OnConflict fs.ConflictMode // Defaults to rename_with_suffix
}

// getConflictMode returns the conflict mode, defaulting to rename_with_suffix.
func (r *Rename) getConflictMode() fs.ConflictMode {
	if r.OnConflict == "" {
		return fs.ConflictRenameWithSuffix
	}
	return r.OnConflict
}

// Execute renames the file to the new name in the same directory.
func (r *Rename) Execute(path string, filesystem fs.FileSystem) (*rules.ExecutionResult, error) {
	newName := r.NewName.ExpandWithNameExt(path).ExpandWithTime().String()

	// Validate that new_name doesn't contain path separators
	if strings.ContainsRune(newName, filepath.Separator) {
		return nil, fmt.Errorf("rename new_name must not contain path separators: %s", newName)
	}

	dir := filepath.Dir(path)
	destPath := filepath.Join(dir, newName)

	// Skip if source and destination are the same
	if path == destPath {
		return nil, nil
	}

	// Check if destination file already exists
	if _, err := filesystem.Stat(destPath); err == nil {
		newDestPath, proceed, err := filesystem.ResolveConflict(r.getConflictMode(), path, destPath)
		if err != nil {
			return nil, err
		}
		if !proceed {
			return &rules.ExecutionResult{ConflictAlreadyExists: true}, nil
		}
		destPath = newDestPath
	}

	if err := filesystem.Rename(path, destPath); err != nil {
		return nil, err
	}

	return &rules.ExecutionResult{
		NewPath: destPath,
	}, nil
}

// deserializeRename creates a Rename action from YAML.
// Supports both "rename: newfile.txt" and "rename: {new_name: newfile.txt, on_conflict: overwrite}".
func deserializeRename(node yaml.Node) (rules.Executable, error) {
	// Try as plain string first
	if node.Kind == yaml.ScalarNode {
		var newName string
		if err := node.Decode(&newName); err != nil {
			return nil, err
		}
		return &Rename{NewName: utils.Template(newName)}, nil
	}

	// Otherwise expect a mapping with "new_name" key and optional "on_conflict"
	var m struct {
		NewName    utils.Template  `yaml:"new_name"`
		OnConflict fs.ConflictMode `yaml:"on_conflict"`
	}
	if err := node.Decode(&m); err != nil {
		return nil, err
	}

	if m.NewName == "" {
		return nil, fmt.Errorf("rename action requires new_name")
	}

	return &Rename{NewName: m.NewName, OnConflict: m.OnConflict}, nil
}
