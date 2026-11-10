// Package filemanagement provides atomic file writes with optional backups.
package filemanagement

import (
	"fmt"
	"os"
	"path/filepath"
)

// AtomicWrite writes data to path via a temporary file in the destination
// directory, fsyncs and renames. When backup is set and path exists, the
// original is copied to path+".bak" before the rename.
func AtomicWrite(path string, data []byte, backup bool) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".dxf-write-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if backup {
		if err := Backup(path); err != nil {
			return err
		}
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}

// Backup copies path to path+".bak" when path exists.
func Backup(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read %s for backup: %w", path, err)
	}
	if err := os.WriteFile(path+".bak", data, 0o644); err != nil {
		return fmt.Errorf("write backup %s.bak: %w", path, err)
	}
	return nil
}
