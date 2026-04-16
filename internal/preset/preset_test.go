package preset

import (
	"testing"

	"github.com/LiukerSun/cc-cli/internal/config"
)

func TestApplyFillsPresetDefaults(t *testing.T) {
	profile, err := Apply(config.Profile{APIKey: "test-key"}, "openai")
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if profile.Name != "Codex OpenAI" {
		t.Fatalf("Name = %q", profile.Name)
	}
	if profile.Provider != "openai" {
		t.Fatalf("Provider = %q", profile.Provider)
	}
	if profile.Command != "codex" {
		t.Fatalf("Command = %q", profile.Command)
	}
	if profile.BaseURL != "https://api.openai.com/v1" {
		t.Fatalf("BaseURL = %q", profile.BaseURL)
	}
	if profile.Model != "gpt-5.4" {
		t.Fatalf("Model = %q", profile.Model)
	}
	if profile.FastModel != "gpt-5.4-mini" {
		t.Fatalf("FastModel = %q", profile.FastModel)
	}
}

func TestApplyKeepsExplicitOverrides(t *testing.T) {
	profile, err := Apply(config.Profile{
		Name:      "Custom OpenAI",
		Provider:  "custom",
		Command:   "codex",
		BaseURL:   "https://relay.example.com/v1",
		Model:     "gpt-5.5-preview",
		FastModel: "gpt-5.5-mini",
		APIKey:    "test-key",
	}, "openai")
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	if profile.Name != "Custom OpenAI" {
		t.Fatalf("Name = %q", profile.Name)
	}
	if profile.Provider != "custom" {
		t.Fatalf("Provider = %q", profile.Provider)
	}
	if profile.BaseURL != "https://relay.example.com/v1" {
		t.Fatalf("BaseURL = %q", profile.BaseURL)
	}
	if profile.Model != "gpt-5.5-preview" {
		t.Fatalf("Model = %q", profile.Model)
	}
	if profile.FastModel != "gpt-5.5-mini" {
		t.Fatalf("FastModel = %q", profile.FastModel)
	}
}

func TestApplyRejectsUnknownPreset(t *testing.T) {
	if _, err := Apply(config.Profile{}, "unknown"); err == nil {
		t.Fatal("expected unknown preset error")
	}
}

func TestLookupReturnsDefinition(t *testing.T) {
	definition, err := Lookup("zhipu")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if definition.Command != "claude" {
		t.Fatalf("Command = %q", definition.Command)
	}
	if definition.BaseURL != "https://open.bigmodel.cn/api/anthropic" {
		t.Fatalf("BaseURL = %q", definition.BaseURL)
	}
}

func TestLookupSupportsAliases(t *testing.T) {
	definition, err := Lookup("qwen")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if definition.Provider != "alibaba" {
		t.Fatalf("Provider = %q", definition.Provider)
	}

	definition, err = Lookup("codex")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if definition.Command != "codex" {
		t.Fatalf("Command = %q", definition.Command)
	}

	definition, err = Lookup("moonshot")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if definition.Provider != "kimi" {
		t.Fatalf("Provider = %q", definition.Provider)
	}
}
