package utils

import (
	"testing"

	"github.com/prettymuchbryce/autotidy/internal/testutil"
)

func TestTemplate_ExpandWithNameExt(t *testing.T) {
	tests := []struct {
		name     string
		template Template
		srcPath  string
		expected string
	}{
		{
			name:     "name variable only",
			template: Template("${name}"),
			srcPath:  testutil.Path("/", "dir", "report.pdf"),
			expected: "report",
		},
		{
			name:     "ext variable only",
			template: Template("${ext}"),
			srcPath:  testutil.Path("/", "dir", "report.pdf"),
			expected: ".pdf",
		},
		{
			name:     "name and ext combined",
			template: Template("${name}_backup${ext}"),
			srcPath:  testutil.Path("/", "dir", "document.docx"),
			expected: "document_backup.docx",
		},
		{
			name:     "multiple dots in filename",
			template: Template("${name}${ext}"),
			srcPath:  testutil.Path("/", "dir", "report.final.pdf"),
			expected: "report.final.pdf",
		},
		{
			name:     "no extension",
			template: Template("${name}${ext}"),
			srcPath:  testutil.Path("/", "dir", "README"),
			expected: "README",
		},
		{
			name:     "hidden file",
			template: Template("${name}${ext}"),
			srcPath:  testutil.Path("/", "dir", ".gitignore"),
			expected: ".gitignore",
		},
		{
			name:     "hidden file with extension",
			template: Template("${name}${ext}"),
			srcPath:  testutil.Path("/", "dir", ".config.json"),
			expected: ".config.json",
		},
		{
			name:     "no variables",
			template: Template("plain-filename.txt"),
			srcPath:  testutil.Path("/", "dir", "ignored.txt"),
			expected: "plain-filename.txt",
		},
		{
			name:     "unknown variable left unchanged",
			template: Template("${unknown}"),
			srcPath:  testutil.Path("/", "dir", "file.txt"),
			expected: "${unknown}",
		},
		{
			name:     "multiple same variables",
			template: Template("${name}-${name}${ext}"),
			srcPath:  testutil.Path("/", "dir", "doc.txt"),
			expected: "doc-doc.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.template.ExpandWithNameExt(tt.srcPath)

			if result.String() != tt.expected {
				t.Errorf("result = %q, want %q", result.String(), tt.expected)
			}
		})
	}
}

func TestTemplate_ExpandWithTime(t *testing.T) {
	// We can't test exact values since time changes, but we can verify format
	template := Template("%Y-%m-%d")
	result := template.ExpandWithTime()

	// Result should be in YYYY-MM-DD format (10 chars)
	if len(result.String()) != 10 {
		t.Errorf("expected 10 char date, got %q", result.String())
	}

	// Should have dashes in right places
	if result.String()[4] != '-' || result.String()[7] != '-' {
		t.Errorf("expected YYYY-MM-DD format, got %q", result.String())
	}
}

func TestTemplate_Chained(t *testing.T) {
	// Test chaining ExpandWithNameExt and ExpandWithTime
	template := Template("${name}_copy${ext}")
	result := template.ExpandWithNameExt(testutil.Path("/", "dir", "report.pdf")).ExpandWithTime()

	if result.String() != "report_copy.pdf" {
		t.Errorf("result = %q, want %q", result.String(), "report_copy.pdf")
	}
}

func TestTemplate_String(t *testing.T) {
	tmpl := Template("${name}${ext}")
	if tmpl.String() != "${name}${ext}" {
		t.Errorf("String() = %q, want %q", tmpl.String(), "${name}${ext}")
	}
}

func TestReplaceVariables(t *testing.T) {
	tests := []struct {
		name     string
		template Template
		vars     map[string]string
		expected string
	}{
		{
			name:     "single variable",
			template: Template("Hello ${name}!"),
			vars:     map[string]string{"name": "World"},
			expected: "Hello World!",
		},
		{
			name:     "multiple variables",
			template: Template("${first} ${last}"),
			vars:     map[string]string{"first": "John", "last": "Doe"},
			expected: "John Doe",
		},
		{
			name:     "missing variable unchanged",
			template: Template("${known} ${unknown}"),
			vars:     map[string]string{"known": "value"},
			expected: "value ${unknown}",
		},
		{
			name:     "empty vars",
			template: Template("${foo}"),
			vars:     map[string]string{},
			expected: "${foo}",
		},
		{
			name:     "no variables in template",
			template: Template("plain text"),
			vars:     map[string]string{"foo": "bar"},
			expected: "plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replaceVariables(tt.template, tt.vars)

			if result.String() != tt.expected {
				t.Errorf("result = %q, want %q", result.String(), tt.expected)
			}
		})
	}
}
