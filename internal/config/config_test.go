package config

import (
	"bytes"
	"testing"
	"text/template"
	"time"

	"github.com/prettymuchbryce/autotidy/internal/testutil"

	"github.com/spf13/afero"
)

// renderYAML renders a YAML template with the given data.
func renderYAML(t *testing.T, tmpl string, data any) string {
	t.Helper()
	var buf bytes.Buffer
	template.Must(template.New("yaml").Parse(tmpl)).Execute(&buf, data)
	return buf.String()
}

func TestLoadWithFs_ValidConfig(t *testing.T) {
	configPath := testutil.Path("/", "config.yaml")
	downloadsPath := testutil.Path("/", "home", "user", "downloads")

	fs := afero.NewMemMapFs()
	// Note: Actions not included here since action deserializers are tested separately
	configYAML := renderYAML(t, `
rules:
  - name: test-rule
    locations:
      - {{.DownloadsPath}}
daemon:
  debounce: 1s
logging:
  level: debug
`, map[string]string{"DownloadsPath": downloadsPath})
	afero.WriteFile(fs, configPath, []byte(configYAML), 0644)

	cfg, err := LoadWithFs(configPath, fs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(cfg.Rules))
	}

	if cfg.Rules[0].Name != "test-rule" {
		t.Errorf("expected rule name 'test-rule', got %q", cfg.Rules[0].Name)
	}

	if cfg.Daemon.Debounce != time.Second {
		t.Errorf("expected debounce 1s, got %v", cfg.Daemon.Debounce)
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("expected logging level 'debug', got %q", cfg.Logging.Level)
	}
}

func TestLoadWithFs_DefaultValues(t *testing.T) {
	configPath := testutil.Path("/", "config.yaml")
	tmpPath := testutil.Path("/", "tmp")

	fs := afero.NewMemMapFs()
	// Minimal config - should use defaults for daemon and logging
	configYAML := renderYAML(t, `
rules:
  - name: minimal-rule
    locations: {{.TmpPath}}
`, map[string]string{"TmpPath": tmpPath})
	afero.WriteFile(fs, configPath, []byte(configYAML), 0644)

	cfg, err := LoadWithFs(configPath, fs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check daemon defaults
	if cfg.Daemon.Debounce != 500*time.Millisecond {
		t.Errorf("expected default debounce 500ms, got %v", cfg.Daemon.Debounce)
	}

	// Check logging defaults
	if cfg.Logging.Level != "warn" {
		t.Errorf("expected default logging level 'warn', got %q", cfg.Logging.Level)
	}
}

func TestLoadWithFs_FileNotFound(t *testing.T) {
	fs := afero.NewMemMapFs()

	_, err := LoadWithFs(testutil.Path("/", "nonexistent.yaml"), fs)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLoadWithFs_InvalidYAML(t *testing.T) {
	configPath := testutil.Path("/", "config.yaml")

	fs := afero.NewMemMapFs()
	invalidYAML := `
rules:
  - name: test
    locations: [unclosed
`
	afero.WriteFile(fs, configPath, []byte(invalidYAML), 0644)

	_, err := LoadWithFs(configPath, fs)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

// Note: Fs is no longer set by LoadWithFs - it's set by cmd/run.go
// based on whether --dry-run is specified.

func TestLoadWithFs_EmptyConfig(t *testing.T) {
	configPath := testutil.Path("/", "config.yaml")

	fs := afero.NewMemMapFs()
	afero.WriteFile(fs, configPath, []byte(""), 0644)

	cfg, err := LoadWithFs(configPath, fs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have no rules but still have defaults
	if len(cfg.Rules) != 0 {
		t.Errorf("expected 0 rules, got %d", len(cfg.Rules))
	}

	// Defaults should still be applied
	if cfg.Daemon.Debounce != 500*time.Millisecond {
		t.Errorf("expected default debounce, got %v", cfg.Daemon.Debounce)
	}
}

func TestLoadWithFs_MultipleRules(t *testing.T) {
	configPath := testutil.Path("/", "config.yaml")
	tmpA := testutil.Path("/", "tmp", "a")
	tmpB1 := testutil.Path("/", "tmp", "b1")
	tmpB2 := testutil.Path("/", "tmp", "b2")
	tmpC := testutil.Path("/", "tmp", "c")

	fs := afero.NewMemMapFs()
	configYAML := renderYAML(t, `
rules:
  - name: rule-a
    locations: {{.TmpA}}
  - name: rule-b
    locations:
      - {{.TmpB1}}
      - {{.TmpB2}}
  - name: rule-c
    locations: {{.TmpC}}
    recursive: true
`, map[string]string{
		"TmpA":  tmpA,
		"TmpB1": tmpB1,
		"TmpB2": tmpB2,
		"TmpC":  tmpC,
	})
	afero.WriteFile(fs, configPath, []byte(configYAML), 0644)

	cfg, err := LoadWithFs(configPath, fs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(cfg.Rules))
	}

	expectedNames := []string{"rule-a", "rule-b", "rule-c"}
	for i, name := range expectedNames {
		if cfg.Rules[i].Name != name {
			t.Errorf("rule %d: expected name %q, got %q", i, name, cfg.Rules[i].Name)
		}
	}

	// Check that rule-b has 2 locations
	if len(cfg.Rules[1].Locations) != 2 {
		t.Errorf("expected rule-b to have 2 locations, got %d", len(cfg.Rules[1].Locations))
	}

	// Check that rule-c is recursive
	if !cfg.Rules[2].IsRecursive() {
		t.Error("expected rule-c to be recursive")
	}
}

func TestDefaultDaemonConfig(t *testing.T) {
	cfg := DefaultDaemonConfig()
	if cfg.Debounce != 500*time.Millisecond {
		t.Errorf("expected debounce 500ms, got %v", cfg.Debounce)
	}
}

func TestDefaultLoggingConfig(t *testing.T) {
	cfg := DefaultLoggingConfig()
	if cfg.Level != "warn" {
		t.Errorf("expected level 'warn', got %q", cfg.Level)
	}
}
