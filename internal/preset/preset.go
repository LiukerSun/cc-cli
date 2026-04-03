package preset

import (
	"fmt"
	"sort"
	"strings"

	"github.com/LiukerSun/cc-cli/internal/config"
)

type Definition struct {
	Name      string
	Provider  string
	Command   string
	BaseURL   string
	Model     string
	FastModel string
}

var definitions = map[string]Definition{
	"anthropic": {
		Name:      "Claude Official",
		Provider:  "anthropic",
		Command:   "claude",
		BaseURL:   "https://api.anthropic.com",
		Model:     "claude-sonnet-4-6",
		FastModel: "claude-haiku-4-5",
	},
	"openai": {
		Name:      "Codex OpenAI",
		Provider:  "openai",
		Command:   "codex",
		BaseURL:   "https://api.openai.com/v1",
		Model:     "gpt-5.4",
		FastModel: "gpt-5.4-mini",
	},
	"zhipu": {
		Name:      "Zhipu Claude-Compatible",
		Provider:  "zhipu",
		Command:   "claude",
		BaseURL:   "https://open.bigmodel.cn/api/anthropic",
		Model:     "glm-5",
		FastModel: "glm-4.7",
	},
	"alibaba": {
		Name:      "Alibaba Coding Plan",
		Provider:  "alibaba",
		Command:   "claude",
		BaseURL:   "https://coding.dashscope.aliyuncs.com/apps/anthropic",
		Model:     "qwen3.5-plus",
		FastModel: "qwen3.5-plus",
	},
}

var aliases = map[string]string{
	"claude":    "anthropic",
	"codex":     "openai",
	"gpt":       "openai",
	"openai":    "openai",
	"zai":       "zhipu",
	"glm":       "zhipu",
	"zhipu":     "zhipu",
	"qwen":      "alibaba",
	"dashscope": "alibaba",
	"tongyi":    "alibaba",
	"alibaba":   "alibaba",
}

func Apply(profile config.Profile, name string) (config.Profile, error) {
	definition, err := Lookup(name)
	if err != nil {
		return profile, err
	}
	if strings.TrimSpace(name) == "" {
		return profile, nil
	}

	if strings.TrimSpace(profile.Name) == "" {
		profile.Name = definition.Name
	}
	if strings.TrimSpace(profile.Provider) == "" {
		profile.Provider = definition.Provider
	}
	if strings.TrimSpace(profile.Command) == "" {
		profile.Command = definition.Command
	}
	if strings.TrimSpace(profile.BaseURL) == "" {
		profile.BaseURL = definition.BaseURL
	}
	if strings.TrimSpace(profile.Model) == "" {
		profile.Model = definition.Model
	}
	if strings.TrimSpace(profile.FastModel) == "" {
		profile.FastModel = definition.FastModel
	}

	return profile, nil
}

func Lookup(name string) (Definition, error) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return Definition{}, nil
	}
	if aliased, ok := aliases[name]; ok {
		name = aliased
	}

	definition, ok := definitions[name]
	if !ok {
		return Definition{}, fmt.Errorf("unknown preset %q (supported: %s)", name, strings.Join(Names(), ", "))
	}
	return definition, nil
}

func Names() []string {
	names := make([]string, 0, len(definitions))
	for name := range definitions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
