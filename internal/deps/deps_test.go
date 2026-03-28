package deps

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionLessThan(t *testing.T) {
	tests := []struct {
		left  string
		right string
		want  bool
	}{
		{"v16.20.0", "18.0.0", true},
		{"18.0.0", "18.0.0", false},
		{"18.1.0", "18.0.0", false},
		{"npm 10.8.0", "10.9.0", true},
	}

	for _, tt := range tests {
		if got := VersionLessThan(tt.left, tt.right); got != tt.want {
			t.Fatalf("VersionLessThan(%q, %q) = %v, want %v", tt.left, tt.right, got, tt.want)
		}
	}
}

func TestEnsureCLIInstallsMissingCommand(t *testing.T) {
	originalPath := os.Getenv("PATH")
	t.Cleanup(func() {
		_ = os.Setenv("PATH", originalPath)
	})

	home := t.TempDir()
	binDir := filepath.Join(home, "bin")
	npmPrefix := filepath.Join(home, ".npm-global")
	npmBinDir := filepath.Join(npmPrefix, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("MkdirAll binDir: %v", err)
	}
	if err := os.MkdirAll(npmBinDir, 0o755); err != nil {
		t.Fatalf("MkdirAll npmBinDir: %v", err)
	}

	nodeScript := "#!/bin/sh\necho v20.11.0\n"
	npmScript := "#!/bin/sh\n" +
		"if [ \"$1\" = \"--version\" ]; then echo 10.8.0; exit 0; fi\n" +
		"if [ \"$1\" = \"config\" ] && [ \"$2\" = \"get\" ] && [ \"$3\" = \"prefix\" ]; then echo \"" + npmPrefix + "\"; exit 0; fi\n" +
		"if [ \"$1\" = \"install\" ] && [ \"$2\" = \"-g\" ] && [ \"$3\" = \"@openai/codex\" ]; then printf '#!/bin/sh\\necho codex-installed\\n' > \"" + filepath.Join(npmBinDir, "codex") + "\"; chmod +x \"" + filepath.Join(npmBinDir, "codex") + "\"; exit 0; fi\n" +
		"echo unexpected npm args: \"$@\" >&2\n" +
		"exit 1\n"

	if err := os.WriteFile(filepath.Join(binDir, "node"), []byte(nodeScript), 0o755); err != nil {
		t.Fatalf("WriteFile node: %v", err)
	}
	if err := os.WriteFile(filepath.Join(binDir, "npm"), []byte(npmScript), 0o755); err != nil {
		t.Fatalf("WriteFile npm: %v", err)
	}

	if err := os.Setenv("PATH", binDir+string(os.PathListSeparator)+"/usr/bin:/bin"); err != nil {
		t.Fatalf("Setenv PATH: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := EnsureCLI("codex", &stdout, &stderr); err != nil {
		t.Fatalf("EnsureCLI returned error: %v", err)
	}

	if _, err := lookPath("codex"); err != nil {
		t.Fatalf("expected codex on PATH after install: %v", err)
	}
	if !strings.Contains(stdout.String(), "Installed codex CLI") {
		t.Fatalf("stdout missing install confirmation: %s", stdout.String())
	}
}

func TestEnsureCLIFailsWhenNodeMissing(t *testing.T) {
	originalPath := os.Getenv("PATH")
	t.Cleanup(func() {
		_ = os.Setenv("PATH", originalPath)
	})

	home := t.TempDir()
	binDir := filepath.Join(home, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("MkdirAll binDir: %v", err)
	}
	npmScript := "#!/bin/sh\necho 10.8.0\n"
	if err := os.WriteFile(filepath.Join(binDir, "npm"), []byte(npmScript), 0o755); err != nil {
		t.Fatalf("WriteFile npm: %v", err)
	}
	if err := os.Setenv("PATH", binDir+string(os.PathListSeparator)+"/usr/bin:/bin"); err != nil {
		t.Fatalf("Setenv PATH: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := EnsureCLI("claude", &stdout, &stderr)
	if err == nil {
		t.Fatal("EnsureCLI returned nil error, want failure")
	}
	if !strings.Contains(err.Error(), "Node.js is not installed") {
		t.Fatalf("unexpected error: %v", err)
	}
}
