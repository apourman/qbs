package cli

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestChecksumForAndExtractBinary(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "release.tar.gz")
	binary := []byte("new qbs binary")

	archive, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	compressed := gzip.NewWriter(archive)
	writer := tar.NewWriter(compressed)
	if err := writer.WriteHeader(&tar.Header{Name: "qbs", Mode: 0o755, Size: int64(len(binary))}); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write(binary); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := compressed.Close(); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}

	digest := sha256.Sum256(mustRead(t, archivePath))
	checksums := filepath.Join(dir, "SHA256SUMS")
	if err := os.WriteFile(checksums, []byte(hex.EncodeToString(digest[:])+"  qbs_1.2.3_linux_amd64.tar.gz\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := verifyChecksum(archivePath, checksums, "qbs_1.2.3_linux_amd64.tar.gz"); err != nil {
		t.Fatal(err)
	}

	extracted, err := extractBinary(archivePath, dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := mustRead(t, extracted); string(got) != string(binary) {
		t.Fatalf("extracted binary = %q", got)
	}
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
