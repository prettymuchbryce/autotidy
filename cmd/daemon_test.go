//go:build integration

package cmd

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"text/template"
	"time"

	"github.com/spf13/afero"
)

func TestDaemon_GracefulShutdown(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}

	tmpl := template.Must(template.New("config").Parse(`
rules:
  - name: no-op
    locations: {{.SourceDir}}
    actions: []

daemon:
  debounce: 50ms
`))
	var buf bytes.Buffer
	tmpl.Execute(&buf, map[string]string{
		"SourceDir": sourceDir,
	})

	fs := afero.NewOsFs()
	if err := afero.WriteFile(fs, configPath, buf.Bytes(), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- RunDaemon(ctx, configPath, fs)
	}()

	time.Sleep(50 * time.Millisecond)

	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("expected nil error on graceful shutdown, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("daemon did not shut down within timeout")
	}
}
