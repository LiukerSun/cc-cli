package deps

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ToolStatus struct {
	Name      string `json:"name"`
	Installed bool   `json:"installed"`
	Path      string `json:"path,omitempty"`
	Version   string `json:"version,omitempty"`
}

type CommandSpec struct {
	PackageName    string
	MinNodeVersion string
}

var (
	lookPath = exec.LookPath
	runCmd   = runCommand
)

func InspectTool(name string, versionArgs ...string) ToolStatus {
	status := ToolStatus{Name: name}

	path, err := lookPath(name)
	if err != nil {
		return status
	}

	status.Installed = true
	status.Path = path

	if len(versionArgs) == 0 {
		versionArgs = []string{"--version"}
	}

	output, err := runCmd(path, versionArgs...)
	if err == nil {
		status.Version = strings.TrimSpace(firstLine(output))
	}

	return status
}

func EnsureCLI(command string, stdout, stderr io.Writer) error {
	if _, err := lookPath(command); err == nil {
		return nil
	}

	spec, ok := SpecFor(command)
	if !ok {
		return fmt.Errorf("unsupported CLI command %q", command)
	}

	nodeStatus := InspectTool("node", "--version")
	if !nodeStatus.Installed {
		return fmt.Errorf("cannot install %s CLI automatically because Node.js is not installed\nCurrent version: not installed\nMinimum required version: Node.js >= %s\nPlease install or upgrade Node.js, then rerun ccc", command, spec.MinNodeVersion)
	}

	npmStatus := InspectTool("npm", "--version")
	if !npmStatus.Installed {
		return fmt.Errorf("cannot install %s CLI automatically because npm is not installed\nCurrent version: Node.js %s\nMinimum required version: Node.js >= %s\nPlease install npm (usually bundled with Node.js), then rerun ccc", command, nodeStatus.Version, spec.MinNodeVersion)
	}

	if VersionLessThan(nodeStatus.Version, spec.MinNodeVersion) {
		return fmt.Errorf("Node.js version is too old to install %s CLI\nCurrent version: %s\nMinimum required version: Node.js >= %s\nPlease upgrade Node.js, then rerun ccc", command, nodeStatus.Version, spec.MinNodeVersion)
	}

	prependNPMGlobalBinToPath()

	fmt.Fprintf(stdout, "%s CLI not found. Attempting to install %s via npm...\n", command, spec.PackageName)
	if _, err := runCmd("npm", "install", "-g", spec.PackageName); err != nil {
		return fmt.Errorf("failed to install %s CLI automatically\nTry manually: npm install -g %s", command, spec.PackageName)
	}

	prependNPMGlobalBinToPath()
	if _, err := lookPath(command); err != nil {
		return fmt.Errorf("%s CLI is still not available after installation\nTry manually: npm install -g %s", command, spec.PackageName)
	}

	fmt.Fprintf(stdout, "Installed %s CLI\n", command)
	_ = stderr
	return nil
}

func SpecFor(command string) (CommandSpec, bool) {
	switch command {
	case "claude":
		return CommandSpec{
			PackageName:    "@anthropic-ai/claude-code",
			MinNodeVersion: "18.0.0",
		}, true
	case "codex":
		return CommandSpec{
			PackageName:    "@openai/codex",
			MinNodeVersion: "16.0.0",
		}, true
	default:
		return CommandSpec{}, false
	}
}

func FirstVersionToken(input string) string {
	for _, field := range strings.Fields(input) {
		clean := strings.TrimPrefix(field, "v")
		if clean != "" && containsDigit(clean) {
			return strings.TrimRight(clean, ",;")
		}
	}
	return ""
}

func VersionLessThan(left, right string) bool {
	leftParts := normalizeVersion(left)
	rightParts := normalizeVersion(right)
	for i := 0; i < 3; i++ {
		if leftParts[i] < rightParts[i] {
			return true
		}
		if leftParts[i] > rightParts[i] {
			return false
		}
	}
	return false
}

func normalizeVersion(input string) [3]int {
	token := FirstVersionToken(input)
	token = strings.TrimSpace(strings.TrimPrefix(token, ">="))
	var out [3]int
	if token == "" {
		return out
	}

	parts := strings.Split(token, ".")
	for i := 0; i < len(parts) && i < 3; i++ {
		value := 0
		for _, r := range parts[i] {
			if r < '0' || r > '9' {
				break
			}
			value = value*10 + int(r-'0')
		}
		out[i] = value
	}
	return out
}

func prependNPMGlobalBinToPath() {
	output, err := runCmd("npm", "config", "get", "prefix")
	if err != nil {
		return
	}

	npmPrefix := strings.TrimSpace(output)
	if npmPrefix == "" || npmPrefix == "undefined" {
		return
	}

	candidates := []string{
		filepath.Join(npmPrefix, "bin"),
		npmPrefix,
	}
	currentPath := os.Getenv("PATH")
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if info, err := os.Stat(candidate); err != nil || !info.IsDir() {
			continue
		}
		if pathContains(currentPath, candidate) {
			return
		}
		_ = os.Setenv("PATH", candidate+string(os.PathListSeparator)+currentPath)
		return
	}
}

func pathContains(pathValue, target string) bool {
	cleanTarget := filepath.Clean(target)
	for _, entry := range filepath.SplitList(pathValue) {
		if filepath.Clean(strings.TrimSpace(entry)) == cleanTarget {
			return true
		}
	}
	return false
}

func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	return output.String(), err
}

func firstLine(input string) string {
	for _, line := range strings.Split(input, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

func containsDigit(input string) bool {
	for _, r := range input {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}
