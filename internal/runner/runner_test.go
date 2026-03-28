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
	if len(plan.Args) == 0 || plan.Args[0] != "--dangerously-bypass-approvals-and-sandbox" {
		t.Fatalf("unexpected args: %#v", plan.Args)
	}
	if got := plan.Env["OPENAI_MODEL"]; got != "gpt-5.4" {
		t.Fatalf("OPENAI_MODEL = %q, want gpt-5.4", got)
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
}

func TestEnvListStableOrder(t *testing.T) {
	env := envList(map[string]string{"B": "2", "A": "1"})
	got := strings.Join(env, ",")
	if got != "A=1,B=2" {
		t.Fatalf("envList order = %q", got)
	}
}
