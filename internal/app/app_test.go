package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/LiukerSun/cc-cli/internal/config"
	"github.com/LiukerSun/cc-cli/internal/platform"
)

func setupTestHome(t *testing.T) string {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(home, "AppData", "Local"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	t.Setenv("XDG_STATE_HOME", filepath.Join(home, ".local", "state"))
	return home
}

func prependTestPath(t *testing.T, dirs ...string) {
	t.Helper()
	t.Setenv("PATH", strings.Join(dirs, string(os.PathListSeparator)))
}

func writeTestCommand(t *testing.T, dir, name, unixContent, windowsContent string) string {
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

func decodeProfileList(t *testing.T, output string) []config.Profile {
	t.Helper()

	var payload struct {
		Profiles []config.Profile `json:"profiles"`
	}
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("decode profile list: %v\noutput: %s", err, output)
	}
	return payload.Profiles
}

func findProfileByID(t *testing.T, profiles []config.Profile, id string) config.Profile {
	t.Helper()

	for _, profile := range profiles {
		if profile.ID == id {
			return profile
		}
	}
	t.Fatalf("profile %q not found in payload: %+v", id, profiles)
	return config.Profile{}
}

func TestProfileAddAndList(t *testing.T) {
	setupTestHome(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{
		"profile", "add",
		"--name", "Codex Relay",
		"--command", "codex",
		"--base-url", "https://relay.example.com/v1",
		"--api-key", "sk-test",
		"--model", "gpt-5.4",
	}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("profile add exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	exitCode = Run([]string{"profile", "list"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("profile list exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Codex Relay") {
		t.Fatalf("profile list output missing profile name: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "codex-relay") {
		t.Fatalf("profile list output missing generated id: %s", stdout.String())
	}
}

func TestTopLevelAddUsesPositionalPresetShortcut(t *testing.T) {
	setupTestHome(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{
		"add", "openai", "sk-test", "gpt-5.4-mini",
	}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("add exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	exitCode = Run([]string{"current"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("current exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Name: Codex OpenAI") {
		t.Fatalf("current output missing preset name: %s", output)
	}
	if !strings.Contains(output, "Provider: openai") {
		t.Fatalf("current output missing provider: %s", output)
	}
	if !strings.Contains(output, "Command: codex") {
		t.Fatalf("current output missing command: %s", output)
	}
	if !strings.Contains(output, "Model: gpt-5.4-mini") {
		t.Fatalf("current output missing positional model override: %s", output)
	}
}

func TestProfileAddAppliesPresetDefaults(t *testing.T) {
	setupTestHome(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{
		"profile", "add",
		"--preset", "zhipu",
		"--api-key", "zhipu-test-key",
	}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("profile add exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	exitCode = Run([]string{"current"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("current exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Name: Zhipu Claude-Compatible") {
		t.Fatalf("current output missing preset name: %s", output)
	}
	if !strings.Contains(output, "Command: claude") {
		t.Fatalf("current output missing preset command: %s", output)
	}
	if !strings.Contains(output, "Provider: zhipu") {
		t.Fatalf("current output missing preset provider: %s", output)
	}
	if !strings.Contains(output, "Base URL: https://open.bigmodel.cn/api/anthropic") {
		t.Fatalf("current output missing preset base url: %s", output)
	}
	if !strings.Contains(output, "Model: glm-5") {
		t.Fatalf("current output missing preset model: %s", output)
	}
}

func TestProfileAddRejectsUnknownPreset(t *testing.T) {
	setupTestHome(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{
		"profile", "add",
		"--preset", "unknown",
		"--api-key", "test-key",
	}, &stdout, &stderr)
	if exitCode == 0 {
		t.Fatal("expected profile add with unknown preset to fail")
	}
	if !strings.Contains(stderr.String(), "failed to apply preset") {
		t.Fatalf("unexpected stderr: %s", stderr.String())
	}
}

func TestProfileUpdateChangesPresetModelAndEnv(t *testing.T) {
	setupTestHome(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if exitCode := Run([]string{
		"profile", "add",
		"--preset", "openai",
		"--name", "Relay",
		"--api-key", "sk-test",
		"--base-url", "https://relay.example.com",
	}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile add exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{
		"profile", "update", "relay",
		"--preset", "zhipu",
		"--model", "glm-4.7-air",
		"--env", "FOO=bar",
		"--no-sync",
	}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile update exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"profile", "list", "--json"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile list exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	profile := findProfileByID(t, decodeProfileList(t, stdout.String()), "relay")
	if profile.Provider != "zhipu" {
		t.Fatalf("updated profile provider = %q, want zhipu", profile.Provider)
	}
	if profile.Command != "claude" {
		t.Fatalf("updated profile command = %q, want claude", profile.Command)
	}
	if profile.Model != "glm-4.7-air" {
		t.Fatalf("updated profile model = %q, want glm-4.7-air", profile.Model)
	}
	if profile.ExtraEnv["FOO"] != "bar" {
		t.Fatalf("updated profile env missing FOO=bar: %+v", profile.ExtraEnv)
	}
	if profile.SyncExternal {
		t.Fatalf("updated profile should disable sync: %+v", profile)
	}
}

func TestProfileUpdateRenamesCurrentProfileID(t *testing.T) {
	setupTestHome(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if exitCode := Run([]string{
		"profile", "add",
		"--preset", "anthropic",
		"--name", "Claude Official",
		"--api-key", "sk-ant-test",
	}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile add exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{
		"profile", "update", "claude-official",
		"--id", "claude-prod",
		"--name", "Claude Prod",
	}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile update exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"current"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("current exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	output := stdout.String()
	if !strings.Contains(output, "ID: claude-prod") {
		t.Fatalf("current output missing updated id: %s", output)
	}
	if !strings.Contains(output, "Name: Claude Prod") {
		t.Fatalf("current output missing updated name: %s", output)
	}
}

func TestConfigShowReadsLegacyConfig(t *testing.T) {
	home := setupTestHome(t)
	legacyPath := filepath.Join(home, ".ccc", "config.json")
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	content := `[
  {
    "name": "Legacy Claude",
    "env": {
      "ANTHROPIC_BASE_URL": "https://api.anthropic.com",
      "ANTHROPIC_AUTH_TOKEN": "token",
      "ANTHROPIC_MODEL": "claude-main"
    }
  }
]`
	if err := os.WriteFile(legacyPath, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"config", "show"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("config show exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"source": "legacy-root"`) {
		t.Fatalf("config show output missing legacy source: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), `"name": "Legacy Claude"`) {
		t.Fatalf("config show output missing profile: %s", stdout.String())
	}
}

func TestProfileUseAndCurrent(t *testing.T) {
	setupTestHome(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	for _, args := range [][]string{
		{"profile", "add", "--name", "Profile One", "--command", "claude", "--base-url", "https://api.anthropic.com", "--api-key", "token-1", "--model", "claude-main"},
		{"profile", "add", "--name", "Profile Two", "--command", "codex", "--base-url", "https://relay.example.com/v1", "--api-key", "token-2", "--model", "gpt-5.4"},
	} {
		stdout.Reset()
		stderr.Reset()
		if exitCode := Run(args, &stdout, &stderr); exitCode != 0 {
			t.Fatalf("Run(%v) exitCode = %d, stderr = %s", args, exitCode, stderr.String())
		}
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"profile", "use", "profile-two"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile use exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"current"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("current exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Profile Two") {
		t.Fatalf("current output missing selected profile: %s", stdout.String())
	}
}

func TestConfigMigrateWritesCurrentConfig(t *testing.T) {
	home := setupTestHome(t)

	legacyPath := filepath.Join(home, ".ccc", "config.json")
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	content := `[
  {
    "name": "Legacy Claude",
    "env": {
      "ANTHROPIC_BASE_URL": "https://api.anthropic.com",
      "ANTHROPIC_AUTH_TOKEN": "token",
      "ANTHROPIC_MODEL": "claude-main"
    }
  }
]`
	if err := os.WriteFile(legacyPath, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if exitCode := Run([]string{"config", "migrate"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("config migrate exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	layout, err := platform.ResolveLayout(runtime.GOOS, home, os.Getenv)
	if err != nil {
		t.Fatalf("ResolveLayout: %v", err)
	}
	newConfigPath := layout.ConfigFile()
	if _, err := os.Stat(newConfigPath); err != nil {
		t.Fatalf("expected migrated config at %s: %v", newConfigPath, err)
	}
}

func TestRunDryRunUsesCurrentProfile(t *testing.T) {
	setupTestHome(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if exitCode := Run([]string{
		"profile", "add",
		"--name", "Codex Relay",
		"--command", "codex",
		"--base-url", "https://relay.example.com/v1",
		"--api-key", "sk-test",
		"--model", "gpt-5.4",
	}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile add exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"run", "--dry-run", "--", "--help"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("run --dry-run exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Command: codex") {
		t.Fatalf("dry-run output missing command: %s", output)
	}
	if !strings.Contains(output, "OPENAI_MODEL=gpt-5.4") {
		t.Fatalf("dry-run output missing model env: %s", output)
	}
	if !strings.Contains(output, "Args: --help") {
		t.Fatalf("dry-run output missing args: %s", output)
	}
}

func TestRunExecutesTargetCommand(t *testing.T) {
	home := setupTestHome(t)
	binDir := filepath.Join(home, "bin")
	prependTestPath(t, binDir)

	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("MkdirAll binDir: %v", err)
	}
	script := "#!/bin/sh\n" +
		"printf 'cmd:%s\\n' \"$0\"\n" +
		"printf 'model:%s\\n' \"$OPENAI_MODEL\"\n" +
		"printf 'arg1:%s\\n' \"$1\"\n"
	scriptWin := "@echo off\r\n" +
		"echo cmd:%~f0\r\n" +
		"echo model:%OPENAI_MODEL%\r\n" +
		"echo arg1:%~1\r\n"
	writeTestCommand(t, binDir, "codex", script, scriptWin)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if exitCode := Run([]string{
		"profile", "add",
		"--name", "Codex Relay",
		"--command", "codex",
		"--base-url", "https://relay.example.com/v1",
		"--api-key", "sk-test",
		"--model", "gpt-5.4",
	}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile add exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"run", "--", "hello"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("run exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "model:gpt-5.4") {
		t.Fatalf("run output missing OPENAI_MODEL: %s", output)
	}
	if !strings.Contains(output, "arg1:hello") {
		t.Fatalf("run output missing arg1: %s", output)
	}

	if _, err := os.Stat(filepath.Join(home, ".codex", "config.toml")); err != nil {
		t.Fatalf("expected codex config to be synced: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".codex", "auth.json")); err != nil {
		t.Fatalf("expected codex auth to be synced: %v", err)
	}
}

func TestDoctorInspectsInstalledTools(t *testing.T) {
	home := setupTestHome(t)
	binDir := filepath.Join(home, "bin")
	prependTestPath(t, binDir)

	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("MkdirAll binDir: %v", err)
	}
	tools := map[string][2]string{
		"node":   {"#!/bin/sh\necho v20.11.0\n", "@echo off\r\necho v20.11.0\r\n"},
		"npm":    {"#!/bin/sh\necho 10.8.0\n", "@echo off\r\necho 10.8.0\r\n"},
		"codex":  {"#!/bin/sh\necho codex 0.1.0\n", "@echo off\r\necho codex 0.1.0\r\n"},
		"claude": {"#!/bin/sh\necho claude 0.1.0\n", "@echo off\r\necho claude 0.1.0\r\n"},
	}
	for name, content := range tools {
		writeTestCommand(t, binDir, name, content[0], content[1])
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if exitCode := Run([]string{
		"profile", "add",
		"--name", "Codex Relay",
		"--command", "codex",
		"--base-url", "https://relay.example.com/v1",
		"--api-key", "sk-test",
		"--model", "gpt-5.4",
	}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile add exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"doctor"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("doctor exitCode = %d, stdout = %s stderr = %s", exitCode, stdout.String(), stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "- node installed: yes") {
		t.Fatalf("doctor output missing node status: %s", output)
	}
	if !strings.Contains(output, "- codex installed: yes") {
		t.Fatalf("doctor output missing codex status: %s", output)
	}
}

func TestRunAutoInstallsMissingCLI(t *testing.T) {
	home := setupTestHome(t)
	binDir := filepath.Join(home, "bin")
	npmPrefix := filepath.Join(home, ".npm-global")
	npmBinDir := filepath.Join(npmPrefix, "bin")
	prependTestPath(t, binDir)

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
		"if [ \"$1\" = \"install\" ] && [ \"$2\" = \"-g\" ] && [ \"$3\" = \"@openai/codex\" ]; then printf '#!/bin/sh\\necho auto-installed-codex\\n' > \"" + filepath.Join(npmBinDir, "codex") + "\"; /bin/chmod +x \"" + filepath.Join(npmBinDir, "codex") + "\"; exit 0; fi\n" +
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
		"    echo echo auto-installed-codex\r\n" +
		"  )\r\n" +
		"  exit /b 0\r\n" +
		")\r\n" +
		"echo unexpected npm args: %* 1>&2\r\n" +
		"exit /b 1\r\n"
	writeTestCommand(t, binDir, "node", nodeScript, nodeScriptWin)
	writeTestCommand(t, binDir, "npm", npmScript, npmScriptWin)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if exitCode := Run([]string{
		"profile", "add",
		"--name", "Codex Relay",
		"--command", "codex",
		"--base-url", "https://relay.example.com",
		"--api-key", "sk-test",
		"--model", "gpt-5.4",
	}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile add exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"run"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("run exitCode = %d, stdout = %s stderr = %s", exitCode, stdout.String(), stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Installed codex CLI") {
		t.Fatalf("run output missing install confirmation: %s", output)
	}
	if !strings.Contains(output, "auto-installed-codex") {
		t.Fatalf("run output missing installed command output: %s", output)
	}
}

func TestSyncDryRunShowsTargets(t *testing.T) {
	setupTestHome(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if exitCode := Run([]string{
		"profile", "add",
		"--name", "Claude Main",
		"--command", "claude",
		"--base-url", "https://api.anthropic.com",
		"--api-key", "token",
		"--model", "claude-main",
	}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile add exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"sync", "--dry-run"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("sync --dry-run exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), filepath.Join(".claude", "settings.json")) {
		t.Fatalf("sync dry-run missing target path: %s", stdout.String())
	}
}

func TestSyncWritesClaudeSettings(t *testing.T) {
	home := setupTestHome(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if exitCode := Run([]string{
		"profile", "add",
		"--name", "Claude Main",
		"--command", "claude",
		"--base-url", "https://api.anthropic.com",
		"--api-key", "token",
		"--model", "claude-main",
		"--fast-model", "claude-fast",
	}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile add exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"sync"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("sync exitCode = %d, stdout = %s stderr = %s", exitCode, stdout.String(), stderr.String())
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile settings: %v", err)
	}
	if !strings.Contains(string(data), "CLAUDE_CODE_MODEL") {
		t.Fatalf("claude settings missing synced env: %s", string(data))
	}
}

func TestUpgradeDryRunShowsResolvedTarget(t *testing.T) {
	setupTestHome(t)
	t.Setenv("CCC_RELEASE_DOWNLOAD_BASE_URL", "https://example.com/releases/download")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if exitCode := Run([]string{"upgrade", "--version", "2.2.1", "--dry-run"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("upgrade --dry-run exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Target version: 2.2.1") {
		t.Fatalf("upgrade --dry-run output missing target version: %s", output)
	}
	expectedAsset := fmt.Sprintf("ccc_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		expectedAsset = fmt.Sprintf("ccc_%s_%s.zip", runtime.GOOS, runtime.GOARCH)
	}
	if !strings.Contains(output, expectedAsset) {
		t.Fatalf("upgrade --dry-run output missing asset name: %s", output)
	}
}
