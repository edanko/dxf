package filemanagement

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWrite(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f.txt")
	if err := AtomicWrite(p, []byte("hello"), false); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("data: %s", data)
	}
}

func TestAtomicWriteBackup(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f.txt")
	if err := os.WriteFile(p, []byte("orig"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := AtomicWrite(p, []byte("new"), true); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new" {
		t.Fatalf("data: %s", data)
	}
	bak, err := os.ReadFile(p + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	if string(bak) != "orig" {
		t.Fatalf("backup: %s", bak)
	}
}
