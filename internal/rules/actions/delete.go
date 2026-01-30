package actions

import (
	"github.com/prettymuchbryce/autotidy/internal/fs"
	"github.com/prettymuchbryce/autotidy/internal/rules"

	"gopkg.in/yaml.v3"
)

func init() {
	rules.RegisterAction("delete", deserializeDelete)
}

// Delete is an action that permanently deletes files.
type Delete struct{}

// Execute deletes the file or directory.
func (d *Delete) Execute(path string, filesystem fs.FileSystem) (*rules.ExecutionResult, error) {
	info, err := filesystem.Stat(path)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		if err := filesystem.RemoveAll(path); err != nil {
			return nil, err
		}
	} else {
		if err := filesystem.Remove(path); err != nil {
			return nil, err
		}
	}

	return &rules.ExecutionResult{
		Deleted: true,
	}, nil
}

// deserializeDelete creates a Delete action from YAML.
// Usage: "delete" or "delete: null" or "delete: {}"
func deserializeDelete(node yaml.Node) (rules.Executable, error) {
	return &Delete{}, nil
}
