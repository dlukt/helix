package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRepoRoot(t *testing.T) {
	root := t.TempDir()

	mustMkdirAll(t, filepath.Join(root, "eventsub"))
	mustMkdirAll(t, filepath.Join(root, "oauth"))
	mustMkdirAll(t, filepath.Join(root, "internal", "cmd", "genmanifest"))
	mustMkdirAll(t, filepath.Join(root, "internal", "manifest"))
	mustWriteFile(t, filepath.Join(root, "internal", "manifest", "eventsub.json"), []byte("{}"))

	tests := map[string]string{
		"repo root":              root,
		"eventsub package":       filepath.Join(root, "eventsub"),
		"oauth package":          filepath.Join(root, "oauth"),
		"internal cmd directory": filepath.Join(root, "internal", "cmd"),
		"generator package":      filepath.Join(root, "internal", "cmd", "genmanifest"),
	}

	for name, start := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := findRepoRoot(start)
			if err != nil {
				t.Fatalf("findRepoRoot(%q) returned error: %v", start, err)
			}

			if got != root {
				t.Fatalf("findRepoRoot(%q) = %q, want %q", start, got, root)
			}
		})
	}
}

func TestFindRepoRootReturnsErrorWhenMarkerMissing(t *testing.T) {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "internal", "cmd", "genmanifest"))

	_, err := findRepoRoot(filepath.Join(root, "internal", "cmd", "genmanifest"))
	if err == nil {
		t.Fatal("findRepoRoot returned nil error, want failure when marker file is missing")
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string, contents []byte) {
	t.Helper()

	if err := os.WriteFile(path, contents, 0o644); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
}
