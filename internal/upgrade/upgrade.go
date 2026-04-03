package upgrade

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultRepoOwner   = "LiukerSun"
	defaultRepoName    = "cc-cli"
	defaultProjectName = "ccc"
)

var (
	renameFile                 = os.Rename
	removeFile                 = os.Remove
	scheduleWindowsCleanupFile = defaultScheduleWindowsCleanupFile
)

type Manager struct {
	RepoOwner       string
	RepoName        string
	ProjectName     string
	CurrentVersion  string
	ExecutablePath  string
	GOOS            string
	GOARCH          string
	APIBaseURL      string
	DownloadBaseURL string
	HTTPClient      *http.Client
}

type Plan struct {
	CurrentVersion string
	TargetVersion  string
	Tag            string
	AssetName      string
	AssetURL       string
	ChecksumsURL   string
	ExecutablePath string
}

type latestReleaseResponse struct {
	TagName string `json:"tag_name"`
}

func DefaultManager(currentVersion, executablePath, goos, goarch string) Manager {
	apiBaseURL := strings.TrimRight(os.Getenv("CCC_RELEASE_API_BASE_URL"), "/")
	if apiBaseURL == "" {
		apiBaseURL = "https://api.github.com"
	}

	downloadBaseURL := strings.TrimRight(os.Getenv("CCC_RELEASE_DOWNLOAD_BASE_URL"), "/")
	if downloadBaseURL == "" {
		downloadBaseURL = "https://github.com/" + defaultRepoOwner + "/" + defaultRepoName + "/releases/download"
	}

	return Manager{
		RepoOwner:       defaultRepoOwner,
		RepoName:        defaultRepoName,
		ProjectName:     defaultProjectName,
		CurrentVersion:  normalizeVersion(currentVersion),
		ExecutablePath:  executablePath,
		GOOS:            goos,
		GOARCH:          goarch,
		APIBaseURL:      apiBaseURL,
		DownloadBaseURL: downloadBaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (m Manager) Plan(ctx context.Context, requestedVersion string) (Plan, error) {
	if m.ProjectName == "" {
		m.ProjectName = defaultProjectName
	}
	if m.RepoOwner == "" {
		m.RepoOwner = defaultRepoOwner
	}
	if m.RepoName == "" {
		m.RepoName = defaultRepoName
	}
	if m.GOOS == "" || m.GOARCH == "" {
		return Plan{}, errors.New("goos and goarch are required")
	}
	if m.ExecutablePath == "" {
		return Plan{}, errors.New("executable path is required")
	}
	if m.HTTPClient == nil {
		m.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}

	targetVersion := normalizeVersion(requestedVersion)
	if targetVersion == "" || targetVersion == "latest" {
		latest, err := m.fetchLatestVersion(ctx)
		if err != nil {
			return Plan{}, err
		}
		targetVersion = latest
	}

	assetName, err := m.assetName()
	if err != nil {
		return Plan{}, err
	}
	tag := "v" + targetVersion
	return Plan{
		CurrentVersion: m.CurrentVersion,
		TargetVersion:  targetVersion,
		Tag:            tag,
		AssetName:      assetName,
		AssetURL:       strings.TrimRight(m.DownloadBaseURL, "/") + "/" + tag + "/" + assetName,
		ChecksumsURL:   strings.TrimRight(m.DownloadBaseURL, "/") + "/" + tag + "/checksums.txt",
		ExecutablePath: m.ExecutablePath,
	}, nil
}

func (m Manager) Upgrade(ctx context.Context, plan Plan) error {
	archiveBytes, err := m.download(ctx, plan.AssetURL)
	if err != nil {
		return fmt.Errorf("download archive: %w", err)
	}
	checksumBytes, err := m.download(ctx, plan.ChecksumsURL)
	if err != nil {
		return fmt.Errorf("download checksums: %w", err)
	}
	if err := verifyChecksum(plan.AssetName, archiveBytes, checksumBytes); err != nil {
		return err
	}

	binaryName := m.ProjectName
	if m.GOOS == "windows" {
		binaryName += ".exe"
	}
	binary, err := extractBinary(plan.AssetName, archiveBytes, binaryName)
	if err != nil {
		return err
	}
	if err := replaceExecutable(plan.ExecutablePath, binary, m.GOOS); err != nil {
		return err
	}
	return nil
}

func (m Manager) fetchLatestVersion(ctx context.Context) (string, error) {
	url := strings.TrimRight(m.APIBaseURL, "/") + "/repos/" + m.RepoOwner + "/" + m.RepoName + "/releases/latest"
	body, err := m.download(ctx, url)
	if err != nil {
		return "", fmt.Errorf("fetch latest release metadata: %w", err)
	}

	var payload latestReleaseResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", fmt.Errorf("decode latest release metadata: %w", err)
	}
	version := normalizeVersion(payload.TagName)
	if version == "" {
		return "", errors.New("latest release metadata did not include a tag_name")
	}
	return version, nil
}

func (m Manager) assetName() (string, error) {
	switch m.GOOS {
	case "linux", "darwin":
		return fmt.Sprintf("%s_%s_%s.tar.gz", m.ProjectName, m.GOOS, m.GOARCH), nil
	case "windows":
		return fmt.Sprintf("%s_%s_%s.zip", m.ProjectName, m.GOOS, m.GOARCH), nil
	default:
		return "", fmt.Errorf("unsupported operating system %q", m.GOOS)
	}
}

func (m Manager) download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", m.ProjectName+"-upgrade")

	resp, err := m.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

func normalizeVersion(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "refs/tags/")
	value = strings.TrimPrefix(value, "v")
	return value
}

func verifyChecksum(assetName string, archiveBytes, checksumBytes []byte) error {
	var expected string
	lines := strings.Split(string(checksumBytes), "\n")
	for _, line := range lines {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) >= 2 && fields[len(fields)-1] == assetName {
			expected = fields[0]
			break
		}
	}
	if expected == "" {
		return fmt.Errorf("checksums.txt did not include %s", assetName)
	}

	sum := sha256.Sum256(archiveBytes)
	actual := hex.EncodeToString(sum[:])
	if !strings.EqualFold(expected, actual) {
		return fmt.Errorf("checksum mismatch for %s", assetName)
	}
	return nil
}

func extractBinary(assetName string, archiveBytes []byte, binaryName string) ([]byte, error) {
	switch {
	case strings.HasSuffix(assetName, ".tar.gz"):
		return extractTarGz(archiveBytes, binaryName)
	case strings.HasSuffix(assetName, ".zip"):
		return extractZip(archiveBytes, binaryName)
	default:
		return nil, fmt.Errorf("unsupported archive format for %s", assetName)
	}
}

func extractTarGz(archiveBytes []byte, binaryName string) ([]byte, error) {
	gzReader, err := gzip.NewReader(bytes.NewReader(archiveBytes))
	if err != nil {
		return nil, fmt.Errorf("open tar.gz archive: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("read tar.gz archive: %w", err)
		}
		if filepath.Base(header.Name) != binaryName {
			continue
		}
		return io.ReadAll(tarReader)
	}
	return nil, fmt.Errorf("archive did not contain %s", binaryName)
}

func extractZip(archiveBytes []byte, binaryName string) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(archiveBytes), int64(len(archiveBytes)))
	if err != nil {
		return nil, fmt.Errorf("open zip archive: %w", err)
	}
	for _, file := range reader.File {
		if filepath.Base(file.Name) != binaryName {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("open %s from zip archive: %w", binaryName, err)
		}
		defer rc.Close()
		return io.ReadAll(rc)
	}
	return nil, fmt.Errorf("archive did not contain %s", binaryName)
}

func replaceExecutable(executablePath string, binary []byte, goos string) error {
	info, err := os.Stat(executablePath)
	if err != nil {
		return fmt.Errorf("stat executable: %w", err)
	}

	dir := filepath.Dir(executablePath)
	tmpPath := filepath.Join(dir, "."+filepath.Base(executablePath)+".upgrade.tmp")
	if err := os.WriteFile(tmpPath, binary, info.Mode()); err != nil {
		return fmt.Errorf("write upgraded binary: %w", err)
	}
	if err := os.Chmod(tmpPath, info.Mode()); err != nil {
		_ = removeFile(tmpPath)
		return fmt.Errorf("chmod upgraded binary: %w", err)
	}
	if goos == "windows" {
		return replaceWindowsExecutable(executablePath, tmpPath)
	}
	if err := renameFile(tmpPath, executablePath); err != nil {
		_ = removeFile(tmpPath)
		return fmt.Errorf("replace executable: %w", err)
	}
	return nil
}

func replaceWindowsExecutable(executablePath, tmpPath string) error {
	backupPath := executablePath + ".old"
	if err := removeIfExists(backupPath); err != nil {
		_ = removeFile(tmpPath)
		return fmt.Errorf("remove stale backup: %w", err)
	}

	if err := renameFile(executablePath, backupPath); err != nil {
		_ = removeFile(tmpPath)
		return fmt.Errorf("backup executable: %w", err)
	}

	if err := renameFile(tmpPath, executablePath); err != nil {
		_ = removeFile(tmpPath)
		if restoreErr := renameFile(backupPath, executablePath); restoreErr != nil {
			return fmt.Errorf("replace executable: %w (restore original: %v)", err, restoreErr)
		}
		return fmt.Errorf("replace executable: %w", err)
	}

	// The upgrade already succeeded; a cleanup failure should not leave the
	// user without the new binary.
	_ = scheduleWindowsCleanupFile(backupPath)

	return nil
}

func removeIfExists(path string) error {
	err := removeFile(path)
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func defaultScheduleWindowsCleanupFile(path string) error {
	cmd := exec.Command("cmd", "/c", fmt.Sprintf("ping 127.0.0.1 -n 3 >NUL && del /f /q \"%s\"", path))
	return cmd.Start()
}
