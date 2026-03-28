package legacy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectWithoutLegacyAssets(t *testing.T) {
	home := t.TempDir()
	detection := Detect(home)

	if detection.NeedsMigration {
		t.Fatal("NeedsMigration = true, want false")
	}
	for _, candidate := range detection.Candidates {
		if candidate.Exists {
			t.Fatalf("candidate %q unexpectedly exists", candidate.Path)
		}
	}
}

func TestDetectWithLegacyAssets(t *testing.T) {
	home := t.TempDir()
	legacyRoot := filepath.Join(home, ".ccc")
	legacyLauncher := filepath.Join(home, "bin", "ccc")

	if err := os.MkdirAll(legacyRoot, 0o755); err != nil {
		t.Fatalf("MkdirAll legacy root: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(legacyLauncher), 0o755); err != nil {
		t.Fatalf("MkdirAll legacy launcher dir: %v", err)
	}
	if err := os.WriteFile(legacyLauncher, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("WriteFile legacy launcher: %v", err)
	}

	detection := Detect(home)
	if !detection.NeedsMigration {
		t.Fatal("NeedsMigration = false, want true")
	}

	found := 0
	for _, candidate := range detection.Candidates {
		if candidate.Exists {
			found++
		}
	}
	if found < 2 {
		t.Fatalf("expected at least 2 legacy candidates, found %d", found)
	}
}
