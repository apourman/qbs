package cli

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	updateRepository = "apourman/qbs"
	updateAPIURL     = "https://api.github.com/repos/" + updateRepository + "/releases/latest"
)

var updateHTTPClient = http.DefaultClient

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func update(requestedVersion string, stdout io.Writer) error {
	if runtime.GOOS == "windows" {
		return errors.New("self-update is not supported on Windows yet; download the new release from GitHub")
	}

	version := strings.TrimPrefix(requestedVersion, "v")
	if version == "" {
		release, err := fetchLatestRelease()
		if err != nil {
			return err
		}
		version = strings.TrimPrefix(release.TagName, "v")
	}
	if version == "" || strings.ContainsAny(version, "/\\") {
		return fmt.Errorf("invalid release version %q", version)
	}

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("find qbs executable: %w", err)
	}
	executable, err = filepath.EvalSymlinks(executable)
	if err != nil {
		return fmt.Errorf("resolve qbs executable: %w", err)
	}

	asset := fmt.Sprintf("qbs_%s_%s_%s.tar.gz", version, runtime.GOOS, runtime.GOARCH)
	baseURL := fmt.Sprintf("https://github.com/%s/releases/download/v%s/", updateRepository, version)
	tempDir, err := os.MkdirTemp(filepath.Dir(executable), ".qbs-update-")
	if err != nil {
		return fmt.Errorf("create update workspace: %w", err)
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(tempDir, asset)
	if err := download(baseURL+asset, archivePath); err != nil {
		return err
	}
	checksums := filepath.Join(tempDir, "SHA256SUMS")
	if err := download(baseURL+"SHA256SUMS", checksums); err != nil {
		return err
	}
	if err := verifyChecksum(archivePath, checksums, asset); err != nil {
		return err
	}

	newBinary, err := extractBinary(archivePath, tempDir)
	if err != nil {
		return err
	}
	mode, err := fileMode(executable)
	if err != nil {
		return fmt.Errorf("read qbs permissions: %w", err)
	}
	if err := os.Chmod(newBinary, mode); err != nil {
		return fmt.Errorf("set updated qbs permissions: %w", err)
	}
	if err := replaceExecutable(executable, newBinary); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "updated qbs to %s\n", version)
	return err
}

func fetchLatestRelease() (githubRelease, error) {
	var release githubRelease
	resp, err := updateHTTPClient.Get(updateAPIURL)
	if err != nil {
		return release, fmt.Errorf("check for the latest release: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return release, fmt.Errorf("check for the latest release: GitHub returned %s", resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return release, fmt.Errorf("read latest release: %w", err)
	}
	return release, nil
}

func download(url, destination string) error {
	resp, err := updateHTTPClient.Get(url)
	if err != nil {
		return fmt.Errorf("download %s: %w", filepath.Base(destination), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: GitHub returned %s", filepath.Base(destination), resp.Status)
	}
	file, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("create download: %w", err)
	}
	defer file.Close()
	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("save download: %w", err)
	}
	return file.Close()
}

func verifyChecksum(filePath, checksumPath, name string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("hash release: %w", err)
	}
	want, err := checksumFor(checksumPath, name)
	if err != nil {
		return err
	}
	got := hex.EncodeToString(hash.Sum(nil))
	if got != want {
		return fmt.Errorf("checksum mismatch for %s", name)
	}
	return nil
}

func checksumFor(path, name string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read checksums: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && filepath.Base(fields[len(fields)-1]) == name {
			return strings.ToLower(fields[0]), nil
		}
	}
	return "", fmt.Errorf("checksum for %s not found", name)
}

func extractBinary(archivePath, directory string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	compressed, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("open release archive: %w", err)
	}
	defer compressed.Close()
	reader := tar.NewReader(compressed)
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read release archive: %w", err)
		}
		if header.Name != "qbs" || header.Typeflag != tar.TypeReg {
			continue
		}
		path := filepath.Join(directory, "qbs-new")
		out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o700)
		if err != nil {
			return "", err
		}
		_, copyErr := io.Copy(out, reader)
		closeErr := out.Close()
		if copyErr != nil {
			return "", copyErr
		}
		if closeErr != nil {
			return "", closeErr
		}
		return path, nil
	}
	return "", errors.New("qbs executable not found in release archive")
}

func fileMode(path string) (os.FileMode, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Mode().Perm(), nil
}

func replaceExecutable(destination, source string) error {
	if err := os.Rename(source, destination); err != nil {
		return fmt.Errorf("replace qbs executable: %w", err)
	}
	return nil
}
