package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LiukerSun/cc-cli/internal/config"
	"github.com/LiukerSun/cc-cli/internal/util"
	toml "github.com/pelletier/go-toml/v2"
)

type Result struct {
	Paths []string `json:"paths"`
}

func Apply(home string, profile config.Profile) (Result, error) {
	switch profile.Command {
	case "codex":
		return applyCodex(home, profile)
	default:
		return applyClaude(home, profile)
	}
}

func TargetPaths(home string, profile config.Profile) []string {
	switch profile.Command {
	case "codex":
		return []string{
			filepath.Join(home, ".codex", "config.toml"),
			filepath.Join(home, ".codex", "auth.json"),
		}
	default:
		return []string{
			filepath.Join(home, ".claude", "settings.json"),
		}
	}
}

func applyClaude(home string, profile config.Profile) (Result, error) {
	settingsPath := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return Result{}, fmt.Errorf("create claude config dir: %w", err)
	}

	doc := map[string]any{}
	if data, err := os.ReadFile(settingsPath); err == nil && len(strings.TrimSpace(string(data))) > 0 {
		if err := json.Unmarshal(data, &doc); err != nil {
			doc = map[string]any{}
		}
	}

	envMap := ensureMap(doc, "env")
	fastModel := util.FirstNonEmpty(profile.FastModel, profile.Model)
	envMap["ANTHROPIC_MODEL"] = profile.Model
	envMap["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = fastModel
	envMap["ANTHROPIC_SMALL_FAST_MODEL"] = fastModel
	envMap["CLAUDE_CODE_MODEL"] = profile.Model
	envMap["CLAUDE_CODE_SMALL_MODEL"] = fastModel
	if profile.SubagentModel != "" {
		envMap["CLAUDE_CODE_SUBAGENT_MODEL"] = profile.SubagentModel
	}
	doc["model"] = profile.Model

	if len(profile.SyncDenyPermissions) > 0 {
		permissions := ensureMap(doc, "permissions")
		for _, permission := range profile.SyncDenyPermissions {
			permissions["deny"] = appendUniqueString(permissions["deny"], permission)
		}
	}

	if err := writeJSON(settingsPath, doc, 0o600); err != nil {
		return Result{}, err
	}
	return Result{Paths: []string{settingsPath}}, nil
}

func applyCodex(home string, profile config.Profile) (Result, error) {
	configPath := filepath.Join(home, ".codex", "config.toml")
	authPath := filepath.Join(home, ".codex", "auth.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return Result{}, fmt.Errorf("create codex config dir: %w", err)
	}

	configDoc := map[string]any{}
	if data, err := os.ReadFile(configPath); err == nil && len(strings.TrimSpace(string(data))) > 0 {
		if err := toml.Unmarshal(data, &configDoc); err != nil {
			configDoc = map[string]any{}
		}
	}

	configDoc["model_provider"] = "codex"
	configDoc["model"] = profile.Model
	if _, ok := configDoc["model_reasoning_effort"]; !ok {
		configDoc["model_reasoning_effort"] = "high"
	}
	if _, ok := configDoc["disable_response_storage"]; !ok {
		configDoc["disable_response_storage"] = true
	}
	delete(configDoc, "openai_base_url")

	providers := ensureMap(configDoc, "model_providers")
	codexProvider := ensureNestedMap(providers, "codex")
	codexProvider["name"] = "codex"
	codexProvider["base_url"] = util.NormalizeCodexBaseURL(profile.BaseURL)
	codexProvider["wire_api"] = "responses"

	if err := writeTOML(configPath, configDoc, 0o600); err != nil {
		return Result{}, err
	}

	authDoc := map[string]any{}
	if data, err := os.ReadFile(authPath); err == nil && len(strings.TrimSpace(string(data))) > 0 {
		if err := json.Unmarshal(data, &authDoc); err != nil {
			authDoc = map[string]any{}
		}
	}
	authDoc["OPENAI_API_KEY"] = profile.APIKey
	if err := writeJSON(authPath, authDoc, 0o600); err != nil {
		return Result{}, err
	}

	return Result{Paths: []string{configPath, authPath}}, nil
}

func ensureMap(doc map[string]any, key string) map[string]any {
	if existing, ok := doc[key].(map[string]any); ok {
		return existing
	}
	newMap := map[string]any{}
	doc[key] = newMap
	return newMap
}

func ensureNestedMap(doc map[string]any, key string) map[string]any {
	if existing, ok := doc[key].(map[string]any); ok {
		return existing
	}
	if existing, ok := doc[key].(map[string]interface{}); ok {
		out := map[string]any{}
		for k, v := range existing {
			out[k] = v
		}
		doc[key] = out
		return out
	}
	newMap := map[string]any{}
	doc[key] = newMap
	return newMap
}

func appendUniqueString(value any, item string) []string {
	items := []string{}
	switch typed := value.(type) {
	case []string:
		items = append(items, typed...)
	case []any:
		for _, entry := range typed {
			if text, ok := entry.(string); ok {
				items = append(items, text)
			}
		}
	}

	for _, existing := range items {
		if existing == item {
			return items
		}
	}
	return append(items, item)
}

func writeJSON(path string, payload any, mode os.FileMode) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json %s: %w", path, err)
	}
	data = append(data, '\n')
	return writeFileAtomically(path, data, mode)
}

func writeTOML(path string, payload any, mode os.FileMode) error {
	data, err := toml.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal toml %s: %w", path, err)
	}
	return writeFileAtomically(path, data, mode)
}

func writeFileAtomically(path string, data []byte, mode os.FileMode) error {
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, mode); err != nil {
		return fmt.Errorf("write temp %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("replace %s: %w", path, err)
	}
	return nil
}
