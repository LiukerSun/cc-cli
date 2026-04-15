package app

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LiukerSun/cc-cli/internal/config"
	"github.com/LiukerSun/cc-cli/internal/platform"
	"github.com/LiukerSun/cc-cli/internal/preset"
	"github.com/LiukerSun/cc-cli/internal/util"
)

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
	subagentModel := fs.String("subagent-model", "", "subagent model (optional, overrides default subagent routing)")
	noSync := fs.Bool("no-sync", false, "disable external sync")
	envVars := kvFlag{}
	denyPermissions := stringListFlag{}
	fs.Var(&envVars, "env", "extra environment variable in KEY=VALUE form; repeatable")
	fs.Var(&denyPermissions, "deny-permission", "append a Claude permissions.deny entry during sync; repeatable")

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
			Name:                *name,
			ID:                  *id,
			BaseURL:             *baseURL,
			Model:               *model,
			FastModel:           *fastModel,
			SubagentModel:       *subagentModel,
			NoSync:              *noSync,
			EnvVars:             envVars.values,
			SyncDenyPermissions: denyPermissions.values,
		})
	}

	return addProfile(stdout, stderr, store, addProfileOptions{
		Name:                *name,
		ID:                  *id,
		PresetName:          *presetName,
		Command:             *command,
		Provider:            *provider,
		BaseURL:             *baseURL,
		APIKey:              *apiKey,
		Model:               *model,
		FastModel:           *fastModel,
		NoSync:              *noSync,
		EnvVars:             envVars.values,
		SyncDenyPermissions: denyPermissions.values,
	})
}

type addProfileOptions struct {
	Name                string
	ID                  string
	PresetName          string
	Command             string
	Provider            string
	BaseURL             string
	APIKey              string
	Model               string
	FastModel           string
	SubagentModel       string
	NoSync              bool
	EnvVars             map[string]string
	SyncDenyPermissions []string
}

func runAddInteractive(stdin io.Reader, stdout, stderr io.Writer, store config.Store, initial addProfileOptions) int {
	fmt.Fprintln(stdout, "ccc add")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Interactive profile configuration")
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "  1) Anthropic Claude")
	fmt.Fprintln(stdout, "  2) OpenAI Codex")
	fmt.Fprintln(stdout, "  3) ZAI / ZHIPU AI")
	fmt.Fprintln(stdout, "  4) Alibaba Coding Plan")
	fmt.Fprintln(stdout, "  5) Manual input")

	reader := bufio.NewReader(stdin)
	choice, err := promptChoice(reader, stdout, "Choice", 5)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read provider choice: %v\n", err)
		return 1
	}

	switch choice {
	case 1:
		return runAddAnthropicInteractive(stdin, reader, stdout, stderr, store, initial)
	case 2:
		return runAddCodexInteractive(stdin, reader, stdout, stderr, store, initial)
	case 3:
		return runAddZhipuInteractive(stdin, reader, stdout, stderr, store, initial)
	case 4:
		return runAddAlibabaInteractive(stdin, reader, stdout, stderr, store, initial)
	case 5:
		return runAddManualInteractive(stdin, reader, stdout, stderr, store, initial)
	default:
		fmt.Fprintf(stderr, "invalid choice %d\n", choice)
		return 1
	}
}

func runAddAnthropicInteractive(stdin io.Reader, reader *bufio.Reader, stdout, stderr io.Writer, store config.Store, initial addProfileOptions) int {
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Anthropic Claude")

	apiKey, err := promptSecretRequired(stdin, reader, stdout, "API key")
	if err != nil {
		fmt.Fprintf(stderr, "failed to read API key: %v\n", err)
		return 1
	}

	mainModel, err := promptModelChoice(reader, stdout, "main", interactiveAnthropicModels, util.FirstNonEmpty(initial.Model, "claude-sonnet-4-6"), true)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read main model: %v\n", err)
		return 1
	}
	fastModel, err := promptModelChoice(reader, stdout, "fast", interactiveAnthropicModels, util.FirstNonEmpty(initial.FastModel, "claude-haiku-4-5"), true)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read fast model: %v\n", err)
		return 1
	}

	subagentModel, _ := promptWithDefault(reader, stdout, "Subagent model (optional, press Enter to skip)", "")
	defaultName := util.FirstNonEmpty(initial.Name, fmt.Sprintf("Claude (%s)", mainModel))
	name, err := promptWithDefault(reader, stdout, "Display name", defaultName)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read display name: %v\n", err)
		return 1
	}

	return addProfile(stdout, stderr, store, addProfileOptions{
		Name:          name,
		ID:            initial.ID,
		PresetName:    "anthropic",
		APIKey:        apiKey,
		Model:         mainModel,
		FastModel:     fastModel,
		SubagentModel: subagentModel,
		NoSync:        initial.NoSync,
		EnvVars:       initial.EnvVars,
	})
}

func runAddZhipuInteractive(stdin io.Reader, reader *bufio.Reader, stdout, stderr io.Writer, store config.Store, initial addProfileOptions) int {
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "ZAI / ZHIPU AI")

	apiKey, err := promptSecretRequired(stdin, reader, stdout, "API key")
	if err != nil {
		fmt.Fprintf(stderr, "failed to read API key: %v\n", err)
		return 1
	}

	models, err := fetchZhipuModels(apiKey)
	if err != nil {
		fmt.Fprintf(stdout, "Failed to fetch ZAI model list, using built-in list: %v\n", err)
		models = append([]string(nil), interactiveZhipuFallbackModels...)
	}

	mainModel, err := promptModelChoice(reader, stdout, "main", models, util.FirstNonEmpty(initial.Model, "glm-5"), true)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read main model: %v\n", err)
		return 1
	}
	fastModel, err := promptModelChoice(reader, stdout, "fast", models, util.FirstNonEmpty(initial.FastModel, mainModel), true)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read fast model: %v\n", err)
		return 1
	}

	subagentModel, _ := promptWithDefault(reader, stdout, "Subagent model (optional, press Enter to skip)", "")
	defaultName := util.FirstNonEmpty(initial.Name, fmt.Sprintf("ZHIPU (%s)", mainModel))
	name, err := promptWithDefault(reader, stdout, "Display name", defaultName)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read display name: %v\n", err)
		return 1
	}

	return addProfile(stdout, stderr, store, addProfileOptions{
		Name:          name,
		ID:            initial.ID,
		PresetName:    "zhipu",
		APIKey:        apiKey,
		Model:         mainModel,
		FastModel:     fastModel,
		SubagentModel: subagentModel,
		NoSync:        initial.NoSync,
		EnvVars:       initial.EnvVars,
	})
}

func runAddAlibabaInteractive(stdin io.Reader, reader *bufio.Reader, stdout, stderr io.Writer, store config.Store, initial addProfileOptions) int {
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Alibaba Coding Plan")

	apiKey, err := promptSecretRequired(stdin, reader, stdout, "API key")
	if err != nil {
		fmt.Fprintf(stderr, "failed to read API key: %v\n", err)
		return 1
	}

	models, err := fetchAlibabaModels(apiKey)
	if err != nil {
		fmt.Fprintf(stdout, "Failed to fetch Alibaba model list, using built-in list: %v\n", err)
		models = append([]string(nil), interactiveAlibabaFallbackModels...)
	}

	defaultMain := util.FirstNonEmpty(initial.Model, interactiveAlibabaFallbackModels[0])
	mainModel, err := promptModelChoice(reader, stdout, "main", models, defaultMain, true)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read main model: %v\n", err)
		return 1
	}
	fastModel, err := promptModelChoice(reader, stdout, "fast", models, util.FirstNonEmpty(initial.FastModel, "qwen3.5-plus"), true)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read fast model: %v\n", err)
		return 1
	}

	defaultName := util.FirstNonEmpty(initial.Name, fmt.Sprintf("Alibaba Coding Plan (%s)", mainModel))
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

func runAddCodexInteractive(stdin io.Reader, reader *bufio.Reader, stdout, stderr io.Writer, store config.Store, initial addProfileOptions) int {
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "OpenAI Codex")

	baseURL, err := promptWithDefault(reader, stdout, "API base URL", util.FirstNonEmpty(initial.BaseURL, "https://api.openai.com/v1"))
	if err != nil {
		fmt.Fprintf(stderr, "failed to read API base URL: %v\n", err)
		return 1
	}
	apiKey, err := promptSecretRequired(stdin, reader, stdout, "API key")
	if err != nil {
		fmt.Fprintf(stderr, "failed to read API key: %v\n", err)
		return 1
	}

	mainModel, err := promptModelChoice(reader, stdout, "main", interactiveCodexModels, util.FirstNonEmpty(initial.Model, "gpt-5.4"), true)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read model: %v\n", err)
		return 1
	}

	defaultName := util.FirstNonEmpty(initial.Name, fmt.Sprintf("Codex (%s)", mainModel))
	name, err := promptWithDefault(reader, stdout, "Display name", defaultName)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read display name: %v\n", err)
		return 1
	}

	return addProfile(stdout, stderr, store, addProfileOptions{
		Name:       name,
		ID:         initial.ID,
		PresetName: "openai",
		BaseURL:    util.NormalizeCodexBaseURL(baseURL),
		APIKey:     apiKey,
		Model:      mainModel,
		FastModel:  mainModel,
		NoSync:     initial.NoSync,
		EnvVars:    initial.EnvVars,
	})
}

func runAddManualInteractive(stdin io.Reader, reader *bufio.Reader, stdout, stderr io.Writer, store config.Store, initial addProfileOptions) int {
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "Manual configuration")

	name, err := promptRequired(reader, stdout, "Display name")
	if err != nil {
		fmt.Fprintf(stderr, "failed to read display name: %v\n", err)
		return 1
	}
	command, err := promptWithDefault(reader, stdout, "Command", util.FirstNonEmpty(initial.Command, "claude"))
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
	apiKey, err := promptSecretRequired(stdin, reader, stdout, "API key")
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
		fastModel, err = promptWithDefault(reader, stdout, "Fast model", util.FirstNonEmpty(initial.FastModel, model))
		if err != nil {
			fmt.Fprintf(stderr, "failed to read fast model: %v\n", err)
			return 1
		}
	}

	var subagentModel string
	if command == "claude" {
		subagentModel, _ = promptWithDefault(reader, stdout, "Subagent model (optional, press Enter to skip)", "")
	}

	if command == "codex" {
		baseURL = util.NormalizeCodexBaseURL(baseURL)
	}

	return addProfile(stdout, stderr, store, addProfileOptions{
		Name:          name,
		ID:            initial.ID,
		Command:       command,
		Provider:      util.FirstNonEmpty(initial.Provider, "custom"),
		BaseURL:       baseURL,
		APIKey:        apiKey,
		Model:         model,
		FastModel:     fastModel,
		SubagentModel: subagentModel,
		NoSync:        initial.NoSync,
		EnvVars:       initial.EnvVars,
	})
}

func addProfile(stdout, stderr io.Writer, store config.Store, options addProfileOptions) int {
	cfg, _, err := store.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}

	profile := config.Profile{
		ID:                  strings.TrimSpace(options.ID),
		Name:                strings.TrimSpace(options.Name),
		Command:             strings.TrimSpace(options.Command),
		Provider:            strings.TrimSpace(options.Provider),
		BaseURL:             strings.TrimSpace(options.BaseURL),
		APIKey:              strings.TrimSpace(options.APIKey),
		Model:               strings.TrimSpace(options.Model),
		FastModel:           strings.TrimSpace(options.FastModel),
		SubagentModel:       strings.TrimSpace(options.SubagentModel),
		ExtraEnv:            options.EnvVars,
		SyncExternal:        !options.NoSync,
		SyncDenyPermissions: append([]string(nil), options.SyncDenyPermissions...),
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
