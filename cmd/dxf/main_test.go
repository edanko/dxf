// Integration tests for the dxf CLI: build the binary and run it against
// fixtures, asserting output contracts and exit codes.
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	dxf "github.com/edanko/dxf"
	"github.com/edanko/dxf/drawing"
)

func buildCLI(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "dxf")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build CLI: %v", err)
	}
	return bin
}

func runCLI(t *testing.T, bin string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		ee, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("run: %v", err)
		}
		code = ee.ExitCode()
	}
	return stdout.String(), stderr.String(), code
}

func TestInfo(t *testing.T) {
	bin := buildCLI(t)
	out, _, code := runCLI(t, bin, "info", "../../testdata/Minimal_DXF_AC1009.dxf")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(out, "AC1009") {
		t.Fatalf("version missing: %s", out)
	}
}

func TestInfoJSON(t *testing.T) {
	bin := buildCLI(t)
	out, _, code := runCLI(t, bin, "info", "--format", "json", "../../testdata/Minimal_DXF_AC1009.dxf")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !json.Valid([]byte(out)) {
		t.Fatalf("invalid json: %s", out)
	}
}

func TestUnknownCommand(t *testing.T) {
	bin := buildCLI(t)
	_, _, code := runCLI(t, bin, "nope")
	if code != 2 {
		t.Fatalf("exit %d", code)
	}
}

func TestInvalidInput(t *testing.T) {
	bin := buildCLI(t)
	_, errOut, code := runCLI(t, bin, "info", "nonexistent.dxf")
	if code != 1 {
		t.Fatalf("exit %d err=%s", code, errOut)
	}
	if errOut == "" {
		t.Fatal("actionable message expected")
	}
}

func TestConvert(t *testing.T) {
	bin := buildCLI(t)
	out := filepath.Join(t.TempDir(), "out.dxf")
	_, _, code := runCLI(t, bin, "convert", "--version", "AC1015", "--output", out, "../../testdata/Minimal_DXF_AC1009.dxf")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	doc, err := dxf.FromFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Header().Version != "AC1015" {
		t.Fatalf("version %s", doc.Header().Version)
	}
}

func TestDrawPNG(t *testing.T) {
	bin := buildCLI(t)
	out := filepath.Join(t.TempDir(), "out.png")
	_, _, code := runCLI(t, bin, "draw", "--backend", "png", "--output", out, "../../testdata/Minimal_DXF_AC1009.dxf")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, []byte("\x89PNG")) {
		t.Fatalf("output is not a PNG")
	}
}

func TestQueryJSON(t *testing.T) {
	bin := buildCLI(t)
	out, _, code := runCLI(t, bin, "query", "--format", "json", "../../testdata/Minimal_DXF_AC1009.dxf", "type == 'LINE'")
	if code != 0 {
		t.Fatalf("exit %d out=%s", code, out)
	}
	if !json.Valid([]byte(out)) {
		t.Fatalf("invalid json: %s", out)
	}
}

func TestStripComments(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	d.SetComments(map[string][]string{"test": {"hello"}})
	var buf bytes.Buffer
	if _, err := d.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	in := filepath.Join(t.TempDir(), "in.dxf")
	if err := os.WriteFile(in, buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "out.dxf")
	bin := buildCLI(t)
	_, _, code := runCLI(t, bin, "strip", "--comments", "--output", out, in)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte("\n999\n")) {
		t.Fatalf("comments remain in output")
	}
}

func TestAuditClean(t *testing.T) {
	bin := buildCLI(t)
	_, _, code := runCLI(t, bin, "audit", "../../testdata/Minimal_DXF_AC1009.dxf")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
}

func TestHelp(t *testing.T) {
	bin := buildCLI(t)
	out, _, code := runCLI(t, bin, "--help")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	for _, name := range []string{"info", "analyze", "difftags", "debug", "audit", "convert", "draw", "query", "explode", "strip", "header", "layers", "browse", "version", "completion"} {
		if !strings.Contains(out, name) {
			t.Fatalf("usage missing %s:\n%s", name, out)
		}
	}
}

func TestVersion(t *testing.T) {
	bin := buildCLI(t)
	out, _, code := runCLI(t, bin, "version")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(out, "dxf") {
		t.Fatalf("version output: %s", out)
	}
}
