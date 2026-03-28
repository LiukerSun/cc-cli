package app

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/LiukerSun/cc-cli/internal/buildinfo"
	"github.com/LiukerSun/cc-cli/internal/config"
	"github.com/LiukerSun/cc-cli/internal/deps"
	"github.com/LiukerSun/cc-cli/internal/legacy"
	"github.com/LiukerSun/cc-cli/internal/platform"
	"github.com/LiukerSun/cc-cli/internal/preset"
	"github.com/LiukerSun/cc-cli/internal/runner"
	cfgsync "github.com/LiukerSun/cc-cli/internal/sync"
	"github.com/LiukerSun/cc-cli/internal/upgrade"
)

type pathsReport struct {
	Layout     platform.Layout  `json:"layout"`
	ConfigFile string           `json:"config_file"`
	Legacy     legacy.Detection `json:"legacy"`
}

func Run(args []string, stdout, stderr io.Writer) int {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "failed to resolve home directory: %v\n", err)
		return 1
	}

	layout, err := platform.ResolveLayout(runtime.GOOS, home, os.Getenv)
	if err != nil {
		fmt.Fprintf(stderr, "failed to resolve directory layout: %v\n", err)
		return 1
	}

	report := pathsReport{
		Layout:     layout,
		ConfigFile: layout.ConfigFile(),
		Legacy:     legacy.Detect(home),
	}

	if len(args) == 0 {
		printHelp(stdout)
		return 0
	}

	switch args[0] {
	case "-h", "--help", "help":
		printHelp(stdout)
		return 0
	case "-V", "--version", "version":
		fmt.Fprintf(stdout, "ccc version %s\n", buildinfo.Version)
		return 0
	case "current":
		return runCurrent(stdout, stderr, home, layout)
	case "run":
		return runRun(stdout, stderr, home, layout, args[1:])
	case "sync":
		return runSync(stdout, stderr, home, layout, args[1:])
	case "profile":
		return runProfile(stdout, stderr, home, layout, args[1:])
	case "paths":
		return runPaths(stdout, stderr, report, args[1:])
	case "doctor":
		return runDoctor(stdout, stderr, home, layout, report)
	case "upgrade":
		return runUpgrade(stdout, stderr, args[1:])
	case "config":
		return runConfig(stdout, stderr, home, layout, report, args[1:])
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", args[0])
		printHelp(stderr)
		return 1
	}
}

func runPaths(stdout, stderr io.Writer, report pathsReport, args []string) int {
	if len(args) == 0 {
		printPaths(stdout, report)
		return 0
	}

	if len(args) == 1 && args[0] == "--json" {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintf(stderr, "failed to encode paths report: %v\n", err)
			return 1
		}
		return 0
	}

	fmt.Fprintln(stderr, "usage: ccc paths [--json]")
	return 1
}

func runConfig(stdout, stderr io.Writer, home string, layout platform.Layout, report pathsReport, args []string) int {
	if len(args) == 1 && args[0] == "path" {
		fmt.Fprintln(stdout, report.ConfigFile)
		return 0
	}
	if len(args) == 1 && args[0] == "show" {
		return runConfigShow(stdout, stderr, home, layout)
	}
	if len(args) == 1 && args[0] == "migrate" {
		return runConfigMigrate(stdout, stderr, home, layout)
	}

	fmt.Fprintln(stderr, "usage: ccc config path")
	fmt.Fprintln(stderr, "       ccc config show")
	fmt.Fprintln(stderr, "       ccc config migrate")
	return 1
}

func runConfigShow(stdout, stderr io.Writer, home string, layout platform.Layout) int {
	store := config.NewStore(home, layout)
	cfg, meta, err := store.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}

	payload := struct {
		Metadata config.Metadata `json:"metadata"`
		Config   config.File     `json:"config"`
	}{
		Metadata: meta,
		Config:   cfg,
	}

	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(payload); err != nil {
		fmt.Fprintf(stderr, "failed to encode config: %v\n", err)
		return 1
	}
	return 0
}

func runConfigMigrate(stdout, stderr io.Writer, home string, layout platform.Layout) int {
	store := config.NewStore(home, layout)
	cfg, meta, err := store.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}

	if meta.Source == "current" {
		fmt.Fprintf(stdout, "Config already uses current schema at %s\n", meta.Path)
		return 0
	}

	if err := store.Save(cfg); err != nil {
		fmt.Fprintf(stderr, "failed to write migrated config: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Migrated %d profile(s) from %s (%s) to %s\n", len(cfg.Profiles), meta.Source, meta.Path, layout.ConfigFile())
	return 0
}

func runDoctor(stdout, stderr io.Writer, home string, layout platform.Layout, report pathsReport) int {
	store := config.NewStore(home, layout)
	cfg, meta, err := store.Load()
	if err != nil {
		fmt.Fprintln(stdout, "ccc doctor")
		fmt.Fprintln(stdout)
		fmt.Fprintf(stdout, "Config error: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout, "ccc doctor")
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(stdout, "Home: %s\n", report.Layout.HomeDir)
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Directory layout:")
	fmt.Fprintf(stdout, "- Binary dir: %s\n", report.Layout.BinDir)
	fmt.Fprintf(stdout, "- Config dir: %s\n", report.Layout.ConfigDir)
	fmt.Fprintf(stdout, "- Config file: %s\n", report.ConfigFile)
	fmt.Fprintf(stdout, "- Data dir: %s\n", report.Layout.DataDir)
	fmt.Fprintf(stdout, "- Cache dir: %s\n", report.Layout.CacheDir)
	fmt.Fprintf(stdout, "- State dir: %s\n", report.Layout.StateDir)
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Checks:")
	fmt.Fprintf(stdout, "- Binary dir on PATH: %s\n", yesNo(pathContains(os.Getenv("PATH"), report.Layout.BinDir)))
	fmt.Fprintf(stdout, "- Config source: %s (%s)\n", meta.Source, meta.Path)
	fmt.Fprintf(stdout, "- Profiles loaded: %d\n", len(cfg.Profiles))
	fmt.Fprintf(stdout, "- Legacy assets detected: %s\n", yesNo(report.Legacy.NeedsMigration))

	nodeStatus := deps.InspectTool("node", "--version")
	npmStatus := deps.InspectTool("npm", "--version")
	claudeStatus := deps.InspectTool("claude", "--version")
	codexStatus := deps.InspectTool("codex", "--version")
	fmt.Fprintf(stdout, "- node installed: %s%s\n", yesNo(nodeStatus.Installed), versionSuffix(nodeStatus.Version))
	fmt.Fprintf(stdout, "- npm installed: %s%s\n", yesNo(npmStatus.Installed), versionSuffix(npmStatus.Version))
	fmt.Fprintf(stdout, "- claude installed: %s%s\n", yesNo(claudeStatus.Installed), pathSuffix(claudeStatus.Path))
	fmt.Fprintf(stdout, "- codex installed: %s%s\n", yesNo(codexStatus.Installed), pathSuffix(codexStatus.Path))

	warnings := []string{}
	if nodeStatus.Installed {
		if nodeStatus.Version != "" && deps.VersionLessThan(nodeStatus.Version, "18.0.0") {
			warnings = append(warnings, fmt.Sprintf("Node.js %s is below the preferred version for claude (>= 18.0.0).", nodeStatus.Version))
		}
	} else {
		warnings = append(warnings, "Node.js is not installed; automatic CLI installation is unavailable.")
	}
	if !npmStatus.Installed {
		warnings = append(warnings, "npm is not installed; automatic CLI installation is unavailable.")
	}

	if cfg.CurrentProfile != "" {
		currentProfile, ok := cfg.FindProfile(cfg.CurrentProfile)
		if ok {
			fmt.Fprintf(stdout, "- current profile: %s (%s)\n", currentProfile.Name, currentProfile.Command)
			if _, err := exec.LookPath(currentProfile.Command); err != nil {
				warnings = append(warnings, fmt.Sprintf("Current profile command %q is not available on PATH.", currentProfile.Command))
			}
		} else {
			warnings = append(warnings, fmt.Sprintf("Current profile %q is missing from config.", cfg.CurrentProfile))
		}
	} else {
		warnings = append(warnings, "No current profile selected.")
	}

	if report.Legacy.NeedsMigration {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Legacy assets:")
		for _, candidate := range report.Legacy.Candidates {
			if candidate.Exists {
				fmt.Fprintf(stdout, "- %s: %s\n", candidate.Label, candidate.Path)
			}
		}
	}

	if meta.Source != "current" && len(cfg.Profiles) > 0 {
		fmt.Fprintln(stdout)
		fmt.Fprintf(stdout, "Note: loaded %d profile(s) from %s. The next write will persist them to %s.\n", len(cfg.Profiles), meta.Source, layout.ConfigFile())
	}

	if len(warnings) > 0 {
		fmt.Fprintln(stdout)
		fmt.Fprintln(stdout, "Warnings:")
		for _, warning := range warnings {
			fmt.Fprintf(stdout, "- %s\n", warning)
		}
	}

	_ = stderr
	if len(warnings) > 0 {
		return 1
	}
	return 0
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "ccc")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Go refactor bootstrap for the next-generation cc-cli.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  ccc help")
	fmt.Fprintln(w, "  ccc version")
	fmt.Fprintln(w, "  ccc current")
	fmt.Fprintln(w, "  ccc run [profile-id-or-name] [--dry-run] [--env-only] [-y|--bypass] [-- cli-args...]")
	fmt.Fprintln(w, "  ccc sync [profile-id-or-name] [--dry-run]")
	fmt.Fprintln(w, "  ccc profile list [--json]")
	fmt.Fprintln(w, "  ccc profile add [--name ...] [--preset anthropic|openai|zhipu] --api-key ...")
	fmt.Fprintln(w, "  ccc profile update <profile-id-or-name> [--preset anthropic|openai|zhipu] [--model ...]")
	fmt.Fprintln(w, "  ccc profile use <profile-id-or-name>")
	fmt.Fprintln(w, "  ccc profile delete <profile-id-or-name>")
	fmt.Fprintln(w, "  ccc paths [--json]")
	fmt.Fprintln(w, "  ccc config path")
	fmt.Fprintln(w, "  ccc config show")
	fmt.Fprintln(w, "  ccc config migrate")
	fmt.Fprintln(w, "  ccc doctor")
	fmt.Fprintln(w, "  ccc upgrade [--version <semver>] [--dry-run]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Compatibility aliases:")
	fmt.Fprintln(w, "  ccc --help")
	fmt.Fprintln(w, "  ccc --version")
}

func printPaths(w io.Writer, report pathsReport) {
	fmt.Fprintln(w, "ccc paths")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Binary dir: %s\n", report.Layout.BinDir)
	fmt.Fprintf(w, "Config dir: %s\n", report.Layout.ConfigDir)
	fmt.Fprintf(w, "Config file: %s\n", report.ConfigFile)
	fmt.Fprintf(w, "Data dir: %s\n", report.Layout.DataDir)
	fmt.Fprintf(w, "Cache dir: %s\n", report.Layout.CacheDir)
	fmt.Fprintf(w, "State dir: %s\n", report.Layout.StateDir)
	fmt.Fprintf(w, "Legacy migration needed: %s\n", yesNo(report.Legacy.NeedsMigration))
}

func runProfile(stdout, stderr io.Writer, home string, layout platform.Layout, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: ccc profile <list|add|update|use|delete>")
		return 1
	}

	store := config.NewStore(home, layout)
	switch args[0] {
	case "list":
		return runProfileList(stdout, stderr, store, args[1:])
	case "add":
		return runProfileAdd(stdout, stderr, store, args[1:])
	case "update":
		return runProfileUpdate(stdout, stderr, store, args[1:])
	case "use":
		return runProfileUse(stdout, stderr, store, args[1:])
	case "delete":
		return runProfileDelete(stdout, stderr, store, args[1:])
	default:
		fmt.Fprintf(stderr, "unknown profile command: %s\n", args[0])
		return 1
	}
}

func runProfileList(stdout, stderr io.Writer, store config.Store, args []string) int {
	fs := flag.NewFlagSet("profile list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	jsonOutput := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(stderr, "failed to parse flags: %v\n", err)
		return 1
	}

	cfg, meta, err := store.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}

	sort.Slice(cfg.Profiles, func(i, j int) bool {
		return strings.ToLower(cfg.Profiles[i].Name) < strings.ToLower(cfg.Profiles[j].Name)
	})

	if *jsonOutput {
		payload := struct {
			Metadata config.Metadata  `json:"metadata"`
			Profiles []config.Profile `json:"profiles"`
		}{
			Metadata: meta,
			Profiles: cfg.Profiles,
		}
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(payload); err != nil {
			fmt.Fprintf(stderr, "failed to encode profiles: %v\n", err)
			return 1
		}
		return 0
	}

	fmt.Fprintln(stdout, "ccc profile list")
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "Config source: %s (%s)\n", meta.Source, meta.Path)
	if len(cfg.Profiles) == 0 {
		fmt.Fprintln(stdout, "No profiles configured.")
		return 0
	}
	fmt.Fprintln(stdout)
	for _, profile := range cfg.Profiles {
		line := fmt.Sprintf("- %s", profile.ID)
		if cfg.CurrentProfile == profile.ID {
			line += " [current]"
		}
		fmt.Fprintln(stdout, line)
		fmt.Fprintf(stdout, "  Name: %s\n", profile.Name)
		fmt.Fprintf(stdout, "  Command: %s\n", profile.Command)
		fmt.Fprintf(stdout, "  Provider: %s\n", profile.Provider)
		fmt.Fprintf(stdout, "  Base URL: %s\n", profile.BaseURL)
		fmt.Fprintf(stdout, "  Model: %s\n", profile.Model)
		if profile.FastModel != "" {
			fmt.Fprintf(stdout, "  Fast model: %s\n", profile.FastModel)
		}
	}
	return 0
}

func runProfileAdd(stdout, stderr io.Writer, store config.Store, args []string) int {
	fs := flag.NewFlagSet("profile add", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	name := fs.String("name", "", "display name")
	id := fs.String("id", "", "profile id")
	presetName := fs.String("preset", "", "provider preset")
	command := fs.String("command", "", "target command")
	provider := fs.String("provider", "", "provider type")
	baseURL := fs.String("base-url", "", "base URL")
	apiKey := fs.String("api-key", "", "API key")
	model := fs.String("model", "", "main model")
	fastModel := fs.String("fast-model", "", "fast model")
	noSync := fs.Bool("no-sync", false, "disable external sync")
	envVars := kvFlag{}
	fs.Var(&envVars, "env", "extra environment variable in KEY=VALUE form; repeatable")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(stderr, "failed to parse flags: %v\n", err)
		return 1
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: ccc profile add [--name ...] [--preset anthropic|openai|zhipu] --api-key ...")
		return 1
	}

	cfg, _, err := store.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}

	profile := config.Profile{
		ID:           strings.TrimSpace(*id),
		Name:         strings.TrimSpace(*name),
		Command:      strings.TrimSpace(*command),
		Provider:     strings.TrimSpace(*provider),
		BaseURL:      strings.TrimSpace(*baseURL),
		APIKey:       strings.TrimSpace(*apiKey),
		Model:        strings.TrimSpace(*model),
		FastModel:    strings.TrimSpace(*fastModel),
		ExtraEnv:     envVars.values,
		SyncExternal: !*noSync,
	}
	profile, err = preset.Apply(profile, *presetName)
	if err != nil {
		fmt.Fprintf(stderr, "failed to apply preset: %v\n", err)
		return 1
	}

	if profile.ID == "" {
		profile.ID = config.MakeProfileID(profile.Name)
	}
	profile = cfg.EnsureUniqueProfileID(profile)
	if err := cfg.UpsertProfile(profile); err != nil {
		fmt.Fprintf(stderr, "failed to add profile: %v\n", err)
		return 1
	}
	if cfg.CurrentProfile == "" {
		cfg.CurrentProfile = profile.ID
	}

	if err := store.Save(cfg); err != nil {
		fmt.Fprintf(stderr, "failed to save config: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Added profile %q with id %q\n", profile.Name, profile.ID)
	fmt.Fprintf(stdout, "Config written to %s\n", store.Layout.ConfigFile())
	return 0
}

func runProfileUpdate(stdout, stderr io.Writer, store config.Store, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: ccc profile update <profile-id-or-name> [--preset anthropic|openai|zhipu] [--model ...]")
		return 1
	}

	identifier := args[0]
	fs := flag.NewFlagSet("profile update", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	name := fs.String("name", "", "display name")
	id := fs.String("id", "", "profile id")
	presetName := fs.String("preset", "", "provider preset")
	command := fs.String("command", "", "target command")
	provider := fs.String("provider", "", "provider type")
	baseURL := fs.String("base-url", "", "base URL")
	apiKey := fs.String("api-key", "", "API key")
	model := fs.String("model", "", "main model")
	fastModel := fs.String("fast-model", "", "fast model")
	clearEnv := fs.Bool("clear-env", false, "remove all extra env entries before applying updates")
	sync := boolFlag{}
	noSync := boolFlag{}
	envVars := kvFlag{}
	unsetEnvVars := stringListFlag{}
	fs.Var(&sync, "sync", "enable external sync")
	fs.Var(&noSync, "no-sync", "disable external sync")
	fs.Var(&envVars, "env", "extra environment variable in KEY=VALUE form; repeatable")
	fs.Var(&unsetEnvVars, "unset-env", "remove an extra environment variable by key; repeatable")

	if err := fs.Parse(args[1:]); err != nil {
		fmt.Fprintf(stderr, "failed to parse flags: %v\n", err)
		return 1
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: ccc profile update <profile-id-or-name> [--preset anthropic|openai|zhipu] [--model ...]")
		return 1
	}
	if sync.set && noSync.set {
		fmt.Fprintln(stderr, "cannot use --sync and --no-sync together")
		return 1
	}

	cfg, _, err := store.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}

	profile, ok := cfg.FindProfile(identifier)
	if !ok {
		fmt.Fprintf(stderr, "profile %q not found\n", identifier)
		return 1
	}
	previousID := profile.ID
	previousName := profile.Name

	if strings.TrimSpace(*presetName) != "" {
		definition, err := preset.Lookup(*presetName)
		if err != nil {
			fmt.Fprintf(stderr, "failed to apply preset: %v\n", err)
			return 1
		}
		profile.Provider = definition.Provider
		profile.Command = definition.Command
		profile.BaseURL = definition.BaseURL
		profile.Model = definition.Model
		profile.FastModel = definition.FastModel
	}

	if strings.TrimSpace(*name) != "" {
		profile.Name = strings.TrimSpace(*name)
	}
	if strings.TrimSpace(*id) != "" {
		profile.ID = strings.TrimSpace(*id)
	}
	if strings.TrimSpace(*command) != "" {
		profile.Command = strings.TrimSpace(*command)
	}
	if strings.TrimSpace(*provider) != "" {
		profile.Provider = strings.TrimSpace(*provider)
	}
	if strings.TrimSpace(*baseURL) != "" {
		profile.BaseURL = strings.TrimSpace(*baseURL)
	}
	if strings.TrimSpace(*apiKey) != "" {
		profile.APIKey = strings.TrimSpace(*apiKey)
	}
	if strings.TrimSpace(*model) != "" {
		profile.Model = strings.TrimSpace(*model)
	}
	if strings.TrimSpace(*fastModel) != "" {
		profile.FastModel = strings.TrimSpace(*fastModel)
	}
	if *clearEnv {
		profile.ExtraEnv = map[string]string{}
	}
	if profile.ExtraEnv == nil {
		profile.ExtraEnv = map[string]string{}
	}
	for _, key := range unsetEnvVars.values {
		delete(profile.ExtraEnv, key)
	}
	for key, value := range envVars.values {
		profile.ExtraEnv[key] = value
	}
	if sync.set {
		profile.SyncExternal = sync.value
	}
	if noSync.set {
		profile.SyncExternal = !noSync.value
	}

	if _, ok, err := cfg.ReplaceProfile(identifier, profile); err != nil {
		fmt.Fprintf(stderr, "failed to update profile: %v\n", err)
		return 1
	} else if !ok {
		fmt.Fprintf(stderr, "profile %q not found\n", identifier)
		return 1
	}

	if err := store.Save(cfg); err != nil {
		fmt.Fprintf(stderr, "failed to save config: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Updated profile %q (%s -> %s)\n", previousName, previousID, profile.ID)
	fmt.Fprintf(stdout, "Config written to %s\n", store.Layout.ConfigFile())
	return 0
}

func runProfileDelete(stdout, stderr io.Writer, store config.Store, args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: ccc profile delete <profile-id-or-name>")
		return 1
	}

	cfg, _, err := store.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}

	removed, ok := cfg.DeleteProfile(args[0])
	if !ok {
		fmt.Fprintf(stderr, "profile %q not found\n", args[0])
		return 1
	}

	if err := store.Save(cfg); err != nil {
		fmt.Fprintf(stderr, "failed to save config: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Deleted profile %q (%s)\n", removed.Name, removed.ID)
	return 0
}

func runProfileUse(stdout, stderr io.Writer, store config.Store, args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: ccc profile use <profile-id-or-name>")
		return 1
	}

	cfg, _, err := store.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}

	profile, ok := cfg.FindProfile(args[0])
	if !ok {
		fmt.Fprintf(stderr, "profile %q not found\n", args[0])
		return 1
	}

	cfg.CurrentProfile = profile.ID
	if err := store.Save(cfg); err != nil {
		fmt.Fprintf(stderr, "failed to save config: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Current profile set to %q (%s)\n", profile.Name, profile.ID)
	return 0
}

func runCurrent(stdout, stderr io.Writer, home string, layout platform.Layout) int {
	store := config.NewStore(home, layout)
	cfg, _, err := store.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}

	if cfg.CurrentProfile == "" {
		fmt.Fprintln(stdout, "No current profile selected.")
		return 0
	}

	profile, ok := cfg.FindProfile(cfg.CurrentProfile)
	if !ok {
		fmt.Fprintf(stderr, "current profile %q not found in config\n", cfg.CurrentProfile)
		return 1
	}

	fmt.Fprintln(stdout, "ccc current")
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "ID: %s\n", profile.ID)
	fmt.Fprintf(stdout, "Name: %s\n", profile.Name)
	fmt.Fprintf(stdout, "Command: %s\n", profile.Command)
	fmt.Fprintf(stdout, "Provider: %s\n", profile.Provider)
	fmt.Fprintf(stdout, "Base URL: %s\n", profile.BaseURL)
	fmt.Fprintf(stdout, "Model: %s\n", profile.Model)
	if profile.FastModel != "" {
		fmt.Fprintf(stdout, "Fast model: %s\n", profile.FastModel)
	}
	return 0
}

func runRun(stdout, stderr io.Writer, home string, layout platform.Layout, args []string) int {
	profileIdentifier, flagArgs, cliArgs := splitRunArgs(args)

	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	dryRun := fs.Bool("dry-run", false, "print the execution plan instead of running it")
	envOnly := fs.Bool("env-only", false, "print the environment variables instead of running the command")
	bypass := fs.Bool("bypass", false, "enable bypass permissions")
	fs.BoolVar(bypass, "y", false, "enable bypass permissions")

	if err := fs.Parse(flagArgs); err != nil {
		fmt.Fprintf(stderr, "failed to parse run flags: %v\n", err)
		return 1
	}

	store := config.NewStore(home, layout)
	cfg, _, err := store.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}

	plan, err := runner.BuildPlan(cfg, profileIdentifier, cliArgs, *bypass)
	if err != nil {
		fmt.Fprintf(stderr, "failed to build run plan: %v\n", err)
		return 1
	}

	if *dryRun {
		printRunPlan(stdout, home, plan)
		return 0
	}
	if *envOnly {
		for _, entry := range runner.EnvList(plan) {
			fmt.Fprintln(stdout, entry)
		}
		return 0
	}

	if err := deps.EnsureCLI(plan.Command, stdout, stderr); err != nil {
		fmt.Fprintf(stderr, "dependency check failed: %v\n", err)
		return 1
	}

	if plan.Profile.SyncExternal {
		if _, err := cfgsync.Apply(home, plan.Profile); err != nil {
			fmt.Fprintf(stderr, "failed to sync external config: %v\n", err)
			return 1
		}
	}

	if err := runner.Run(plan, stdout, stderr); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(stderr, "failed to run %s: %v\n", plan.Command, err)
		return 1
	}
	return 0
}

func runSync(stdout, stderr io.Writer, home string, layout platform.Layout, args []string) int {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	dryRun := fs.Bool("dry-run", false, "print target paths without writing files")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(stderr, "failed to parse sync flags: %v\n", err)
		return 1
	}

	identifier := ""
	if fs.NArg() > 1 {
		fmt.Fprintln(stderr, "usage: ccc sync [profile-id-or-name] [--dry-run]")
		return 1
	}
	if fs.NArg() == 1 {
		identifier = fs.Arg(0)
	}

	store := config.NewStore(home, layout)
	cfg, _, err := store.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}

	plan, err := runner.BuildPlan(cfg, identifier, nil, false)
	if err != nil {
		fmt.Fprintf(stderr, "failed to resolve profile for sync: %v\n", err)
		return 1
	}

	if !plan.Profile.SyncExternal {
		fmt.Fprintf(stdout, "Profile %q does not require external sync.\n", plan.Profile.Name)
		return 0
	}

	paths := cfgsync.TargetPaths(home, plan.Profile)
	if *dryRun {
		fmt.Fprintln(stdout, "ccc sync --dry-run")
		fmt.Fprintln(stdout)
		fmt.Fprintf(stdout, "Profile: %s (%s)\n", plan.Profile.Name, plan.Profile.ID)
		fmt.Fprintln(stdout, "Target paths:")
		for _, path := range paths {
			fmt.Fprintf(stdout, "- %s\n", path)
		}
		return 0
	}

	result, err := cfgsync.Apply(home, plan.Profile)
	if err != nil {
		fmt.Fprintf(stderr, "failed to sync profile %q: %v\n", plan.Profile.Name, err)
		return 1
	}

	fmt.Fprintf(stdout, "Synced external config for %q\n", plan.Profile.Name)
	for _, path := range result.Paths {
		fmt.Fprintf(stdout, "- %s\n", path)
	}
	return 0
}

func runUpgrade(stdout, stderr io.Writer, args []string) int {
	fs := flag.NewFlagSet("upgrade", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	requestedVersion := fs.String("version", "", "target version in semver form")
	dryRun := fs.Bool("dry-run", false, "show the release asset plan without downloading")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(stderr, "failed to parse upgrade flags: %v\n", err)
		return 1
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: ccc upgrade [--version <semver>] [--dry-run]")
		return 1
	}

	executablePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(stderr, "failed to resolve current executable: %v\n", err)
		return 1
	}

	manager := upgrade.DefaultManager(buildinfo.Version, executablePath, runtime.GOOS, runtime.GOARCH)
	plan, err := manager.Plan(context.Background(), *requestedVersion)
	if err != nil {
		fmt.Fprintf(stderr, "failed to plan upgrade: %v\n", err)
		return 1
	}

	if *dryRun {
		fmt.Fprintln(stdout, "ccc upgrade --dry-run")
		fmt.Fprintln(stdout)
		fmt.Fprintf(stdout, "Current version: %s\n", plan.CurrentVersion)
		fmt.Fprintf(stdout, "Target version: %s\n", plan.TargetVersion)
		fmt.Fprintf(stdout, "Asset: %s\n", plan.AssetName)
		fmt.Fprintf(stdout, "Download URL: %s\n", plan.AssetURL)
		fmt.Fprintf(stdout, "Checksums URL: %s\n", plan.ChecksumsURL)
		fmt.Fprintf(stdout, "Executable path: %s\n", plan.ExecutablePath)
		if runtime.GOOS == "windows" {
			fmt.Fprintln(stdout, "Note: Windows currently requires reinstalling via install.ps1 instead of in-place self-upgrade.")
		}
		return 0
	}

	if plan.CurrentVersion != "" && plan.CurrentVersion != "dev" && plan.CurrentVersion == plan.TargetVersion {
		fmt.Fprintf(stdout, "ccc is already at version %s\n", plan.CurrentVersion)
		return 0
	}

	if err := manager.Upgrade(context.Background(), plan); err != nil {
		fmt.Fprintf(stderr, "failed to upgrade ccc: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Upgraded ccc from %s to %s\n", plan.CurrentVersion, plan.TargetVersion)
	fmt.Fprintf(stdout, "Executable updated at %s\n", plan.ExecutablePath)
	return 0
}

type kvFlag struct {
	values map[string]string
}

type stringListFlag struct {
	values []string
}

type boolFlag struct {
	value bool
	set   bool
}

func (f *kvFlag) String() string {
	return ""
}

func (f *kvFlag) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid env value %q: expected KEY=VALUE", value)
	}
	key := strings.TrimSpace(parts[0])
	if key == "" {
		return fmt.Errorf("invalid env value %q: key is empty", value)
	}
	if f.values == nil {
		f.values = map[string]string{}
	}
	f.values[key] = parts[1]
	return nil
}

func (f *stringListFlag) String() string {
	return strings.Join(f.values, ",")
}

func (f *stringListFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("value cannot be empty")
	}
	f.values = append(f.values, value)
	return nil
}

func (f *boolFlag) String() string {
	if !f.set {
		return ""
	}
	if f.value {
		return "true"
	}
	return "false"
}

func (f *boolFlag) Set(value string) error {
	f.set = true
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "true", "1", "yes":
		f.value = true
	case "false", "0", "no":
		f.value = false
	default:
		return fmt.Errorf("invalid boolean value %q", value)
	}
	return nil
}

func (f *boolFlag) IsBoolFlag() bool {
	return true
}

func pathContains(pathValue, target string) bool {
	cleanTarget := filepath.Clean(target)
	for _, entry := range filepath.SplitList(pathValue) {
		if filepath.Clean(strings.TrimSpace(entry)) == cleanTarget {
			return true
		}
	}
	return false
}

func yesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func versionSuffix(version string) string {
	if version == "" {
		return ""
	}
	return " (" + version + ")"
}

func pathSuffix(path string) string {
	if path == "" {
		return ""
	}
	return " (" + path + ")"
}

func splitRunArgs(args []string) (string, []string, []string) {
	cliIndex := len(args)
	for i, arg := range args {
		if arg == "--" {
			cliIndex = i
			break
		}
	}

	before := append([]string{}, args[:cliIndex]...)
	after := []string{}
	if cliIndex < len(args) {
		after = append(after, args[cliIndex+1:]...)
	}

	profileIdentifier := ""
	if len(before) > 0 && !strings.HasPrefix(before[0], "-") {
		profileIdentifier = before[0]
		before = before[1:]
	}

	return profileIdentifier, before, after
}

func printRunPlan(stdout io.Writer, home string, plan runner.Plan) {
	fmt.Fprintln(stdout, "ccc run --dry-run")
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "Profile: %s (%s)\n", plan.Profile.Name, plan.Profile.ID)
	fmt.Fprintf(stdout, "Command: %s\n", plan.Command)
	fmt.Fprintf(stdout, "External sync: %s\n", yesNo(plan.Profile.SyncExternal))
	if plan.Profile.SyncExternal {
		fmt.Fprintln(stdout, "Sync targets:")
		for _, path := range cfgsync.TargetPaths(home, plan.Profile) {
			fmt.Fprintf(stdout, "- %s\n", path)
		}
	}
	if len(plan.Args) == 0 {
		fmt.Fprintln(stdout, "Args: <none>")
	} else {
		fmt.Fprintf(stdout, "Args: %s\n", strings.Join(plan.Args, " "))
	}
	fmt.Fprintln(stdout, "Environment:")
	for _, entry := range runner.EnvList(plan) {
		fmt.Fprintf(stdout, "- %s\n", entry)
	}
}
