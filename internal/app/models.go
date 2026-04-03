package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

var fetchZhipuModels = defaultFetchZhipuModels
var fetchAlibabaModels = defaultFetchAlibabaModels

var interactiveAnthropicModels = []string{
	"claude-opus-4-6",
	"claude-sonnet-4-6",
	"claude-haiku-4-5",
	"claude-3-7-sonnet",
	"claude-3-5-haiku",
}

var interactiveCodexModels = []string{
	"gpt-5.4",
	"gpt-5.4-mini",
	"gpt-5.3-codex",
	"gpt-5.3-codex-spark",
	"gpt-5.2-codex",
	"gpt-5.2",
	"gpt-5.1-codex-max",
	"gpt-5.1",
	"gpt-5.1-codex",
	"gpt-5",
	"gpt-5-codex",
	"gpt-5-codex-mini",
}

var interactiveZhipuFallbackModels = []string{
	"glm-5",
	"glm-4.7",
	"glm-4.7-air",
	"glm-4.7-airx",
}

var interactiveAlibabaFallbackModels = []string{
	"qwen3.5-plus",
	"qwen3-max-2026-01-23",
	"qwen3-coder-next",
	"qwen3-coder-plus",
	"glm-5",
	"glm-4.7",
	"kimi-k2.5",
	"minimax-m2.5",
}

func defaultFetchZhipuModels(apiKey string) ([]string, error) {
	models, err := fetchModels("GET", "https://open.bigmodel.cn/api/paas/v4/models", apiKey, parseIDModelList)
	if err != nil {
		return nil, err
	}
	return uniqueStrings(append(models, interactiveZhipuFallbackModels...)), nil
}

func defaultFetchAlibabaModels(apiKey string) ([]string, error) {
	models, err := fetchModels("GET", "https://dashscope.aliyuncs.com/api/v1/models", apiKey, parseAlibabaModelList)
	if err != nil {
		return nil, err
	}
	filtered := make([]string, 0, len(models))
	for _, model := range models {
		normalized := strings.ToLower(strings.TrimSpace(model))
		if strings.HasPrefix(normalized, "qwen") || strings.HasPrefix(normalized, "glm") || strings.HasPrefix(normalized, "kimi") || strings.HasPrefix(normalized, "minimax") {
			filtered = append(filtered, model)
		}
	}
	return uniqueStrings(append(filtered, interactiveAlibabaFallbackModels...)), nil
}

func fetchModels(method, url, apiKey string, parser func([]byte) ([]string, error)) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}
	return parser(body)
}

func parseIDModelList(body []byte) ([]string, error) {
	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	models := make([]string, 0, len(payload.Data))
	for _, item := range payload.Data {
		if strings.TrimSpace(item.ID) != "" {
			models = append(models, item.ID)
		}
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("no models found")
	}
	sort.Strings(models)
	return models, nil
}

func parseAlibabaModelList(body []byte) ([]string, error) {
	var payload struct {
		Data []struct {
			ID    string `json:"id"`
			Model string `json:"model"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	models := make([]string, 0, len(payload.Data))
	for _, item := range payload.Data {
		if strings.TrimSpace(item.Model) != "" {
			models = append(models, item.Model)
			continue
		}
		if strings.TrimSpace(item.ID) != "" {
			models = append(models, item.ID)
		}
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("no models found")
	}
	sort.Strings(models)
	return models, nil
}

func uniqueStrings(values []string) []string {
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
	return out
}

func indexOf(values []string, target string) int {
	for i, value := range values {
		if value == target {
			return i
		}
	}
	return -1
}
