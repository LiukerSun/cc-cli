package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/LiukerSun/cc-cli/internal/platform"
)

const CurrentVersion = 1

var profileIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
var validCommandNames = []string{"claude", "codex"}

type File struct {
	Version        int       `json:"version"`
	CurrentProfile string    `json:"current_profile,omitempty"`
	Profiles       []Profile `json:"profiles"`
}

type Profile struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	Provider            string            `json:"provider,omitempty"`
	Command             string            `json:"command"`
	BaseURL             string            `json:"base_url"`
	APIKey              string            `json:"api_key"`
	Model               string            `json:"model"`
	FastModel           string            `json:"fast_model,omitempty"`
	ExtraEnv            map[string]string `json:"env,omitempty"`
	SyncExternal        bool              `json:"sync_external"`
	SyncDenyPermissions []string          `json:"sync_deny_permissions,omitempty"`
}

type Metadata struct {
	Path   string `json:"path"`
	Source string `json:"source"`
	Exists bool   `json:"exists"`
}

type Store struct {
	Home   string
	Layout platform.Layout
}

type legacyProfile struct {
	Name    string            `json:"name"`
	Command string            `json:"command"`
	Env     map[string]string `json:"env"`
}

func NewStore(home string, layout platform.Layout) Store {
	return Store{Home: home, Layout: layout}
}

func DefaultFile() File {
	return File{
		Version:  CurrentVersion,
		Profiles: []Profile{},
	}
}

func (s Store) Load() (File, Metadata, error) {
	newPath := s.Layout.ConfigFile()
	if fileExists(newPath) {
		cfg, err := decodeFile(newPath)
		return cfg, Metadata{Path: newPath, Source: "current", Exists: true}, err
	}

	legacyCandidates := []Metadata{
		{Path: filepath.Join(s.Home, ".ccc", "config.json"), Source: "legacy-root"},
		{Path: filepath.Join(s.Home, ".cc-config.json"), Source: "legacy-file"},
	}

	for _, candidate := range legacyCandidates {
		if fileExists(candidate.Path) {
			cfg, err := decodeFile(candidate.Path)
			if err != nil {
				return File{}, Metadata{}, err
			}
			return cfg, Metadata{Path: candidate.Path, Source: candidate.Source, Exists: true}, nil
		}
	}

	return DefaultFile(), Metadata{Path: newPath, Source: "default", Exists: false}, nil
}

func (s Store) Save(cfg File) error {
	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return err
	}

	if err := os.MkdirAll(s.Layout.ConfigDir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	data = append(data, '\n')

	tmpPath := s.Layout.ConfigFile() + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := os.Rename(tmpPath, s.Layout.ConfigFile()); err != nil {
		return fmt.Errorf("replace config: %w", err)
	}
	return nil
}

func (f *File) normalize() {
	if f.Version == 0 {
		f.Version = CurrentVersion
	}
	for i := range f.Profiles {
		f.Profiles[i].normalize()
	}
}

func (f File) Validate() error {
	if f.Version <= 0 {
		return errors.New("config version must be greater than 0")
	}

	seen := map[string]struct{}{}
	for i, profile := range f.Profiles {
		if err := profile.Validate(); err != nil {
			return fmt.Errorf("profile %d: %w", i+1, err)
		}
		if _, ok := seen[profile.ID]; ok {
			return fmt.Errorf("duplicate profile id %q", profile.ID)
		}
		seen[profile.ID] = struct{}{}
	}

	if f.CurrentProfile != "" {
		if _, ok := seen[f.CurrentProfile]; !ok {
			return fmt.Errorf("current profile %q does not exist", f.CurrentProfile)
		}
	}
	return nil
}

func (p *Profile) normalize() {
	p.Name = strings.TrimSpace(p.Name)
	p.ID = strings.TrimSpace(p.ID)
	p.Provider = strings.TrimSpace(p.Provider)
	p.Command = strings.TrimSpace(strings.ToLower(p.Command))
	p.BaseURL = strings.TrimSpace(p.BaseURL)
	p.APIKey = strings.TrimSpace(p.APIKey)
	p.Model = strings.TrimSpace(p.Model)
	p.FastModel = strings.TrimSpace(p.FastModel)
	p.SyncDenyPermissions = normalizeStringList(p.SyncDenyPermissions)

	if p.Command == "" {
		p.Command = "claude"
	}
	if p.Provider == "" {
		p.Provider = "custom"
	}
	if p.ID == "" {
		p.ID = MakeProfileID(p.Name)
	}
	if p.ExtraEnv == nil {
		p.ExtraEnv = map[string]string{}
	}
}

func (p Profile) Validate() error {
	if p.Name == "" {
		return errors.New("name is required")
	}
	if p.ID == "" {
		return errors.New("id is required")
	}
	if !profileIDPattern.MatchString(p.ID) {
		return fmt.Errorf("id %q must match [a-z0-9-]", p.ID)
	}
	if !slices.Contains(validCommandNames, p.Command) {
		return fmt.Errorf("unsupported command %q", p.Command)
	}
	if p.BaseURL == "" {
		return errors.New("base_url is required")
	}
	if p.APIKey == "" {
		return errors.New("api_key is required")
	}
	if p.Model == "" {
		return errors.New("model is required")
	}
	return nil
}

func (f File) Redacted() File {
	redacted := f
	redacted.Profiles = make([]Profile, len(f.Profiles))
	for i, profile := range f.Profiles {
		redacted.Profiles[i] = profile.Redacted()
	}
	return redacted
}

func (p Profile) Redacted() Profile {
	redacted := p
	redacted.APIKey = MaskSecret(redacted.APIKey)
	return redacted
}

func MaskSecret(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= 4 {
		return "****"
	}
	return "****" + string(runes[len(runes)-4:])
}

func (f *File) UpsertProfile(profile Profile) error {
	profile.normalize()
	if err := profile.Validate(); err != nil {
		return err
	}

	for i := range f.Profiles {
		if f.Profiles[i].ID == profile.ID {
			f.Profiles[i] = profile
			return nil
		}
	}

	f.Profiles = append(f.Profiles, profile)
	return nil
}

func (f *File) DeleteProfile(identifier string) (Profile, bool) {
	for i, profile := range f.Profiles {
		if profile.ID == identifier || profile.Name == identifier {
			f.Profiles = append(f.Profiles[:i], f.Profiles[i+1:]...)
			if f.CurrentProfile == profile.ID {
				f.CurrentProfile = ""
			}
			return profile, true
		}
	}
	return Profile{}, false
}

func (f File) FindProfile(identifier string) (Profile, bool) {
	for _, profile := range f.Profiles {
		if profile.ID == identifier || profile.Name == identifier {
			return profile, true
		}
	}
	return Profile{}, false
}

func (f *File) ReplaceProfile(identifier string, profile Profile) (Profile, bool, error) {
	profile.normalize()
	if err := profile.Validate(); err != nil {
		return Profile{}, false, err
	}

	index := -1
	var previous Profile
	for i, existing := range f.Profiles {
		if existing.ID == identifier || existing.Name == identifier {
			index = i
			previous = existing
			break
		}
	}
	if index == -1 {
		return Profile{}, false, nil
	}

	for i, existing := range f.Profiles {
		if i == index {
			continue
		}
		if existing.ID == profile.ID {
			return previous, true, fmt.Errorf("duplicate profile id %q", profile.ID)
		}
	}

	f.Profiles[index] = profile
	if f.CurrentProfile == previous.ID {
		f.CurrentProfile = profile.ID
	}
	return previous, true, nil
}

func (f File) EnsureUniqueProfileID(profile Profile) Profile {
	profile.normalize()
	baseID := profile.ID
	if baseID == "" {
		baseID = "profile"
	}

	id := baseID
	index := 2
	for {
		if _, exists := f.FindProfile(id); !exists {
			profile.ID = id
			return profile
		}
		id = fmt.Sprintf("%s-%d", baseID, index)
		index++
	}
}

func MakeProfileID(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return "profile"
	}

	var builder strings.Builder
	lastDash := false
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
			lastDash = false
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				builder.WriteByte('-')
				lastDash = true
			}
		}
	}

	id := strings.Trim(builder.String(), "-")
	if id == "" {
		return "profile"
	}
	return id
}

func decodeFile(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return File{}, fmt.Errorf("read config %s: %w", path, err)
	}

	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return DefaultFile(), nil
	}

	if strings.HasPrefix(trimmed, "[") {
		var legacyProfiles []legacyProfile
		if err := json.Unmarshal(data, &legacyProfiles); err != nil {
			return File{}, fmt.Errorf("parse legacy config %s: %w", path, err)
		}
		cfg := DefaultFile()
		for _, legacyProfile := range legacyProfiles {
			cfg.Profiles = append(cfg.Profiles, convertLegacyProfile(legacyProfile))
		}
		cfg.normalize()
		if err := cfg.Validate(); err != nil {
			return File{}, err
		}
		return cfg, nil
	}

	var cfg File
	if err := json.Unmarshal(data, &cfg); err != nil {
		return File{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return File{}, err
	}
	return cfg, nil
}

func convertLegacyProfile(old legacyProfile) Profile {
	command := strings.TrimSpace(strings.ToLower(old.Command))
	if command == "" {
		command = "claude"
	}

	profile := Profile{
		ID:           MakeProfileID(old.Name),
		Name:         strings.TrimSpace(old.Name),
		Command:      command,
		Provider:     "legacy-import",
		ExtraEnv:     map[string]string{},
		SyncExternal: true,
	}

	switch command {
	case "codex":
		profile.BaseURL = strings.TrimSpace(old.Env["OPENAI_BASE_URL"])
		profile.APIKey = strings.TrimSpace(old.Env["OPENAI_API_KEY"])
		profile.Model = strings.TrimSpace(old.Env["OPENAI_MODEL"])
		profile.FastModel = strings.TrimSpace(old.Env["OPENAI_SMALL_FAST_MODEL"])
		for key, value := range old.Env {
			if key == "OPENAI_BASE_URL" || key == "OPENAI_API_KEY" || key == "OPENAI_MODEL" || key == "OPENAI_SMALL_FAST_MODEL" {
				continue
			}
			profile.ExtraEnv[key] = value
		}
	default:
		profile.BaseURL = strings.TrimSpace(old.Env["ANTHROPIC_BASE_URL"])
		profile.APIKey = strings.TrimSpace(old.Env["ANTHROPIC_AUTH_TOKEN"])
		profile.Model = strings.TrimSpace(old.Env["ANTHROPIC_MODEL"])
		profile.FastModel = strings.TrimSpace(old.Env["ANTHROPIC_SMALL_FAST_MODEL"])
		for key, value := range old.Env {
			if key == "ANTHROPIC_BASE_URL" || key == "ANTHROPIC_AUTH_TOKEN" || key == "ANTHROPIC_MODEL" || key == "ANTHROPIC_SMALL_FAST_MODEL" {
				continue
			}
			profile.ExtraEnv[key] = value
		}
	}

	profile.normalize()
	return profile
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func normalizeStringList(values []string) []string {
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
	if len(out) == 0 {
		return nil
	}
	return out
}
