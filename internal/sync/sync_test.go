package sync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/LiukerSun/cc-cli/internal/config"
	toml "github.com/pelletier/go-toml/v2"
)

func TestApplyClaudePreservesUnknownFields(t *testing.T) {
	home := t.TempDir()
	settingsPath := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	initial := `{
  "custom": true,
  "permissions": {
    "deny": ["Existing"]
  }
}`
	if err := os.WriteFile(settingsPath, []byte(initial), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := Apply(home, config.Profile{
		Command:   "claude",
		BaseURL:   "https://api.anthropic.com",
		APIKey:    "token",
		Model:     "claude-main",
		FastModel: "claude-fast",
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if doc["custom"] != true {
		t.Fatalf("custom field not preserved: %#v", doc["custom"])
	}
	env := doc["env"].(map[string]any)
	if env["CLAUDE_CODE_MODEL"] != "claude-main" {
		t.Fatalf("unexpected CLAUDE_CODE_MODEL: %#v", env["CLAUDE_CODE_MODEL"])
	}
	permissions := doc["permissions"].(map[string]any)
	deny := permissions["deny"].([]any)
	if len(deny) != 1 || deny[0] != "Existing" {
		t.Fatalf("unexpected deny permissions: %#v", deny)
	}
}

func TestApplyClaudeAppendsConfiguredDenyPermissions(t *testing.T) {
	home := t.TempDir()
	settingsPath := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(settingsPath, []byte(`{"permissions":{"deny":["Existing"]}}`), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := Apply(home, config.Profile{
		Command:             "claude",
		BaseURL:             "https://api.anthropic.com",
		APIKey:              "token",
		Model:               "claude-main",
		FastModel:           "claude-fast",
		SyncDenyPermissions: []string{"Agent(Explore)", "Bash(rm -rf)"},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	permissions := doc["permissions"].(map[string]any)
	deny := permissions["deny"].([]any)
	want := []any{"Existing", "Agent(Explore)", "Bash(rm -rf)"}
	if len(deny) != len(want) {
		t.Fatalf("deny len = %d, want %d (%#v)", len(deny), len(want), deny)
	}
	for i := range want {
		if deny[i] != want[i] {
			t.Fatalf("deny[%d] = %#v, want %#v", i, deny[i], want[i])
		}
	}
}

func TestApplyClaudeRemovesStaleSubagentModelWhenUnset(t *testing.T) {
	home := t.TempDir()
	settingsPath := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	initial := `{
  "env": {
    "CLAUDE_CODE_SUBAGENT_MODEL": "glm-5",
    "CLAUDE_CODE_MODEL": "glm-5"
  },
  "model": "glm-5"
}`
	if err := os.WriteFile(settingsPath, []byte(initial), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := Apply(home, config.Profile{
		Command:   "claude",
		BaseURL:   "https://api.anthropic.com",
		APIKey:    "token",
		Model:     "claude-main",
		FastModel: "claude-fast",
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	env := doc["env"].(map[string]any)
	if _, ok := env["CLAUDE_CODE_SUBAGENT_MODEL"]; ok {
		t.Fatalf("stale CLAUDE_CODE_SUBAGENT_MODEL should be removed: %#v", env)
	}
	if env["CLAUDE_CODE_MODEL"] != "claude-main" {
		t.Fatalf("unexpected CLAUDE_CODE_MODEL: %#v", env["CLAUDE_CODE_MODEL"])
	}
	if doc["model"] != "claude-main" {
		t.Fatalf("model = %#v, want claude-main", doc["model"])
	}
}

func TestApplyClaudeSupportsDeepSeekAnthropicEndpoint(t *testing.T) {
	home := t.TempDir()

	_, err := Apply(home, config.Profile{
		Provider:  "deepseek",
		Command:   "claude",
		BaseURL:   "https://api.deepseek.com/anthropic",
		APIKey:    "deepseek-key",
		Model:     "deepseek-v4-pro",
		FastModel: "deepseek-v4-flash",
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	env := doc["env"].(map[string]any)
	if env["ANTHROPIC_MODEL"] != "deepseek-v4-pro" {
		t.Fatalf("ANTHROPIC_MODEL = %#v, want deepseek-v4-pro", env["ANTHROPIC_MODEL"])
	}
	if env["ANTHROPIC_SMALL_FAST_MODEL"] != "deepseek-v4-flash" {
		t.Fatalf("ANTHROPIC_SMALL_FAST_MODEL = %#v, want deepseek-v4-flash", env["ANTHROPIC_SMALL_FAST_MODEL"])
	}
	if env["CLAUDE_CODE_MODEL"] != "deepseek-v4-pro" {
		t.Fatalf("CLAUDE_CODE_MODEL = %#v, want deepseek-v4-pro", env["CLAUDE_CODE_MODEL"])
	}
	if doc["model"] != "deepseek-v4-pro" {
		t.Fatalf("model = %#v, want deepseek-v4-pro", doc["model"])
	}
}

func TestApplyCodexPreservesUnknownFields(t *testing.T) {
	home := t.TempDir()
	configPath := filepath.Join(home, ".codex", "config.toml")
	authPath := filepath.Join(home, ".codex", "auth.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	initialConfig := "custom_value = \"keep\"\n\n[model_providers.other]\nname = \"other\"\n"
	if err := os.WriteFile(configPath, []byte(initialConfig), 0o600); err != nil {
		t.Fatalf("WriteFile config: %v", err)
	}
	if err := os.WriteFile(authPath, []byte(`{"extra":"keep"}`), 0o600); err != nil {
		t.Fatalf("WriteFile auth: %v", err)
	}

	_, err := Apply(home, config.Profile{
		Command: "codex",
		BaseURL: "https://relay.example.com",
		APIKey:  "sk-test",
		Model:   "gpt-5.4",
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile config: %v", err)
	}
	var doc map[string]any
	if err := toml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("Unmarshal TOML: %v", err)
	}
	if doc["custom_value"] != "keep" {
		t.Fatalf("custom_value not preserved: %#v", doc["custom_value"])
	}
	modelProviders := doc["model_providers"].(map[string]any)
	codexProvider := modelProviders["codex"].(map[string]any)
	if codexProvider["base_url"] != "https://relay.example.com/v1" {
		t.Fatalf("unexpected codex base_url: %#v", codexProvider["base_url"])
	}

	authData, err := os.ReadFile(authPath)
	if err != nil {
		t.Fatalf("ReadFile auth: %v", err)
	}
	var authDoc map[string]any
	if err := json.Unmarshal(authData, &authDoc); err != nil {
		t.Fatalf("Unmarshal auth: %v", err)
	}
	if authDoc["extra"] != "keep" {
		t.Fatalf("auth extra field not preserved: %#v", authDoc["extra"])
	}
	if authDoc["OPENAI_API_KEY"] != "sk-test" {
		t.Fatalf("unexpected api key: %#v", authDoc["OPENAI_API_KEY"])
	}
}

func TestTargetPaths(t *testing.T) {
	home := "/tmp/home"
	paths := TargetPaths(home, config.Profile{Command: "codex"})
	if len(paths) != 2 {
		t.Fatalf("len(paths) = %d, want 2", len(paths))
	}
}

func TestApplyUsesAtomicWritesWithoutLeavingTempFiles(t *testing.T) {
	home := t.TempDir()

	_, err := Apply(home, config.Profile{
		Command: "codex",
		BaseURL: "https://relay.example.com",
		APIKey:  "sk-test",
		Model:   "gpt-5.4",
	})
	if err != nil {
		t.Fatalf("Apply codex: %v", err)
	}

	for _, path := range []string{
		filepath.Join(home, ".codex", "config.toml"),
		filepath.Join(home, ".codex", "auth.json"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
		if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
			t.Fatalf("expected no temp file for %s, got err=%v", path, err)
		}
	}
}
