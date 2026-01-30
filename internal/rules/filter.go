package rules

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Evaluable is the interface that filters implement.
type Evaluable interface {
	Evaluate(path string) (bool, error)
}

// FilterDeserializer is a function that creates an Evaluable from a YAML value.
type FilterDeserializer func(value yaml.Node) (Evaluable, error)

// filterRegistry holds registered filter deserializers.
var filterRegistry = map[string]FilterDeserializer{}

// RegisterFilter registers a filter deserializer by name.
func RegisterFilter(name string, deserializer FilterDeserializer) {
	filterRegistry[name] = deserializer
}

// Filter wraps an Evaluable with its name for debugging.
type Filter struct {
	Name  string
	Inner Evaluable
}

// Evaluate delegates to the inner Evaluable.
func (f *Filter) Evaluate(path string) (bool, error) {
	return f.Inner.Evaluate(path)
}

// UnmarshalYAML implements custom YAML unmarshaling for Filter.
// It expects a mapping with exactly one key that matches a registered filter name.
func (f *Filter) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("filter must be a mapping, got %v", node.Kind)
	}

	if len(node.Content) != 2 {
		return fmt.Errorf("filter must have exactly one key, got %d", len(node.Content)/2)
	}

	var name string
	if err := node.Content[0].Decode(&name); err != nil {
		return fmt.Errorf("failed to decode filter name: %w", err)
	}

	deserializer, ok := filterRegistry[name]
	if !ok {
		available := make([]string, 0, len(filterRegistry))
		for k := range filterRegistry {
			available = append(available, k)
		}
		return fmt.Errorf("unknown filter %q, available: %v", name, available)
	}

	inner, err := deserializer(*node.Content[1])
	if err != nil {
		return fmt.Errorf("failed to deserialize filter %q: %w", name, err)
	}

	f.Name = name
	f.Inner = inner
	return nil
}
