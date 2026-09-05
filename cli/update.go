package cli

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
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	defaultLatestURL   = "https://rudrakshkarpe.com/downloads/agentsmd/latest"
	defaultReleaseBase = "https://rudrakshkarpe.com/downloads/agentsmd"
	maxArchiveBytes    = 64 << 20
	maxBinaryBytes     = 32 << 20
)

var releaseVersionPattern = regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)

type updateRelease struct {
	Version      string
	ArchiveName  string
	ArchiveURL   string
	ChecksumsURL string
}

func (a *app) updateCommand() *cobra.Command {
	var checkOnly bool
	command := &cobra.Command{
		Use:     "update",
		Aliases: []string{"upgrade"},
		Short:   "Update agentsmd CLI to the latest release",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ui := uiFor(cmd)
			if ui.interactive {
				fmt.Fprintln(cmd.OutOrStdout(), ui.icon("🔄")+ui.brand("agentsmd CLI · update"))
				fmt.Fprintln(cmd.OutOrStdout(), ui.muted("Checking the release registry…"))
				fmt.Fprintln(cmd.OutOrStdout())
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Minute)
			defer cancel()
			client := &http.Client{Timeout: 3 * time.Minute}
			release, err := discoverUpdate(ctx, client, defaultLatestURL, defaultReleaseBase, runtime.GOOS, runtime.GOARCH)
			if err != nil {
				return err
			}
			current := normalizedVersion(Version)
			if current != "dev" && compareReleaseVersions(current, release.Version) >= 0 {
				writeSuccess(cmd, fmt.Sprintf("already up to date (%s)", release.Version))
				return nil
			}
			writeInfo(cmd, fmt.Sprintf("current %s · latest %s", current, release.Version))
			if checkOnly {
				writeInfo(cmd, "Run `agentsmd update` to install it.")
				return nil
			}
			target, err := executableForUpdate(os.Args[0])
			if err != nil {
				return err
			}
			if err := installUpdate(ctx, client, release, target); err != nil {
				return err
			}
			writeSuccess(cmd, fmt.Sprintf("updated agentsmd CLI to %s", release.Version))
			writeInfo(cmd, "The next command will use the new version.")
			return nil
		},
	}
	command.Flags().BoolVar(&checkOnly, "check", false, "check for an update without installing it")
	return command
}

func executableForUpdate(invokedAs string) (string, error) {
	invokedPath := invokedAs
	var err error
	if !strings.ContainsRune(invokedPath, os.PathSeparator) {
		invokedPath, err = exec.LookPath(invokedPath)
		if err != nil {
			return "", fmt.Errorf("locate agentsmd on PATH: %w", err)
		}
	}
	if info, statErr := os.Lstat(invokedPath); statErr == nil && info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("agentsmd is managed through a symlink; update it with its package manager or reinstall from rudrakshkarpe.com")
	}
	target, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locate agentsmd executable: %w", err)
	}
	return target, nil
}

func normalizedVersion(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "dev" {
		return "dev"
	}
	if !strings.HasPrefix(value, "v") {
		return "v" + value
	}
	return value
}

func compareReleaseVersions(left, right string) int {
	parse := func(value string) [3]int {
		var result [3]int
		parts := strings.Split(strings.TrimPrefix(value, "v"), ".")
		for index := 0; index < len(result) && index < len(parts); index++ {
			result[index], _ = strconv.Atoi(parts[index])
		}
		return result
	}
	a, b := parse(left), parse(right)
	for index := range a {
		if a[index] < b[index] {
			return -1
		}
		if a[index] > b[index] {
			return 1
		}
	}
	return 0
}

func discoverUpdate(ctx context.Context, client *http.Client, latestURL, releaseBase, goos, goarch string) (updateRelease, error) {
	if goos != "darwin" && goos != "linux" {
		return updateRelease{}, fmt.Errorf("self-update is not supported on %s; use the package or install script that installed agentsmd", goos)
	}
	if goarch != "amd64" && goarch != "arm64" {
		return updateRelease{}, fmt.Errorf("self-update is not supported on %s/%s", goos, goarch)
	}
	data, err := download(ctx, client, latestURL, 128)
	if err != nil {
		return updateRelease{}, fmt.Errorf("check latest agentsmd version: %w", err)
	}
	version := strings.TrimSpace(string(data))
	if !releaseVersionPattern.MatchString(version) {
		return updateRelease{}, fmt.Errorf("release registry returned invalid version %q", version)
	}
	archive := fmt.Sprintf("agentsmd_%s_%s.tar.gz", goos, goarch)
	base := strings.TrimRight(releaseBase, "/") + "/" + version
	return updateRelease{
		Version:      version,
		ArchiveName:  archive,
		ArchiveURL:   base + "/" + archive,
		ChecksumsURL: base + "/checksums.txt",
	}, nil
}

func installUpdate(ctx context.Context, client *http.Client, release updateRelease, executablePath string) error {
	info, err := os.Lstat(executablePath)
	if err != nil {
		return fmt.Errorf("inspect current executable: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("current executable is managed through a symlink; update it with its package manager or reinstall from rudrakshkarpe.com")
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("current executable is not a regular file: %s", executablePath)
	}

	checksums, err := download(ctx, client, release.ChecksumsURL, 1<<20)
	if err != nil {
		return fmt.Errorf("download checksums: %w", err)
	}
	expected, err := checksumFor(checksums, release.ArchiveName)
	if err != nil {
		return err
	}
	archive, err := download(ctx, client, release.ArchiveURL, maxArchiveBytes)
	if err != nil {
		return fmt.Errorf("download %s: %w", release.ArchiveName, err)
	}
	actual := sha256.Sum256(archive)
	if !bytes.Equal(actual[:], expected) {
		return fmt.Errorf("checksum verification failed for %s", release.ArchiveName)
	}
	binary, err := extractAgentsmd(archive)
	if err != nil {
		return err
	}
	return replaceExecutable(ctx, executablePath, binary, release.Version, info.Mode().Perm())
}

func download(ctx context.Context, client *http.Client, url string, limit int64) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "agentsmd-cli/"+Version)
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned %s", url, response.Status)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("response from %s exceeds %d bytes", url, limit)
	}
	return data, nil
}

func checksumFor(data []byte, archiveName string) ([]byte, error) {
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || strings.TrimPrefix(fields[1], "*") != archiveName {
			continue
		}
		value, err := hex.DecodeString(fields[0])
		if err != nil || len(value) != sha256.Size {
			return nil, fmt.Errorf("invalid checksum for %s", archiveName)
		}
		return value, nil
	}
	return nil, fmt.Errorf("checksum not found for %s", archiveName)
}

func extractAgentsmd(archive []byte) ([]byte, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("open release archive: %w", err)
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read release archive: %w", err)
		}
		if header.Typeflag != tar.TypeReg || filepath.Clean(header.Name) != "agentsmd" {
			continue
		}
		if header.Size < 1 || header.Size > maxBinaryBytes {
			return nil, fmt.Errorf("release binary has invalid size %d", header.Size)
		}
		binary, err := io.ReadAll(io.LimitReader(tarReader, maxBinaryBytes+1))
		if err != nil {
			return nil, fmt.Errorf("extract release binary: %w", err)
		}
		if int64(len(binary)) != header.Size {
			return nil, fmt.Errorf("release binary size mismatch")
		}
		return binary, nil
	}
	return nil, fmt.Errorf("release archive does not contain agentsmd")
}

func replaceExecutable(ctx context.Context, target string, binary []byte, version string, mode os.FileMode) (returnErr error) {
	temporary, err := os.CreateTemp(filepath.Dir(target), ".agentsmd-update-*")
	if err != nil {
		return fmt.Errorf("create update beside %s: %w", target, err)
	}
	temporaryPath := temporary.Name()
	defer func() {
		_ = temporary.Close()
		if returnErr != nil {
			_ = os.Remove(temporaryPath)
		}
	}()
	if _, err := temporary.Write(binary); err != nil {
		return fmt.Errorf("write updated executable: %w", err)
	}
	if err := temporary.Chmod(mode | 0o111); err != nil {
		return fmt.Errorf("mark updated executable runnable: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close updated executable: %w", err)
	}
	verifyContext, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	output, err := exec.CommandContext(verifyContext, temporaryPath, "--version").CombinedOutput()
	if err != nil {
		return fmt.Errorf("verify updated executable: %w: %s", err, strings.TrimSpace(string(output)))
	}
	want := "agentsmd version " + strings.TrimPrefix(version, "v")
	if strings.TrimSpace(string(output)) != want {
		return fmt.Errorf("updated executable reported %q, expected %q", strings.TrimSpace(string(output)), want)
	}
	if err := os.Rename(temporaryPath, target); err != nil {
		return fmt.Errorf("replace %s: %w", target, err)
	}
	return nil
}
