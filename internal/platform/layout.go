package platform

import (
	"errors"
	"path/filepath"
)

type EnvLookup func(string) string

type Layout struct {
	HomeDir   string `json:"home_dir"`
	BinDir    string `json:"bin_dir"`
	ConfigDir string `json:"config_dir"`
	DataDir   string `json:"data_dir"`
	CacheDir  string `json:"cache_dir"`
	StateDir  string `json:"state_dir"`
}

func ResolveLayout(goos, home string, getenv EnvLookup) (Layout, error) {
	if home == "" {
		return Layout{}, errors.New("home directory is required")
	}

	if getenv == nil {
		getenv = func(string) string { return "" }
	}

	if goos == "windows" {
		appData := firstNonEmpty(getenv("APPDATA"), filepath.Join(home, "AppData", "Roaming"))
		localAppData := firstNonEmpty(getenv("LOCALAPPDATA"), filepath.Join(home, "AppData", "Local"))

		return Layout{
			HomeDir:   home,
			BinDir:    filepath.Join(localAppData, "Programs", "ccc", "bin"),
			ConfigDir: filepath.Join(appData, "ccc"),
			DataDir:   filepath.Join(localAppData, "ccc", "data"),
			CacheDir:  filepath.Join(localAppData, "ccc", "cache"),
			StateDir:  filepath.Join(localAppData, "ccc", "state"),
		}, nil
	}

	configBase := firstNonEmpty(getenv("XDG_CONFIG_HOME"), filepath.Join(home, ".config"))
	dataBase := firstNonEmpty(getenv("XDG_DATA_HOME"), filepath.Join(home, ".local", "share"))
	cacheBase := firstNonEmpty(getenv("XDG_CACHE_HOME"), filepath.Join(home, ".cache"))
	stateBase := firstNonEmpty(getenv("XDG_STATE_HOME"), filepath.Join(home, ".local", "state"))

	return Layout{
		HomeDir:   home,
		BinDir:    filepath.Join(home, ".local", "bin"),
		ConfigDir: filepath.Join(configBase, "ccc"),
		DataDir:   filepath.Join(dataBase, "ccc"),
		CacheDir:  filepath.Join(cacheBase, "ccc"),
		StateDir:  filepath.Join(stateBase, "ccc"),
	}, nil
}

func (l Layout) ConfigFile() string {
	return filepath.Join(l.ConfigDir, "config.json")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
