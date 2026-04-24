package runner

import (
	"strings"
	"testing"

	"github.com/LiukerSun/cc-cli/internal/config"
)

func TestBuildPlanUsesCurrentProfile(t *testing.T) {
	cfg := config.File{
		Version:        1,
		CurrentProfile: "codex-relay",
		Profiles: []config.Profile{
			{
				ID:           "codex-relay",
				Name:         "Codex Relay",
				Command:      "codex",
				BaseURL:      "https://relay.example.com/v1",
				APIKey:       "sk-test",
				Model:        "gpt-5.4",
				SyncExternal: true,
			},
		},
	}

	plan, err := BuildPlan(cfg, "", []string{"--help"}, true)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if plan.Command != "codex" {
		t.Fatalf("plan.Command = %q, want codex", plan.Command)
	}
	if len(plan.Args) == 0 || plan.Args[0] != "--yolo" {
		t.Fatalf("unexpected args: %#v", plan.Args)
	}
	if got := plan.Env["OPENAI_MODEL"]; got != "gpt-5.4" {
		t.Fatalf("OPENAI_MODEL = %q, want gpt-5.4", got)
	}
	if _, ok := plan.Env["OPENAI_BASE_URL"]; ok {
		t.Fatalf("OPENAI_BASE_URL should not be set when external sync is enabled")
	}
}

func TestBuildPlanKeepsCodexBaseURLWithoutExternalSync(t *testing.T) {
	cfg := config.File{
		Version:        1,
		CurrentProfile: "codex-relay",
		Profiles: []config.Profile{
			{
				ID:           "codex-relay",
				Name:         "Codex Relay",
				Command:      "codex",
				BaseURL:      "https://relay.example.com/v1",
				APIKey:       "sk-test",
				Model:        "gpt-5.4",
				SyncExternal: false,
			},
		},
	}

	plan, err := BuildPlan(cfg, "", nil, false)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if got := plan.Env["OPENAI_BASE_URL"]; got != "https://relay.example.com/v1" {
		t.Fatalf("OPENAI_BASE_URL = %q, want https://relay.example.com/v1", got)
	}
}

func TestBuildPlanSetsClaudeBypassEnv(t *testing.T) {
	cfg := config.File{
		Version:        1,
		CurrentProfile: "zhipu-main",
		Profiles: []config.Profile{
			{
				ID:           "zhipu-main",
				Name:         "Zhipu Main",
				Command:      "claude",
				BaseURL:      "https://open.bigmodel.cn/api/anthropic",
				APIKey:       "test-key",
				Model:        "glm-5",
				FastModel:    "glm-4.7",
				SyncExternal: true,
			},
		},
	}

	plan, err := BuildPlan(cfg, "", nil, true)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if len(plan.Args) == 0 || plan.Args[0] != "--dangerously-skip-permissions" {
		t.Fatalf("unexpected args: %#v", plan.Args)
	}
	if plan.Env["CLAUDE_SKIP_PERMISSIONS"] != "1" {
		t.Fatalf("CLAUDE_SKIP_PERMISSIONS = %q", plan.Env["CLAUDE_SKIP_PERMISSIONS"])
	}
	if plan.Env["IS_SANDBOX"] != "1" {
		t.Fatalf("IS_SANDBOX = %q", plan.Env["IS_SANDBOX"])
	}
	if plan.Env["ANTHROPIC_API_KEY"] != "test-key" {
		t.Fatalf("ANTHROPIC_API_KEY = %q, want test-key", plan.Env["ANTHROPIC_API_KEY"])
	}
	if _, ok := plan.Env["ANTHROPIC_AUTH_TOKEN"]; ok {
		t.Fatalf("ANTHROPIC_AUTH_TOKEN should not be set for anthropic profiles")
	}
}

func TestBuildPlanOmitsAnthropicAPIKeyForKimi(t *testing.T) {
	cfg := config.File{
		Version:        1,
		CurrentProfile: "kimi-main",
		Profiles: []config.Profile{
			{
				ID:           "kimi-main",
				Name:         "Kimi Main",
				Provider:     "kimi",
				Command:      "claude",
				BaseURL:      "https://api.kimi.com/coding/",
				APIKey:       "kimi-key",
				Model:        "K2.6-code-preview",
				FastModel:    "K2.6-code-preview",
				SyncExternal: true,
			},
		},
	}

	plan, err := BuildPlan(cfg, "", nil, false)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if plan.Env["ANTHROPIC_AUTH_TOKEN"] != "kimi-key" {
		t.Fatalf("ANTHROPIC_AUTH_TOKEN = %q, want kimi-key", plan.Env["ANTHROPIC_AUTH_TOKEN"])
	}
	if _, ok := plan.Env["ANTHROPIC_API_KEY"]; ok {
		t.Fatalf("ANTHROPIC_API_KEY should not be set for kimi profiles")
	}
}

func TestBuildPlanUsesAnthropicEnvForDeepSeek(t *testing.T) {
	cfg := config.File{
		Version:        1,
		CurrentProfile: "deepseek-main",
		Profiles: []config.Profile{
			{
				ID:           "deepseek-main",
				Name:         "DeepSeek Main",
				Provider:     "deepseek",
				Command:      "claude",
				BaseURL:      "https://api.deepseek.com/anthropic",
				APIKey:       "deepseek-key",
				Model:        "deepseek-v4-pro",
				FastModel:    "deepseek-v4-flash",
				SyncExternal: true,
			},
		},
	}

	plan, err := BuildPlan(cfg, "", nil, false)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	if plan.Command != "claude" {
		t.Fatalf("plan.Command = %q, want claude", plan.Command)
	}
	if plan.Env["ANTHROPIC_BASE_URL"] != "https://api.deepseek.com/anthropic" {
		t.Fatalf("ANTHROPIC_BASE_URL = %q, want https://api.deepseek.com/anthropic", plan.Env["ANTHROPIC_BASE_URL"])
	}
	if plan.Env["ANTHROPIC_AUTH_TOKEN"] != "deepseek-key" {
		t.Fatalf("ANTHROPIC_AUTH_TOKEN = %q, want deepseek-key", plan.Env["ANTHROPIC_AUTH_TOKEN"])
	}
	if _, ok := plan.Env["ANTHROPIC_API_KEY"]; ok {
		t.Fatalf("ANTHROPIC_API_KEY should not be set for deepseek profiles")
	}
	if plan.Env["ANTHROPIC_MODEL"] != "deepseek-v4-pro" {
		t.Fatalf("ANTHROPIC_MODEL = %q, want deepseek-v4-pro", plan.Env["ANTHROPIC_MODEL"])
	}
	if plan.Env["ANTHROPIC_SMALL_FAST_MODEL"] != "deepseek-v4-flash" {
		t.Fatalf("ANTHROPIC_SMALL_FAST_MODEL = %q, want deepseek-v4-flash", plan.Env["ANTHROPIC_SMALL_FAST_MODEL"])
	}
}

func TestEnvListStableOrder(t *testing.T) {
	env := envList(map[string]string{"B": "2", "A": "1"})
	got := strings.Join(env, ",")
	if got != "A=1,B=2" {
		t.Fatalf("envList order = %q", got)
	}
}

func TestCommandEnvRemovesStaleManagedClaudeVars(t *testing.T) {
	t.Setenv("ANTHROPIC_MODEL", "stale-model")
	t.Setenv("ANTHROPIC_API_KEY", "stale-key")
	t.Setenv("CLAUDE_CODE_SUBAGENT_MODEL", "glm-5")

	env := commandEnv(Plan{
		Command: "claude",
		Env: map[string]string{
			"ANTHROPIC_MODEL":   "claude-main",
			"ANTHROPIC_API_KEY": "claude-key",
			"CLAUDE_CODE_MODEL": "claude-main",
		},
	})

	values := map[string]string{}
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			t.Fatalf("invalid env entry: %q", entry)
		}
		values[key] = value
	}

	if values["ANTHROPIC_MODEL"] != "claude-main" {
		t.Fatalf("ANTHROPIC_MODEL = %q, want claude-main", values["ANTHROPIC_MODEL"])
	}
	if values["ANTHROPIC_API_KEY"] != "claude-key" {
		t.Fatalf("ANTHROPIC_API_KEY = %q, want claude-key", values["ANTHROPIC_API_KEY"])
	}
	if _, ok := values["CLAUDE_CODE_SUBAGENT_MODEL"]; ok {
		t.Fatalf("CLAUDE_CODE_SUBAGENT_MODEL should be cleared when not set in the plan")
	}
}

func TestCommandEnvClearsStaleAnthropicAPIKeyWhenPlanOmitsIt(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "stale-key")

	env := commandEnv(Plan{
		Command: "claude",
		Env: map[string]string{
			"ANTHROPIC_AUTH_TOKEN": "kimi-key",
			"ANTHROPIC_MODEL":      "K2.6-code-preview",
			"CLAUDE_CODE_MODEL":    "K2.6-code-preview",
		},
	})

	values := map[string]string{}
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			t.Fatalf("invalid env entry: %q", entry)
		}
		values[key] = value
	}

	if values["ANTHROPIC_AUTH_TOKEN"] != "kimi-key" {
		t.Fatalf("ANTHROPIC_AUTH_TOKEN = %q, want kimi-key", values["ANTHROPIC_AUTH_TOKEN"])
	}
	if _, ok := values["ANTHROPIC_API_KEY"]; ok {
		t.Fatalf("ANTHROPIC_API_KEY should be removed when omitted from the plan")
	}
}

func TestCommandEnvClearsStaleAnthropicAuthTokenWhenPlanOmitsIt(t *testing.T) {
	t.Setenv("ANTHROPIC_AUTH_TOKEN", "stale-token")

	env := commandEnv(Plan{
		Command: "claude",
		Env: map[string]string{
			"ANTHROPIC_API_KEY": "anthropic-key",
			"ANTHROPIC_MODEL":   "claude-main",
			"CLAUDE_CODE_MODEL": "claude-main",
		},
	})

	values := map[string]string{}
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			t.Fatalf("invalid env entry: %q", entry)
		}
		values[key] = value
	}

	if values["ANTHROPIC_API_KEY"] != "anthropic-key" {
		t.Fatalf("ANTHROPIC_API_KEY = %q, want anthropic-key", values["ANTHROPIC_API_KEY"])
	}
	if _, ok := values["ANTHROPIC_AUTH_TOKEN"]; ok {
		t.Fatalf("ANTHROPIC_AUTH_TOKEN should be removed when omitted from the plan")
	}
}
