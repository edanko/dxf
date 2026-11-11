package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestZipUnzipRoundTripCLI(t *testing.T) {
	data, err := os.ReadFile("../../testdata/Minimal_DXF_AC1009.dxf")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	input := filepath.Join(dir, "drawing.dxf")
	if err := os.WriteFile(input, data, 0o644); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(dir, "drawing.zip")

	bin := buildCLI(t)
	if out, errOut, code := runCLI(t, bin, "zip", archive, input); code != 0 {
		t.Fatalf("zip failed: exit %d out=%s err=%s", code, out, errOut)
	}
	target := filepath.Join(dir, "out")
	if out, errOut, code := runCLI(t, bin, "unzip", archive, target); code != 0 {
		t.Fatalf("unzip failed: exit %d out=%s err=%s", code, out, errOut)
	}
	extracted, err := os.ReadFile(filepath.Join(target, "drawing.dxf"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(extracted, data) {
		t.Fatal("unzipped file is not byte-identical to the input")
	}
}
