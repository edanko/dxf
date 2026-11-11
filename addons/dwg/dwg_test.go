package dwg

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeRunner struct {
	calls  int
	write  []byte
	err    error
	stderr []byte
}

func (r *fakeRunner) Run(name string, args []string, stdin []byte) ([]byte, []byte, error) {
	r.calls++
	if r.err != nil {
		return nil, r.stderr, r.err
	}
	if len(args) > 1 && r.write != nil {
		if err := os.WriteFile(args[1], r.write, 0o600); err != nil {
			return nil, nil, err
		}
	}
	return nil, nil, nil
}

func TestConvertMissingToolLeavesNoOutput(t *testing.T) {
	output := filepath.Join(t.TempDir(), "out.dxf")
	err := New(nil, "").Convert("input.dwg", output)
	if err == nil {
		t.Fatal("expected missing converter error")
	}
	if !strings.Contains(err.Error(), "DWG converter not available") {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, statErr := os.Stat(output); !os.IsNotExist(statErr) {
		t.Fatalf("output should not exist, stat error: %v", statErr)
	}
}

func TestConvertFailureLeavesNoPartialOutput(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "out.dxf")
	runner := &fakeRunner{err: errors.New("converter crashed"), stderr: []byte("bad input")}
	err := New(runner, "ODAFileConverter").Convert("input.dwg", output)
	if err == nil {
		t.Fatal("expected conversion error")
	}
	if !strings.Contains(err.Error(), "bad input") {
		t.Fatalf("expected stderr in error, got %v", err)
	}
	if _, statErr := os.Stat(output); !os.IsNotExist(statErr) {
		t.Fatalf("output should not exist, stat error: %v", statErr)
	}
	if matches, _ := filepath.Glob(filepath.Join(dir, ".*.tmp")); len(matches) != 0 {
		t.Fatalf("temporary files left behind: %v", matches)
	}
}

func TestConvertSuccessWritesOutput(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "out.dxf")
	runner := &fakeRunner{write: []byte("DXF")}
	if err := New(runner, "ODAFileConverter").Convert("input.dwg", output); err != nil {
		t.Fatalf("convert: %v", err)
	}
	if runner.calls != 1 {
		t.Fatalf("runner calls = %d, want 1", runner.calls)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if string(data) != "DXF" {
		t.Fatalf("output = %q", data)
	}
	if matches, _ := filepath.Glob(filepath.Join(dir, ".*.tmp")); len(matches) != 0 {
		t.Fatalf("temporary files left behind: %v", matches)
	}
}
