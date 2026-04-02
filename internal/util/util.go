package util

import (
	"path/filepath"
	"strings"
)

func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func NormalizeCodexBaseURL(baseURL string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	switch {
	case strings.HasSuffix(trimmed, "/v1"):
		return trimmed
	case strings.HasSuffix(trimmed, "/v1/models"):
		return strings.TrimSuffix(trimmed, "/models")
	case strings.HasSuffix(trimmed, "/models"):
		return strings.TrimSuffix(trimmed, "/models") + "/v1"
	case strings.HasSuffix(trimmed, "/responses"):
		return strings.TrimSuffix(trimmed, "/responses") + "/v1"
	case trimmed == "":
		return "/v1"
	default:
		return trimmed + "/v1"
	}
}

func PathContains(pathValue, target string) bool {
	cleanTarget := filepath.Clean(target)
	for _, entry := range filepath.SplitList(pathValue) {
		if filepath.Clean(strings.TrimSpace(entry)) == cleanTarget {
			return true
		}
	}
	return false
}
