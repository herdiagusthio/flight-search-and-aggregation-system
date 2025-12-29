// Package testutil provides test helper functions for unit and integration tests.
package testutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// LoadTestJSON loads a JSON file from the testdata directory.
// The filename should be relative to the testdata directory.
func LoadTestJSON(t *testing.T, filename string) []byte {
	t.Helper()

	// Get the path to testdata relative to this file
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed to get current file path")
	}

	// Navigate to project root (testutil is in test/testutil)
	projectRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")
	testDataPath := filepath.Join(projectRoot, "test", "testdata", filename)

	data, err := os.ReadFile(testDataPath)
	if err != nil {
		t.Fatalf("Failed to load test file %s: %v", filename, err)
	}
	return data
}

// LoadMockJSON loads a JSON file from the docs/response-mock directory.
// This is a convenience function for loading provider mock responses.
func LoadMockJSON(t *testing.T, filename string) []byte {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed to get current file path")
	}

	projectRoot := filepath.Join(filepath.Dir(currentFile), "..", "..")
	mockPath := filepath.Join(projectRoot, "docs", "response-mock", filename)

	data, err := os.ReadFile(mockPath)
	if err != nil {
		t.Fatalf("Failed to load mock file %s: %v", filename, err)
	}
	return data
}

// MustParseTime parses a time string in RFC3339 format.
// It fails the test if parsing fails.
func MustParseTime(t *testing.T, dateStr string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		t.Fatalf("Failed to parse time %s: %v", dateStr, err)
	}
	return parsed
}

// MustParseDate parses a date string in YYYY-MM-DD format.
// It fails the test if parsing fails.
func MustParseDate(t *testing.T, dateStr string) time.Time {
	t.Helper()
	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		t.Fatalf("Failed to parse date %s: %v", dateStr, err)
	}
	return parsed
}

// Ptr returns a pointer to the given value.
// Useful for creating pointers to literals in tests.
func Ptr[T any](v T) *T {
	return &v
}

// FloatPtr returns a pointer to a float64.
// Convenience function for filter option tests.
func FloatPtr(f float64) *float64 {
	return &f
}

// IntPtr returns a pointer to an int.
// Convenience function for filter option tests.
func IntPtr(i int) *int {
	return &i
}

// StringSlice returns a slice of strings.
// Convenience function for airline filter tests.
func StringSlice(s ...string) []string {
	return s
}
