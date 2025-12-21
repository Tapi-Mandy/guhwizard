// FILE: internal/fs/safefs.go
package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CopyFile copies a file from src to dst.
// It ensures the destination directory exists and uses safe path handling.
func CopyFile(src, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return err
	}

	return nil
}

// BackupAndCopy copies src to dst, but backs up dst if it already exists.
// Backup format: filename.bak.20060102150405
func BackupAndCopy(src, dst string) error {
	dst = filepath.Clean(dst)
	if _, err := os.Stat(dst); err == nil {
		// File exists, backup!
		timestamp := time.Now().Format("20060102150405")
		backupPath := fmt.Sprintf("%s.bak.%s", dst, timestamp)

		if err := os.Rename(dst, backupPath); err != nil {
			return fmt.Errorf("failed to backup existing file %s: %w", dst, err)
		}
	}

	return CopyFile(src, dst)
}

// AtomicWrite writes content to a file atomically by writing to a temp file and renaming.
// It also ensures the destination directory exists.
func AtomicWrite(path string, content []byte, mode os.FileMode) error {
	path = filepath.Clean(path)
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create temp file in the same directory to ensure atomic rename works across partitions
	tmpFile, err := os.CreateTemp(dir, "guhwizard-tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name()) // Clean up if something goes wrong

	if _, err := tmpFile.Write(content); err != nil {
		tmpFile.Close()
		return err
	}

	// Set perms
	if err := tmpFile.Chmod(mode); err != nil {
		tmpFile.Close()
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	return os.Rename(tmpFile.Name(), path)
}

// ExpandHome expands the "~" in a path to the user's home directory.
func ExpandHome(path string) (string, error) {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}
