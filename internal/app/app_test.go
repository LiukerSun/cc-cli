package app

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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

func stubInteractiveModelFetchers(t *testing.T, zhipuModels, alibabaModels []string) {
	t.Helper()

	previousZhipu := fetchZhipuModels
	previousAlibaba := fetchAlibabaModels
	fetchZhipuModels = func(string) ([]string, error) {
		return append([]string(nil), zhipuModels...), nil
	}
	fetchAlibabaModels = func(string) ([]string, error) {
		return append([]string(nil), alibabaModels...), nil
	}
	t.Cleanup(func() {
		fetchZhipuModels = previousZhipu
		fetchAlibabaModels = previousAlibaba
	})
}

func stubInteractiveInput(t *testing.T, interactive bool) {
	t.Helper()

	previous := stdinIsInteractive
	stdinIsInteractive = func(io.Reader) bool {
		return interactive
	}
	t.Cleanup(func() {
		stdinIsInteractive = previous
	})
}

func stubSecretPrompt(t *testing.T, supported bool, value string, err error) *int {
	t.Helper()

	previousSupported := secretInputSupported
	previousRead := readSecretInput
	calls := 0
	secretInputSupported = func(io.Reader) bool {
		return supported
	}
	readSecretInput = func(io.Reader, io.Writer, string) (string, error) {
		calls++
		return value, err
	}
	t.Cleanup(func() {
		secretInputSupported = previousSupported
		readSecretInput = previousRead
	})
	return &calls
}

func stubArrowSelector(t *testing.T, enabled bool) {
	t.Helper()

	previousEnabled := arrowSelectorEnabled
	previousRaw := makeRawSelectorInput
	arrowSelectorEnabled = func(io.Reader, io.Writer) bool {
		return enabled
	}
	makeRawSelectorInput = func(io.Reader) (func(), error) {
		return func() {}, nil
	}
	t.Cleanup(func() {
		arrowSelectorEnabled = previousEnabled
		makeRawSelectorInput = previousRaw
	})
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

func TestTopLevelAddInteractiveZhipu(t *testing.T) {
	home := setupTestHome(t)
	layout, err := platform.ResolveLayout(runtime.GOOS, home, os.Getenv)
	if err != nil {
		t.Fatalf("ResolveLayout: %v", err)
	}

	stubInteractiveModelFetchers(t, []string{"glm-5", "glm-4.7"}, interactiveAlibabaFallbackModels)

	store := config.NewStore(home, layout)
	input := strings.NewReader("3\nzhipu-test-key\n2\n1\n\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := runAddInteractive(input, &stdout, &stderr, store, addProfileOptions{})
	if exitCode != 0 {
		t.Fatalf("runAddInteractive exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	cfg, _, err := store.Load()
	if err != nil {
		t.Fatalf("store.Load: %v", err)
	}

	profile, ok := cfg.FindProfile("zhipu-glm-4-7")
	if !ok {
		t.Fatalf("expected interactive add to create zhipu-glm-4-7 profile: %+v", cfg.Profiles)
	}
	if profile.Provider != "zhipu" {
		t.Fatalf("Provider = %q, want zhipu", profile.Provider)
	}
	if profile.APIKey != "zhipu-test-key" {
		t.Fatalf("APIKey = %q, want zhipu-test-key", profile.APIKey)
	}
	if profile.Model != "glm-4.7" {
		t.Fatalf("Model = %q, want glm-4.7", profile.Model)
	}
	if profile.FastModel != "glm-5" {
		t.Fatalf("FastModel = %q, want glm-5", profile.FastModel)
	}
	if profile.Name != "ZHIPU (glm-4.7)" {
		t.Fatalf("Name = %q, want ZHIPU (glm-4.7)", profile.Name)
	}
}

func TestTopLevelAddInteractiveAlibaba(t *testing.T) {
	home := setupTestHome(t)
	layout, err := platform.ResolveLayout(runtime.GOOS, home, os.Getenv)
	if err != nil {
		t.Fatalf("ResolveLayout: %v", err)
	}

	stubInteractiveModelFetchers(t, interactiveZhipuFallbackModels, []string{"qwen3.6-plus", "qwen3.5-plus", "qwen3-coder-next"})

	store := config.NewStore(home, layout)
	input := strings.NewReader("4\nalibaba-test-key\n3\n\n\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := runAddInteractive(input, &stdout, &stderr, store, addProfileOptions{})
	if exitCode != 0 {
		t.Fatalf("runAddInteractive exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	cfg, _, err := store.Load()
	if err != nil {
		t.Fatalf("store.Load: %v", err)
	}

	profile, ok := cfg.FindProfile("alibaba-coding-plan-qwen3-coder-next")
	if !ok {
		t.Fatalf("expected interactive add to create alibaba coding plan profile: %+v", cfg.Profiles)
	}
	if profile.Provider != "alibaba" {
		t.Fatalf("Provider = %q, want alibaba", profile.Provider)
	}
	if profile.APIKey != "alibaba-test-key" {
		t.Fatalf("APIKey = %q, want alibaba-test-key", profile.APIKey)
	}
	if profile.Model != "qwen3-coder-next" {
		t.Fatalf("Model = %q, want qwen3-coder-next", profile.Model)
	}
	if profile.FastModel != "qwen3.5-plus" {
		t.Fatalf("FastModel = %q, want qwen3.5-plus", profile.FastModel)
	}
}

func TestTopLevelAddInteractiveKimi(t *testing.T) {
	home := setupTestHome(t)
	layout, err := platform.ResolveLayout(runtime.GOOS, home, os.Getenv)
	if err != nil {
		t.Fatalf("ResolveLayout: %v", err)
	}

	store := config.NewStore(home, layout)
	input := strings.NewReader("5\nkimi-test-key\n\n\n\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := runAddInteractive(input, &stdout, &stderr, store, addProfileOptions{})
	if exitCode != 0 {
		t.Fatalf("runAddInteractive exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	cfg, _, err := store.Load()
	if err != nil {
		t.Fatalf("store.Load: %v", err)
	}

	profile, ok := cfg.FindProfile("kimi-coding-plan-k2-6-code-preview")
	if !ok {
		t.Fatalf("expected interactive add to create kimi profile: %+v", cfg.Profiles)
	}
	if profile.Provider != "kimi" {
		t.Fatalf("Provider = %q, want kimi", profile.Provider)
	}
	if profile.BaseURL != "https://api.kimi.com/coding/" {
		t.Fatalf("BaseURL = %q, want https://api.kimi.com/coding/", profile.BaseURL)
	}
	if profile.APIKey != "kimi-test-key" {
		t.Fatalf("APIKey = %q, want kimi-test-key", profile.APIKey)
	}
	if profile.Model != "K2.6-code-preview" {
		t.Fatalf("Model = %q, want K2.6-code-preview", profile.Model)
	}
	if profile.FastModel != "K2.6-code-preview" {
		t.Fatalf("FastModel = %q, want K2.6-code-preview", profile.FastModel)
	}
}

func TestTopLevelAddInteractiveCodex(t *testing.T) {
	home := setupTestHome(t)
	layout, err := platform.ResolveLayout(runtime.GOOS, home, os.Getenv)
	if err != nil {
		t.Fatalf("ResolveLayout: %v", err)
	}

	store := config.NewStore(home, layout)
	input := strings.NewReader("2\n\nsk-openai-test\n2\n\n")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := runAddInteractive(input, &stdout, &stderr, store, addProfileOptions{})
	if exitCode != 0 {
		t.Fatalf("runAddInteractive exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	cfg, _, err := store.Load()
	if err != nil {
		t.Fatalf("store.Load: %v", err)
	}

	profile, ok := cfg.FindProfile("codex-gpt-5-4-mini")
	if !ok {
		t.Fatalf("expected interactive add to create codex profile: %+v", cfg.Profiles)
	}
	if profile.Command != "codex" {
		t.Fatalf("Command = %q, want codex", profile.Command)
	}
	if profile.BaseURL != "https://api.openai.com/v1" {
		t.Fatalf("BaseURL = %q, want https://api.openai.com/v1", profile.BaseURL)
	}
	if profile.APIKey != "sk-openai-test" {
		t.Fatalf("APIKey = %q, want sk-openai-test", profile.APIKey)
	}
	if profile.Model != "gpt-5.4-mini" {
		t.Fatalf("Model = %q, want gpt-5.4-mini", profile.Model)
	}
	if profile.Name != "Codex (gpt-5.4-mini)" {
		t.Fatalf("Name = %q, want Codex (gpt-5.4-mini)", profile.Name)
	}
}

func TestPromptSecretRequiredUsesHiddenInputWhenAvailable(t *testing.T) {
	calls := stubSecretPrompt(t, true, "sk-hidden", nil)

	var stdout bytes.Buffer
	value, err := promptSecretRequired(strings.NewReader("ignored\n"), bufio.NewReader(strings.NewReader("ignored\n")), &stdout, "API key")
	if err != nil {
		t.Fatalf("promptSecretRequired returned error: %v", err)
	}
	if value != "sk-hidden" {
		t.Fatalf("value = %q, want sk-hidden", value)
	}
	if *calls != 1 {
		t.Fatalf("hidden secret prompt calls = %d, want 1", *calls)
	}
}

func TestPromptSecretRequiredFallsBackToBufferedReader(t *testing.T) {
	calls := stubSecretPrompt(t, false, "", nil)

	input := strings.NewReader("sk-buffered\n")
	var stdout bytes.Buffer
	value, err := promptSecretRequired(input, bufio.NewReader(input), &stdout, "API key")
	if err != nil {
		t.Fatalf("promptSecretRequired returned error: %v", err)
	}
	if value != "sk-buffered" {
		t.Fatalf("value = %q, want sk-buffered", value)
	}
	if *calls != 0 {
		t.Fatalf("hidden secret prompt calls = %d, want 0", *calls)
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

func TestProfileAddAppliesAlibabaPresetDefaults(t *testing.T) {
	setupTestHome(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{
		"profile", "add",
		"--preset", "qwen",
		"--api-key", "alibaba-test-key",
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
	if !strings.Contains(output, "Name: Alibaba Coding Plan") {
		t.Fatalf("current output missing preset name: %s", output)
	}
	if !strings.Contains(output, "Provider: alibaba") {
		t.Fatalf("current output missing preset provider: %s", output)
	}
	if !strings.Contains(output, "Base URL: https://coding.dashscope.aliyuncs.com/apps/anthropic") {
		t.Fatalf("current output missing preset base url: %s", output)
	}
	if !strings.Contains(output, "Model: qwen3.6-plus") {
		t.Fatalf("current output missing preset model: %s", output)
	}
}

func TestProfileAddAppliesKimiPresetDefaults(t *testing.T) {
	setupTestHome(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{
		"profile", "add",
		"--preset", "kimi",
		"--api-key", "kimi-test-key",
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
	if !strings.Contains(output, "Name: Kimi Coding Plan") {
		t.Fatalf("current output missing preset name: %s", output)
	}
	if !strings.Contains(output, "Provider: kimi") {
		t.Fatalf("current output missing preset provider: %s", output)
	}
	if !strings.Contains(output, "Base URL: https://api.kimi.com/coding/") {
		t.Fatalf("current output missing preset base url: %s", output)
	}
	if !strings.Contains(output, "Model: K2.6-code-preview") {
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

func TestProfileAddAndUpdateManageSyncDenyPermissions(t *testing.T) {
	setupTestHome(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if exitCode := Run([]string{
		"profile", "add",
		"--preset", "anthropic",
		"--name", "Claude Restricted",
		"--api-key", "sk-ant-test",
		"--deny-permission", "Agent(Explore)",
		"--deny-permission", "Bash(rm -rf)",
	}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile add exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{
		"profile", "update", "claude-restricted",
		"--unset-deny-permission", "Agent(Explore)",
		"--deny-permission", "Read(/etc/shadow)",
	}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile update exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"profile", "list", "--json"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile list exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	profile := findProfileByID(t, decodeProfileList(t, stdout.String()), "claude-restricted")
	want := []string{"Bash(rm -rf)", "Read(/etc/shadow)"}
	if len(profile.SyncDenyPermissions) != len(want) {
		t.Fatalf("SyncDenyPermissions = %#v, want %#v", profile.SyncDenyPermissions, want)
	}
	for i := range want {
		if profile.SyncDenyPermissions[i] != want[i] {
			t.Fatalf("SyncDenyPermissions[%d] = %q, want %q", i, profile.SyncDenyPermissions[i], want[i])
		}
	}
}

func TestProfileDuplicateCreatesUniqueCopy(t *testing.T) {
	setupTestHome(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if exitCode := Run([]string{
		"profile", "add",
		"--preset", "openai",
		"--name", "Relay",
		"--api-key", "sk-test",
	}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile add exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{
		"profile", "duplicate", "relay",
		"--name", "Relay Backup",
	}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile duplicate exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"profile", "list", "--json", "--show-secrets"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile list exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	profiles := decodeProfileList(t, stdout.String())
	if len(profiles) != 2 {
		t.Fatalf("len(profiles) = %d, want 2", len(profiles))
	}
	profile := findProfileByID(t, profiles, "relay-backup")
	if profile.Name != "Relay Backup" {
		t.Fatalf("duplicated profile name = %q, want Relay Backup", profile.Name)
	}
	if profile.APIKey != "sk-test" {
		t.Fatalf("duplicated profile api key = %q, want sk-test", profile.APIKey)
	}
}

func TestProfileListJSONMasksSecretsByDefault(t *testing.T) {
	setupTestHome(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if exitCode := Run([]string{
		"profile", "add",
		"--preset", "openai",
		"--name", "Masked Relay",
		"--api-key", "sk-secret-1234",
	}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile add exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"profile", "list", "--json"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile list exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	if strings.Contains(stdout.String(), "sk-secret-1234") {
		t.Fatalf("profile list JSON should mask secrets: %s", stdout.String())
	}
	profile := findProfileByID(t, decodeProfileList(t, stdout.String()), "masked-relay")
	if profile.APIKey != "****1234" {
		t.Fatalf("masked profile api key = %q, want ****1234", profile.APIKey)
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"profile", "list", "--json", "--show-secrets"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile list --show-secrets exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	profile = findProfileByID(t, decodeProfileList(t, stdout.String()), "masked-relay")
	if profile.APIKey != "sk-secret-1234" {
		t.Fatalf("unmasked profile api key = %q, want sk-secret-1234", profile.APIKey)
	}
}

func TestProfileExportAndImportRoundTrip(t *testing.T) {
	exportPath := filepath.Join(t.TempDir(), "profiles.json")

	setupTestHome(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if exitCode := Run([]string{
		"profile", "add",
		"--preset", "anthropic",
		"--name", "Claude Export",
		"--api-key", "sk-ant-test",
		"--deny-permission", "Agent(Explore)",
	}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile add exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{
		"profile", "export", "claude-export",
		"--output", exportPath,
	}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile export exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	setupTestHome(t)
	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{
		"profile", "import",
		"--input", exportPath,
	}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile import exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"profile", "list", "--json"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("profile list exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	profile := findProfileByID(t, decodeProfileList(t, stdout.String()), "claude-export")
	if profile.Name != "Claude Export" {
		t.Fatalf("imported profile name = %q, want Claude Export", profile.Name)
	}
	if len(profile.SyncDenyPermissions) != 1 || profile.SyncDenyPermissions[0] != "Agent(Explore)" {
		t.Fatalf("imported deny permissions = %#v", profile.SyncDenyPermissions)
	}
}

func TestCompatibilityAliases(t *testing.T) {
	setupTestHome(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if exitCode := Run([]string{"--add", "openai", "sk-test", "gpt-5.4-mini"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("--add exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"--list"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("--list exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Codex OpenAI") {
		t.Fatalf("--list output missing profile: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"--current"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("--current exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "ID: codex-openai") {
		t.Fatalf("--current output missing current profile: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"-e", "codex-openai"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("-e exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "OPENAI_API_KEY=sk-test") {
		t.Fatalf("-e output missing env vars: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := Run([]string{"--delete", "codex-openai"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("--delete exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
}

func TestUpgradeCheckReportsLatestVersion(t *testing.T) {
	setupTestHome(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/LiukerSun/cc-cli/releases/latest":
			_, _ = io.WriteString(w, `{"tag_name":"v2.3.4"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("CCC_RELEASE_API_BASE_URL", server.URL)
	t.Setenv("CCC_RELEASE_DOWNLOAD_BASE_URL", server.URL+"/LiukerSun/cc-cli/releases/download")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"upgrade", "--check"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("upgrade --check exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Latest version: 2.3.4") {
		t.Fatalf("upgrade --check output missing latest version: %s", stdout.String())
	}
}

func TestUpgradeCheckFallsBackWhenGitHubAPIIsForbidden(t *testing.T) {
	setupTestHome(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/LiukerSun/cc-cli/releases/latest":
			http.Error(w, "rate limited", http.StatusForbidden)
		case "/LiukerSun/cc-cli/releases/latest":
			http.Redirect(w, r, "/LiukerSun/cc-cli/releases/tag/v2.3.5", http.StatusFound)
		case "/LiukerSun/cc-cli/releases/tag/v2.3.5":
			_, _ = io.WriteString(w, "release page")
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("CCC_RELEASE_API_BASE_URL", server.URL)
	t.Setenv("CCC_RELEASE_DOWNLOAD_BASE_URL", server.URL+"/LiukerSun/cc-cli/releases/download")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	exitCode := Run([]string{"upgrade", "--check"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("upgrade --check exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Latest version: 2.3.5") {
		t.Fatalf("upgrade --check output missing fallback latest version: %s", stdout.String())
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
	if strings.Contains(stdout.String(), `"api_key": "token"`) {
		t.Fatalf("config show should mask api_key by default: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), `"api_key": "****oken"`) {
		t.Fatalf("config show output missing masked api_key: %s", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	exitCode = Run([]string{"config", "show", "--show-secrets"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("config show --show-secrets exitCode = %d, stderr = %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"api_key": "token"`) {
		t.Fatalf("config show --show-secrets output missing raw api_key: %s", stdout.String())
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

func TestCompletionBashOutputsScript(t *testing.T) {
	setupTestHome(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if exitCode := Run([]string{"completion", "bash"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("completion bash exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "complete -o default -F _ccc_completion ccc") {
		t.Fatalf("bash completion output missing complete hook: %s", output)
	}
	if !strings.Contains(output, "ccc __complete") {
		t.Fatalf("bash completion output missing __complete bridge: %s", output)
	}
}

func TestInternalCompleteTopLevelCommands(t *testing.T) {
	setupTestHome(t)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if exitCode := Run([]string{"__complete"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("__complete exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "completion\n") {
		t.Fatalf("completion output missing completion command: %s", output)
	}
	if !strings.Contains(output, "run\n") {
		t.Fatalf("completion output missing run command: %s", output)
	}
}

func TestInternalCompleteRunProfiles(t *testing.T) {
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
	if exitCode := Run([]string{"__complete", "run", ""}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("__complete run exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "codex-relay\n") {
		t.Fatalf("completion output missing profile id: %s", output)
	}
	if !strings.Contains(output, "--auto-sync\n") {
		t.Fatalf("completion output missing run flag: %s", output)
	}
}

func TestInternalCompleteProfileUpdateIdentifiers(t *testing.T) {
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
	if exitCode := Run([]string{"__complete", "profile", "update", ""}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("__complete profile update exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	if !strings.Contains(stdout.String(), "claude-main\n") {
		t.Fatalf("completion output missing profile id: %s", stdout.String())
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
	if !strings.Contains(output, "Missing CLI policy: fail") {
		t.Fatalf("dry-run output missing missing-CLI policy: %s", output)
	}
	if !strings.Contains(output, "OPENAI_MODEL=gpt-5.4") {
		t.Fatalf("dry-run output missing model env: %s", output)
	}
	if !strings.Contains(output, "Args: --help") {
		t.Fatalf("dry-run output missing args: %s", output)
	}
}

func TestRunDryRunPromptsForProfileSelection(t *testing.T) {
	home := setupTestHome(t)
	layout, err := platform.ResolveLayout(runtime.GOOS, home, os.Getenv)
	if err != nil {
		t.Fatalf("ResolveLayout: %v", err)
	}
	store := config.NewStore(home, layout)

	cfg := config.DefaultFile()
	if err := cfg.UpsertProfile(config.Profile{
		ID:           "zhipu-main",
		Name:         "ZHIPU Main",
		Command:      "claude",
		Provider:     "zhipu",
		BaseURL:      "https://open.bigmodel.cn/api/anthropic",
		APIKey:       "zhipu-key",
		Model:        "glm-5",
		FastModel:    "glm-4.7",
		SyncExternal: true,
	}); err != nil {
		t.Fatalf("UpsertProfile #1: %v", err)
	}
	if err := cfg.UpsertProfile(config.Profile{
		ID:           "codex-relay",
		Name:         "Codex Relay",
		Command:      "codex",
		Provider:     "openai",
		BaseURL:      "https://relay.example.com/v1",
		APIKey:       "sk-test",
		Model:        "gpt-5.4-mini",
		FastModel:    "gpt-5.4-mini",
		SyncExternal: true,
	}); err != nil {
		t.Fatalf("UpsertProfile #2: %v", err)
	}
	cfg.CurrentProfile = "zhipu-main"
	if err := store.Save(cfg); err != nil {
		t.Fatalf("store.Save: %v", err)
	}

	stubInteractiveInput(t, true)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	input := strings.NewReader("2\n")
	if exitCode := runRunWithInput(input, &stdout, &stderr, home, layout, []string{"--dry-run", "--", "--help"}); exitCode != 0 {
		t.Fatalf("runRunWithInput exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Select a profile:") {
		t.Fatalf("output missing selection prompt: %s", output)
	}
	if !strings.Contains(output, "Codex Relay") {
		t.Fatalf("output missing listed profile: %s", output)
	}
	if !strings.Contains(output, "Command: codex") {
		t.Fatalf("dry-run output missing selected command: %s", output)
	}
	if !strings.Contains(output, "OPENAI_MODEL=gpt-5.4-mini") {
		t.Fatalf("dry-run output missing selected model env: %s", output)
	}

	savedCfg, _, err := store.Load()
	if err != nil {
		t.Fatalf("store.Load: %v", err)
	}
	if savedCfg.CurrentProfile != "codex-relay" {
		t.Fatalf("CurrentProfile = %q, want codex-relay", savedCfg.CurrentProfile)
	}
}

func TestRunDryRunSupportsArrowSelection(t *testing.T) {
	home := setupTestHome(t)
	layout, err := platform.ResolveLayout(runtime.GOOS, home, os.Getenv)
	if err != nil {
		t.Fatalf("ResolveLayout: %v", err)
	}
	store := config.NewStore(home, layout)

	cfg := config.DefaultFile()
	if err := cfg.UpsertProfile(config.Profile{
		ID:           "zhipu-main",
		Name:         "ZHIPU Main",
		Command:      "claude",
		Provider:     "zhipu",
		BaseURL:      "https://open.bigmodel.cn/api/anthropic",
		APIKey:       "zhipu-key",
		Model:        "glm-5",
		FastModel:    "glm-4.7",
		SyncExternal: true,
	}); err != nil {
		t.Fatalf("UpsertProfile #1: %v", err)
	}
	if err := cfg.UpsertProfile(config.Profile{
		ID:           "codex-relay",
		Name:         "Codex Relay",
		Command:      "codex",
		Provider:     "openai",
		BaseURL:      "https://relay.example.com/v1",
		APIKey:       "sk-test",
		Model:        "gpt-5.4-mini",
		FastModel:    "gpt-5.4-mini",
		SyncExternal: true,
	}); err != nil {
		t.Fatalf("UpsertProfile #2: %v", err)
	}
	cfg.CurrentProfile = "zhipu-main"
	if err := store.Save(cfg); err != nil {
		t.Fatalf("store.Save: %v", err)
	}

	stubInteractiveInput(t, true)
	stubArrowSelector(t, true)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	input := strings.NewReader("\x1b[B\r")
	if exitCode := runRunWithInput(input, &stdout, &stderr, home, layout, []string{"--dry-run"}); exitCode != 0 {
		t.Fatalf("runRunWithInput exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Use Up/Down to choose a profile") {
		t.Fatalf("output missing arrow selector help: %s", output)
	}
	if !strings.Contains(output, "ccc\r\n\r\nUse Up/Down to choose a profile") {
		t.Fatalf("output missing CRLF selector header formatting: %q", output)
	}
	if !strings.Contains(output, "Command: codex") {
		t.Fatalf("dry-run output missing selected command: %s", output)
	}
}

func TestRunWithoutArgsExecutesOnlyProfile(t *testing.T) {
	home := setupTestHome(t)
	binDir := filepath.Join(home, "bin")
	prependTestPath(t, binDir)

	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("MkdirAll binDir: %v", err)
	}
	script := "#!/bin/sh\nprintf 'ran:%s\\n' \"$OPENAI_MODEL\"\n"
	scriptWin := "@echo off\r\necho ran:%OPENAI_MODEL%\r\n"
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
	if exitCode := Run([]string{}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("Run with no args exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	if !strings.Contains(stdout.String(), "ran:gpt-5.4") {
		t.Fatalf("output missing executed model: %s", stdout.String())
	}
}

func TestTopLevelBypassShortcutExecutesRun(t *testing.T) {
	home := setupTestHome(t)
	binDir := filepath.Join(home, "bin")
	prependTestPath(t, binDir)

	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("MkdirAll binDir: %v", err)
	}
	script := "#!/bin/sh\nprintf 'arg1:%s\\n' \"$1\"\n"
	scriptWin := "@echo off\r\necho arg1:%~1\r\n"
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
	if exitCode := Run([]string{"-y"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("Run -y exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	if !strings.Contains(stdout.String(), "arg1:--yolo") {
		t.Fatalf("output missing bypass arg: %s", stdout.String())
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
		t.Fatalf("expected codex config to be synced before run: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".codex", "auth.json")); err != nil {
		t.Fatalf("expected codex auth to be synced before run: %v", err)
	}
}

func TestRunAutoSyncWritesExternalConfig(t *testing.T) {
	home := setupTestHome(t)
	binDir := filepath.Join(home, "bin")
	prependTestPath(t, binDir)

	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("MkdirAll binDir: %v", err)
	}
	script := "#!/bin/sh\nprintf 'synced-run:%s\\n' \"$OPENAI_MODEL\"\n"
	scriptWin := "@echo off\r\necho synced-run:%OPENAI_MODEL%\r\n"
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
	if exitCode := Run([]string{"run", "--auto-sync"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("run --auto-sync exitCode = %d, stderr = %s", exitCode, stderr.String())
	}

	if !strings.Contains(stdout.String(), "synced-run:gpt-5.4") {
		t.Fatalf("run output missing executed model: %s", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(home, ".codex", "config.toml")); err != nil {
		t.Fatalf("expected codex config to be synced with --auto-sync: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".codex", "auth.json")); err != nil {
		t.Fatalf("expected codex auth to be synced with --auto-sync: %v", err)
	}
}

func TestRunSyncsClaudeSettingsBeforeExecution(t *testing.T) {
	home := setupTestHome(t)
	binDir := filepath.Join(home, "bin")
	prependTestPath(t, binDir)

	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("MkdirAll binDir: %v", err)
	}
	script := "#!/bin/sh\nwhile IFS= read -r line; do\n  printf '%s\\n' \"$line\"\ndone < \"$HOME/.claude/settings.json\"\n"
	scriptWin := "@echo off\r\ntype \"%USERPROFILE%\\.claude\\settings.json\"\r\n"
	writeTestCommand(t, binDir, "claude", script, scriptWin)

	settingsPath := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll settings dir: %v", err)
	}
	initial := `{
  "model": "glm-5",
  "env": {
    "CLAUDE_CODE_MODEL": "glm-5"
  }
}`
	if err := os.WriteFile(settingsPath, []byte(initial), 0o600); err != nil {
		t.Fatalf("WriteFile settings: %v", err)
	}

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
	if exitCode := Run([]string{"run"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("run exitCode = %d, stdout = %s stderr = %s", exitCode, stdout.String(), stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, `"model": "claude-main"`) {
		t.Fatalf("claude settings were not synced before run: %s", output)
	}
	if !strings.Contains(output, `"CLAUDE_CODE_MODEL": "claude-main"`) {
		t.Fatalf("claude env model was not synced before run: %s", output)
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

func TestDoctorWarnsWhenCurrentProfileCommandMissing(t *testing.T) {
	setupTestHome(t)
	t.Setenv("PATH", "")

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
	exitCode := Run([]string{"doctor"}, &stdout, &stderr)
	if exitCode == 0 {
		t.Fatalf("doctor exitCode = %d, want warning exit", exitCode)
	}
	if !strings.Contains(stdout.String(), "use 'ccc run --auto-install'") {
		t.Fatalf("doctor output missing auto-install guidance: %s", stdout.String())
	}
}

func TestRunFailsWhenCLIMissingWithoutAutoInstall(t *testing.T) {
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
	if exitCode := Run([]string{"run"}, &stdout, &stderr); exitCode == 0 {
		t.Fatalf("run exitCode = %d, want failure when CLI is missing", exitCode)
	}
	if !strings.Contains(stderr.String(), "rerun with --auto-install") {
		t.Fatalf("stderr missing auto-install guidance: %s", stderr.String())
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
	if exitCode := Run([]string{"run", "--auto-install"}, &stdout, &stderr); exitCode != 0 {
		t.Fatalf("run --auto-install exitCode = %d, stdout = %s stderr = %s", exitCode, stdout.String(), stderr.String())
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
