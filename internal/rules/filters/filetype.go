package filters

import (
	"fmt"
	"os"

	"github.com/prettymuchbryce/autotidy/internal/rules"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

func init() {
	rules.RegisterFilter("file_type", deserializeFileType)
}

// FileType is a filter that matches by file type (file, directory, symlink).
type FileType struct {
	Types []string
	Fs    afero.Fs
}

// Evaluate checks if the path's type matches any of the specified types.
func (f *FileType) Evaluate(path string) (bool, error) {
	fs := f.Fs
	if fs == nil {
		fs = afero.NewOsFs()
	}

	// Use Lstat to not follow symlinks
	info, _, err := fs.(afero.Lstater).LstatIfPossible(path)
	if err != nil {
		return false, err
	}

	mode := info.Mode()
	var actualType string

	switch {
	case mode&os.ModeSymlink != 0:
		actualType = "symlink"
	case mode.IsDir():
		actualType = "directory"
	case mode.IsRegular():
		actualType = "file"
	default:
		actualType = "other"
	}

	for _, t := range f.Types {
		// Normalize type aliases
		normalized := normalizeFileType(t)
		if normalized == actualType {
			return true, nil
		}
	}

	return false, nil
}

// normalizeFileType converts type aliases to canonical names.
func normalizeFileType(t string) string {
	switch t {
	case "dir", "folder":
		return "directory"
	default:
		return t
	}
}

// deserializeFileType creates a FileType filter from YAML.
// Supports:
//   - "file_type: file"
//   - "file_type: [file, directory]"
//   - "file_type: {types: file}"
//   - "file_type: {types: [file, directory]}"
func deserializeFileType(node yaml.Node) (rules.Evaluable, error) {
	// Try as plain string first
	if node.Kind == yaml.ScalarNode {
		var fileType string
		if err := node.Decode(&fileType); err != nil {
			return nil, err
		}
		if err := validateFileType(fileType); err != nil {
			return nil, err
		}
		return &FileType{Types: []string{fileType}, Fs: afero.NewOsFs()}, nil
	}

	// Try as sequence of strings
	if node.Kind == yaml.SequenceNode {
		var types []string
		if err := node.Decode(&types); err != nil {
			return nil, err
		}
		for _, t := range types {
			if err := validateFileType(t); err != nil {
				return nil, err
			}
		}
		return &FileType{Types: types, Fs: afero.NewOsFs()}, nil
	}

	// Otherwise expect a mapping with "types" key
	var m struct {
		Types rules.StringList `yaml:"types"`
	}
	if err := node.Decode(&m); err != nil {
		return nil, err
	}

	if len(m.Types) == 0 {
		return nil, fmt.Errorf("file_type filter requires at least one type")
	}

	for _, t := range m.Types {
		if err := validateFileType(t); err != nil {
			return nil, err
		}
	}

	return &FileType{Types: m.Types, Fs: afero.NewOsFs()}, nil
}

// validateFileType checks if the given type is valid.
func validateFileType(t string) error {
	normalized := normalizeFileType(t)
	switch normalized {
	case "file", "directory", "symlink":
		return nil
	default:
		return fmt.Errorf("invalid file type %q: must be one of file, directory, symlink (or aliases: dir, folder, link)", t)
	}
}
