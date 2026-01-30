package filters

import (
	"fmt"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/gabriel-vasile/mimetype"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
	"github.com/prettymuchbryce/autotidy/internal/rules"
)

func init() {
	rules.RegisterFilter("mime_type", deserializeMimeType)
}

// MimeType is a filter that matches files by MIME type.
type MimeType struct {
	MimeTypes []string
	Fs        afero.Fs
}

// Evaluate checks if the file's MIME type matches any of the patterns.
func (m *MimeType) Evaluate(path string) (bool, error) {
	fs := m.Fs
	if fs == nil {
		fs = afero.NewOsFs()
	}

	// Check if file exists and is not a directory
	info, err := fs.Stat(path)
	if err != nil {
		return false, err
	}
	if info.IsDir() {
		return false, nil
	}

	// Open file and detect MIME type using reader
	f, err := fs.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	detected, err := mimetype.DetectReader(f)
	if err != nil {
		return false, err
	}

	mimeType := detected.String()

	// Check against each pattern
	for _, pattern := range m.MimeTypes {
		matched, err := doublestar.Match(pattern, mimeType)
		if err != nil {
			return false, fmt.Errorf("invalid mime_type pattern %q: %w", pattern, err)
		}
		if matched {
			return true, nil
		}
	}

	return false, nil
}

// deserializeMimeType creates a MimeType filter from YAML.
// Supports:
//   - "mime_type: image/*"
//   - "mime_type: [image/*, video/*]"
//   - "mime_type: {mime_types: image/*}"
//   - "mime_type: {mime_types: [image/*, video/*]}"
func deserializeMimeType(node yaml.Node) (rules.Evaluable, error) {
	// Try as plain string first
	if node.Kind == yaml.ScalarNode {
		var mimeType string
		if err := node.Decode(&mimeType); err != nil {
			return nil, err
		}
		return &MimeType{MimeTypes: []string{mimeType}, Fs: afero.NewOsFs()}, nil
	}

	// Try as sequence of strings
	if node.Kind == yaml.SequenceNode {
		var mimeTypes []string
		if err := node.Decode(&mimeTypes); err != nil {
			return nil, err
		}
		return &MimeType{MimeTypes: mimeTypes, Fs: afero.NewOsFs()}, nil
	}

	// Otherwise expect a mapping with "mime_types" key
	var m struct {
		MimeTypes rules.StringList `yaml:"mime_types"`
	}
	if err := node.Decode(&m); err != nil {
		return nil, err
	}

	if len(m.MimeTypes) == 0 {
		return nil, fmt.Errorf("mime_type filter requires at least one mime type pattern")
	}

	return &MimeType{MimeTypes: m.MimeTypes, Fs: afero.NewOsFs()}, nil
}
