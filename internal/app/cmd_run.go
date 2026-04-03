package app

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/LiukerSun/cc-cli/internal/config"
	"github.com/LiukerSun/cc-cli/internal/deps"
	"github.com/LiukerSun/cc-cli/internal/platform"
	"github.com/LiukerSun/cc-cli/internal/runner"
	cfgsync "github.com/LiukerSun/cc-cli/internal/sync"
)

func runCurrent(stdout, stderr io.Writer, home string, layout platform.Layout) int {
	store := config.NewStore(home, layout)
	cfg, _, err := store.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}

	if cfg.CurrentProfile == "" {
		fmt.Fprintln(stdout, "No current profile selected. Use 'ccc profile use <id>' or run 'ccc' to select one interactively.")
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
	if profile.SubagentModel != "" {
		fmt.Fprintf(stdout, "Subagent model: %s\n", profile.SubagentModel)
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
	autoInstall := fs.Bool("auto-install", false, "install the target CLI automatically when missing")
	autoSync := fs.Bool("auto-sync", false, "write external Claude/Codex config before running")
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
	if profileIdentifier == "" && len(cfg.Profiles) == 0 {
		fmt.Fprintln(stderr, "no profiles configured; use 'ccc add' to add one")
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
		printRunPlan(stdout, home, plan, *autoInstall, *autoSync)
		return 0
	}
	if *envOnly {
		for _, entry := range runner.EnvList(plan) {
			fmt.Fprintln(stdout, entry)
		}
		return 0
	}

	if err := ensureRunCLI(plan.Command, *autoInstall, stdout, stderr); err != nil {
		fmt.Fprintf(stderr, "dependency check failed: %v\n", err)
		return 1
	}

	if plan.Profile.SyncExternal && *autoSync {
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

func ensureRunCLI(command string, autoInstall bool, stdout, stderr io.Writer) error {
	if deps.InspectTool(command).Installed {
		return nil
	}
	if autoInstall {
		return deps.EnsureCLI(command, stdout, stderr)
	}

	spec, ok := deps.SpecFor(command)
	if !ok {
		return fmt.Errorf("%s CLI not found on PATH", command)
	}
	return fmt.Errorf("%s CLI not found on PATH; rerun with --auto-install or install manually: npm install -g %s", command, spec.PackageName)
}

func printRunPlan(stdout io.Writer, home string, plan runner.Plan, autoInstall, autoSync bool) {
	fmt.Fprintln(stdout, "ccc run --dry-run")
	fmt.Fprintln(stdout)
	fmt.Fprintf(stdout, "Profile: %s (%s)\n", plan.Profile.Name, plan.Profile.ID)
	fmt.Fprintf(stdout, "Command: %s\n", plan.Command)
	fmt.Fprintf(stdout, "Missing CLI policy: %s\n", runPolicy(autoInstall, "install automatically", "fail"))
	fmt.Fprintf(stdout, "External sync: %s\n", yesNo(plan.Profile.SyncExternal))
	fmt.Fprintf(stdout, "External sync policy: %s\n", runPolicy(plan.Profile.SyncExternal && autoSync, "write before run", "skip"))
	if plan.Profile.SyncExternal {
		fmt.Fprintln(stdout, "Sync targets:")
		for _, path := range cfgsync.TargetPaths(home, plan.Profile) {
			fmt.Fprintf(stdout, "- %s\n", path)
		}
		if !autoSync {
			fmt.Fprintf(stdout, "Note: external sync is configured but skipped by default. Use 'ccc sync %s' or rerun with --auto-sync.\n", plan.Profile.ID)
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

func runPolicy(enabled bool, whenEnabled, whenDisabled string) string {
	if enabled {
		return whenEnabled
	}
	return whenDisabled
}
