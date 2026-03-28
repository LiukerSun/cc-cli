package runner

import (
	"fmt"
	"os"
	"os/exec"
	"sort"

	"github.com/LiukerSun/cc-cli/internal/config"
)

type Plan struct {
	Profile config.Profile
	Command string
	Args    []string
	Env     map[string]string
}

func BuildPlan(cfg config.File, identifier string, cliArgs []string, bypass bool) (Plan, error) {
	profile, err := resolveProfile(cfg, identifier)
	if err != nil {
		return Plan{}, err
	}

	env := profileEnv(profile)
	args := append([]string{}, cliArgs...)
	if bypass {
		args = append([]string{bypassFlag(profile.Command)}, args...)
	}

	return Plan{
		Profile: profile,
		Command: profile.Command,
		Args:    args,
		Env:     env,
	}, nil
}

func Run(plan Plan, stdout, stderr interface {
	Write([]byte) (int, error)
}) error {
	cmd := exec.Command(plan.Command, plan.Args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = append(os.Environ(), envList(plan.Env)...)
	return cmd.Run()
}

func EnvList(plan Plan) []string {
	return envList(plan.Env)
}

func resolveProfile(cfg config.File, identifier string) (config.Profile, error) {
	if identifier != "" {
		profile, ok := cfg.FindProfile(identifier)
		if !ok {
			return config.Profile{}, fmt.Errorf("profile %q not found", identifier)
		}
		return profile, nil
	}

	if cfg.CurrentProfile != "" {
		profile, ok := cfg.FindProfile(cfg.CurrentProfile)
		if !ok {
			return config.Profile{}, fmt.Errorf("current profile %q not found", cfg.CurrentProfile)
		}
		return profile, nil
	}

	if len(cfg.Profiles) == 1 {
		return cfg.Profiles[0], nil
	}
	if len(cfg.Profiles) == 0 {
		return config.Profile{}, fmt.Errorf("no profiles configured")
	}
	return config.Profile{}, fmt.Errorf("no current profile selected; use 'ccc profile use <id>' or pass a profile to 'ccc run'")
}

func profileEnv(profile config.Profile) map[string]string {
	env := map[string]string{}

	for key, value := range profile.ExtraEnv {
		env[key] = value
	}

	switch profile.Command {
	case "codex":
		env["OPENAI_BASE_URL"] = profile.BaseURL
		env["OPENAI_API_KEY"] = profile.APIKey
		env["OPENAI_MODEL"] = profile.Model
		env["OPENAI_SMALL_FAST_MODEL"] = firstNonEmpty(profile.FastModel, profile.Model)
	default:
		env["ANTHROPIC_BASE_URL"] = profile.BaseURL
		env["ANTHROPIC_AUTH_TOKEN"] = profile.APIKey
		env["ANTHROPIC_MODEL"] = profile.Model
		env["ANTHROPIC_SMALL_FAST_MODEL"] = firstNonEmpty(profile.FastModel, profile.Model)
		env["CLAUDE_CODE_MODEL"] = profile.Model
		env["CLAUDE_CODE_SMALL_MODEL"] = firstNonEmpty(profile.FastModel, profile.Model)
		env["CLAUDE_CODE_SUBAGENT_MODEL"] = profile.Model
	}

	return env
}

func bypassFlag(command string) string {
	if command == "codex" {
		return "--dangerously-bypass-approvals-and-sandbox"
	}
	return "--dangerously-skip-permissions"
}

func envList(env map[string]string) []string {
	keys := make([]string, 0, len(env))
	for key := range env {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := make([]string, 0, len(keys))
	for _, key := range keys {
		out = append(out, fmt.Sprintf("%s=%s", key, env[key]))
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
