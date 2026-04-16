package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/LiukerSun/cc-cli/internal/buildinfo"
	"github.com/LiukerSun/cc-cli/internal/config"
	"github.com/LiukerSun/cc-cli/internal/deps"
	"github.com/LiukerSun/cc-cli/internal/legacy"
	"github.com/LiukerSun/cc-cli/internal/platform"
	"github.com/LiukerSun/cc-cli/internal/util"
)

type pathsReport struct {
	Layout     platform.Layout  `json:"layout"`
	ConfigFile string           `json:"config_file"`
	Legacy     legacy.Detection `json:"legacy"`
}

func Run(args []string, stdout, stderr io.Writer) int {
	args = normalizeCompatibilityArgs(args)

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
		return runRun(stdout, stderr, home, layout, nil)
	}
	if isTopLevelRunShortcut(args[0]) {
		return runRun(stdout, stderr, home, layout, args)
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
	case "completion":
		return runCompletion(stdout, stderr, args[1:])
	case "__complete":
		return runInternalComplete(stdout, stderr, home, layout, args[1:])
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

func normalizeCompatibilityArgs(args []string) []string {
	if len(args) == 0 {
		return args
	}

	switch args[0] {
	case "--list":
		return []string{"profile", "list"}
	case "--add":
		return append([]string{"add"}, args[1:]...)
	case "--delete":
		return append([]string{"profile", "delete", "--force"}, args[1:]...)
	case "--current":
		return []string{"current"}
	case "-e":
		return append([]string{"run", "--env-only"}, args[1:]...)
	default:
		return args
	}
}

func isTopLevelRunShortcut(arg string) bool {
	switch arg {
	case "-y", "--bypass", "--dry-run", "--env-only", "--auto-install", "--auto-sync", "--":
		return true
	default:
		return false
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
	if len(args) >= 1 && args[0] == "show" {
		return runConfigShow(stdout, stderr, home, layout, args[1:])
	}
	if len(args) == 1 && args[0] == "migrate" {
		return runConfigMigrate(stdout, stderr, home, layout)
	}

	fmt.Fprintln(stderr, "usage: ccc config path")
	fmt.Fprintln(stderr, "       ccc config show [--show-secrets]")
	fmt.Fprintln(stderr, "       ccc config migrate")
	return 1
}

func runConfigShow(stdout, stderr io.Writer, home string, layout platform.Layout, args []string) int {
	fs := flag.NewFlagSet("config show", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	showSecrets := fs.Bool("show-secrets", false, "show API keys in config output")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(stderr, "failed to parse flags: %v\n", err)
		return 1
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: ccc config show [--show-secrets]")
		return 1
	}

	store := config.NewStore(home, layout)
	cfg, meta, err := store.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}
	if !*showSecrets {
		cfg = cfg.Redacted()
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
	fmt.Fprintf(stdout, "- Binary dir on PATH: %s\n", yesNo(util.PathContains(os.Getenv("PATH"), report.Layout.BinDir)))
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
				warnings = append(warnings, fmt.Sprintf("Current profile command %q is not available on PATH. Install it first or use 'ccc run --auto-install'.", currentProfile.Command))
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
	fmt.Fprintln(w, "Profile manager and launcher for Claude/Codex-compatible CLIs.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  ccc help")
	fmt.Fprintln(w, "  ccc version")
	fmt.Fprintln(w, "  ccc current")
	fmt.Fprintln(w, "  ccc add [<preset> <api-key> [model]] [--name ...] [--id ...]")
	fmt.Fprintln(w, "  ccc run [profile-id-or-name] [--dry-run] [--env-only] [--auto-install] [--auto-sync] [-y|--bypass] [-- cli-args...]")
	fmt.Fprintln(w, "  ccc completion <bash|zsh|fish|powershell>")
	fmt.Fprintln(w, "  ccc sync [profile-id-or-name] [--dry-run]")
	fmt.Fprintln(w, "  ccc profile list [--json] [--show-secrets]")
	fmt.Fprintln(w, "  ccc profile add [--name ...] [--preset anthropic|openai|zhipu|alibaba|kimi] --api-key ...")
	fmt.Fprintln(w, "  ccc profile update <profile-id-or-name> [--preset anthropic|openai|zhipu|alibaba|kimi] [--model ...]")
	fmt.Fprintln(w, "  ccc profile duplicate <profile-id-or-name> [--name ...] [--id ...]")
	fmt.Fprintln(w, "  ccc profile export [profile-id-or-name] [--output <path>]")
	fmt.Fprintln(w, "  ccc profile import [--input <path>] [--replace]")
	fmt.Fprintln(w, "  ccc profile use <profile-id-or-name>")
	fmt.Fprintln(w, "  ccc profile delete <profile-id-or-name> [--force]")
	fmt.Fprintln(w, "  ccc paths [--json]")
	fmt.Fprintln(w, "  ccc config path")
	fmt.Fprintln(w, "  ccc config show [--show-secrets]")
	fmt.Fprintln(w, "  ccc config migrate")
	fmt.Fprintln(w, "  ccc doctor")
	fmt.Fprintln(w, "  ccc upgrade [--version <semver>] [--check] [--dry-run]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Compatibility aliases:")
	fmt.Fprintln(w, "  ccc --help")
	fmt.Fprintln(w, "  ccc --version")
	fmt.Fprintln(w, "  ccc --list")
	fmt.Fprintln(w, "  ccc --add ...")
	fmt.Fprintln(w, "  ccc --delete <profile-id-or-name>")
	fmt.Fprintln(w, "  ccc --current")
	fmt.Fprintln(w, "  ccc -e [profile-id-or-name]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Notes:")
	fmt.Fprintln(w, "  ccc without arguments opens the profile selector and runs the selected profile.")
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
