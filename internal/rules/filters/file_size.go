package filters

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/prettymuchbryce/autotidy/internal/rules"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

func init() {
	rules.RegisterFilter("file_size", deserializeSize)
}

const (
	bytesPerKB = 1024
	bytesPerMB = 1024 * 1024
	bytesPerGB = 1024 * 1024 * 1024
	bytesPerTB = 1024 * 1024 * 1024 * 1024
)

// SizeSpec represents a size value with a unit.
// Exactly one field should be set.
type SizeSpec struct {
	B  *float64 `yaml:"b"`
	KB *float64 `yaml:"kb"`
	MB *float64 `yaml:"mb"`
	GB *float64 `yaml:"gb"`
	TB *float64 `yaml:"tb"`
}

// ToBytes converts the SizeSpec to bytes.
func (s *SizeSpec) ToBytes() (int64, error) {
	switch {
	case s.B != nil:
		return int64(*s.B), nil
	case s.KB != nil:
		return int64(*s.KB * bytesPerKB), nil
	case s.MB != nil:
		return int64(*s.MB * bytesPerMB), nil
	case s.GB != nil:
		return int64(*s.GB * bytesPerGB), nil
	case s.TB != nil:
		return int64(*s.TB * bytesPerTB), nil
	default:
		return 0, fmt.Errorf("no size unit specified")
	}
}

// Validate checks that exactly one size unit is set.
func (s *SizeSpec) Validate() error {
	count := 0
	if s.B != nil {
		count++
	}
	if s.KB != nil {
		count++
	}
	if s.MB != nil {
		count++
	}
	if s.GB != nil {
		count++
	}
	if s.TB != nil {
		count++
	}

	if count == 0 {
		return fmt.Errorf("no size unit specified")
	}
	if count > 1 {
		return fmt.Errorf("only one size unit allowed")
	}

	return nil
}

// SizeBetween represents a size range.
type SizeBetween struct {
	Min SizeSpec `yaml:"min"`
	Max SizeSpec `yaml:"max"`
}

// Size is a filter that matches files by size.
type Size struct {
	GreaterThan *SizeSpec    `yaml:"greater_than"`
	LessThan    *SizeSpec    `yaml:"less_than"`
	AtLeast     *SizeSpec    `yaml:"at_least"`
	AtMost      *SizeSpec    `yaml:"at_most"`
	Between     *SizeBetween `yaml:"between"`
	Fs          afero.Fs
}

// Evaluate checks if the file size matches the criteria.
// Returns false for directories since they don't have a meaningful file size.
func (s *Size) Evaluate(path string) (bool, error) {
	fs := s.Fs
	if fs == nil {
		fs = afero.NewOsFs()
	}

	info, err := fs.Stat(path)
	if err != nil {
		return false, err
	}

	if info.IsDir() {
		return false, nil
	}

	fileSize := info.Size()

	if s.GreaterThan != nil {
		threshold, err := s.GreaterThan.ToBytes()
		if err != nil {
			return false, err
		}
		return fileSize > threshold, nil
	}

	if s.LessThan != nil {
		threshold, err := s.LessThan.ToBytes()
		if err != nil {
			return false, err
		}
		return fileSize < threshold, nil
	}

	if s.AtLeast != nil {
		threshold, err := s.AtLeast.ToBytes()
		if err != nil {
			return false, err
		}
		return fileSize >= threshold, nil
	}

	if s.AtMost != nil {
		threshold, err := s.AtMost.ToBytes()
		if err != nil {
			return false, err
		}
		return fileSize <= threshold, nil
	}

	if s.Between != nil {
		min, err := s.Between.Min.ToBytes()
		if err != nil {
			return false, err
		}
		max, err := s.Between.Max.ToBytes()
		if err != nil {
			return false, err
		}
		return fileSize >= min && fileSize <= max, nil
	}

	return false, fmt.Errorf("no size comparison specified")
}

// shorthand pattern: "> 10mb", ">=2.5gb", "< 500 kb", etc.
var sizeShorthandPattern = regexp.MustCompile(`(?i)^\s*(>=?|<=?)\s*(\d+(?:\.\d+)?)\s*(b|kb|mb|gb|tb)\s*$`)

// parseShorthand parses a size shorthand string like "> 10mb".
func parseShorthand(s string) (*Size, error) {
	matches := sizeShorthandPattern.FindStringSubmatch(s)
	if matches == nil {
		return nil, fmt.Errorf("invalid size shorthand %q: expected format like \"> 10mb\"", s)
	}

	op := matches[1]
	value, err := strconv.ParseFloat(matches[2], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid size value %q: %w", matches[2], err)
	}
	unit := strings.ToLower(matches[3])

	spec := &SizeSpec{}
	switch unit {
	case "b":
		spec.B = &value
	case "kb":
		spec.KB = &value
	case "mb":
		spec.MB = &value
	case "gb":
		spec.GB = &value
	case "tb":
		spec.TB = &value
	}

	size := &Size{Fs: afero.NewOsFs()}
	switch op {
	case ">":
		size.GreaterThan = spec
	case ">=":
		size.AtLeast = spec
	case "<":
		size.LessThan = spec
	case "<=":
		size.AtMost = spec
	}

	return size, nil
}

// deserializeSize creates a Size filter from YAML.
// Supports:
//   - Shorthand: "size: \"> 10mb\""
//   - Explicit: "size: {greater_than: {mb: 10}}"
func deserializeSize(node yaml.Node) (rules.Evaluable, error) {
	// Try shorthand string first
	if node.Kind == yaml.ScalarNode {
		var shorthand string
		if err := node.Decode(&shorthand); err != nil {
			return nil, err
		}
		return parseShorthand(shorthand)
	}

	// Otherwise expect a mapping
	var s Size
	if err := node.Decode(&s); err != nil {
		return nil, err
	}
	s.Fs = afero.NewOsFs()

	// Validate that exactly one comparison is set
	count := 0
	if s.GreaterThan != nil {
		if err := s.GreaterThan.Validate(); err != nil {
			return nil, fmt.Errorf("greater_than: %w", err)
		}
		count++
	}
	if s.LessThan != nil {
		if err := s.LessThan.Validate(); err != nil {
			return nil, fmt.Errorf("less_than: %w", err)
		}
		count++
	}
	if s.AtLeast != nil {
		if err := s.AtLeast.Validate(); err != nil {
			return nil, fmt.Errorf("at_least: %w", err)
		}
		count++
	}
	if s.AtMost != nil {
		if err := s.AtMost.Validate(); err != nil {
			return nil, fmt.Errorf("at_most: %w", err)
		}
		count++
	}
	if s.Between != nil {
		if err := s.Between.Min.Validate(); err != nil {
			return nil, fmt.Errorf("between.min: %w", err)
		}
		if err := s.Between.Max.Validate(); err != nil {
			return nil, fmt.Errorf("between.max: %w", err)
		}
		count++
	}

	if count == 0 {
		return nil, fmt.Errorf("size filter requires a comparison (greater_than, less_than, at_least, at_most, or between)")
	}
	if count > 1 {
		return nil, fmt.Errorf("size filter can only have one comparison")
	}

	return &s, nil
}
