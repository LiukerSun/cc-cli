package app

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

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

var fetchZhipuModels = defaultFetchZhipuModels
var fetchAlibabaModels = defaultFetchAlibabaModels
var stdinIsInteractive = isInteractiveInput

var interactiveCodexModels = []string{
	"gpt-5.4",
	"gpt-5.4-mini",
	"gpt-5.3-codex",
	"gpt-5.3-codex-spark",
	"gpt-5.2-codex",
	"gpt-5.2",
	"gpt-5.1-codex-max",
	"gpt-5.1",
	"gpt-5.1-codex",
	"gpt-5",
	"gpt-5-codex",
	"gpt-5-codex-mini",
}

var interactiveZhipuFallbackModels = []string{
	"glm-5",
	"glm-4.7",
	"glm-4.7-air",
	"glm-4.7-airx",
}

var interactiveAlibabaFallbackModels = []string{
	"qwen3.5-plus",
	"qwen3-max-2026-01-23",
	"qwen3-coder-next",
	"qwen3-coder-plus",
	"glm-5",
	"glm-4.7",
	"kimi-k2.5",
	"minimax-m2.5",
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
	case "add":
		return runAdd(stdout, stderr, home, layout, args[1:])
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
	fmt.Fprintln(w, "  ccc add [<preset> <api-key> [model]] [--name ...] [--id ...]")
	fmt.Fprintln(w, "  ccc run [profile-id-or-name] [--dry-run] [--env-only] [-y|--bypass] [-- cli-args...]")
	fmt.Fprintln(w, "  ccc sync [profile-id-or-name] [--dry-run]")
	fmt.Fprintln(w, "  ccc profile list [--json]")
	fmt.Fprintln(w, "  ccc profile add [--name ...] [--preset anthropic|openai|zhipu|alibaba] --api-key ...")
	fmt.Fprintln(w, "  ccc profile update <profile-id-or-name> [--preset anthropic|openai|zhipu|alibaba] [--model ...]")
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
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Notes:")
	fmt.Fprintln(w, "  ccc run without a profile opens an interactive selector when used in a terminal.")
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

func runAdd(stdout, stderr io.Writer, home string, layout platform.Layout, args []string) int {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
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

	positionalArgs := args
	flagArgs := []string{}
	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			positionalArgs = args[:i]
			flagArgs = args[i:]
			break
		}
	}

	if err := fs.Parse(flagArgs); err != nil {
		fmt.Fprintf(stderr, "failed to parse flags: %v\n", err)
		return 1
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: ccc add [<preset> <api-key> [model]] [--name ...] [--id ...]")
		return 1
	}
	if len(positionalArgs) > 3 {
		fmt.Fprintln(stderr, "usage: ccc add [<preset> <api-key> [model]] [--name ...] [--id ...]")
		return 1
	}

	if len(positionalArgs) >= 1 && strings.TrimSpace(*presetName) == "" {
		*presetName = positionalArgs[0]
	}
	if len(positionalArgs) >= 2 && strings.TrimSpace(*apiKey) == "" {
		*apiKey = positionalArgs[1]
	}
	if len(positionalArgs) >= 3 && strings.TrimSpace(*model) == "" {
		*model = positionalArgs[2]
	}

	store := config.NewStore(home, layout)
	if len(positionalArgs) == 0 && strings.TrimSpace(*presetName) == "" && strings.TrimSpace(*apiKey) == "" {
		return runAddInteractive(os.Stdin, stdout, stderr, store, addProfileOptions{
			Name:      *name,
			ID:        *id,
			BaseURL:   *baseURL,
			Model:     *model,
			FastModel: *fastModel,
			NoSync:    *noSync,
			EnvVars:   envVars.values,
		})
	}

	return addProfile(stdout, stderr, store, addProfileOptions{
		Name:       *name,
		ID:         *id,
		PresetName: *presetName,
		Command:    *command,
		Provider:   *provider,
		BaseURL:    *baseURL,
		APIKey:     *apiKey,
		Model:      *model,
		FastModel:  *fastModel,
		NoSync:     *noSync,
		EnvVars:    envVars.values,
	})
}

type addProfileOptions struct {
	Name       string
	ID         string
	PresetName string
	Command    string
	Provider   string
	BaseURL    string
	APIKey     string
	Model      string
	FastModel  string
	NoSync     bool
	EnvVars    map[string]string
}

func runAddInteractive(stdin io.Reader, stdout, stderr io.Writer, store config.Store, initial addProfileOptions) int {
	fmt.Fprintln(stdout, "ccc add")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Interactive model configuration")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "  1) ZAI / ZHIPU AI")
	fmt.Fprintln(stdout, "  2) Alibaba Coding Plan")
	fmt.Fprintln(stdout, "  3) OpenAI Codex")
	fmt.Fprintln(stdout, "  4) Manual input")

	reader := bufio.NewReader(stdin)
	choice, err := promptChoice(reader, stdout, "Choice", 4)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read provider choice: %v\n", err)
		return 1
	}

	switch choice {
	case 1:
		return runAddZhipuInteractive(reader, stdout, stderr, store, initial)
	case 2:
		return runAddAlibabaInteractive(reader, stdout, stderr, store, initial)
	case 3:
		return runAddCodexInteractive(reader, stdout, stderr, store, initial)
	case 4:
		return runAddManualInteractive(reader, stdout, stderr, store, initial)
	default:
		fmt.Fprintf(stderr, "invalid choice %d\n", choice)
		return 1
	}
}

func runAddZhipuInteractive(reader *bufio.Reader, stdout, stderr io.Writer, store config.Store, initial addProfileOptions) int {
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "ZAI / ZHIPU AI")

	apiKey, err := promptRequired(reader, stdout, "API key")
	if err != nil {
		fmt.Fprintf(stderr, "failed to read API key: %v\n", err)
		return 1
	}

	models, err := fetchZhipuModels(apiKey)
	if err != nil {
		fmt.Fprintf(stdout, "Failed to fetch ZAI model list, using built-in list: %v\n", err)
		models = append([]string(nil), interactiveZhipuFallbackModels...)
	}

	mainModel, err := promptModelChoice(reader, stdout, "main", models, firstNonEmpty(initial.Model, "glm-5"), true)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read main model: %v\n", err)
		return 1
	}
	fastModel, err := promptModelChoice(reader, stdout, "fast", models, firstNonEmpty(initial.FastModel, mainModel), true)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read fast model: %v\n", err)
		return 1
	}

	defaultName := firstNonEmpty(initial.Name, fmt.Sprintf("ZHIPU (%s)", mainModel))
	name, err := promptWithDefault(reader, stdout, "Display name", defaultName)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read display name: %v\n", err)
		return 1
	}

	return addProfile(stdout, stderr, store, addProfileOptions{
		Name:       name,
		ID:         initial.ID,
		PresetName: "zhipu",
		APIKey:     apiKey,
		Model:      mainModel,
		FastModel:  fastModel,
		NoSync:     initial.NoSync,
		EnvVars:    initial.EnvVars,
	})
}

func runAddAlibabaInteractive(reader *bufio.Reader, stdout, stderr io.Writer, store config.Store, initial addProfileOptions) int {
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Alibaba Coding Plan")

	apiKey, err := promptRequired(reader, stdout, "API key")
	if err != nil {
		fmt.Fprintf(stderr, "failed to read API key: %v\n", err)
		return 1
	}

	models, err := fetchAlibabaModels(apiKey)
	if err != nil {
		fmt.Fprintf(stdout, "Failed to fetch Alibaba model list, using built-in list: %v\n", err)
		models = append([]string(nil), interactiveAlibabaFallbackModels...)
	}

	defaultMain := firstNonEmpty(initial.Model, interactiveAlibabaFallbackModels[0])
	mainModel, err := promptModelChoice(reader, stdout, "main", models, defaultMain, true)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read main model: %v\n", err)
		return 1
	}
	fastModel, err := promptModelChoice(reader, stdout, "fast", models, firstNonEmpty(initial.FastModel, mainModel), true)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read fast model: %v\n", err)
		return 1
	}

	defaultName := firstNonEmpty(initial.Name, fmt.Sprintf("Alibaba Coding Plan (%s)", mainModel))
	name, err := promptWithDefault(reader, stdout, "Display name", defaultName)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read display name: %v\n", err)
		return 1
	}

	return addProfile(stdout, stderr, store, addProfileOptions{
		Name:       name,
		ID:         initial.ID,
		PresetName: "alibaba",
		APIKey:     apiKey,
		Model:      mainModel,
		FastModel:  fastModel,
		NoSync:     initial.NoSync,
		EnvVars:    initial.EnvVars,
	})
}

func runAddCodexInteractive(reader *bufio.Reader, stdout, stderr io.Writer, store config.Store, initial addProfileOptions) int {
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "OpenAI Codex")

	baseURL, err := promptWithDefault(reader, stdout, "API base URL", firstNonEmpty(initial.BaseURL, "https://api.openai.com/v1"))
	if err != nil {
		fmt.Fprintf(stderr, "failed to read API base URL: %v\n", err)
		return 1
	}
	apiKey, err := promptRequired(reader, stdout, "API key")
	if err != nil {
		fmt.Fprintf(stderr, "failed to read API key: %v\n", err)
		return 1
	}

	mainModel, err := promptModelChoice(reader, stdout, "main", interactiveCodexModels, firstNonEmpty(initial.Model, "gpt-5.4"), true)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read model: %v\n", err)
		return 1
	}

	defaultName := firstNonEmpty(initial.Name, fmt.Sprintf("Codex (%s)", mainModel))
	name, err := promptWithDefault(reader, stdout, "Display name", defaultName)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read display name: %v\n", err)
		return 1
	}

	return addProfile(stdout, stderr, store, addProfileOptions{
		Name:       name,
		ID:         initial.ID,
		PresetName: "openai",
		BaseURL:    normalizeCodexBaseURL(baseURL),
		APIKey:     apiKey,
		Model:      mainModel,
		FastModel:  mainModel,
		NoSync:     initial.NoSync,
		EnvVars:    initial.EnvVars,
	})
}

func runAddManualInteractive(reader *bufio.Reader, stdout, stderr io.Writer, store config.Store, initial addProfileOptions) int {
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Manual configuration")

	name, err := promptRequired(reader, stdout, "Display name")
	if err != nil {
		fmt.Fprintf(stderr, "failed to read display name: %v\n", err)
		return 1
	}
	command, err := promptWithDefault(reader, stdout, "Command", firstNonEmpty(initial.Command, "claude"))
	if err != nil {
		fmt.Fprintf(stderr, "failed to read command: %v\n", err)
		return 1
	}
	command = strings.ToLower(strings.TrimSpace(command))
	if command != "claude" && command != "codex" {
		fmt.Fprintf(stderr, "unsupported command %q\n", command)
		return 1
	}

	baseURL, err := promptRequired(reader, stdout, "API base URL")
	if err != nil {
		fmt.Fprintf(stderr, "failed to read API base URL: %v\n", err)
		return 1
	}
	apiKey, err := promptRequired(reader, stdout, "API key")
	if err != nil {
		fmt.Fprintf(stderr, "failed to read API key: %v\n", err)
		return 1
	}
	model, err := promptRequired(reader, stdout, "Main model")
	if err != nil {
		fmt.Fprintf(stderr, "failed to read main model: %v\n", err)
		return 1
	}

	fastModel := ""
	if command == "claude" {
		fastModel, err = promptWithDefault(reader, stdout, "Fast model", firstNonEmpty(initial.FastModel, model))
		if err != nil {
			fmt.Fprintf(stderr, "failed to read fast model: %v\n", err)
			return 1
		}
	}

	if command == "codex" {
		baseURL = normalizeCodexBaseURL(baseURL)
	}

	return addProfile(stdout, stderr, store, addProfileOptions{
		Name:      name,
		ID:        initial.ID,
		Command:   command,
		Provider:  firstNonEmpty(initial.Provider, "custom"),
		BaseURL:   baseURL,
		APIKey:    apiKey,
		Model:     model,
		FastModel: fastModel,
		NoSync:    initial.NoSync,
		EnvVars:   initial.EnvVars,
	})
}

func promptRequired(reader *bufio.Reader, stdout io.Writer, label string) (string, error) {
	for {
		value, err := promptWithDefault(reader, stdout, label, "")
		if err != nil {
			return "", err
		}
		value = strings.TrimSpace(value)
		if value != "" {
			return value, nil
		}
		fmt.Fprintf(stdout, "%s is required.\n", label)
	}
}

func promptWithDefault(reader *bufio.Reader, stdout io.Writer, label, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Fprintf(stdout, "%s [%s]: ", label, defaultValue)
	} else {
		fmt.Fprintf(stdout, "%s: ", label)
	}

	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}

	value := strings.TrimSpace(line)
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func promptChoice(reader *bufio.Reader, stdout io.Writer, label string, max int) (int, error) {
	for {
		value, err := promptRequired(reader, stdout, fmt.Sprintf("%s [1-%d]", label, max))
		if err != nil {
			return 0, err
		}
		var choice int
		if _, err := fmt.Sscanf(value, "%d", &choice); err == nil && choice >= 1 && choice <= max {
			return choice, nil
		}
		fmt.Fprintf(stdout, "Please enter a number between 1 and %d.\n", max)
	}
}

func promptModelChoice(reader *bufio.Reader, stdout io.Writer, kind string, models []string, defaultValue string, allowCustom bool) (string, error) {
	choices := uniqueStrings(models)
	if allowCustom {
		choices = append(choices, "Custom model ID")
	}

	fmt.Fprintf(stdout, "Available %s models:\n", kind)
	for i, model := range choices {
		fmt.Fprintf(stdout, "  %2d) %s\n", i+1, model)
	}

	if defaultValue != "" {
		if idx := indexOf(choices, defaultValue); idx >= 0 {
			value, err := promptWithDefault(reader, stdout, fmt.Sprintf("Select %s model [1-%d]", kind, len(choices)), fmt.Sprintf("%d", idx+1))
			if err != nil {
				return "", err
			}
			return resolveModelChoice(reader, stdout, choices, value, kind)
		}
		if allowCustom {
			value, err := promptWithDefault(reader, stdout, fmt.Sprintf("Select %s model [1-%d]", kind, len(choices)), fmt.Sprintf("%d", len(choices)))
			if err != nil {
				return "", err
			}
			model, err := resolveModelChoice(reader, stdout, choices, value, kind)
			if err != nil {
				return "", err
			}
			if model == "Custom model ID" {
				return promptRequired(reader, stdout, "Custom model ID")
			}
			return model, nil
		}
	}

	for {
		value, err := promptRequired(reader, stdout, fmt.Sprintf("Select %s model [1-%d]", kind, len(choices)))
		if err != nil {
			return "", err
		}
		model, err := resolveModelChoice(reader, stdout, choices, value, kind)
		if err != nil {
			fmt.Fprintln(stdout, err.Error())
			continue
		}
		if model == "Custom model ID" {
			return promptRequired(reader, stdout, "Custom model ID")
		}
		return model, nil
	}
}

func resolveModelChoice(reader *bufio.Reader, stdout io.Writer, choices []string, value, kind string) (string, error) {
	var choice int
	if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &choice); err != nil || choice < 1 || choice > len(choices) {
		return "", fmt.Errorf("please enter a number between 1 and %d", len(choices))
	}
	model := choices[choice-1]
	if model == "Custom model ID" {
		customModel, err := promptRequired(reader, stdout, "Custom model ID")
		if err != nil {
			return "", err
		}
		return customModel, nil
	}
	_ = kind
	return model, nil
}

func defaultFetchZhipuModels(apiKey string) ([]string, error) {
	models, err := fetchModels("GET", "https://open.bigmodel.cn/api/paas/v4/models", apiKey, parseIDModelList)
	if err != nil {
		return nil, err
	}
	return uniqueStrings(append(models, interactiveZhipuFallbackModels...)), nil
}

func defaultFetchAlibabaModels(apiKey string) ([]string, error) {
	models, err := fetchModels("GET", "https://dashscope.aliyuncs.com/api/v1/models", apiKey, parseAlibabaModelList)
	if err != nil {
		return nil, err
	}
	filtered := make([]string, 0, len(models))
	for _, model := range models {
		normalized := strings.ToLower(strings.TrimSpace(model))
		if strings.HasPrefix(normalized, "qwen") || strings.HasPrefix(normalized, "glm") || strings.HasPrefix(normalized, "kimi") || strings.HasPrefix(normalized, "minimax") {
			filtered = append(filtered, model)
		}
	}
	return uniqueStrings(append(filtered, interactiveAlibabaFallbackModels...)), nil
}

func fetchModels(method, url, apiKey string, parser func([]byte) ([]string, error)) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}
	return parser(body)
}

func parseIDModelList(body []byte) ([]string, error) {
	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	models := make([]string, 0, len(payload.Data))
	for _, item := range payload.Data {
		if strings.TrimSpace(item.ID) != "" {
			models = append(models, item.ID)
		}
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("no models found")
	}
	sort.Strings(models)
	return models, nil
}

func parseAlibabaModelList(body []byte) ([]string, error) {
	var payload struct {
		Data []struct {
			ID    string `json:"id"`
			Model string `json:"model"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	models := make([]string, 0, len(payload.Data))
	for _, item := range payload.Data {
		if strings.TrimSpace(item.Model) != "" {
			models = append(models, item.Model)
			continue
		}
		if strings.TrimSpace(item.ID) != "" {
			models = append(models, item.ID)
		}
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("no models found")
	}
	sort.Strings(models)
	return models, nil
}

func normalizeCodexBaseURL(baseURL string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	switch {
	case strings.HasSuffix(trimmed, "/v1"):
		return trimmed
	case strings.HasSuffix(trimmed, "/v1/models"):
		return strings.TrimSuffix(trimmed, "/models")
	case strings.HasSuffix(trimmed, "/models"):
		return strings.TrimSuffix(trimmed, "/models") + "/v1"
	case strings.HasSuffix(trimmed, "/responses"):
		return strings.TrimSuffix(trimmed, "/responses") + "/v1"
	case trimmed == "":
		return "/v1"
	default:
		return trimmed + "/v1"
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func indexOf(values []string, target string) int {
	for i, value := range values {
		if value == target {
			return i
		}
	}
	return -1
}

func addProfile(stdout, stderr io.Writer, store config.Store, options addProfileOptions) int {
	cfg, _, err := store.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}

	profile := config.Profile{
		ID:           strings.TrimSpace(options.ID),
		Name:         strings.TrimSpace(options.Name),
		Command:      strings.TrimSpace(options.Command),
		Provider:     strings.TrimSpace(options.Provider),
		BaseURL:      strings.TrimSpace(options.BaseURL),
		APIKey:       strings.TrimSpace(options.APIKey),
		Model:        strings.TrimSpace(options.Model),
		FastModel:    strings.TrimSpace(options.FastModel),
		ExtraEnv:     options.EnvVars,
		SyncExternal: !options.NoSync,
	}
	profile, err = preset.Apply(profile, options.PresetName)
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
		fmt.Fprintln(stderr, "usage: ccc profile add [--name ...] [--preset anthropic|openai|zhipu|alibaba] --api-key ...")
		return 1
	}
	return addProfile(stdout, stderr, store, addProfileOptions{
		Name:       *name,
		ID:         *id,
		PresetName: *presetName,
		Command:    *command,
		Provider:   *provider,
		BaseURL:    *baseURL,
		APIKey:     *apiKey,
		Model:      *model,
		FastModel:  *fastModel,
		NoSync:     *noSync,
		EnvVars:    envVars.values,
	})
}

func runProfileUpdate(stdout, stderr io.Writer, store config.Store, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: ccc profile update <profile-id-or-name> [--preset anthropic|openai|zhipu|alibaba] [--model ...]")
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
		fmt.Fprintln(stderr, "usage: ccc profile update <profile-id-or-name> [--preset anthropic|openai|zhipu|alibaba] [--model ...]")
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
	return runRunWithInput(os.Stdin, stdout, stderr, home, layout, args)
}

func runRunWithInput(stdin io.Reader, stdout, stderr io.Writer, home string, layout platform.Layout, args []string) int {
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

	if profileIdentifier == "" && len(cfg.Profiles) > 1 && stdinIsInteractive(stdin) {
		selectedID, err := selectRunProfile(stdin, stdout, cfg)
		if err != nil {
			fmt.Fprintf(stderr, "failed to select profile: %v\n", err)
			return 1
		}
		profileIdentifier = selectedID
		if cfg.CurrentProfile != selectedID {
			cfg.CurrentProfile = selectedID
			if err := store.Save(cfg); err != nil {
				fmt.Fprintf(stderr, "failed to save selected profile: %v\n", err)
				return 1
			}
		}
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

func selectRunProfile(stdin io.Reader, stdout io.Writer, cfg config.File) (string, error) {
	reader := bufio.NewReader(stdin)

	fmt.Fprintln(stdout, "ccc run")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Select a model:")
	for i, profile := range cfg.Profiles {
		current := ""
		if profile.ID == cfg.CurrentProfile {
			current = " [current]"
		}
		fmt.Fprintf(stdout, "  %2d) %s%s\n", i+1, profile.Name, current)
		fmt.Fprintf(stdout, "      %s | %s | %s\n", profile.Command, profile.Provider, profile.Model)
	}

	defaultChoice := "1"
	if cfg.CurrentProfile != "" {
		for i, profile := range cfg.Profiles {
			if profile.ID == cfg.CurrentProfile {
				defaultChoice = fmt.Sprintf("%d", i+1)
				break
			}
		}
	}

	for {
		value, err := promptWithDefault(reader, stdout, fmt.Sprintf("Choice [1-%d]", len(cfg.Profiles)), defaultChoice)
		if err != nil {
			return "", err
		}
		var choice int
		if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &choice); err == nil && choice >= 1 && choice <= len(cfg.Profiles) {
			return cfg.Profiles[choice-1].ID, nil
		}
		fmt.Fprintf(stdout, "Please enter a number between 1 and %d.\n", len(cfg.Profiles))
	}
}

func isInteractiveInput(stdin io.Reader) bool {
	file, ok := stdin.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
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
