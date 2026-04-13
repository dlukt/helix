package main

import (
	"bytes"
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

func TestValidateManifest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		manifest manifest
		wantErr  string
	}{
		{
			name: "valid",
			manifest: manifest{
				Events: []eventDef{{
					SubscriptionType: "channel.follow",
					Version:          "2",
					GoType:           "ChannelFollowEvent",
				}},
			},
		},
		{
			name: "missing subscription type",
			manifest: manifest{
				Events: []eventDef{{Version: "2", GoType: "ChannelFollowEvent"}},
			},
			wantErr: "subscription_type is required",
		},
		{
			name: "missing version",
			manifest: manifest{
				Events: []eventDef{{SubscriptionType: "channel.follow", GoType: "ChannelFollowEvent"}},
			},
			wantErr: "version is required",
		},
		{
			name: "missing go type",
			manifest: manifest{
				Events: []eventDef{{SubscriptionType: "channel.follow", Version: "2"}},
			},
			wantErr: "go_type is required",
		},
		{
			name: "duplicate type version",
			manifest: manifest{
				Events: []eventDef{
					{SubscriptionType: "channel.follow", Version: "2", GoType: "ChannelFollowEvent"},
					{SubscriptionType: "channel.follow", Version: "2", GoType: "AnotherType"},
				},
			},
			wantErr: `duplicate event "channel.follow@2"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateManifest(tt.manifest)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("validateManifest() error = %v", err)
			}
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("validateManifest() error = nil, want %q", tt.wantErr)
				}
				if got := err.Error(); got == "" || !bytes.Contains([]byte(got), []byte(tt.wantErr)) {
					t.Fatalf("validateManifest() error = %q, want substring %q", got, tt.wantErr)
				}
			}
		})
	}
}

func TestRenderRegistryMatchesCheckedInGeneratedFile(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}

	root, err := findRepoRoot(wd)
	if err != nil {
		t.Fatalf("findRepoRoot() error = %v", err)
	}

	data, err := loadManifest(filepath.Join(root, "internal", "manifest", "eventsub.json"))
	if err != nil {
		t.Fatalf("loadManifest() error = %v", err)
	}

	got, err := renderRegistry(data)
	if err != nil {
		t.Fatalf("renderRegistry() error = %v", err)
	}

	want, err := os.ReadFile(filepath.Join(root, "eventsub", "generated_registry.go"))
	if err != nil {
		t.Fatalf("ReadFile(generated_registry.go) error = %v", err)
	}

	if !bytes.Equal(got, want) {
		t.Fatal("rendered registry does not match checked-in generated file")
	}
}
