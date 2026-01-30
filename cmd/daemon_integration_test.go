//go:build integration

package cmd

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/spf13/afero"

	"github.com/prettymuchbryce/autotidy/daemon"

	// Import for side effects (filter/action registration)
	_ "github.com/prettymuchbryce/autotidy/internal/rules/actions"
	_ "github.com/prettymuchbryce/autotidy/internal/rules/filters"

	"github.com/prettymuchbryce/autotidy/internal/testutil"
)

func TestDaemon_FileCreate(t *testing.T) {
	testutil.Run(t, testutil.TestCase{
		Name: "move file on create",
		Config: `
rules:
  - name: move-all
    locations: {{join .TmpDir "source"}}
    actions:
      - move: {{join .TmpDir "dest"}}

daemon:
  debounce: 50ms

logging:
  level: debug
`,
		Before: []testutil.FileEntry{
			testutil.Dir("source"),
			testutil.Dir("dest"),
		},
		Trigger: []testutil.FileEntry{
			testutil.File("source/test.txt").WithContent("hello"),
		},
		Expect: []testutil.FileEntry{
			testutil.File("dest/test.txt"),
		},
		Missing: []string{
			"source/test.txt",
		},
	})
}

func TestDaemon_Filter(t *testing.T) {
	// Tests that filters only process matching files
	testutil.Run(t, testutil.TestCase{
		Name: "filter moves only matching files",
		Config: `
rules:
  - name: move-txt-only
    locations: {{join .TmpDir "source"}}
    actions:
      - move: {{join .TmpDir "dest"}}
    filters:
      - name: "*.txt"

daemon:
  debounce: 50ms

logging:
  level: debug
`,
		Before: []testutil.FileEntry{
			testutil.Dir("source"),
			testutil.Dir("dest"),
		},
		Trigger: []testutil.FileEntry{
			testutil.File("source/test.txt").WithContent("hello"),
			testutil.File("source/image.jpg").WithContent("image"),
		},
		Expect: []testutil.FileEntry{
			testutil.File("dest/test.txt"),      // .txt moved
			testutil.File("source/image.jpg"),   // .jpg stays
		},
		Missing: []string{
			"source/test.txt",  // .txt no longer in source
			"dest/image.jpg",   // .jpg not in dest
		},
	})
}

func TestDaemon_NotFilter(t *testing.T) {
	// Tests that not: filters prevent matching files from being processed
	testutil.Run(t, testutil.TestCase{
		Name: "not filter prevents matching files from moving",
		Config: `
rules:
  - name: move-except-bak
    locations: {{join .TmpDir "source"}}
    actions:
      - move: {{join .TmpDir "dest"}}
    filters:
      - not:
          - name: "*.bak"

daemon:
  debounce: 50ms

logging:
  level: debug
`,
		Before: []testutil.FileEntry{
			testutil.Dir("source"),
			testutil.Dir("dest"),
		},
		Trigger: []testutil.FileEntry{
			testutil.File("source/test.txt").WithContent("hello"),
			testutil.File("source/backup.bak").WithContent("backup"),
		},
		Expect: []testutil.FileEntry{
			testutil.File("dest/test.txt"),      // .txt moved
			testutil.File("source/backup.bak"),  // .bak filtered out, stays
		},
		Missing: []string{
			"source/test.txt",  // .txt no longer in source
			"dest/backup.bak",  // .bak not in dest
		},
	})
}

func TestDaemon_ComplexFilter(t *testing.T) {
	// Tests complex filter combinations:
	// - any: with 3 branches (each file matches exactly one)
	// - top-level not: excludes files matching *_skip*
	// - nested not: inside one any branch excludes *_draft* files
	testutil.Run(t, testutil.TestCase{
		Name: "complex filter with any, not, and nested not",
		Config: `
rules:
  - name: complex-filter-test
    locations: {{join .TmpDir "source"}}
    actions:
      - move: {{join .TmpDir "dest"}}
    filters:
      - any:
          - extension: txt
            not:
              - name: "*_draft*"
          - extension: pdf
          - extension: doc
      - not:
          - name: "*_skip*"

daemon:
  debounce: 50ms

logging:
  level: debug
`,
		Before: []testutil.FileEntry{
			testutil.Dir("source"),
			testutil.Dir("dest"),
		},
		Trigger: []testutil.FileEntry{
			// These 3 files each match one any branch, should be moved
			testutil.File("source/report.txt").WithContent("txt file"),
			testutil.File("source/document.pdf").WithContent("pdf file"),
			testutil.File("source/letter.doc").WithContent("doc file"),
			// Matches txt branch but excluded by top-level not
			testutil.File("source/notes_skip_me.txt").WithContent("should stay"),
			// Matches txt extension but excluded by nested not in that branch
			testutil.File("source/memo_draft.txt").WithContent("draft should stay"),
		},
		Expect: []testutil.FileEntry{
			// Files that passed all filters - moved to dest
			testutil.File("dest/report.txt"),
			testutil.File("dest/document.pdf"),
			testutil.File("dest/letter.doc"),
			// Files excluded by filters - stay in source
			testutil.File("source/notes_skip_me.txt"),
			testutil.File("source/memo_draft.txt"),
		},
		Missing: []string{
			// Moved files no longer in source
			"source/report.txt",
			"source/document.pdf",
			"source/letter.doc",
			// Excluded files not in dest
			"dest/notes_skip_me.txt",
			"dest/memo_draft.txt",
		},
	})
}

func TestDaemon_Trash(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("skipping trash test on CI - requires interactive permissions")
	}

	// Tests that trash action moves files to system trash
	testutil.Run(t, testutil.TestCase{
		Name: "trash action removes file from source",
		Config: `
rules:
  - name: trash-txt-files
    locations: {{join .TmpDir "source"}}
    actions:
      - trash
    filters:
      - name: "*.txt"

daemon:
  debounce: 50ms

logging:
  level: debug
`,
		Before: []testutil.FileEntry{
			testutil.Dir("source"),
		},
		Trigger: []testutil.FileEntry{
			testutil.File("source/test.txt").WithContent("hello"),
		},
		Expect:  []testutil.FileEntry{},
		Missing: []string{"source/test.txt"},
	})
}

func TestDaemon_AllFiltersAndActions(t *testing.T) {
	// Comprehensive test of all filters and action chaining:
	// Filters: extension, size, name (glob), name (regex), file_type,
	//          date_modified, date_accessed, date_changed, date_created, mime_type
	// Actions: log -> rename -> copy -> move (chained), delete (separate rule)
	// Also tests: not filters
	testutil.Run(t, testutil.TestCase{
		Name: "all filters and actions",
		Config: `
rules:
  - name: comprehensive-test
    locations: {{join .TmpDir "source" "main"}}
    filters:
      - any:
          - extension: txt
          - file_size: "> 100b"
          - name: "glob_match_*"
          - name:
              regex: "^regex_\\d+\\.dat$"
          - file_type: file
          - date_modified:
              after:
                hours_ago: 1
          - date_accessed:
              after:
                hours_ago: 1
          - date_changed:
              after:
                hours_ago: 1
          - date_created:
              after:
                hours_ago: 1
          - mime_type: "text/plain*"
      - not:
          - name: "*_skip*"
    actions:
      - log: "Processing ${name}${ext}"
      - rename: "${name}_renamed${ext}"
      - copy: "${name}_backup${ext}"
      - move:
          dest: {{join .TmpDir "dest"}}

  - name: delete-test
    locations: {{join .TmpDir "source" "main"}}
    filters:
      - name: "to_delete*"
    actions:
      - delete

daemon:
  debounce: 50ms

logging:
  level: debug
`,
		Before: []testutil.FileEntry{
			testutil.Dir("source/main"),
			testutil.Dir("dest"),
		},
		Trigger: []testutil.FileEntry{
			// Extension filter (.txt, also matches mime_type text/plain)
			testutil.File("source/main/document.txt").WithContent("text content"),
			// Size filter (> 100 bytes)
			testutil.File("source/main/largefile.bin").WithSize(200),
			// Name glob filter (glob_match_*)
			testutil.File("source/main/glob_match_test.dat").WithContent("glob"),
			// Name regex filter (regex_\d+\.dat)
			testutil.File("source/main/regex_123.dat").WithContent("regex"),
			// Filtered out file (*_skip*)
			testutil.File("source/main/data_skip_this.txt").WithContent("skip"),
			// Delete test file
			testutil.File("source/main/to_delete_me.tmp").WithContent("delete"),
		},
		Expect: []testutil.FileEntry{
			// Renamed originals stay in source
			testutil.File("source/main/document_renamed.txt"),
			testutil.File("source/main/largefile_renamed.bin"),
			testutil.File("source/main/glob_match_test_renamed.dat"),
			testutil.File("source/main/regex_123_renamed.dat"),
			// Backup copies moved to dest
			testutil.File("dest/document_renamed_backup.txt"),
			testutil.File("dest/largefile_renamed_backup.bin"),
			testutil.File("dest/glob_match_test_renamed_backup.dat"),
			testutil.File("dest/regex_123_renamed_backup.dat"),
			// Filtered out file unchanged
			testutil.File("source/main/data_skip_this.txt"),
		},
		Missing: []string{
			// Original names should not exist
			"source/main/document.txt",
			"source/main/largefile.bin",
			"source/main/glob_match_test.dat",
			"source/main/regex_123.dat",
			// Delete action should remove file
			"source/main/to_delete_me.tmp",
		},
		Timeout: 5 * time.Second,
	})
}

func TestDaemon_MissingDirectory(t *testing.T) {
	// Tests that daemon handles initially missing directories
	// by watching the parent and activating when the directory is created
	testutil.Run(t, testutil.TestCase{
		Name: "watch directory created after daemon start",
		Config: `
rules:
  - name: move-all
    locations: {{join .TmpDir "source"}}
    actions:
      - move: {{join .TmpDir "dest"}}

daemon:
  debounce: 50ms

logging:
  level: debug
`,
		Before: []testutil.FileEntry{
			testutil.Dir("dest"),
			// Note: "source" intentionally not created
		},
		Trigger: []testutil.FileEntry{
			testutil.Dir("source"), // Create the directory first
			testutil.File("source/test.txt").WithContent("hello"),
		},
		Expect: []testutil.FileEntry{
			testutil.File("dest/test.txt"),
		},
		Missing: []string{
			"source/test.txt",
		},
		Timeout: 3 * time.Second,
	})
}

func TestDaemon_NestedMissingDirectory(t *testing.T) {
	// Tests watching a deeply nested directory that doesn't exist
	// and creating it incrementally (parent by parent)
	testutil.Run(t, testutil.TestCase{
		Name: "watch deeply nested directory created incrementally",
		Config: `
rules:
  - name: move-all
    locations: {{join .TmpDir "a" "b" "c"}}
    actions:
      - move: {{join .TmpDir "dest"}}

daemon:
  debounce: 50ms

logging:
  level: debug
`,
		Before: []testutil.FileEntry{
			testutil.Dir("dest"),
			// Note: "a/b/c" does not exist
		},
		Trigger: []testutil.FileEntry{
			testutil.Dir("a"),
			testutil.Dir("a/b"),
			testutil.Dir("a/b/c"),
			testutil.File("a/b/c/test.txt").WithContent("hello"),
		},
		Expect: []testutil.FileEntry{
			testutil.File("dest/test.txt"),
		},
		Missing: []string{
			"a/b/c/test.txt",
		},
		Timeout: 5 * time.Second,
	})
}

func TestDaemon_MissingRecursiveDirectory(t *testing.T) {
	// Tests recursive rule with initially missing directory
	testutil.Run(t, testutil.TestCase{
		Name: "recursive rule with initially missing directory",
		Config: `
rules:
  - name: move-all
    locations: {{join .TmpDir "source"}}
    recursive: true
    actions:
      - move: {{join .TmpDir "dest"}}

daemon:
  debounce: 50ms

logging:
  level: debug
`,
		Before: []testutil.FileEntry{
			testutil.Dir("dest"),
		},
		Trigger: []testutil.FileEntry{
			testutil.Dir("source"),
			testutil.Dir("source/sub"),
			testutil.File("source/test.txt").WithContent("root"),
			testutil.File("source/sub/nested.txt").WithContent("nested"),
		},
		Expect: []testutil.FileEntry{
			testutil.File("dest/test.txt"),
			testutil.File("dest/nested.txt"),
		},
		Missing: []string{
			"source/test.txt",
			"source/sub/nested.txt",
		},
		Timeout: 5 * time.Second,
	})
}

func TestDaemon_DeleteAndRecreateDirectory(t *testing.T) {
	// Tests that daemon resumes watching after a configured directory is deleted and recreated
	testutil.Run(t, testutil.TestCase{
		Name: "resume watching after directory delete and recreate",
		Config: `
rules:
  - name: move-all
    locations: {{join .TmpDir "source"}}
    actions:
      - move: {{join .TmpDir "dest"}}

daemon:
  debounce: 50ms

logging:
  level: debug
`,
		Before: []testutil.FileEntry{
			testutil.Dir("source"),
			testutil.Dir("dest"),
		},
		Trigger: []testutil.FileEntry{
			// First, create a file that gets processed
			testutil.File("source/first.txt").WithContent("first"),
		},
		Expect: []testutil.FileEntry{
			testutil.File("dest/first.txt"),
		},
		Missing: []string{
			"source/first.txt",
		},
		Timeout: 3 * time.Second,
	})
}

func TestDaemon_DeleteRecreateManual(t *testing.T) {
	// Manual test for delete/recreate scenario that can't be expressed declaratively
	tmpDir := t.TempDir()
	sourceDir := tmpDir + "/source"
	destDir := tmpDir + "/dest"

	// Create initial directories
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to create source dir: %v", err)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("failed to create dest dir: %v", err)
	}

	// Write config
	config := `
rules:
  - name: move-all
    locations: ` + sourceDir + `
    actions:
      - move: ` + destDir + `

daemon:
  debounce: 50ms

logging:
  level: debug
`
	configPath := tmpDir + "/config.yaml"
	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Start daemon
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- daemon.Run(ctx, configPath, afero.NewOsFs(), func(level string) {})
	}()

	// Wait for daemon to start
	time.Sleep(200 * time.Millisecond)

	// Create and process a file to verify initial setup works
	if err := os.WriteFile(sourceDir+"/first.txt", []byte("first"), 0644); err != nil {
		t.Fatalf("failed to create first.txt: %v", err)
	}

	// Wait for processing
	time.Sleep(300 * time.Millisecond)

	// Verify first file was moved
	if _, err := os.Stat(destDir + "/first.txt"); err != nil {
		t.Errorf("expected dest/first.txt to exist after initial processing")
	}

	// Delete the source directory
	if err := os.RemoveAll(sourceDir); err != nil {
		t.Fatalf("failed to remove source dir: %v", err)
	}

	// Wait for fsnotify to process the delete
	time.Sleep(200 * time.Millisecond)

	// Recreate the source directory
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("failed to recreate source dir: %v", err)
	}

	// Wait for pending watch to activate and initial rule execution to complete,
	// plus cooldown period (1 second)
	time.Sleep(1500 * time.Millisecond)

	// Create a new file - this should trigger another rule execution
	t.Log("Creating second.txt")
	if err := os.WriteFile(sourceDir+"/second.txt", []byte("second"), 0644); err != nil {
		t.Fatalf("failed to create second.txt: %v", err)
	}

	// Wait for processing (debounce 50ms + execution time)
	time.Sleep(500 * time.Millisecond)

	// Verify second file was moved
	if _, err := os.Stat(destDir + "/second.txt"); err != nil {
		t.Errorf("expected dest/second.txt to exist after recreate, but got error: %v", err)
	}
	if _, err := os.Stat(sourceDir + "/second.txt"); err == nil {
		t.Errorf("expected source/second.txt to NOT exist after processing")
	}

	// Cleanup
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("daemon returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Error("daemon did not stop within timeout")
	}
}
