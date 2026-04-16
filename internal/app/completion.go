package app

import (
	"fmt"
	"io"
	"strings"

	"github.com/LiukerSun/cc-cli/internal/config"
	"github.com/LiukerSun/cc-cli/internal/platform"
)

var completionShells = []string{"bash", "zsh", "fish", "powershell"}

var topLevelCommands = []string{
	"help",
	"version",
	"current",
	"add",
	"run",
	"sync",
	"profile",
	"paths",
	"doctor",
	"upgrade",
	"config",
	"completion",
}

var topLevelCompatArgs = []string{
	"--help",
	"--version",
	"--list",
	"--add",
	"--delete",
	"--current",
	"-e",
	"-y",
	"--bypass",
	"--dry-run",
	"--env-only",
	"--auto-install",
	"--auto-sync",
}

var runFlags = []string{
	"--dry-run",
	"--env-only",
	"--auto-install",
	"--auto-sync",
	"--bypass",
	"-y",
	"--",
}

var syncFlags = []string{
	"--dry-run",
}

var addFlags = []string{
	"--name",
	"--id",
	"--preset",
	"--command",
	"--provider",
	"--base-url",
	"--api-key",
	"--model",
	"--fast-model",
	"--subagent-model",
	"--no-sync",
	"--env",
	"--deny-permission",
}

var profileSubcommands = []string{
	"list",
	"add",
	"update",
	"duplicate",
	"export",
	"import",
	"use",
	"delete",
}

var profileListFlags = []string{
	"--json",
	"--show-secrets",
}

var profileAddFlags = []string{
	"--name",
	"--id",
	"--preset",
	"--command",
	"--provider",
	"--base-url",
	"--api-key",
	"--model",
	"--fast-model",
	"--subagent-model",
	"--no-sync",
	"--env",
	"--deny-permission",
}

var profileUpdateFlags = []string{
	"--name",
	"--id",
	"--preset",
	"--command",
	"--provider",
	"--base-url",
	"--api-key",
	"--model",
	"--fast-model",
	"--subagent-model",
	"--clear-env",
	"--clear-deny-permissions",
	"--sync",
	"--no-sync",
	"--env",
	"--unset-env",
	"--deny-permission",
	"--unset-deny-permission",
}

var profileDuplicateFlags = []string{
	"--name",
	"--id",
}

var profileExportFlags = []string{
	"--output",
}

var profileImportFlags = []string{
	"--input",
	"--replace",
}

var configSubcommands = []string{
	"path",
	"show",
	"migrate",
}

var configShowFlags = []string{
	"--show-secrets",
}

var pathFlags = []string{
	"--json",
}

var upgradeFlags = []string{
	"--version",
	"--check",
	"--dry-run",
}

var providerPresets = []string{
	"anthropic",
	"openai",
	"zhipu",
	"alibaba",
	"claude",
	"codex",
	"gpt",
	"zai",
	"glm",
	"qwen",
	"dashscope",
	"tongyi",
	"kimi",
	"moonshot",
}

var commandValues = []string{
	"claude",
	"codex",
}

func runCompletion(stdout, stderr io.Writer, args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: ccc completion <bash|zsh|fish|powershell>")
		return 1
	}

	script, err := completionScript(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "failed to build completion script: %v\n", err)
		return 1
	}
	fmt.Fprint(stdout, script)
	return 0
}

func runInternalComplete(stdout, stderr io.Writer, home string, layout platform.Layout, args []string) int {
	candidates, err := completionCandidates(home, layout, args)
	if err != nil {
		fmt.Fprintf(stderr, "failed to complete args: %v\n", err)
		return 1
	}
	for _, candidate := range candidates {
		fmt.Fprintln(stdout, candidate)
	}
	return 0
}

func completionScript(shell string) (string, error) {
	switch shell {
	case "bash":
		return bashCompletionScript(), nil
	case "zsh":
		return zshCompletionScript(), nil
	case "fish":
		return fishCompletionScript(), nil
	case "powershell":
		return powershellCompletionScript(), nil
	default:
		return "", fmt.Errorf("unsupported shell %q", shell)
	}
}

func completionCandidates(home string, layout platform.Layout, args []string) ([]string, error) {
	args = normalizeCompletionArgs(args)
	if len(args) == 0 {
		return mergeCompletionGroups(topLevelCommands, topLevelCompatArgs), nil
	}

	if isTopLevelRunShortcut(args[0]) {
		return completeRun(home, layout, args)
	}

	switch args[0] {
	case "help", "version", "current", "doctor":
		return nil, nil
	case "completion":
		return completeSimpleArgs(args[1:], completionShells), nil
	case "paths":
		return completeFlagOnlyArgs(args[1:], pathFlags), nil
	case "upgrade":
		return completeFlagOnlyArgs(args[1:], upgradeFlags), nil
	case "config":
		return completeConfig(args[1:]), nil
	case "add":
		return completeAdd(args[1:]), nil
	case "run":
		return completeRun(home, layout, args[1:])
	case "sync":
		return completeSync(home, layout, args[1:])
	case "profile":
		return completeProfile(home, layout, args[1:])
	default:
		if len(args) == 1 {
			return mergeCompletionGroups(topLevelCommands, topLevelCompatArgs), nil
		}
	}

	return nil, nil
}

func normalizeCompletionArgs(args []string) []string {
	if len(args) == 0 {
		return args
	}
	if args[0] == "ccc" {
		return args[1:]
	}
	return args
}

func completeSimpleArgs(args []string, values []string) []string {
	if len(args) <= 1 {
		return values
	}
	return nil
}

func completeFlagOnlyArgs(args []string, flags []string) []string {
	if len(args) == 0 {
		return flags
	}
	current := args[len(args)-1]
	if strings.HasPrefix(current, "-") || current == "" {
		return flags
	}
	return nil
}

func completeConfig(args []string) []string {
	if len(args) == 0 {
		return configSubcommands
	}
	if len(args) == 1 && !strings.HasPrefix(args[0], "-") {
		return configSubcommands
	}
	switch args[0] {
	case "show":
		return completeFlagOnlyArgs(args[1:], configShowFlags)
	default:
		return nil
	}
}

func completeAdd(args []string) []string {
	if len(args) == 0 {
		return mergeCompletionGroups(providerPresets, addFlags)
	}
	prev := previousArg(args)
	if prev == "--preset" {
		return providerPresets
	}
	if prev == "--command" {
		return commandValues
	}
	current := args[len(args)-1]
	if len(args) == 1 && !strings.HasPrefix(current, "-") {
		return mergeCompletionGroups(providerPresets, addFlags)
	}
	if strings.HasPrefix(current, "-") || current == "" {
		return addFlags
	}
	return nil
}

func completeRun(home string, layout platform.Layout, args []string) ([]string, error) {
	if len(args) == 0 {
		return mergeCompletionGroups(runFlags, loadProfileIDs(home, layout)), nil
	}
	if containsDoubleDash(args) {
		return nil, nil
	}
	current := args[len(args)-1]
	if strings.HasPrefix(current, "-") || current == "" {
		if !hasRunProfileArg(args) {
			return mergeCompletionGroups(runFlags, loadProfileIDs(home, layout)), nil
		}
		return runFlags, nil
	}
	if !hasRunProfileArg(args) {
		return loadProfileIDs(home, layout), nil
	}
	return nil, nil
}

func completeSync(home string, layout platform.Layout, args []string) ([]string, error) {
	if len(args) == 0 {
		return mergeCompletionGroups(syncFlags, loadProfileIDs(home, layout)), nil
	}
	current := args[len(args)-1]
	if strings.HasPrefix(current, "-") || current == "" {
		if !hasPositionalArg(args) {
			return mergeCompletionGroups(syncFlags, loadProfileIDs(home, layout)), nil
		}
		return syncFlags, nil
	}
	if !hasPositionalArg(args) {
		return loadProfileIDs(home, layout), nil
	}
	return nil, nil
}

func completeProfile(home string, layout platform.Layout, args []string) ([]string, error) {
	if len(args) == 0 {
		return profileSubcommands, nil
	}
	if len(args) == 1 && !strings.HasPrefix(args[0], "-") {
		return profileSubcommands, nil
	}

	switch args[0] {
	case "list":
		return completeFlagOnlyArgs(args[1:], profileListFlags), nil
	case "add":
		return completeProfileAdd(args[1:]), nil
	case "update":
		return completeProfileWithIdentifier(args[1:], home, layout, profileUpdateFlags), nil
	case "duplicate":
		return completeProfileWithIdentifier(args[1:], home, layout, profileDuplicateFlags), nil
	case "export":
		return completeProfileWithIdentifier(args[1:], home, layout, profileExportFlags), nil
	case "import":
		return completeFlagOnlyArgs(args[1:], profileImportFlags), nil
	case "use", "delete":
		return completeProfileWithIdentifier(args[1:], home, layout, nil), nil
	default:
		return nil, nil
	}
}

func completeProfileAdd(args []string) []string {
	if len(args) == 0 {
		return profileAddFlags
	}
	prev := previousArg(args)
	if prev == "--preset" {
		return providerPresets
	}
	if prev == "--command" {
		return commandValues
	}
	current := args[len(args)-1]
	if strings.HasPrefix(current, "-") || current == "" {
		return profileAddFlags
	}
	return nil
}

func completeProfileWithIdentifier(args []string, home string, layout platform.Layout, flags []string) []string {
	if len(args) == 0 {
		return mergeCompletionGroups(loadProfileIDs(home, layout), flags)
	}
	prev := previousArg(args)
	if prev == "--preset" {
		return providerPresets
	}
	if prev == "--command" {
		return commandValues
	}
	current := args[len(args)-1]
	if !hasPositionalArg(args) {
		if strings.HasPrefix(current, "-") || current == "" {
			return mergeCompletionGroups(loadProfileIDs(home, layout), flags)
		}
		return loadProfileIDs(home, layout)
	}
	if strings.HasPrefix(current, "-") || current == "" {
		return flags
	}
	return nil
}

func hasRunProfileArg(args []string) bool {
	for _, arg := range args {
		if arg == "--" {
			return true
		}
		if strings.HasPrefix(arg, "-") || strings.TrimSpace(arg) == "" {
			continue
		}
		return true
	}
	return false
}

func hasPositionalArg(args []string) bool {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") || strings.TrimSpace(arg) == "" {
			continue
		}
		return true
	}
	return false
}

func containsDoubleDash(args []string) bool {
	for _, arg := range args {
		if arg == "--" {
			return true
		}
	}
	return false
}

func previousArg(args []string) string {
	if len(args) < 2 {
		return ""
	}
	return args[len(args)-2]
}

func loadProfileIDs(home string, layout platform.Layout) []string {
	store := config.NewStore(home, layout)
	cfg, _, err := store.Load()
	if err != nil {
		return nil
	}
	ids := make([]string, 0, len(cfg.Profiles))
	for _, profile := range cfg.Profiles {
		if strings.TrimSpace(profile.ID) != "" {
			ids = append(ids, profile.ID)
		}
	}
	return ids
}

func mergeCompletionGroups(groups ...[]string) []string {
	var merged []string
	for _, group := range groups {
		merged = append(merged, group...)
	}
	return uniqueStrings(merged)
}

func bashCompletionScript() string {
	return `#!/usr/bin/env bash
_ccc_completion() {
  local cur results
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  results="$(ccc __complete "${COMP_WORDS[@]}")"
  while IFS= read -r line; do
    [ -n "$line" ] || continue
    COMPREPLY+=("$line")
  done < <(compgen -W "$results" -- "$cur")
}

complete -o default -F _ccc_completion ccc
`
}

func zshCompletionScript() string {
	return `#compdef ccc
_ccc_completion() {
  local -a replies
  replies=("${(@f)$(ccc __complete "${words[@]}")}")
  _describe 'values' replies
}

compdef _ccc_completion ccc
`
}

func fishCompletionScript() string {
	return `complete -c ccc -f -a "(ccc __complete (commandline -cp))"
`
}

func powershellCompletionScript() string {
	return `Register-ArgumentCompleter -CommandName ccc -ScriptBlock {
  param($wordToComplete, $commandAst, $cursorPosition)

  $elements = $commandAst.CommandElements | ForEach-Object { $_.Extent.Text }
  ccc __complete @($elements) | ForEach-Object {
    [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
  }
}
`
}
