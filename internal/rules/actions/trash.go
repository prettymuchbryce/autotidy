package actions

import (
	"github.com/prettymuchbryce/autotidy/internal/fs"
	"github.com/prettymuchbryce/autotidy/internal/rules"

	"gopkg.in/yaml.v3"
)

func init() {
	rules.RegisterAction("trash", deserializeTrash)
}

// Trash is an action that moves files to the system trash.
type Trash struct{}

// Execute moves the file or directory to the system trash.
func (t *Trash) Execute(path string, filesystem fs.FileSystem) (*rules.ExecutionResult, error) {
	if err := filesystem.Trash(path); err != nil {
		return nil, err
	}

	return &rules.ExecutionResult{
		Deleted: true,
	}, nil
}

// deserializeTrash creates a Trash action from YAML.
// Usage: "trash" or "trash: null" or "trash: {}"
func deserializeTrash(node yaml.Node) (rules.Executable, error) {
	return &Trash{}, nil
}
