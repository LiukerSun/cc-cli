package util

import (
	"os"
	"testing"
)

func TestFirstNonEmpty(t *testing.T) {
	if got := FirstNonEmpty("", "  ", " demo ", "fallback"); got != "demo" {
		t.Fatalf("FirstNonEmpty() = %q, want demo", got)
	}
}

func TestNormalizeCodexBaseURL(t *testing.T) {
	tests := map[string]string{
		"https://api.openai.com/v1":        "https://api.openai.com/v1",
		"https://relay.example.com/models": "https://relay.example.com/v1",
		"https://relay.example.com/v1/":    "https://relay.example.com/v1",
		"":                                 "/v1",
	}
	for input, want := range tests {
		if got := NormalizeCodexBaseURL(input); got != want {
			t.Fatalf("NormalizeCodexBaseURL(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestPathContains(t *testing.T) {
	pathValue := "/tmp/one" + string(os.PathListSeparator) + " /tmp/two "
	if !PathContains(pathValue, "/tmp/two") {
		t.Fatal("expected PathContains to find normalized path")
	}
	if PathContains(pathValue, "/tmp/three") {
		t.Fatal("expected PathContains to reject missing path")
	}
}
