package deps

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func prependPath(t *testing.T, dirs ...string) {
	t.Helper()
	if err := os.Setenv("PATH", strings.Join(dirs, string(os.PathListSeparator))); err != nil {
		t.Fatalf("Setenv PATH: %v", err)
	}
}

func writeFakeCommand(t *testing.T, dir, name, unixContent, windowsContent string) string {
	t.Helper()

	fileName := name
	content := unixContent
	mode := os.FileMode(0o755)
	if runtime.GOOS == "windows" {
		fileName += ".cmd"
		content = windowsContent
		mode = 0o644
	}

	path := filepath.Join(dir, fileName)
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatalf("WriteFile %s: %v", path, err)
	}
	return path
}

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
		"if [ \"$1\" = \"install\" ] && [ \"$2\" = \"-g\" ] && [ \"$3\" = \"@openai/codex\" ]; then printf '#!/bin/sh\\necho codex-installed\\n' > \"" + filepath.Join(npmBinDir, "codex") + "\"; /bin/chmod +x \"" + filepath.Join(npmBinDir, "codex") + "\"; exit 0; fi\n" +
		"echo unexpected npm args: \"$@\" >&2\n" +
		"exit 1\n"
	nodeScriptWin := "@echo off\r\necho v20.11.0\r\n"
	npmScriptWin := "@echo off\r\n" +
		"if \"%~1\"==\"--version\" (\r\n" +
		"  echo 10.8.0\r\n" +
		"  exit /b 0\r\n" +
		")\r\n" +
		"if \"%~1\"==\"config\" if \"%~2\"==\"get\" if \"%~3\"==\"prefix\" (\r\n" +
		"  echo " + npmPrefix + "\r\n" +
		"  exit /b 0\r\n" +
		")\r\n" +
		"if \"%~1\"==\"install\" if \"%~2\"==\"-g\" if \"%~3\"==\"@openai/codex\" (\r\n" +
		"  > \"" + filepath.Join(npmBinDir, "codex.cmd") + "\" (\r\n" +
		"    echo @echo off\r\n" +
		"    echo echo codex-installed\r\n" +
		"  )\r\n" +
		"  exit /b 0\r\n" +
		")\r\n" +
		"echo unexpected npm args: %* 1>&2\r\n" +
		"exit /b 1\r\n"

	writeFakeCommand(t, binDir, "node", nodeScript, nodeScriptWin)
	writeFakeCommand(t, binDir, "npm", npmScript, npmScriptWin)
	prependPath(t, binDir)

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
	npmScriptWin := "@echo off\r\necho 10.8.0\r\n"
	writeFakeCommand(t, binDir, "npm", npmScript, npmScriptWin)
	prependPath(t, binDir)

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
