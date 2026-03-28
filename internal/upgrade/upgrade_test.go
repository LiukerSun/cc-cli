package upgrade

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
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

func TestUpgradeRejectsWindowsSelfReplace(t *testing.T) {
	manager := Manager{
		RepoOwner:       defaultRepoOwner,
		RepoName:        defaultRepoName,
		ProjectName:     defaultProjectName,
		CurrentVersion:  "2.2.0",
		ExecutablePath:  `C:\ccc.exe`,
		GOOS:            "windows",
		GOARCH:          "amd64",
		APIBaseURL:      "https://api.github.com",
		DownloadBaseURL: "https://github.com/LiukerSun/cc-cli/releases/download",
	}

	err := manager.Upgrade(context.Background(), Plan{ExecutablePath: manager.ExecutablePath})
	if err == nil || !strings.Contains(err.Error(), "not supported on Windows") {
		t.Fatalf("Upgrade() error = %v", err)
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
