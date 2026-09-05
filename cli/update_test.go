package cli

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCompareReleaseVersions(t *testing.T) {
	tests := []struct {
		left, right string
		want        int
	}{
		{"v0.3.1", "v0.3.2", -1},
		{"v0.3.2", "v0.3.2", 0},
		{"v1.0.0", "v0.99.99", 1},
	}
	for _, test := range tests {
		if got := compareReleaseVersions(test.left, test.right); got != test.want {
			t.Fatalf("compareReleaseVersions(%q, %q)=%d want %d", test.left, test.right, got, test.want)
		}
	}
}

func TestDiscoverUpdateUsesPlatformArchive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/latest" {
			http.NotFound(response, request)
			return
		}
		_, _ = io.WriteString(response, "v0.3.2\n")
	}))
	defer server.Close()
	release, err := discoverUpdate(context.Background(), server.Client(), server.URL+"/latest", server.URL+"/downloads", "darwin", "arm64")
	if err != nil {
		t.Fatal(err)
	}
	if release.ArchiveName != "agentsmd_darwin_arm64.tar.gz" || release.Version != "v0.3.2" {
		t.Fatalf("release=%+v", release)
	}
	if release.ArchiveURL != server.URL+"/downloads/v0.3.2/agentsmd_darwin_arm64.tar.gz" {
		t.Fatalf("archive url=%q", release.ArchiveURL)
	}
}

func TestDiscoverUpdateRejectsInvalidRegistryValue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(response, "../../unexpected")
	}))
	defer server.Close()
	_, err := discoverUpdate(context.Background(), server.Client(), server.URL, server.URL, "linux", "amd64")
	if err == nil || !strings.Contains(err.Error(), "invalid version") {
		t.Fatalf("err=%v", err)
	}
}

func TestExecutableForUpdateRejectsPackageManagerSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink creation may require elevated Windows privileges")
	}
	directory := t.TempDir()
	target := filepath.Join(directory, "agentsmd-real")
	link := filepath.Join(directory, "agentsmd")
	if err := os.WriteFile(target, []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if _, err := executableForUpdate(link); err == nil || !strings.Contains(err.Error(), "package manager") {
		t.Fatalf("err=%v", err)
	}
}

func TestInstallUpdateVerifiesAndReplacesExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("release self-update currently supports macOS and Linux")
	}
	newBinary := []byte("#!/bin/sh\necho 'agentsmd version 0.3.2'\n")
	archive := testReleaseArchive(t, newBinary)
	hash := sha256.Sum256(archive)
	archiveName := "agentsmd_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/checksums.txt":
			_, _ = fmt.Fprintf(response, "%x  %s\n", hash, archiveName)
		case "/" + archiveName:
			_, _ = response.Write(archive)
		default:
			http.NotFound(response, request)
		}
	}))
	defer server.Close()
	target := filepath.Join(t.TempDir(), "agentsmd")
	if err := os.WriteFile(target, []byte("#!/bin/sh\necho old\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	release := updateRelease{Version: "v0.3.2", ArchiveName: archiveName, ArchiveURL: server.URL + "/" + archiveName, ChecksumsURL: server.URL + "/checksums.txt"}
	if err := installUpdate(context.Background(), server.Client(), release, target); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(target, "--version").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(output)) != "agentsmd version 0.3.2" {
		t.Fatalf("version=%q", output)
	}
}

func TestInstallUpdateRejectsChecksumMismatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("release self-update currently supports macOS and Linux")
	}
	archive := testReleaseArchive(t, []byte("#!/bin/sh\necho 'agentsmd version 0.3.2'\n"))
	archiveName := "agentsmd_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if strings.HasSuffix(request.URL.Path, "checksums.txt") {
			_, _ = fmt.Fprintf(response, "%064d  %s\n", 0, archiveName)
			return
		}
		_, _ = response.Write(archive)
	}))
	defer server.Close()
	target := filepath.Join(t.TempDir(), "agentsmd")
	oldBinary := []byte("#!/bin/sh\necho old\n")
	if err := os.WriteFile(target, oldBinary, 0o755); err != nil {
		t.Fatal(err)
	}
	release := updateRelease{Version: "v0.3.2", ArchiveName: archiveName, ArchiveURL: server.URL + "/archive", ChecksumsURL: server.URL + "/checksums.txt"}
	err := installUpdate(context.Background(), server.Client(), release, target)
	if err == nil || !strings.Contains(err.Error(), "checksum verification failed") {
		t.Fatalf("err=%v", err)
	}
	data, readErr := os.ReadFile(target)
	if readErr != nil || !bytes.Equal(data, oldBinary) {
		t.Fatalf("current executable changed after rejected update")
	}
}

func testReleaseArchive(t *testing.T, binary []byte) []byte {
	t.Helper()
	var output bytes.Buffer
	gzipWriter := gzip.NewWriter(&output)
	tarWriter := tar.NewWriter(gzipWriter)
	if err := tarWriter.WriteHeader(&tar.Header{Name: "agentsmd", Mode: 0o755, Size: int64(len(binary)), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write(binary); err != nil {
		t.Fatal(err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}
