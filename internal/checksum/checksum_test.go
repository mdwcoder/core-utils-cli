package checksum

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	data := []byte("hello world")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	good := String(data)
	if err := ValidateFile(path, good); err != nil {
		t.Fatalf("expected valid checksum, got %v", err)
	}

	if err := ValidateFile(path, "bad"); err == nil {
		t.Error("expected checksum mismatch error")
	}
}

func TestString(t *testing.T) {
	h := String([]byte("test"))
	if len(h) != 64 {
		t.Errorf("expected hex length 64, got %d", len(h))
	}
}
