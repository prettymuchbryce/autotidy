//go:build integration

package testutil

import "time"

// Time helpers for readable test data

// HoursAgo returns a negative duration representing n hours in the past.
// Use with ModifiedAt: File("x").ModifiedAt(HoursAgo(2))
func HoursAgo(n float64) time.Duration {
	return -time.Duration(n * float64(time.Hour))
}

// DaysAgo returns a negative duration representing n days in the past.
// Use with ModifiedAt: File("x").ModifiedAt(DaysAgo(7))
func DaysAgo(n float64) time.Duration {
	return -time.Duration(n * 24 * float64(time.Hour))
}

// MinutesAgo returns a negative duration representing n minutes in the past.
func MinutesAgo(n float64) time.Duration {
	return -time.Duration(n * float64(time.Minute))
}

// File builder helpers for fluent API

// File creates a FileEntry for a file at the given path.
// Path should use forward slashes regardless of OS.
func File(path string) FileEntry {
	return FileEntry{Path: path, IsDir: false}
}

// Dir creates a FileEntry for a directory at the given path.
// Path should use forward slashes regardless of OS.
func Dir(path string) FileEntry {
	return FileEntry{Path: path, IsDir: true}
}

// WithContent sets the file content.
func (f FileEntry) WithContent(content string) FileEntry {
	f.Content = content
	return f
}

// WithSize sets the file size (creates file filled with zero bytes).
func (f FileEntry) WithSize(size int64) FileEntry {
	f.Size = size
	return f
}

// ModifiedAt sets the modification time relative to now.
// Use HoursAgo, DaysAgo, or MinutesAgo helpers.
func (f FileEntry) ModifiedAt(d time.Duration) FileEntry {
	f.ModTime = d
	return f
}

// AccessedAt sets the access time relative to now.
// Use HoursAgo, DaysAgo, or MinutesAgo helpers.
func (f FileEntry) AccessedAt(d time.Duration) FileEntry {
	f.AccessTime = d
	return f
}
