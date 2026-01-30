package actions

import (
	"fmt"
	"path/filepath"

	"github.com/prettymuchbryce/autotidy/internal/fs"
	"github.com/prettymuchbryce/autotidy/internal/rules"
	"github.com/prettymuchbryce/autotidy/internal/utils"

	"gopkg.in/yaml.v3"
)

func init() {
	rules.RegisterAction("move", deserializeMove)
}

// Move is an action that moves files to a destination directory.
type Move struct {
	Dest       utils.Template
	OnConflict fs.ConflictMode // Defaults to rename_with_suffix
}

// getConflictMode returns the conflict mode, defaulting to rename_with_suffix.
func (m *Move) getConflictMode() fs.ConflictMode {
	if m.OnConflict == "" {
		return fs.ConflictRenameWithSuffix
	}
	return m.OnConflict
}

// Execute moves the file to the destination directory.
// The destination must be a directory, not a file path.
func (m *Move) Execute(path string, filesystem fs.FileSystem) (*rules.ExecutionResult, error) {
	destDir := m.Dest.ExpandTilde().ExpandWithNameExt(path).ExpandWithTime().String()
	filename := filepath.Base(path)
	destPath := filepath.Join(destDir, filename)

	// Skip if source and destination are the same
	if path == destPath {
		return nil, nil
	}

	// Check if destination directory path exists and is a file (not allowed)
	if info, err := filesystem.Stat(destDir); err == nil && !info.IsDir() {
		return nil, fmt.Errorf("move destination must be a directory, not a file: %s", destDir)
	}

	// Create destination directory if it doesn't exist
	if err := filesystem.MkdirAll(destDir, 0755); err != nil {
		return nil, err
	}

	// Check if destination file already exists
	if _, err := filesystem.Stat(destPath); err == nil {
		newDestPath, proceed, err := filesystem.ResolveConflict(m.getConflictMode(), path, destPath)
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

// deserializeMove creates a Move action from YAML.
// Supports both "move: ~/dest" and "move: {dest: ~/dest, on_conflict: skip}".
func deserializeMove(node yaml.Node) (rules.Executable, error) {
	// Try as plain string first
	if node.Kind == yaml.ScalarNode {
		var dest string
		if err := node.Decode(&dest); err != nil {
			return nil, err
		}
		return &Move{Dest: utils.Template(dest)}, nil
	}

	// Otherwise expect a mapping with "dest" key and optional "on_conflict"
	var m struct {
		Dest       utils.Template  `yaml:"dest"`
		OnConflict fs.ConflictMode `yaml:"on_conflict"`
	}
	if err := node.Decode(&m); err != nil {
		return nil, err
	}
	return &Move{Dest: m.Dest, OnConflict: m.OnConflict}, nil
}
