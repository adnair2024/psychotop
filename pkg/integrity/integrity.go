package integrity

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

var errMaxFiles = errors.New("max files reached")

// FileHash stores the path and its SHA-256 hash.
type FileHash struct {
	Path string
	Hash string
}

// ChecksumDirectory calculates the SHA-256 hashes for all files in a directory, up to a limit.
func ChecksumDirectory(root string) ([]FileHash, error) {
	var hashes []FileHash
	const maxFiles = 1000

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if len(hashes) >= maxFiles {
			return errMaxFiles
		}
		// Skip directories and non-regular files (symlinks, special files, etc.)
		if !info.Mode().IsRegular() {
			return nil
		}

		hash, err := HashFile(path)
		if err != nil {
			// Skip files we can't read
			return nil
		}

		hashes = append(hashes, FileHash{
			Path: path,
			Hash: hash,
		})
		return nil
	})

	if err != nil && err != errMaxFiles {
		return nil, err
	}

	// Sort by path for consistent ordering
	sort.Slice(hashes, func(i, j int) bool {
		return hashes[i].Path < hashes[j].Path
	})

	return hashes, nil
}

// HashFile calculates the SHA-256 hash of a single file.
func HashFile(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("not a regular file: %s", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
