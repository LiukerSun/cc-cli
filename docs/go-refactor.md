# Go Refactor Status

This document tracks the current state of the Go rewrite proposed in issue `#27`.

## What Is Already Done

- A real Go module at the repository root
- A new `cmd/ccc` entrypoint
- A platform-aware directory model for the new install layout
- Legacy path detection and config migration support
- A real command surface for:
  - `ccc version`
  - `ccc current`
  - `ccc run`
  - `ccc sync`
  - `ccc profile list/add/use/delete`
  - `ccc paths`
  - `ccc config path/show/migrate`
  - `ccc doctor`
- Real dependency inspection for `node`, `npm`, `claude`, and `codex`
- Runtime auto-install for missing `claude` / `codex`
- Claude and Codex external config sync
- A first `ccc upgrade` command for release-based binary upgrades
- Thin installers for Unix and Windows
- GoReleaser config plus CI and tag-based release workflow
- Shell and Go tests covering config, runtime planning, installers, and release asset consistency

## What Changed Architecturally

- The Go CLI is now the primary implementation.
- `install.sh` and `install.ps1` now install the Go binary instead of bootstrapping a shell-heavy runtime.
- `bin/ccc` and `bin/ccc.ps1` were reduced to compatibility wrappers and no longer define the product architecture.
- The default install and config paths now follow the new cross-platform layout described below.

## What Is Still Transitional

- The repository still contains compatibility wrappers for legacy entrypoints.
- Some historical notes remain in `CHANGELOG.md` because it records earlier script-based releases.
- Provider-specific helpers and profile bootstrapping are still thin compared with the final target experience.

## New Directory Model

### Linux

- Binary: `~/.local/bin/ccc`
- Config: `~/.config/ccc/`
- Data: `~/.local/share/ccc/`
- Cache: `~/.cache/ccc/`
- State: `~/.local/state/ccc/`

### macOS

Current bootstrap implementation also uses the XDG-style layout above.

That keeps the first implementation simple and gives us a single cross-platform Unix path model. If we later decide to support `~/Library/Application Support/ccc`, that should remain a compatibility layer rather than a new internal abstraction leak.

### Windows

- Binary: `%LOCALAPPDATA%\\Programs\\ccc\\bin`
- Config: `%APPDATA%\\ccc`
- Data: `%LOCALAPPDATA%\\ccc\\data`
- Cache: `%LOCALAPPDATA%\\ccc\\cache`
- State: `%LOCALAPPDATA%\\ccc\\state`

## Next Steps

1. Add provider-specific presets and profile bootstrap commands.
2. Continue shrinking compatibility surface around old script-era entrypoints.
3. Refine release, upgrade, and migration UX now that the Go binary is the primary product path.
