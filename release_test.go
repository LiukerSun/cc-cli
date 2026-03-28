package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseAssetsStayAligned(t *testing.T) {
	root := "."

	goreleaser := readFile(t, filepath.Join(root, ".goreleaser.yaml"))
	installSh := readFile(t, filepath.Join(root, "install.sh"))
	installPs1 := readFile(t, filepath.Join(root, "install.ps1"))

	assertContains(t, goreleaser, "project_name: ccc")
	assertContains(t, goreleaser, "name_template: >-")
	assertContains(t, goreleaser, "ccc_{{ .Os }}_{{ .Arch }}")
	assertContains(t, goreleaser, "goos: windows")
	assertContains(t, goreleaser, "format: zip")
	assertContains(t, goreleaser, "name_template: checksums.txt")

	assertContains(t, installSh, `asset_name="${PROJECT_NAME}_${os}_${arch}.tar.gz"`)
	assertContains(t, installSh, `checksum_url="${base_url}/checksums.txt"`)

	assertContains(t, installPs1, `$assetName = "ccc_{0}_{1}.zip" -f $osArch.Os, $osArch.Arch`)
	assertContains(t, installPs1, `Invoke-WebRequest -Uri "$baseUrl/checksums.txt" -OutFile $checksumsPath`)
}

func TestInstallersReferenceExpectedPlatforms(t *testing.T) {
	installSh := readFile(t, "install.sh")
	installPs1 := readFile(t, "install.ps1")

	for _, platform := range []string{"linux", "darwin"} {
		if !strings.Contains(installSh, `echo "`+platform+`"`) {
			t.Fatalf("install.sh no longer advertises support for %s", platform)
		}
	}

	assertContains(t, installPs1, `Os = "windows"`)
	for _, arch := range []string{"amd64", "arm64"} {
		if !strings.Contains(installSh, `echo "`+arch+`"`) {
			t.Fatalf("install.sh no longer advertises support for %s", arch)
		}
		if !strings.Contains(installPs1, `"Arm64" { $goArch = "arm64" }`) && arch == "arm64" {
			t.Fatalf("install.ps1 no longer maps arm64")
		}
		if !strings.Contains(installPs1, `"X64" { $goArch = "amd64" }`) && arch == "amd64" {
			t.Fatalf("install.ps1 no longer maps amd64")
		}
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func assertContains(t *testing.T, content, needle string) {
	t.Helper()
	if !strings.Contains(content, needle) {
		t.Fatalf("expected content to contain %q", needle)
	}
}
