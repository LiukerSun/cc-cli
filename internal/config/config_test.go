package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LiukerSun/cc-cli/internal/platform"
)

func TestLoadDefaultWhenNoConfigExists(t *testing.T) {
	home := t.TempDir()
	layout, err := platform.ResolveLayout("linux", home, func(string) string { return "" })
	if err != nil {
		t.Fatalf("ResolveLayout: %v", err)
	}

	store := NewStore(home, layout)
	cfg, meta, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if meta.Source != "default" {
		t.Fatalf("meta.Source = %q, want default", meta.Source)
	}
	if len(cfg.Profiles) != 0 {
		t.Fatalf("len(cfg.Profiles) = %d, want 0", len(cfg.Profiles))
	}
}

func TestLoadLegacyConfig(t *testing.T) {
	home := t.TempDir()
	layout, err := platform.ResolveLayout("linux", home, func(string) string { return "" })
	if err != nil {
		t.Fatalf("ResolveLayout: %v", err)
	}

	legacyPath := filepath.Join(home, ".ccc", "config.json")
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	content := `[
  {
    "name": "Claude (Official)",
    "env": {
      "ANTHROPIC_BASE_URL": "https://api.anthropic.com",
      "ANTHROPIC_AUTH_TOKEN": "token",
      "ANTHROPIC_MODEL": "claude-main",
      "ANTHROPIC_SMALL_FAST_MODEL": "claude-fast"
    }
  }
]`
	if err := os.WriteFile(legacyPath, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	store := NewStore(home, layout)
	cfg, meta, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if meta.Source != "legacy-root" {
		t.Fatalf("meta.Source = %q, want legacy-root", meta.Source)
	}
	if len(cfg.Profiles) != 1 {
		t.Fatalf("len(cfg.Profiles) = %d, want 1", len(cfg.Profiles))
	}
	profile := cfg.Profiles[0]
	if profile.Command != "claude" {
		t.Fatalf("profile.Command = %q, want claude", profile.Command)
	}
	if profile.Model != "claude-main" {
		t.Fatalf("profile.Model = %q, want claude-main", profile.Model)
	}
}

func TestSaveWritesCurrentSchema(t *testing.T) {
	home := t.TempDir()
	layout, err := platform.ResolveLayout("linux", home, func(string) string { return "" })
	if err != nil {
		t.Fatalf("ResolveLayout: %v", err)
	}

	store := NewStore(home, layout)
	cfg := DefaultFile()
	if err := cfg.UpsertProfile(Profile{
		Name:         "Codex Relay",
		Command:      "codex",
		BaseURL:      "https://relay.example.com/v1",
		APIKey:       "sk-test",
		Model:        "gpt-5.4",
		Provider:     "custom",
		SyncExternal: true,
	}); err != nil {
		t.Fatalf("UpsertProfile: %v", err)
	}

	if err := store.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(layout.ConfigFile())
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if len(data) == 0 || data[0] != '{' {
		t.Fatalf("config not written as object schema: %q", string(data))
	}
}

func TestEnsureUniqueProfileID(t *testing.T) {
	cfg := DefaultFile()
	if err := cfg.UpsertProfile(Profile{
		ID:           "demo",
		Name:         "Demo",
		Command:      "claude",
		BaseURL:      "https://api.anthropic.com",
		APIKey:       "token",
		Model:        "main",
		SyncExternal: true,
	}); err != nil {
		t.Fatalf("UpsertProfile: %v", err)
	}

	profile := cfg.EnsureUniqueProfileID(Profile{
		ID:           "demo",
		Name:         "Demo Two",
		Command:      "claude",
		BaseURL:      "https://api.anthropic.com",
		APIKey:       "token",
		Model:        "main",
		SyncExternal: true,
	})
	if profile.ID != "demo-2" {
		t.Fatalf("profile.ID = %q, want demo-2", profile.ID)
	}
}

func TestReplaceProfileUpdatesCurrentProfileID(t *testing.T) {
	cfg := DefaultFile()
	cfg.CurrentProfile = "demo"
	if err := cfg.UpsertProfile(Profile{
		ID:           "demo",
		Name:         "Demo",
		Command:      "claude",
		BaseURL:      "https://api.anthropic.com",
		APIKey:       "token",
		Model:        "main",
		SyncExternal: true,
	}); err != nil {
		t.Fatalf("UpsertProfile: %v", err)
	}

	_, ok, err := cfg.ReplaceProfile("demo", Profile{
		ID:           "demo-updated",
		Name:         "Demo Updated",
		Command:      "claude",
		BaseURL:      "https://api.anthropic.com",
		APIKey:       "token-2",
		Model:        "main-2",
		SyncExternal: false,
	})
	if err != nil {
		t.Fatalf("ReplaceProfile: %v", err)
	}
	if !ok {
		t.Fatal("expected profile to be replaced")
	}
	if cfg.CurrentProfile != "demo-updated" {
		t.Fatalf("CurrentProfile = %q, want demo-updated", cfg.CurrentProfile)
	}
	profile, found := cfg.FindProfile("demo-updated")
	if !found {
		t.Fatal("expected updated profile to be present")
	}
	if profile.Name != "Demo Updated" {
		t.Fatalf("profile.Name = %q", profile.Name)
	}
}
