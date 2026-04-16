package upgrade

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlanResolvesLatestRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/LiukerSun/cc-cli/releases/latest":
			_, _ = io.WriteString(w, `{"tag_name":"v2.2.1"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	manager := Manager{
		RepoOwner:       defaultRepoOwner,
		RepoName:        defaultRepoName,
		ProjectName:     defaultProjectName,
		CurrentVersion:  "2.2.0",
		ExecutablePath:  "/tmp/ccc",
		GOOS:            "linux",
		GOARCH:          "amd64",
		APIBaseURL:      server.URL,
		DownloadBaseURL: server.URL + "/LiukerSun/cc-cli/releases/download",
		HTTPClient:      server.Client(),
	}

	plan, err := manager.Plan(context.Background(), "")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if plan.TargetVersion != "2.2.1" {
		t.Fatalf("TargetVersion = %q, want 2.2.1", plan.TargetVersion)
	}
	if plan.AssetName != "ccc_linux_amd64.tar.gz" {
		t.Fatalf("AssetName = %q", plan.AssetName)
	}
	if !strings.HasSuffix(plan.AssetURL, "/v2.2.1/ccc_linux_amd64.tar.gz") {
		t.Fatalf("AssetURL = %q", plan.AssetURL)
	}
}

func TestPlanFallsBackWhenLatestReleaseAPIIsForbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/LiukerSun/cc-cli/releases/latest":
			http.Error(w, "rate limited", http.StatusForbidden)
		case "/LiukerSun/cc-cli/releases/latest":
			http.Redirect(w, r, "/LiukerSun/cc-cli/releases/tag/v2.2.2", http.StatusFound)
		case "/LiukerSun/cc-cli/releases/tag/v2.2.2":
			_, _ = io.WriteString(w, "release page")
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	manager := Manager{
		RepoOwner:       defaultRepoOwner,
		RepoName:        defaultRepoName,
		ProjectName:     defaultProjectName,
		CurrentVersion:  "2.2.1",
		ExecutablePath:  "/tmp/ccc",
		GOOS:            "linux",
		GOARCH:          "amd64",
		APIBaseURL:      server.URL,
		DownloadBaseURL: server.URL + "/LiukerSun/cc-cli/releases/download",
		HTTPClient:      server.Client(),
	}

	plan, err := manager.Plan(context.Background(), "")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if plan.TargetVersion != "2.2.2" {
		t.Fatalf("TargetVersion = %q, want 2.2.2", plan.TargetVersion)
	}
}

func TestUpgradeReplacesUnixBinary(t *testing.T) {
	archiveBytes := makeTarGzArchive(t, "ccc", []byte("#!/bin/sh\necho upgraded\n"))
	sum := sha256.Sum256(archiveBytes)
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), "ccc_linux_amd64.tar.gz")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/LiukerSun/cc-cli/releases/download/v2.2.1/ccc_linux_amd64.tar.gz":
			_, _ = w.Write(archiveBytes)
		case "/LiukerSun/cc-cli/releases/download/v2.2.1/checksums.txt":
			_, _ = io.WriteString(w, checksums)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	executablePath := filepath.Join(t.TempDir(), "ccc")
	if err := os.WriteFile(executablePath, []byte("#!/bin/sh\necho old\n"), 0o755); err != nil {
		t.Fatalf("WriteFile executable: %v", err)
	}

	manager := Manager{
		RepoOwner:       defaultRepoOwner,
		RepoName:        defaultRepoName,
		ProjectName:     defaultProjectName,
		CurrentVersion:  "2.2.0",
		ExecutablePath:  executablePath,
		GOOS:            "linux",
		GOARCH:          "amd64",
		APIBaseURL:      server.URL,
		DownloadBaseURL: server.URL + "/LiukerSun/cc-cli/releases/download",
		HTTPClient:      server.Client(),
	}

	plan, err := manager.Plan(context.Background(), "2.2.1")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if err := manager.Upgrade(context.Background(), plan); err != nil {
		t.Fatalf("Upgrade: %v", err)
	}

	data, err := os.ReadFile(executablePath)
	if err != nil {
		t.Fatalf("ReadFile executable: %v", err)
	}
	if !strings.Contains(string(data), "upgraded") {
		t.Fatalf("upgraded executable content mismatch: %s", string(data))
	}
}

func TestUpgradeReplacesWindowsBinary(t *testing.T) {
	archiveBytes := makeZipArchive(t, "ccc.exe", []byte("upgraded windows binary\n"))
	sum := sha256.Sum256(archiveBytes)
	checksums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), "ccc_windows_amd64.zip")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/LiukerSun/cc-cli/releases/download/v2.2.1/ccc_windows_amd64.zip":
			_, _ = w.Write(archiveBytes)
		case "/LiukerSun/cc-cli/releases/download/v2.2.1/checksums.txt":
			_, _ = io.WriteString(w, checksums)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	executablePath := filepath.Join(t.TempDir(), "ccc.exe")
	if err := os.WriteFile(executablePath, []byte("old windows binary\n"), 0o755); err != nil {
		t.Fatalf("WriteFile executable: %v", err)
	}

	originalCleanup := scheduleWindowsCleanupFile
	defer func() {
		scheduleWindowsCleanupFile = originalCleanup
	}()

	var scheduledPath string
	scheduleWindowsCleanupFile = func(path string) error {
		scheduledPath = path
		return os.Remove(path)
	}

	manager := Manager{
		RepoOwner:       defaultRepoOwner,
		RepoName:        defaultRepoName,
		ProjectName:     defaultProjectName,
		CurrentVersion:  "2.2.0",
		ExecutablePath:  executablePath,
		GOOS:            "windows",
		GOARCH:          "amd64",
		APIBaseURL:      server.URL,
		DownloadBaseURL: server.URL + "/LiukerSun/cc-cli/releases/download",
		HTTPClient:      server.Client(),
	}

	plan, err := manager.Plan(context.Background(), "2.2.1")
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if err := manager.Upgrade(context.Background(), plan); err != nil {
		t.Fatalf("Upgrade: %v", err)
	}

	data, err := os.ReadFile(executablePath)
	if err != nil {
		t.Fatalf("ReadFile executable: %v", err)
	}
	if !strings.Contains(string(data), "upgraded windows binary") {
		t.Fatalf("upgraded executable content mismatch: %s", string(data))
	}

	if scheduledPath != executablePath+".old" {
		t.Fatalf("scheduled cleanup path = %q, want %q", scheduledPath, executablePath+".old")
	}
	if _, err := os.Stat(executablePath + ".old"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("backup file should be cleaned up, stat err = %v", err)
	}
}

func TestReplaceExecutableWindowsRestoresOriginalOnSwapFailure(t *testing.T) {
	executablePath := filepath.Join(t.TempDir(), "ccc.exe")
	if err := os.WriteFile(executablePath, []byte("old windows binary\n"), 0o755); err != nil {
		t.Fatalf("WriteFile executable: %v", err)
	}

	originalRename := renameFile
	originalCleanup := scheduleWindowsCleanupFile
	defer func() {
		renameFile = originalRename
		scheduleWindowsCleanupFile = originalCleanup
	}()

	renameCalls := 0
	renameFile = func(oldPath, newPath string) error {
		renameCalls++
		switch renameCalls {
		case 1, 3:
			return os.Rename(oldPath, newPath)
		case 2:
			return errors.New("swap failed")
		default:
			return os.Rename(oldPath, newPath)
		}
	}
	scheduleWindowsCleanupFile = func(path string) error {
		t.Fatalf("cleanup should not be scheduled when swap fails")
		return nil
	}

	err := replaceExecutable(executablePath, []byte("new windows binary\n"), "windows")
	if err == nil || !strings.Contains(err.Error(), "replace executable") {
		t.Fatalf("replaceExecutable() error = %v", err)
	}

	data, readErr := os.ReadFile(executablePath)
	if readErr != nil {
		t.Fatalf("ReadFile executable: %v", readErr)
	}
	if !strings.Contains(string(data), "old windows binary") {
		t.Fatalf("original executable should be restored, got: %s", string(data))
	}
	if _, statErr := os.Stat(executablePath + ".old"); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("backup file should be restored away, stat err = %v", statErr)
	}
}

func makeTarGzArchive(t *testing.T, name string, content []byte) []byte {
	t.Helper()

	var archive bytes.Buffer
	gzWriter := gzip.NewWriter(&archive)
	tarWriter := tar.NewWriter(gzWriter)

	header := &tar.Header{
		Name: name,
		Mode: 0o755,
		Size: int64(len(content)),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("Close tarWriter: %v", err)
	}
	if err := gzWriter.Close(); err != nil {
		t.Fatalf("Close gzWriter: %v", err)
	}

	return archive.Bytes()
}

func makeZipArchive(t *testing.T, name string, content []byte) []byte {
	t.Helper()

	var archive bytes.Buffer
	zipWriter := zip.NewWriter(&archive)

	fileWriter, err := zipWriter.Create(name)
	if err != nil {
		t.Fatalf("Create zip entry: %v", err)
	}
	if _, err := fileWriter.Write(content); err != nil {
		t.Fatalf("Write zip entry: %v", err)
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Close zipWriter: %v", err)
	}

	return archive.Bytes()
}
