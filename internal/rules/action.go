package rules

import (
	"fmt"

	"gopkg.in/yaml.v3"
	"github.com/prettymuchbryce/autotidy/internal/fs"
)

// ExecutionResult represents the outcome of executing an action.
// A nil result means the file was not modified and processing should continue.
type ExecutionResult struct {
	NewPath string // New path if file was moved/renamed
	Deleted bool   // True if file was deleted (stop processing actions)
	ConflictAlreadyExists bool // True if skipped because destination already exists
}

// Executable is the interface that actions implement.
// The FileSystem parameter determines behavior - real operations or dry-run logging.
type Executable interface {
	Execute(path string, filesystem fs.FileSystem) (*ExecutionResult, error)
}

// ActionDeserializer is a function that creates an Executable from a YAML value.
type ActionDeserializer func(value yaml.Node) (Executable, error)

// actionRegistry holds registered action deserializers.
var actionRegistry = map[string]ActionDeserializer{}

// RegisterAction registers an action deserializer by name.
func RegisterAction(name string, deserializer ActionDeserializer) {
	actionRegistry[name] = deserializer
}

// Action wraps an Executable with its name for debugging.
type Action struct {
	Name  string
	Inner Executable
}

// Execute delegates to the inner Executable.
func (a *Action) Execute(path string, filesystem fs.FileSystem) (*ExecutionResult, error) {
	return a.Inner.Execute(path, filesystem)
}

// UnmarshalYAML implements custom YAML unmarshaling for Action.
// It supports two formats:
//   - Scalar: "delete" or "trash" (for actions with no arguments)
//   - Mapping: "move: ~/dest" (for actions with arguments)
func (a *Action) UnmarshalYAML(node *yaml.Node) error {
	var name string
	var valueNode yaml.Node

	switch node.Kind {
	case yaml.ScalarNode:
		// Bare action name like "delete" or "trash"
		if err := node.Decode(&name); err != nil {
			return fmt.Errorf("failed to decode action name: %w", err)
		}
		// Create a null node as the value
		valueNode = yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null"}

	case yaml.MappingNode:
		if len(node.Content) != 2 {
			return fmt.Errorf("action must have exactly one key, got %d", len(node.Content)/2)
		}
		if err := node.Content[0].Decode(&name); err != nil {
			return fmt.Errorf("failed to decode action name: %w", err)
		}
		valueNode = *node.Content[1]

	default:
		return fmt.Errorf("action must be a string or mapping, got %v", node.Kind)
	}

	deserializer, ok := actionRegistry[name]
	if !ok {
		available := make([]string, 0, len(actionRegistry))
		for k := range actionRegistry {
			available = append(available, k)
		}
		return fmt.Errorf("unknown action %q, available: %v", name, available)
	}

	inner, err := deserializer(valueNode)
	if err != nil {
		return fmt.Errorf("failed to deserialize action %q: %w", name, err)
	}

	a.Name = name
	a.Inner = inner
	return nil
}
