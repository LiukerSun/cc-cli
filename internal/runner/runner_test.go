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

func TestEnvListStableOrder(t *testing.T) {
	env := envList(map[string]string{"B": "2", "A": "1"})
	got := strings.Join(env, ",")
	if got != "A=1,B=2" {
		t.Fatalf("envList order = %q", got)
	}
}
