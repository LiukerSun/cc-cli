package app

import (
	"flag"
	"fmt"
	"io"

	"github.com/LiukerSun/cc-cli/internal/config"
	"github.com/LiukerSun/cc-cli/internal/platform"
	"github.com/LiukerSun/cc-cli/internal/runner"
	cfgsync "github.com/LiukerSun/cc-cli/internal/sync"
)

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
