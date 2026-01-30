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
	rules.RegisterAction("copy", deserializeCopy)
}

// Copy is an action that copies files in the same directory.
type Copy struct {
	NewName    utils.Template
	OnConflict fs.ConflictMode // Defaults to rename_with_suffix
}

// getConflictMode returns the conflict mode, defaulting to rename_with_suffix.
func (c *Copy) getConflictMode() fs.ConflictMode {
	if c.OnConflict == "" {
		return fs.ConflictRenameWithSuffix
	}
	return c.OnConflict
}

// Execute copies the file to a new name in the same directory.
func (c *Copy) Execute(path string, filesystem fs.FileSystem) (*rules.ExecutionResult, error) {
	newName := c.NewName.ExpandWithNameExt(path).ExpandWithTime().String()

	// Validate that new_name doesn't contain path separators
	if strings.ContainsRune(newName, filepath.Separator) {
		return nil, fmt.Errorf("copy new_name must not contain path separators: %s", newName)
	}

	dir := filepath.Dir(path)
	destPath := filepath.Join(dir, newName)

	// Skip if source and destination are the same
	if path == destPath {
		return nil, nil
	}

	// Check if destination file already exists
	if _, err := filesystem.Stat(destPath); err == nil {
		newDestPath, proceed, err := filesystem.ResolveConflict(c.getConflictMode(), path, destPath)
		if err != nil {
			return nil, err
		}
		if !proceed {
			return &rules.ExecutionResult{ConflictAlreadyExists: true}, nil
		}
		destPath = newDestPath
	}

	if err := filesystem.Copy(path, destPath); err != nil {
		return nil, err
	}

	return &rules.ExecutionResult{
		NewPath: destPath,
	}, nil
}

// deserializeCopy creates a Copy action from YAML.
// Supports both "copy: backup.txt" and "copy: {new_name: backup.txt, on_conflict: overwrite}".
func deserializeCopy(node yaml.Node) (rules.Executable, error) {
	// Try as plain string first
	if node.Kind == yaml.ScalarNode {
		var newName string
		if err := node.Decode(&newName); err != nil {
			return nil, err
		}
		return &Copy{NewName: utils.Template(newName)}, nil
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
		return nil, fmt.Errorf("copy action requires new_name")
	}

	return &Copy{NewName: m.NewName, OnConflict: m.OnConflict}, nil
}
