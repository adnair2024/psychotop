package integrity

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHashFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("hello psychotop")
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Expected SHA-256 for "hello psychotop"
	expected := "a1ac8efd5fc7c6650dfc55fbf86c4d9c903c4704c816fa568c117368afa0ef5d"

	got, err := HashFile(tmpFile)
	if err != nil {
		t.Fatalf("HashFile failed: %v", err)
	}

	if got != expected {
		t.Errorf("HashFile = %s; want %s", got, expected)
	}
}

func TestChecksumDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	files := map[string]string{
		"a.txt": "content a",
		"b.txt": "content b",
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to create temp file %s: %v", name, err)
		}
	}

	hashes, err := ChecksumDirectory(tmpDir)
	if err != nil {
		t.Fatalf("ChecksumDirectory failed: %v", err)
	}

	if len(hashes) != 2 {
		t.Errorf("expected 2 hashes, got %d", len(hashes))
	}

	// Verify sorting
	if hashes[0].Path > hashes[1].Path {
		t.Errorf("hashes are not sorted by path")
	}
}
