package rules

// StringList handles YAML fields that can be a single string or a list of strings.
type StringList []string

func (s *StringList) UnmarshalYAML(unmarshal func(any) error) error {
	// Try single string first
	var single string
	if err := unmarshal(&single); err == nil {
		*s = []string{single}
		return nil
	}

	// Try list of strings
	var list []string
	if err := unmarshal(&list); err != nil {
		return err
	}
	*s = list
	return nil
}
