package runner

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/LiukerSun/cc-cli/internal/config"
	"github.com/LiukerSun/cc-cli/internal/util"
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
		applyBypass(profile.Command, env, &args)
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
	cmd.Env = commandEnv(plan)
	return cmd.Run()
}

func EnvList(plan Plan) []string {
	return envList(plan.Env)
}

func resolveProfile(cfg config.File, identifier string) (config.Profile, error) {
	if identifier != "" {
		profile, ok := cfg.FindProfile(identifier)
		if !ok {
			return config.Profile{}, fmt.Errorf("profile %q not found. Run 'ccc profile list' to see available profiles", identifier)
		}
		return profile, nil
	}

	if cfg.CurrentProfile != "" {
		profile, ok := cfg.FindProfile(cfg.CurrentProfile)
		if !ok {
			return config.Profile{}, fmt.Errorf("current profile %q not found. Run 'ccc profile list' to see available profiles", cfg.CurrentProfile)
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
		if !profile.SyncExternal {
			env["OPENAI_BASE_URL"] = profile.BaseURL
		}
		env["OPENAI_API_KEY"] = profile.APIKey
		env["OPENAI_MODEL"] = profile.Model
		env["OPENAI_SMALL_FAST_MODEL"] = util.FirstNonEmpty(profile.FastModel, profile.Model)
	default:
		env["ANTHROPIC_BASE_URL"] = profile.BaseURL
		if usesAnthropicAuthToken(profile.Provider) {
			env["ANTHROPIC_AUTH_TOKEN"] = profile.APIKey
		} else {
			env["ANTHROPIC_API_KEY"] = profile.APIKey
		}
		env["ANTHROPIC_MODEL"] = profile.Model
		fastModel := util.FirstNonEmpty(profile.FastModel, profile.Model)
		env["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = fastModel
		env["ANTHROPIC_SMALL_FAST_MODEL"] = fastModel
		env["CLAUDE_CODE_MODEL"] = profile.Model
		env["CLAUDE_CODE_SMALL_MODEL"] = fastModel
		if profile.SubagentModel != "" {
			env["CLAUDE_CODE_SUBAGENT_MODEL"] = profile.SubagentModel
		}
	}

	return env
}

func usesAnthropicAuthToken(provider string) bool {
	switch strings.TrimSpace(strings.ToLower(provider)) {
	case "kimi", "deepseek":
		return true
	default:
		return false
	}
}

func bypassFlag(command string) string {
	if command == "codex" {
		return "--yolo"
	}
	return "--dangerously-skip-permissions"
}

func applyBypass(command string, env map[string]string, args *[]string) {
	if command == "codex" {
		*args = append([]string{bypassFlag(command)}, (*args)...)
		return
	}

	env["CLAUDE_SKIP_PERMISSIONS"] = "1"
	env["IS_SANDBOX"] = "1"
	*args = append([]string{bypassFlag(command)}, (*args)...)
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

func commandEnv(plan Plan) []string {
	blocked := map[string]struct{}{}
	for _, key := range managedEnvKeys(plan.Command) {
		blocked[key] = struct{}{}
	}
	for key := range plan.Env {
		blocked[key] = struct{}{}
	}

	base := make([]string, 0, len(os.Environ()))
	for _, entry := range os.Environ() {
		key := entry
		if idx := strings.IndexByte(entry, '='); idx >= 0 {
			key = entry[:idx]
		}
		if _, skip := blocked[key]; skip {
			continue
		}
		base = append(base, entry)
	}

	return append(base, envList(plan.Env)...)
}

func managedEnvKeys(command string) []string {
	switch command {
	case "codex":
		return []string{
			"OPENAI_BASE_URL",
			"OPENAI_API_KEY",
			"OPENAI_MODEL",
			"OPENAI_SMALL_FAST_MODEL",
		}
	default:
		return []string{
			"ANTHROPIC_BASE_URL",
			"ANTHROPIC_AUTH_TOKEN",
			"ANTHROPIC_API_KEY",
			"ANTHROPIC_MODEL",
			"ANTHROPIC_DEFAULT_HAIKU_MODEL",
			"ANTHROPIC_SMALL_FAST_MODEL",
			"CLAUDE_CODE_MODEL",
			"CLAUDE_CODE_SMALL_MODEL",
			"CLAUDE_CODE_SUBAGENT_MODEL",
			"CLAUDE_SKIP_PERMISSIONS",
			"IS_SANDBOX",
		}
	}
}
