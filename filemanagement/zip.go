package filemanagement

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ZipEntry is one named entry in a ZIP archive.
type ZipEntry struct {
	Name string
	Data []byte
}

// CreateZip writes entries to a ZIP archive. The archive is written atomically,
// preserving each entry's recorded name and contents.
func CreateZip(archivePath string, entries []ZipEntry) error {
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	seen := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if entry.Name == "" {
			return fmt.Errorf("filemanagement: empty ZIP entry name")
		}
		if seen[entry.Name] {
			return fmt.Errorf("filemanagement: duplicate ZIP entry %q", entry.Name)
		}
		seen[entry.Name] = true
		part, err := writer.Create(entry.Name)
		if err != nil {
			writer.Close()
			return fmt.Errorf("filemanagement: create ZIP entry %q: %w", entry.Name, err)
		}
		if _, err := part.Write(entry.Data); err != nil {
			writer.Close()
			return fmt.Errorf("filemanagement: write ZIP entry %q: %w", entry.Name, err)
		}
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("filemanagement: finalize ZIP: %w", err)
	}
	return AtomicWrite(archivePath, buf.Bytes(), false)
}

// CreateZipFromFiles writes files to a ZIP archive using each file's base name
// as the entry name. It is a convenience wrapper around CreateZip.
func CreateZipFromFiles(archivePath string, files []string) error {
	entries := make([]ZipEntry, 0, len(files))
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("filemanagement: read %q: %w", file, err)
		}
		entries = append(entries, ZipEntry{Name: filepath.Base(file), Data: data})
	}
	return CreateZip(archivePath, entries)
}

// ReadZip reads every entry from a ZIP archive.
func ReadZip(archivePath string) ([]ZipEntry, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("filemanagement: open ZIP %q: %w", archivePath, err)
	}
	defer reader.Close()

	entries := make([]ZipEntry, 0, len(reader.File))
	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		part, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("filemanagement: open ZIP entry %q: %w", file.Name, err)
		}
		data, readErr := io.ReadAll(part)
		closeErr := part.Close()
		if readErr != nil {
			return nil, fmt.Errorf("filemanagement: read ZIP entry %q: %w", file.Name, readErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("filemanagement: close ZIP entry %q: %w", file.Name, closeErr)
		}
		entries = append(entries, ZipEntry{Name: file.Name, Data: data})
	}
	return entries, nil
}

// ExtractZip extracts an archive into targetDir, creating directories as needed.
// Every entry path is validated before anything is written; absolute paths and
// paths that escape targetDir are rejected.
func ExtractZip(archivePath, targetDir string) ([]string, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("filemanagement: open ZIP %q: %w", archivePath, err)
	}
	defer reader.Close()

	root, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("filemanagement: resolve target directory %q: %w", targetDir, err)
	}

	type destination struct {
		file *zip.File
		path string
	}
	destinations := make([]destination, 0, len(reader.File))
	for _, file := range reader.File {
		relative, err := safeZipPath(root, file.Name)
		if err != nil {
			return nil, err
		}
		if file.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("filemanagement: refusing symlink ZIP entry %q", file.Name)
		}
		destinations = append(destinations, destination{file: file, path: relative})
	}

	extracted := make([]string, 0, len(destinations))
	for _, item := range destinations {
		if item.file.FileInfo().IsDir() || strings.HasSuffix(item.file.Name, "/") {
			if err := os.MkdirAll(item.path, 0o755); err != nil {
				return nil, fmt.Errorf("filemanagement: create directory %q: %w", item.path, err)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(item.path), 0o755); err != nil {
			return nil, fmt.Errorf("filemanagement: create directory %q: %w", filepath.Dir(item.path), err)
		}
		part, err := item.file.Open()
		if err != nil {
			return nil, fmt.Errorf("filemanagement: open ZIP entry %q: %w", item.file.Name, err)
		}
		data, readErr := io.ReadAll(part)
		closeErr := part.Close()
		if readErr != nil {
			return nil, fmt.Errorf("filemanagement: read ZIP entry %q: %w", item.file.Name, readErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("filemanagement: close ZIP entry %q: %w", item.file.Name, closeErr)
		}
		if err := AtomicWrite(item.path, data, false); err != nil {
			return nil, fmt.Errorf("filemanagement: extract ZIP entry %q: %w", item.file.Name, err)
		}
		extracted = append(extracted, item.path)
	}
	return extracted, nil
}

// safeZipPath validates a ZIP entry name and returns its absolute destination.
func safeZipPath(root, name string) (string, error) {
	normalized := strings.ReplaceAll(name, "\\", "/")
	clean := path.Clean(normalized)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") {
		return "", fmt.Errorf("filemanagement: unsafe ZIP entry path %q", name)
	}
	if filepath.IsAbs(clean) {
		return "", fmt.Errorf("filemanagement: unsafe ZIP entry path %q", name)
	}
	dest := filepath.Join(root, filepath.FromSlash(clean))
	relative, err := filepath.Rel(root, dest)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) || filepath.IsAbs(relative) {
		return "", fmt.Errorf("filemanagement: unsafe ZIP entry path %q", name)
	}
	return dest, nil
}
