package utils

import (
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/prettymuchbryce/autotidy/internal/pathutil"

	"github.com/itchyny/timefmt-go"
)

// variablePattern matches ${var} patterns.
var variablePattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// Template is a string that supports template expansion.
// It can contain ${name}, ${ext} variables and strftime tokens like %Y, %m, %d.
// Unlike TemplatePath, it does not perform tilde expansion.
type Template string

func (t Template) ExpandTilde() Template {
	return Template(pathutil.ExpandTilde(string(t)))
}

func (t Template) ExpandWithNameExt(path string) Template {
	pathExpanded := pathutil.ExpandTilde(path)
	base := filepath.Base(pathExpanded)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return replaceVariables(t, map[string]string{
		"name": name,
		"ext":  ext,
	})
}

func (t Template) ExpandWithTime() Template {
	now := time.Now()
	return Template(timefmt.Format(now, string(t)))
}

func (t Template) String() string {
	return string(t)
}

func replaceVariables(template Template, vars map[string]string) Template {
	result := variablePattern.ReplaceAllStringFunc(string(template), func(match string) string {
		// Extract variable name from ${name}
		varName := match[2 : len(match)-1]
		if val, ok := vars[varName]; ok {
			return val
		}
		return match // leave unchanged if not found
	})
	return Template(result)
}
