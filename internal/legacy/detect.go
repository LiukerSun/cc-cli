package legacy

import (
	"os"
	"path/filepath"
)

type Candidate struct {
	Label  string `json:"label"`
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
}

type Detection struct {
	NeedsMigration bool        `json:"needs_migration"`
	Candidates     []Candidate `json:"candidates"`
}

func Detect(home string) Detection {
	candidates := []Candidate{
		newCandidate("legacy root directory", filepath.Join(home, ".ccc")),
		newCandidate("legacy install directory", filepath.Join(home, ".cc-cli")),
		newCandidate("legacy config file", filepath.Join(home, ".cc-config.json")),
		newCandidate("legacy launcher", filepath.Join(home, "bin", "ccc")),
	}

	detection := Detection{Candidates: candidates}
	for _, candidate := range candidates {
		if candidate.Exists {
			detection.NeedsMigration = true
			break
		}
	}

	return detection
}

func newCandidate(label, path string) Candidate {
	_, err := os.Stat(path)
	return Candidate{
		Label:  label,
		Path:   path,
		Exists: err == nil,
	}
}
