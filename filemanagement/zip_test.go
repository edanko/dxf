package filemanagement

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestZipRoundTripPreservesEntries(t *testing.T) {
	entries := []ZipEntry{
		{Name: "drawing.dxf", Data: []byte("0\nSECTION\n2\nHEADER\n0\nENDSEC\n0\nEOF\n")},
		{Name: "nested/readme.txt", Data: []byte("nested entry\n")},
		{Name: "binary.bin", Data: []byte{0x00, 0x01, 0xfe, 0xff}},
	}
	archive := filepath.Join(t.TempDir(), "drawings.zip")
	if err := CreateZip(archive, entries); err != nil {
		t.Fatal(err)
	}

	got, err := ReadZip(archive)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(entries) {
		t.Fatalf("ReadZip returned %d entries, want %d", len(got), len(entries))
	}
	for i := range entries {
		if got[i].Name != entries[i].Name {
			t.Fatalf("entry %d name = %q, want %q", i, got[i].Name, entries[i].Name)
		}
		if !bytes.Equal(got[i].Data, entries[i].Data) {
			t.Fatalf("entry %d data differs", i)
		}
	}

	target := filepath.Join(t.TempDir(), "out")
	extracted, err := ExtractZip(archive, target)
	if err != nil {
		t.Fatal(err)
	}
	if len(extracted) != len(entries) {
		t.Fatalf("ExtractZip returned %d paths, want %d", len(extracted), len(entries))
	}
	for _, entry := range entries {
		data, err := os.ReadFile(filepath.Join(target, filepath.FromSlash(entry.Name)))
		if err != nil {
			t.Fatalf("read extracted %q: %v", entry.Name, err)
		}
		if !bytes.Equal(data, entry.Data) {
			t.Fatalf("extracted %q data differs", entry.Name)
		}
	}
}

func TestExtractZipRejectsTraversal(t *testing.T) {
	tests := []struct {
		name  string
		entry string
	}{
		{name: "parent", entry: "../escape.txt"},
		{name: "nested parent", entry: "sub/../../escape.txt"},
		{name: "absolute", entry: "/absolute.txt"},
		{name: "backslash parent", entry: `..\escape.txt`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			archive := filepath.Join(root, "archive.zip")
			writeZipEntries(t, archive, []ZipEntry{
				{Name: "safe.txt", Data: []byte("safe")},
				{Name: test.entry, Data: []byte("unsafe")},
			})
			target := filepath.Join(root, "out")
			_, err := ExtractZip(archive, target)
			if err == nil {
				t.Fatal("expected unsafe path error")
			}
			if !strings.Contains(err.Error(), "unsafe ZIP entry path") {
				t.Fatalf("unexpected error: %v", err)
			}
			if _, statErr := os.Stat(filepath.Join(target, "safe.txt")); !os.IsNotExist(statErr) {
				t.Fatalf("safe entry was written before validation completed: %v", statErr)
			}
			if _, statErr := os.Stat(filepath.Join(root, "escape.txt")); !os.IsNotExist(statErr) {
				t.Fatalf("entry escaped target directory: %v", statErr)
			}
		})
	}
}

func TestExtractZipRejectsSymlink(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "symlink.zip")
	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	header := &zip.FileHeader{Name: "link"}
	header.SetMode(os.ModeSymlink | 0o777)
	part, err := writer.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("target.txt")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := ExtractZip(archive, t.TempDir()); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("expected symlink rejection, got %v", err)
	}
}

func TestCreateZipRejectsInvalidNames(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "invalid.zip")
	if err := CreateZip(archive, []ZipEntry{{Name: "", Data: nil}}); err == nil {
		t.Fatal("expected empty name error")
	}
	if err := CreateZip(archive, []ZipEntry{{Name: "a.dxf"}, {Name: "a.dxf"}}); err == nil {
		t.Fatal("expected duplicate name error")
	}
}

func writeZipEntries(t *testing.T, archive string, entries []ZipEntry) {
	t.Helper()
	file, err := os.Create(archive)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for _, entry := range entries {
		part, err := writer.Create(entry.Name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(entry.Data); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}
