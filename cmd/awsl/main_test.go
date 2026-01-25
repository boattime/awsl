package main

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// update is a flag to update golden files with current output.
// Run with: make test-update
var update = flag.Bool("update", false, "update golden files")

func TestRun_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := run([]string{"awsl"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(stderr.String(), "usage:") {
		t.Errorf("expected usage message in stderr, got %q", stderr.String())
	}
}

func TestRun_Version(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := run([]string{"awsl", "--version"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout.String(), "awsl version") {
		t.Errorf("expected version output, got %q", stdout.String())
	}
}

func TestRun_FileNotFound(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := run([]string{"awsl", "nonexistent.awsl"}, &stdout, &stderr)

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}

	if !strings.Contains(stderr.String(), "error reading file") {
		t.Errorf("expected file error in stderr, got %q", stderr.String())
	}
}

func TestRun_GoldenFiles(t *testing.T) {
	testFiles, err := filepath.Glob("../../testdata/*.awsl")
	if err != nil {
		t.Fatalf("failed to find test files: %v", err)
	}

	if len(testFiles) == 0 {
		t.Fatal("no test files found in testdata/")
	}

	for _, testFile := range testFiles {
		// Extract test name from filename
		name := strings.TrimSuffix(filepath.Base(testFile), ".awsl")

		t.Run(name, func(t *testing.T) {
			goldenFile := strings.TrimSuffix(testFile, ".awsl") + ".golden"

			var stdout, stderr bytes.Buffer
			exitCode := run([]string{"awsl", testFile}, &stdout, &stderr)

			// Combine stdout and stderr for comparison
			// Format: exit code on first line, then output
			var actual bytes.Buffer
			actual.WriteString(stdout.String())
			if stderr.Len() > 0 {
				actual.WriteString("--- stderr ---\n")
				actual.WriteString(stderr.String())
			}
			actual.WriteString("--- exit code: ")
			actual.WriteString(itoa(exitCode))
			actual.WriteString(" ---\n")

			if *update {
				err := os.WriteFile(goldenFile, actual.Bytes(), 0644)
				if err != nil {
					t.Fatalf("failed to update golden file: %v", err)
				}
				t.Logf("updated %s", goldenFile)
				return
			}

			expected, err := os.ReadFile(goldenFile)
			if err != nil {
				t.Fatalf("failed to read golden file (run with -update to create): %v", err)
			}

			if !bytes.Equal(actual.Bytes(), expected) {
				t.Errorf("output mismatch for %s\n\nexpected:\n%s\n\nactual:\n%s",
					testFile, string(expected), actual.String())
			}
		})
	}
}

// itoa converts an int to a string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	if n < 0 {
		return "-" + itoa(-n)
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
