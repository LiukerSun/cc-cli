package app

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LiukerSun/cc-cli/internal/config"
	"github.com/LiukerSun/cc-cli/internal/platform"
	"github.com/LiukerSun/cc-cli/internal/preset"
)

func runProfile(stdout, stderr io.Writer, home string, layout platform.Layout, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: ccc profile <list|add|update|duplicate|export|import|use|delete>")
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
	case "duplicate":
		return runProfileDuplicate(stdout, stderr, store, args[1:])
	case "export":
		return runProfileExport(stdout, stderr, store, args[1:])
	case "import":
		return runProfileImport(stdout, stderr, store, args[1:])
	case "use":
		return runProfileUse(stdout, stderr, store, args[1:])
	case "delete":
		return runProfileDelete(os.Stdin, stdout, stderr, store, args[1:])
	default:
		fmt.Fprintf(stderr, "unknown profile command: %s\n", args[0])
		return 1
	}
}

func runProfileList(stdout, stderr io.Writer, store config.Store, args []string) int {
	fs := flag.NewFlagSet("profile list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	jsonOutput := fs.Bool("json", false, "print JSON")
	showSecrets := fs.Bool("show-secrets", false, "show API keys in output")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(stderr, "failed to parse flags: %v\n", err)
		return 1
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: ccc profile list [--json] [--show-secrets]")
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
		profiles := append([]config.Profile(nil), cfg.Profiles...)
		if !*showSecrets {
			profiles = make([]config.Profile, len(cfg.Profiles))
			for i, profile := range cfg.Profiles {
				profiles[i] = profile.Redacted()
			}
		}
		payload := struct {
			Metadata config.Metadata  `json:"metadata"`
			Profiles []config.Profile `json:"profiles"`
		}{
			Metadata: meta,
			Profiles: profiles,
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
			fmt.Fprintf(stdout, "  API key: %s\n", config.MaskSecret(profile.APIKey))
		if profile.FastModel != "" {
			fmt.Fprintf(stdout, "  Fast model: %s\n", profile.FastModel)
		}
		if profile.SubagentModel != "" {
			fmt.Fprintf(stdout, "  Subagent model: %s\n", profile.SubagentModel)
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
	subagentModel := fs.String("subagent-model", "", "subagent model (optional, overrides default subagent routing)")
	noSync := fs.Bool("no-sync", false, "disable external sync")
	envVars := kvFlag{}
	denyPermissions := stringListFlag{}
	fs.Var(&envVars, "env", "extra environment variable in KEY=VALUE form; repeatable")
	fs.Var(&denyPermissions, "deny-permission", "append a Claude permissions.deny entry during sync; repeatable")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(stderr, "failed to parse flags: %v\n", err)
		return 1
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: ccc profile add [--name ...] [--preset anthropic|openai|zhipu|alibaba] --api-key ...")
		return 1
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
		SubagentModel:       *subagentModel,
		NoSync:              *noSync,
		EnvVars:             envVars.values,
		SyncDenyPermissions: denyPermissions.values,
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
	subagentModel := fs.String("subagent-model", "", "subagent model (optional, overrides default subagent routing)")
	clearEnv := fs.Bool("clear-env", false, "remove all extra env entries before applying updates")
	clearDenyPermissions := fs.Bool("clear-deny-permissions", false, "remove all configured sync deny permissions")
	sync := boolFlag{}
	noSync := boolFlag{}
	envVars := kvFlag{}
	unsetEnvVars := stringListFlag{}
	denyPermissions := stringListFlag{}
	unsetDenyPermissions := stringListFlag{}
	fs.Var(&sync, "sync", "enable external sync")
	fs.Var(&noSync, "no-sync", "disable external sync")
	fs.Var(&envVars, "env", "extra environment variable in KEY=VALUE form; repeatable")
	fs.Var(&unsetEnvVars, "unset-env", "remove an extra environment variable by key; repeatable")
	fs.Var(&denyPermissions, "deny-permission", "append a Claude permissions.deny entry during sync; repeatable")
	fs.Var(&unsetDenyPermissions, "unset-deny-permission", "remove a configured sync deny permission; repeatable")

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
		fmt.Fprintf(stderr, "profile %q not found. Run 'ccc profile list' to see available profiles.\n", identifier)
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
	if strings.TrimSpace(*subagentModel) != "" {
		profile.SubagentModel = strings.TrimSpace(*subagentModel)
	}
	if *clearEnv {
		profile.ExtraEnv = map[string]string{}
	}
	if *clearDenyPermissions {
		profile.SyncDenyPermissions = nil
	}
	if profile.ExtraEnv == nil {
		profile.ExtraEnv = map[string]string{}
	}
	for _, key := range unsetEnvVars.values {
		delete(profile.ExtraEnv, key)
	}
	for _, permission := range unsetDenyPermissions.values {
		profile.SyncDenyPermissions = removeString(profile.SyncDenyPermissions, permission)
	}
	for key, value := range envVars.values {
		profile.ExtraEnv[key] = value
	}
	profile.SyncDenyPermissions = append(profile.SyncDenyPermissions, denyPermissions.values...)
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
		fmt.Fprintf(stderr, "profile %q not found. Run 'ccc profile list' to see available profiles.\n", identifier)
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

type profileTransferFile struct {
	Version        int              `json:"version,omitempty"`
	CurrentProfile string           `json:"current_profile,omitempty"`
	Profiles       []config.Profile `json:"profiles"`
}

func runProfileDuplicate(stdout, stderr io.Writer, store config.Store, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: ccc profile duplicate <profile-id-or-name> [--name ...] [--id ...]")
		return 1
	}

	identifier := args[0]
	fs := flag.NewFlagSet("profile duplicate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	name := fs.String("name", "", "display name for the duplicated profile")
	id := fs.String("id", "", "profile id for the duplicated profile")
	if err := fs.Parse(args[1:]); err != nil {
		fmt.Fprintf(stderr, "failed to parse flags: %v\n", err)
		return 1
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: ccc profile duplicate <profile-id-or-name> [--name ...] [--id ...]")
		return 1
	}

	cfg, _, err := store.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}

	source, ok := cfg.FindProfile(identifier)
	if !ok {
		fmt.Fprintf(stderr, "profile %q not found. Run 'ccc profile list' to see available profiles.\n", identifier)
		return 1
	}

	duplicate := source
	duplicate.Name = strings.TrimSpace(*name)
	if duplicate.Name == "" {
		duplicate.Name = source.Name + " Copy"
	}
	if trimmedID := strings.TrimSpace(*id); trimmedID != "" {
		if trimmedID == source.ID {
			fmt.Fprintf(stderr, "failed to duplicate profile: duplicate profile id %q\n", trimmedID)
			return 1
		}
		if existing, exists := cfg.FindProfile(trimmedID); exists && existing.ID != source.ID {
			fmt.Fprintf(stderr, "failed to duplicate profile: duplicate profile id %q\n", trimmedID)
			return 1
		}
		duplicate.ID = trimmedID
	} else {
		duplicate.ID = config.MakeProfileID(duplicate.Name)
		duplicate = cfg.EnsureUniqueProfileID(duplicate)
	}

	if err := cfg.UpsertProfile(duplicate); err != nil {
		fmt.Fprintf(stderr, "failed to duplicate profile: %v\n", err)
		return 1
	}
	if err := store.Save(cfg); err != nil {
		fmt.Fprintf(stderr, "failed to save config: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Duplicated profile %q as %q (%s)\n", source.Name, duplicate.Name, duplicate.ID)
	fmt.Fprintf(stdout, "Config written to %s\n", store.Layout.ConfigFile())
	return 0
}

func runProfileExport(stdout, stderr io.Writer, store config.Store, args []string) int {
	identifier := ""
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		identifier = args[0]
		args = args[1:]
	}

	fs := flag.NewFlagSet("profile export", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	outputPath := fs.String("output", "", "write exported profiles to a file instead of stdout")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(stderr, "failed to parse flags: %v\n", err)
		return 1
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: ccc profile export [profile-id-or-name] [--output <path>]")
		return 1
	}

	cfg, _, err := store.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}

	payload := profileTransferFile{
		Version:        config.CurrentVersion,
		CurrentProfile: cfg.CurrentProfile,
		Profiles:       append([]config.Profile(nil), cfg.Profiles...),
	}
	if identifier != "" {
		profile, ok := cfg.FindProfile(identifier)
		if !ok {
			fmt.Fprintf(stderr, "profile %q not found. Run 'ccc profile list' to see available profiles.\n", identifier)
			return 1
		}
		payload.Profiles = []config.Profile{profile}
		if cfg.CurrentProfile != profile.ID {
			payload.CurrentProfile = ""
		}
	}

	if err := writeJSONPayload(*outputPath, payload, stdout); err != nil {
		fmt.Fprintf(stderr, "failed to export profiles: %v\n", err)
		return 1
	}
	if strings.TrimSpace(*outputPath) != "" {
		fmt.Fprintf(stdout, "Exported %d profile(s) to %s\n", len(payload.Profiles), *outputPath)
	}
	return 0
}

func runProfileImport(stdout, stderr io.Writer, store config.Store, args []string) int {
	fs := flag.NewFlagSet("profile import", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	inputPath := fs.String("input", "", "read profiles from a JSON file instead of stdin")
	replace := fs.Bool("replace", false, "replace existing profiles when imported ids already exist")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(stderr, "failed to parse flags: %v\n", err)
		return 1
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: ccc profile import [--input <path>] [--replace]")
		return 1
	}

	data, err := readImportPayload(*inputPath)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read import payload: %v\n", err)
		return 1
	}
	payload, err := decodeProfileTransferFile(data)
	if err != nil {
		fmt.Fprintf(stderr, "failed to decode import payload: %v\n", err)
		return 1
	}
	if len(payload.Profiles) == 0 {
		fmt.Fprintln(stderr, "import payload did not include any profiles")
		return 1
	}

	cfg, _, err := store.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}

	importedIDs := map[string]string{}
	imported := make([]config.Profile, 0, len(payload.Profiles))
	for _, profile := range payload.Profiles {
		originalID := profile.ID
		if *replace {
			if _, ok := cfg.FindProfile(profile.ID); ok {
				if _, _, err := cfg.ReplaceProfile(profile.ID, profile); err != nil {
					fmt.Fprintf(stderr, "failed to import profile %q: %v\n", profile.ID, err)
					return 1
				}
			} else if err := cfg.UpsertProfile(profile); err != nil {
				fmt.Fprintf(stderr, "failed to import profile %q: %v\n", profile.ID, err)
				return 1
			}
		} else {
			if _, ok := cfg.FindProfile(profile.ID); ok {
				profile = cfg.EnsureUniqueProfileID(profile)
			}
			if err := cfg.UpsertProfile(profile); err != nil {
				fmt.Fprintf(stderr, "failed to import profile %q: %v\n", profile.ID, err)
				return 1
			}
		}
		importedIDs[originalID] = profile.ID
		imported = append(imported, profile)
	}

	if cfg.CurrentProfile == "" {
		if mappedID, ok := importedIDs[payload.CurrentProfile]; ok {
			cfg.CurrentProfile = mappedID
		} else if len(imported) > 0 {
			cfg.CurrentProfile = imported[0].ID
		}
	}

	if err := store.Save(cfg); err != nil {
		fmt.Fprintf(stderr, "failed to save config: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Imported %d profile(s)\n", len(imported))
	fmt.Fprintf(stdout, "Config written to %s\n", store.Layout.ConfigFile())
	return 0
}

func writeJSONPayload(outputPath string, payload any, stdout io.Writer) error {
	if strings.TrimSpace(outputPath) == "" {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(payload)
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outputPath, data, 0o600)
}

func readImportPayload(inputPath string) ([]byte, error) {
	if strings.TrimSpace(inputPath) != "" {
		return os.ReadFile(inputPath)
	}
	return io.ReadAll(os.Stdin)
}

func decodeProfileTransferFile(data []byte) (profileTransferFile, error) {
	var wrapper profileTransferFile
	if err := json.Unmarshal(data, &wrapper); err == nil && wrapper.Profiles != nil {
		return wrapper, nil
	}

	var profiles []config.Profile
	if err := json.Unmarshal(data, &profiles); err == nil {
		return profileTransferFile{
			Version:  config.CurrentVersion,
			Profiles: profiles,
		}, nil
	}

	var profile config.Profile
	if err := json.Unmarshal(data, &profile); err == nil {
		return profileTransferFile{
			Version:  config.CurrentVersion,
			Profiles: []config.Profile{profile},
		}, nil
	}

	return profileTransferFile{}, fmt.Errorf("unsupported import format")
}

func runProfileDelete(stdin io.Reader, stdout, stderr io.Writer, store config.Store, args []string) int {
	identifier := ""
	force := false
	for _, arg := range args {
		if arg == "--force" || arg == "-f" {
			force = true
		} else if identifier == "" {
			identifier = arg
		} else {
			fmt.Fprintln(stderr, "usage: ccc profile delete <profile-id-or-name> [--force]")
			return 1
		}
	}
	if identifier == "" {
		fmt.Fprintln(stderr, "usage: ccc profile delete <profile-id-or-name> [--force]")
		return 1
	}

	cfg, _, err := store.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}

	profile, ok := cfg.FindProfile(identifier)
	if !ok {
		fmt.Fprintf(stderr, "profile %q not found. Run 'ccc profile list' to see available profiles.\n", identifier)
		return 1
	}

	if !force && stdinIsInteractive(stdin) {
		reader := bufio.NewReader(stdin)
		confirmed, err := promptWithDefault(reader, stdout, fmt.Sprintf("Delete profile %q (%s)? [y/N]", profile.Name, profile.ID), "n")
		if err != nil {
			return 1
		}
		if strings.ToLower(strings.TrimSpace(confirmed)) != "y" {
			fmt.Fprintln(stdout, "Cancelled.")
			return 0
		}
	}

	removed, _ := cfg.DeleteProfile(identifier)
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
		fmt.Fprintf(stderr, "profile %q not found. Run 'ccc profile list' to see available profiles.\n", args[0])
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
