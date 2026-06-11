package archive

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func createTestZip(dir string) string {
	path := filepath.Join(dir, "test.zip")
	f, _ := os.Create(path)
	w := zip.NewWriter(f)

	fw, _ := w.Create("subdir/hello.txt")
	fw.Write([]byte("hello zip"))

	fw2, _ := w.Create("../../etc/passwd")
	fw2.Write([]byte("bad"))

	w.Close()
	f.Close()
	return path
}

func createTestTarGz(dir string) string {
	path := filepath.Join(dir, "test.tar.gz")
	f, _ := os.Create(path)
	// Write minimal gzip bytes so that gzip reader succeeds but tar reader fails gracefully
	f.Write([]byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03})
	f.Close()
	return path
}

func TestExtractZip(t *testing.T) {
	dir := t.TempDir()
	zippath := createTestZip(dir)
	dest := filepath.Join(dir, "dest")

	if err := Extract(zippath, dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dest, "subdir", "hello.txt"))
	if err != nil {
		t.Fatalf("expected extracted file: %v", err)
	}
	if string(data) != "hello zip" {
		t.Errorf("unexpected content: %s", string(data))
	}

	// path traversal file should be skipped
	if _, err := os.Stat(filepath.Join(dest, "..", "..", "etc", "passwd")); err == nil {
		t.Error("path traversal file should not exist")
	}
}

func TestExtractUnsupported(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("plain"), 0644)
	if err := Extract(path, dir); err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestSafePath(t *testing.T) {
	dest := t.TempDir()
	good, err := safePath(dest, "a/b/c.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(good) {
		t.Error("expected absolute path")
	}

	_, err = safePath(dest, "../outside.txt")
	if err == nil {
		t.Error("expected path traversal error")
	}
}

func TestExtractTarGzInvalid(t *testing.T) {
	dir := t.TempDir()
	tarpath := createTestTarGz(dir)
	dest := filepath.Join(dir, "dest")
	if err := Extract(tarpath, dest); err == nil {
		t.Error("expected error for invalid tar.gz")
	}
}
