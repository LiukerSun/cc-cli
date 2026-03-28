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
