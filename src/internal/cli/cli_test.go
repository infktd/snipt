package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI_RoundTrip(t *testing.T) {
	// Use temp DB
	dbFile := filepath.Join(t.TempDir(), "test.db")

	// Helper to run CLI commands
	run := func(args ...string) (string, error) {
		root := NewRootCmd("test")
		buf := new(bytes.Buffer)
		root.SetOut(buf)
		root.SetErr(buf)
		root.SetArgs(append([]string{"--db", dbFile}, args...))
		err := root.Execute()
		return buf.String(), err
	}

	// Add a snippet via file
	tmpFile := filepath.Join(t.TempDir(), "test.go")
	os.WriteFile(tmpFile, []byte("package main\n\nfunc main() {}"), 0o644)

	out, err := run("add", tmpFile, "--title", "test snippet", "--tags", "test,go")
	if err != nil {
		t.Fatalf("add failed: %v\noutput: %s", err, out)
	}
	if !strings.Contains(out, "saved") {
		t.Errorf("expected 'saved' in output, got: %s", out)
	}

	// List
	out, err = run("list")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(out, "test snippet") {
		t.Errorf("expected 'test snippet' in list output, got: %s", out)
	}

	// Stats
	out, err = run("stats")
	if err != nil {
		t.Fatalf("stats failed: %v", err)
	}
	if !strings.Contains(out, "Total snippets: 1") {
		t.Errorf("expected 1 snippet in stats, got: %s", out)
	}
}
