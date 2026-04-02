package app

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/LiukerSun/cc-cli/internal/buildinfo"
	"github.com/LiukerSun/cc-cli/internal/upgrade"
)

func runUpgrade(stdout, stderr io.Writer, args []string) int {
	fs := flag.NewFlagSet("upgrade", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	requestedVersion := fs.String("version", "", "target version in semver form")
	checkOnly := fs.Bool("check", false, "check whether a newer version is available without upgrading")
	dryRun := fs.Bool("dry-run", false, "show the release asset plan without downloading")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(stderr, "failed to parse upgrade flags: %v\n", err)
		return 1
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: ccc upgrade [--version <semver>] [--check] [--dry-run]")
		return 1
	}
	if *checkOnly && *dryRun {
		fmt.Fprintln(stderr, "cannot use --check and --dry-run together")
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

	if *checkOnly {
		fmt.Fprintln(stdout, "ccc upgrade --check")
		fmt.Fprintln(stdout)
		currentVersion := plan.CurrentVersion
		if currentVersion == "" {
			currentVersion = "unknown"
		}
		fmt.Fprintf(stdout, "Current version: %s\n", currentVersion)
		fmt.Fprintf(stdout, "Latest version: %s\n", plan.TargetVersion)
		if plan.CurrentVersion != "" && plan.CurrentVersion != "dev" && plan.CurrentVersion == plan.TargetVersion {
			fmt.Fprintln(stdout, "Status: up to date")
		} else {
			fmt.Fprintln(stdout, "Status: update available")
		}
		return 0
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
