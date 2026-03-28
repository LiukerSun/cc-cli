package platform

import (
	"path/filepath"
	"testing"
)

func TestResolveLayoutLinuxDefaults(t *testing.T) {
	home := "/home/tester"
	layout, err := ResolveLayout("linux", home, func(string) string { return "" })
	if err != nil {
		t.Fatalf("ResolveLayout returned error: %v", err)
	}

	if got, want := layout.BinDir, filepath.Join(home, ".local", "bin"); got != want {
		t.Fatalf("BinDir = %q, want %q", got, want)
	}
	if got, want := layout.ConfigDir, filepath.Join(home, ".config", "ccc"); got != want {
		t.Fatalf("ConfigDir = %q, want %q", got, want)
	}
	if got, want := layout.DataDir, filepath.Join(home, ".local", "share", "ccc"); got != want {
		t.Fatalf("DataDir = %q, want %q", got, want)
	}
	if got, want := layout.CacheDir, filepath.Join(home, ".cache", "ccc"); got != want {
		t.Fatalf("CacheDir = %q, want %q", got, want)
	}
	if got, want := layout.StateDir, filepath.Join(home, ".local", "state", "ccc"); got != want {
		t.Fatalf("StateDir = %q, want %q", got, want)
	}
	if got, want := layout.ConfigFile(), filepath.Join(home, ".config", "ccc", "config.json"); got != want {
		t.Fatalf("ConfigFile = %q, want %q", got, want)
	}
}

func TestResolveLayoutLinuxRespectsXDG(t *testing.T) {
	home := "/home/tester"
	env := map[string]string{
		"XDG_CONFIG_HOME": "/xdg/config",
		"XDG_DATA_HOME":   "/xdg/data",
		"XDG_CACHE_HOME":  "/xdg/cache",
		"XDG_STATE_HOME":  "/xdg/state",
	}

	layout, err := ResolveLayout("linux", home, func(key string) string { return env[key] })
	if err != nil {
		t.Fatalf("ResolveLayout returned error: %v", err)
	}

	if got, want := layout.ConfigDir, filepath.Join("/xdg/config", "ccc"); got != want {
		t.Fatalf("ConfigDir = %q, want %q", got, want)
	}
	if got, want := layout.DataDir, filepath.Join("/xdg/data", "ccc"); got != want {
		t.Fatalf("DataDir = %q, want %q", got, want)
	}
	if got, want := layout.CacheDir, filepath.Join("/xdg/cache", "ccc"); got != want {
		t.Fatalf("CacheDir = %q, want %q", got, want)
	}
	if got, want := layout.StateDir, filepath.Join("/xdg/state", "ccc"); got != want {
		t.Fatalf("StateDir = %q, want %q", got, want)
	}
}

func TestResolveLayoutWindowsDefaults(t *testing.T) {
	home := `C:\Users\tester`
	layout, err := ResolveLayout("windows", home, func(string) string { return "" })
	if err != nil {
		t.Fatalf("ResolveLayout returned error: %v", err)
	}

	if got, want := layout.BinDir, filepath.Join(home, "AppData", "Local", "Programs", "ccc", "bin"); got != want {
		t.Fatalf("BinDir = %q, want %q", got, want)
	}
	if got, want := layout.ConfigDir, filepath.Join(home, "AppData", "Roaming", "ccc"); got != want {
		t.Fatalf("ConfigDir = %q, want %q", got, want)
	}
	if got, want := layout.DataDir, filepath.Join(home, "AppData", "Local", "ccc", "data"); got != want {
		t.Fatalf("DataDir = %q, want %q", got, want)
	}
}
