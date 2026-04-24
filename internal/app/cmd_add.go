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
	f := registerAddFlags(fs)

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

	if len(positionalArgs) >= 1 && strings.TrimSpace(*f.PresetName) == "" {
		*f.PresetName = positionalArgs[0]
	}
	if len(positionalArgs) >= 2 && strings.TrimSpace(*f.APIKey) == "" {
		*f.APIKey = positionalArgs[1]
	}
	if len(positionalArgs) >= 3 && strings.TrimSpace(*f.Model) == "" {
		*f.Model = positionalArgs[2]
	}

	store := config.NewStore(home, layout)
	if len(positionalArgs) == 0 && strings.TrimSpace(*f.PresetName) == "" && strings.TrimSpace(*f.APIKey) == "" {
		return runAddInteractive(os.Stdin, stdout, stderr, store, f.toOptions())
	}

	return addProfile(stdout, stderr, store, f.toOptions())
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

type addFlagBindings struct {
	Name            *string
	ID              *string
	PresetName      *string
	Command         *string
	Provider        *string
	BaseURL         *string
	APIKey          *string
	Model           *string
	FastModel       *string
	SubagentModel   *string
	NoSync          *bool
	EnvVars         *kvFlag
	DenyPermissions *stringListFlag
}

func registerAddFlags(fs *flag.FlagSet) addFlagBindings {
	b := addFlagBindings{
		EnvVars:         &kvFlag{},
		DenyPermissions: &stringListFlag{},
	}
	b.Name = fs.String("name", "", "display name")
	b.ID = fs.String("id", "", "profile id")
	b.PresetName = fs.String("preset", "", "provider preset")
	b.Command = fs.String("command", "", "target command")
	b.Provider = fs.String("provider", "", "provider type")
	b.BaseURL = fs.String("base-url", "", "base URL")
	b.APIKey = fs.String("api-key", "", "API key")
	b.Model = fs.String("model", "", "main model")
	b.FastModel = fs.String("fast-model", "", "fast model")
	b.SubagentModel = fs.String("subagent-model", "", "subagent model (optional, overrides default subagent routing)")
	b.NoSync = fs.Bool("no-sync", false, "disable external sync")
	fs.Var(b.EnvVars, "env", "extra environment variable in KEY=VALUE form; repeatable")
	fs.Var(b.DenyPermissions, "deny-permission", "append a Claude permissions.deny entry during sync; repeatable")
	return b
}

func (b addFlagBindings) toOptions() addProfileOptions {
	return addProfileOptions{
		Name:                *b.Name,
		ID:                  *b.ID,
		PresetName:          *b.PresetName,
		Command:             *b.Command,
		Provider:            *b.Provider,
		BaseURL:             *b.BaseURL,
		APIKey:              *b.APIKey,
		Model:               *b.Model,
		FastModel:           *b.FastModel,
		SubagentModel:       *b.SubagentModel,
		NoSync:              *b.NoSync,
		EnvVars:             b.EnvVars.values,
		SyncDenyPermissions: b.DenyPermissions.values,
	}
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
	fmt.Fprintln(stdout, "  5) Kimi Coding Plan")
	fmt.Fprintln(stdout, "  6) DeepSeek")
	fmt.Fprintln(stdout, "  7) Manual input")

	reader := bufio.NewReader(stdin)
	choice, err := promptChoice(reader, stdout, "Choice", 7)
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
		return runAddKimiInteractive(stdin, reader, stdout, stderr, store, initial)
	case 6:
		return runAddDeepSeekInteractive(stdin, reader, stdout, stderr, store, initial)
	case 7:
		return runAddManualInteractive(stdin, reader, stdout, stderr, store, initial)
	default:
		fmt.Fprintf(stderr, "invalid choice %d\n", choice)
		return 1
	}
}

type interactivePreset struct {
	label          string
	presetName     string
	nameFmt        string
	models         []string
	fetchModels    func(apiKey string) ([]string, error)
	fallbackModels []string
	defaultMain    string
	defaultFast    string
	askSubagent    bool
	askBaseURL     bool
	fastIsMain     bool
}

var anthropicInteractivePreset = interactivePreset{
	label:       "Anthropic Claude",
	presetName:  "anthropic",
	nameFmt:     "Claude (%s)",
	models:      interactiveAnthropicModels,
	defaultMain: "claude-sonnet-4-6",
	defaultFast: "claude-haiku-4-5",
	askSubagent: true,
}

var zhipuInteractivePreset = interactivePreset{
	label:       "ZAI / ZHIPU AI",
	presetName:  "zhipu",
	nameFmt:     "ZHIPU (%s)",
	models:      interactiveZhipuFallbackModels,
	defaultMain: "glm-5",
	askSubagent: true,
}

var alibabaInteractivePreset = interactivePreset{
	label:       "Alibaba Coding Plan",
	presetName:  "alibaba",
	nameFmt:     "Alibaba Coding Plan (%s)",
	models:      interactiveAlibabaFallbackModels,
	defaultMain: interactiveAlibabaFallbackModels[0],
	defaultFast: "qwen3.5-plus",
}

var kimiInteractivePreset = interactivePreset{
	label:       "Kimi Coding Plan",
	presetName:  "kimi",
	nameFmt:     "Kimi Coding Plan (%s)",
	models:      interactiveKimiFallbackModels,
	defaultMain: interactiveKimiFallbackModels[0],
}

var deepSeekInteractivePreset = interactivePreset{
	label:       "DeepSeek",
	presetName:  "deepseek",
	nameFmt:     "DeepSeek (%s)",
	models:      interactiveDeepSeekModels,
	defaultMain: "deepseek-v4-pro",
	defaultFast: "deepseek-v4-flash",
	askSubagent: true,
}

var codexInteractivePreset = interactivePreset{
	label:       "OpenAI Codex",
	presetName:  "openai",
	nameFmt:     "Codex (%s)",
	models:      interactiveCodexModels,
	defaultMain: "gpt-5.4",
	askBaseURL:  true,
	fastIsMain:  true,
}

func runAddAnthropicInteractive(stdin io.Reader, reader *bufio.Reader, stdout, stderr io.Writer, store config.Store, initial addProfileOptions) int {
	return runPresetInteractive(stdin, reader, stdout, stderr, store, initial, anthropicInteractivePreset)
}

func runAddZhipuInteractive(stdin io.Reader, reader *bufio.Reader, stdout, stderr io.Writer, store config.Store, initial addProfileOptions) int {
	preset := zhipuInteractivePreset
	preset.fetchModels = fetchZhipuModels
	preset.fallbackModels = interactiveZhipuFallbackModels
	return runPresetInteractive(stdin, reader, stdout, stderr, store, initial, preset)
}

func runAddAlibabaInteractive(stdin io.Reader, reader *bufio.Reader, stdout, stderr io.Writer, store config.Store, initial addProfileOptions) int {
	preset := alibabaInteractivePreset
	preset.fetchModels = fetchAlibabaModels
	preset.fallbackModels = interactiveAlibabaFallbackModels
	return runPresetInteractive(stdin, reader, stdout, stderr, store, initial, preset)
}

func runAddKimiInteractive(stdin io.Reader, reader *bufio.Reader, stdout, stderr io.Writer, store config.Store, initial addProfileOptions) int {
	return runPresetInteractive(stdin, reader, stdout, stderr, store, initial, kimiInteractivePreset)
}

func runAddDeepSeekInteractive(stdin io.Reader, reader *bufio.Reader, stdout, stderr io.Writer, store config.Store, initial addProfileOptions) int {
	return runPresetInteractive(stdin, reader, stdout, stderr, store, initial, deepSeekInteractivePreset)
}

func runAddCodexInteractive(stdin io.Reader, reader *bufio.Reader, stdout, stderr io.Writer, store config.Store, initial addProfileOptions) int {
	return runPresetInteractive(stdin, reader, stdout, stderr, store, initial, codexInteractivePreset)
}

func runPresetInteractive(stdin io.Reader, reader *bufio.Reader, stdout, stderr io.Writer, store config.Store, initial addProfileOptions, preset interactivePreset) int {
	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, preset.label)

	var baseURL string
	if preset.askBaseURL {
		var err error
		baseURL, err = promptWithDefault(reader, stdout, "API base URL", util.FirstNonEmpty(initial.BaseURL, defaultInteractiveBaseURL(preset)))
		if err != nil {
			fmt.Fprintf(stderr, "failed to read API base URL: %v\n", err)
			return 1
		}
	}

	apiKey, err := promptSecretRequired(stdin, reader, stdout, "API key")
	if err != nil {
		fmt.Fprintf(stderr, "failed to read API key: %v\n", err)
		return 1
	}

	models := preset.models
	if preset.fetchModels != nil {
		fetched, err := preset.fetchModels(apiKey)
		if err != nil {
			fmt.Fprintf(stdout, "Failed to fetch %s model list, using built-in list: %v\n", preset.label, err)
			models = append([]string(nil), preset.fallbackModels...)
		} else {
			models = fetched
		}
	}

	defaultMain := preset.defaultMain
	if defaultMain == "" && len(models) > 0 {
		defaultMain = models[0]
	}
	mainModel, err := promptModelChoice(reader, stdout, "main", models, util.FirstNonEmpty(initial.Model, defaultMain), true)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read main model: %v\n", err)
		return 1
	}

	var fastModel string
	if preset.fastIsMain {
		fastModel = mainModel
	} else {
		defaultFast := preset.defaultFast
		if defaultFast == "" {
			defaultFast = mainModel
		}
		fastModel, err = promptModelChoice(reader, stdout, "fast", models, util.FirstNonEmpty(initial.FastModel, defaultFast), true)
		if err != nil {
			fmt.Fprintf(stderr, "failed to read fast model: %v\n", err)
			return 1
		}
	}

	var subagentModel string
	if preset.askSubagent {
		subagentModel, _ = promptWithDefault(reader, stdout, "Subagent model (optional, press Enter to skip)", "")
	}

	defaultName := util.FirstNonEmpty(initial.Name, fmt.Sprintf(preset.nameFmt, mainModel))
	name, err := promptWithDefault(reader, stdout, "Display name", defaultName)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read display name: %v\n", err)
		return 1
	}

	opts := addProfileOptions{
		Name:       name,
		ID:         initial.ID,
		PresetName: preset.presetName,
		APIKey:     apiKey,
		Model:      mainModel,
		FastModel:  fastModel,
		NoSync:     initial.NoSync,
		EnvVars:    initial.EnvVars,
	}
	if preset.askBaseURL {
		opts.BaseURL = util.NormalizeCodexBaseURL(baseURL)
	}
	if preset.askSubagent {
		opts.SubagentModel = subagentModel
	}
	return addProfile(stdout, stderr, store, opts)
}

func defaultInteractiveBaseURL(preset interactivePreset) string {
	return "https://api.openai.com/v1"
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
