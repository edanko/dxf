package dxf

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
	"github.com/edanko/dxf/table"
)

// --- Task 1.1: version profile table ---

func TestVersionProfiles(t *testing.T) {
	tests := []struct {
		version          string
		release          string
		objects, classes bool
		binary           bool
		utf8             bool
		trueColor        bool
	}{
		{format.AC1009, "R12", false, false, false, false, false},
		{format.AC1014, "R14", true, true, true, false, false},
		{format.AC1015, "R2000", true, true, true, false, false},
		{format.AC1018, "R2004", true, true, true, false, false},
		{format.AC1021, "R2007", true, true, true, true, true},
		{format.AC1024, "R2010", true, true, true, true, true},
		{format.AC1027, "R2013", true, true, true, true, true},
		{format.AC1032, "R2018", true, true, true, true, true},
	}
	for _, tt := range tests {
		p := format.Lookup(tt.version)
		if p == nil {
			t.Fatalf("Lookup(%s) = nil", tt.version)
		}
		if p.Release != tt.release {
			t.Errorf("%s release = %q, want %q", tt.version, p.Release, tt.release)
		}
		if p.SupportsObjects != tt.objects {
			t.Errorf("%s SupportsObjects = %v, want %v", tt.version, p.SupportsObjects, tt.objects)
		}
		if p.SupportsClasses != tt.classes {
			t.Errorf("%s SupportsClasses = %v, want %v", tt.version, p.SupportsClasses, tt.classes)
		}
		if p.SupportsBinary != tt.binary {
			t.Errorf("%s SupportsBinary = %v, want %v", tt.version, p.SupportsBinary, tt.binary)
		}
		if p.UTF8 != tt.utf8 {
			t.Errorf("%s UTF8 = %v, want %v", tt.version, p.UTF8, tt.utf8)
		}
		if p.SupportsTrueColor != tt.trueColor {
			t.Errorf("%s SupportsTrueColor = %v, want %v", tt.version, p.SupportsTrueColor, tt.trueColor)
		}
	}
	if got := len(format.SupportedVersions()); got != 8 {
		t.Errorf("SupportedVersions() = %d, want 8", got)
	}
}

// --- Task 1.2: $ACADVER detection per fixture version ---

func TestVersionDetectionFixtures(t *testing.T) {
	tests := []struct{ file, want string }{
		{"testdata/Minimal_DXF_AC1006.dxf", format.AC1009}, // normalized from AC1006
		{"testdata/Minimal_DXF_AC1009.dxf", format.AC1009},
		{"testdata/mtext_columns_R2000.dxf", format.AC1015},
		{"testdata/columns_R2007.dxf", format.AC1021},
		{"testdata/parity_R2013.dxf", format.AC1027},
		{"testdata/columns_R2018.dxf", format.AC1032},
	}
	for _, tt := range tests {
		data, err := os.ReadFile(tt.file)
		if err != nil {
			t.Fatal(err)
		}
		d, err := FromStringData(string(data))
		if err != nil {
			t.Fatalf("%s: %v", tt.file, err)
		}
		if got := d.Header().Version; got != tt.want {
			t.Errorf("%s normalized version = %s, want %s", tt.file, got, tt.want)
		}
	}
}

// --- Task 1.3: normalization of pre-R12 and R14 input + load report ---

func TestVersionNormalization(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		changed bool
	}{
		{"AC1006", format.AC1009, true},
		{"AC1012", format.AC1009, true},
		{"AC1009", format.AC1009, false},
		{"AC1014", format.AC1015, true},
		{"AC1015", format.AC1015, false},
		{"AC1032", format.AC1032, false},
		{"AC9999", format.AC1015, true},
	}
	for _, tt := range tests {
		got, changed := format.NormalizeVersion(tt.in)
		if got != tt.want || changed != tt.changed {
			t.Errorf("NormalizeVersion(%s) = (%s,%v), want (%s,%v)", tt.in, got, changed, tt.want, tt.changed)
		}
	}

	d, err := FromFile("testdata/Minimal_DXF_AC1006.dxf")
	if err != nil {
		t.Fatal(err)
	}
	if d.SourceVersion() != "AC1006" {
		t.Errorf("SourceVersion = %q, want AC1006", d.SourceVersion())
	}
	if d.NormalizedVersion() != format.AC1009 {
		t.Errorf("NormalizedVersion = %q, want AC1009", d.NormalizedVersion())
	}
	rep := d.Conversion()
	if rep == nil {
		t.Fatal("Conversion report missing for normalized load")
	}
	if rep.SourceVersion != "AC1006" || rep.NormalizedVersion != format.AC1009 {
		t.Errorf("report = %+v", rep)
	}
}

// --- Task 2.1: codepage-aware ASCII decoding ---

func TestCodepageDecoding(t *testing.T) {
	// ANSI_1251 (cp1251): "Тест" in Windows-1251 bytes.
	raw := []byte{0xD2, 0xE5, 0xF1, 0xF2}
	got := lldxf.DecodeCodepageBytes(raw, "ANSI_1251")
	if got != "Тест" {
		t.Errorf("DecodeCodepageBytes(cp1251) = %q, want Тест", got)
	}
	back := lldxf.EncodeCodepageBytes(got, "ANSI_1251")
	if !bytes.Equal(back, raw) {
		t.Errorf("EncodeCodepageBytes round trip = %v, want %v", back, raw)
	}
	if n := lldxf.CodepageName("ANSI_1252"); n != "cp1252" {
		t.Errorf("CodepageName(ANSI_1252) = %s", n)
	}
	if n := lldxf.CodepageName(""); n != "cp1252" {
		t.Errorf("CodepageName(empty) = %s, want cp1252 default", n)
	}
}

func TestBinaryCodepageRoundTrip(t *testing.T) {
	// Build an R2000 drawing with a non-ASCII layer name in codepage cp1251.
	d, err := newDrawingWithLayerName("Тест")
	if err != nil {
		t.Fatal(err)
	}
	d.SetCodepage("ANSI_1251")
	d.SetTargetVersion(format.AC1015)

	var bin bytes.Buffer
	if _, err := d.WriteBinaryTo(&bin); err != nil {
		t.Fatal(err)
	}
	second, err := FromBinaryReader(bytes.NewReader(bin.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, l := range second.Layers {
		if l.Name() == "Тест" {
			found = true
		}
	}
	if !found {
		names := make([]string, 0, len(second.Layers))
		for n := range second.Layers {
			names = append(names, n)
		}
		t.Errorf("layer name not preserved through binary cp1251 round trip: %v", names)
	}
}

// --- Task 2.2: UTF-8 output for R2007+ and codepage output for older targets ---

func TestUTF8AndCodepageOutput(t *testing.T) {
	d, err := newDrawingWithLayerName("Ünïcödé")
	if err != nil {
		t.Fatal(err)
	}

	// R2013 target: UTF-8 bytes in output.
	d.SetTargetVersion(format.AC1027)
	var out bytes.Buffer
	if _, err := d.WriteTo(&out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Ünïcödé") {
		t.Error("R2013 ASCII output is not UTF-8")
	}

	// R2000 target with cp1252 codepage: Latin-1 bytes in output.
	d.SetCodepage("ANSI_1252")
	d.SetTargetVersion(format.AC1015)
	out.Reset()
	if _, err := d.WriteTo(&out); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "Ünïcödé") {
		t.Error("R2000 output should be codepage-encoded, not UTF-8")
	}
	enc := lldxf.EncodeCodepageBytes("Ünïcödé", "ANSI_1252")
	if !bytes.Contains(out.Bytes(), enc) {
		t.Error("R2000 output missing cp1252-encoded layer name")
	}

	// Round trip through the codepage target preserves the name.
	second, err := FromStringData(out.String())
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := second.Layers["Ünïcödé"]; !ok {
		t.Errorf("layer not restored after codepage round trip: %v", keys(second.Layers))
	}
}

// --- Task 2.3: complete binary writer + ASCII/binary equivalence ---

func TestBinaryASCIIEquivalence(t *testing.T) {
	data, err := os.ReadFile("testdata/columns_R2018.dxf")
	if err != nil {
		t.Fatal(err)
	}
	first, err := FromStringData(string(data))
	if err != nil {
		t.Fatal(err)
	}

	var asciiOut bytes.Buffer
	if _, err := first.WriteTo(&asciiOut); err != nil {
		t.Fatal(err)
	}
	ascii, err := FromStringData(asciiOut.String())
	if err != nil {
		t.Fatal(err)
	}

	var binOut bytes.Buffer
	if n, err := first.WriteBinaryTo(&binOut); err != nil {
		t.Fatal(err)
	} else if n == 0 {
		t.Fatal("binary writer produced no bytes")
	}
	binData := binOut.Bytes()
	if !bytes.HasPrefix(binData, []byte("AutoCAD Binary DXF\r\n\x1a\x00")) {
		t.Error("missing binary signature")
	}

	binary, err := FromBinaryReader(bytes.NewReader(binData))
	if err != nil {
		t.Fatal(err)
	}

	a, b := inventoryOf(first), inventoryOf(binary)
	equalStrings(t, "entities", a.Entities, b.Entities)
	equalStrings(t, "tables", a.Tables, b.Tables)
	equalStrings(t, "blocks", a.Blocks, b.Blocks)
	if len(a.Objects) > 0 && len(b.Objects) == 0 {
		t.Errorf("OBJECTS section lost in binary round trip: %d -> 0", len(a.Objects))
	}
	if !strings.HasSuffix(asciiOut.String(), "EOF\n") {
		t.Error("ASCII output missing EOF")
	}

	// Binary reload must see the same header version as ASCII reload.
	if ascii.Header().Version != binary.Header().Version {
		t.Errorf("version ascii=%s binary=%s", ascii.Header().Version, binary.Header().Version)
	}
}

// --- Task 3.1: save to requested target version ---

func TestSaveToTargetVersion(t *testing.T) {
	d, err := FromFile("testdata/Minimal_DXF_AC1009.dxf")
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{format.AC1015, format.AC1021, format.AC1032} {
		d.SetTargetVersion(target)
		var out bytes.Buffer
		if _, err := d.WriteTo(&out); err != nil {
			t.Fatalf("target %s: %v", target, err)
		}
		if !strings.Contains(out.String(), "1\n"+target+"\n") {
			t.Errorf("target %s: $ACADVER not declared in output", target)
		}
		rep := d.Conversion()
		if rep == nil || rep.SourceVersion != format.AC1009 || rep.NormalizedVersion != target {
			t.Errorf("target %s: conversion report = %+v", target, rep)
		}
	}

	// Unsupported target must error.
	d.SetTargetVersion("AC1099")
	if _, err := d.WriteTo(&bytes.Buffer{}); err == nil {
		t.Error("expected error for unsupported target version")
	}
	d.SetTargetVersion("")

	// Binary R12 must error.
	d.SetTargetVersion(format.AC1009)
	if _, err := d.WriteBinaryTo(&bytes.Buffer{}); err == nil {
		t.Error("expected error writing binary R12")
	}
	d.SetTargetVersion("")
}

// --- Task 3.2: R12 sanitizer + conversion report (downgrade R2018 fixture) ---

func TestR12Downgrade(t *testing.T) {
	data, err := os.ReadFile("testdata/columns_R2018.dxf")
	if err != nil {
		t.Fatal(err)
	}
	d, err := FromStringData(string(data))
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Entities()) == 0 {
		t.Fatal("fixture has no entities")
	}
	d.SetTargetVersion(format.AC1009)

	var out bytes.Buffer
	if _, err := d.WriteTo(&out); err != nil {
		t.Fatal(err)
	}
	s := out.String()

	// R12 restrictions: no OBJECTS/CLASSES sections.
	if strings.Contains(s, "2\nOBJECTS") {
		t.Error("R12 output contains OBJECTS section")
	}
	if strings.Contains(s, "2\nCLASSES") {
		t.Error("R12 output contains CLASSES section")
	}
	// No true color codes: 420 must not appear as a group code (value line after
	// a code line). Float values equal to 420 (e.g. $LIMMAX) are legitimate.
	if hasTrueColorCodeLine(s) {
		t.Error("R12 output contains true color group code 420")
	}

	rep := d.Conversion()
	if rep == nil || len(rep.Dropped) == 0 {
		t.Fatalf("expected conversion report with dropped content, got %+v", rep)
	}

	// R12 output must reload.
	r12, err := FromStringData(s)
	if err != nil {
		t.Fatalf("R12 output does not reload: %v", err)
	}
	sawEntity := false
	for _, e := range r12.Entities() {
		if !isR12EntityType(e.DXFType()) {
			t.Errorf("non-R12 entity survived in R12 output: %s", e.DXFType())
		}
		sawEntity = true
	}
	_ = sawEntity

	// Every dropped note must be non-empty.
	for _, note := range rep.Dropped {
		if note == "" {
			t.Error("empty dropped note")
		}
	}
}

func isR12EntityType(t string) bool {
	switch strings.ToUpper(t) {
	case "POLYLINE", "VERTEX", "3DFACE", "LINE", "ENDMK", "INSERT", "ARC", "SHAPE",
		"SOLID", "TRACE", "SEQEND", "TEXT", "CIRCLE", "POINT", "3DLINE":
		return true
	}
	return false
}

// --- Task 3.3: upgrade/downgrade round-trip checks ---

func TestVersionRoundTrips(t *testing.T) {
	data, err := os.ReadFile("testdata/columns_R2018.dxf")
	if err != nil {
		t.Fatal(err)
	}
	d, err := FromStringData(string(data))
	if err != nil {
		t.Fatal(err)
	}

	// Upgrade: R2000 -> R2013 preserves the typed inventory.
	d.SetTargetVersion(format.AC1015)
	var mid bytes.Buffer
	if _, err := d.WriteTo(&mid); err != nil {
		t.Fatal(err)
	}
	midD, err := FromStringData(mid.String())
	if err != nil {
		t.Fatal(err)
	}
	midD.SetTargetVersion(format.AC1027)
	var up bytes.Buffer
	if _, err := midD.WriteTo(&up); err != nil {
		t.Fatal(err)
	}
	upD, err := FromStringData(up.String())
	if err != nil {
		t.Fatal(err)
	}
	if upD.Header().Version != format.AC1027 {
		t.Errorf("upgrade version = %s, want AC1027", upD.Header().Version)
	}
	a, b := inventoryOf(midD), inventoryOf(upD)
	equalStrings(t, "entities (R2000->R2013)", a.Entities, b.Entities)

	// Downgrade: R2018 -> R12 records dropped content.
	down, err := FromStringData(string(data))
	if err != nil {
		t.Fatal(err)
	}
	down.SetTargetVersion(format.AC1009)
	var dOut bytes.Buffer
	if _, err := down.WriteTo(&dOut); err != nil {
		t.Fatal(err)
	}
	rep := down.Conversion()
	if rep == nil || len(rep.Dropped) == 0 {
		t.Errorf("downgrade report = %+v, want dropped notes", rep)
	}

	// Binary upgrade: R2018 -> binary R2007 reloads with same inventory.
	bin, err := FromStringData(string(data))
	if err != nil {
		t.Fatal(err)
	}
	bin.SetTargetVersion(format.AC1021)
	var bOut bytes.Buffer
	if _, err := bin.WriteBinaryTo(&bOut); err != nil {
		t.Fatal(err)
	}
	binD, err := FromBinaryReader(bytes.NewReader(bOut.Bytes()))
	if err != nil {
		t.Fatal(err)
	}
	c := inventoryOf(binD)
	equalStrings(t, "entities (R2018->binary R2007)", a.Entities, c.Entities)
}

// --- Task 4.1: cross-version fixtures load and save ---

func TestCrossVersionFixtures(t *testing.T) {
	files := []string{
		"testdata/Minimal_DXF_AC1009.dxf",
		"testdata/mtext_columns_R2000.dxf",
		"testdata/columns_R2007.dxf",
		"testdata/parity_R2013.dxf",
		"testdata/columns_R2018.dxf",
	}
	for _, fn := range files {
		data, err := os.ReadFile(fn)
		if err != nil {
			t.Fatal(err)
		}
		d, err := FromStringData(string(data))
		if err != nil {
			t.Fatalf("%s: %v", fn, err)
		}
		var out bytes.Buffer
		if _, err := d.WriteTo(&out); err != nil {
			t.Fatalf("%s save: %v", fn, err)
		}
		second, err := FromStringData(out.String())
		if err != nil {
			t.Fatalf("%s reload: %v", fn, err)
		}
		a, b := inventoryOf(d), inventoryOf(second)
		equalStrings(t, fn+" entities", a.Entities, b.Entities)

		// Binary save/reload for R14+ fixtures.
		if format.Profile(d.Header().Version).SupportsBinary {
			var bin bytes.Buffer
			if _, err := d.WriteBinaryTo(&bin); err != nil {
				t.Fatalf("%s binary save: %v", fn, err)
			}
			binD, err := FromBinaryReader(bytes.NewReader(bin.Bytes()))
			if err != nil {
				t.Fatalf("%s binary reload: %v", fn, err)
			}
			c := inventoryOf(binD)
			equalStrings(t, fn+" binary entities", a.Entities, c.Entities)
		}
	}

	// Binary fixture written by this package (R2018 -> binary R2018).
	data, err := os.ReadFile("testdata/columns_R2018.dxf")
	if err != nil {
		t.Fatal(err)
	}
	d, err := FromStringData(string(data))
	if err != nil {
		t.Fatal(err)
	}
	var bin bytes.Buffer
	if _, err := d.WriteBinaryTo(&bin); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("testdata/parity_binary_R2018.dxfbin", bin.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	binD, err := FromBinaryFile("testdata/parity_binary_R2018.dxfbin")
	if err != nil {
		t.Fatal(err)
	}
	a, c := inventoryOf(d), inventoryOf(binD)
	equalStrings(t, "binary fixture entities", a.Entities, c.Entities)
}

// --- helpers ---

func newDrawingWithLayerName(name string) (*drawing.Drawing, error) {
	d, err := drawing.New()
	if err != nil {
		return nil, err
	}
	l := table.NewLayer(name, 7, d.Layers["0"].LineType)
	d.Layers[name] = l
	for _, tbl := range d.Tables() {
		if tbl.Name() == "LAYER" {
			tbl.Add(l)
		}
	}
	return d, nil
}

func keys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// hasTrueColorCodeLine reports whether the ASCII stream contains a line equal
// to "420" that follows a group-code line, i.e. 420 appears as a group code
// (true color). Float values of exactly 420 (e.g. $LIMMAX coordinates) are
// written in %.6f form and never match.
func hasTrueColorCodeLine(s string) bool {
	lines := strings.Split(s, "\n")
	for i := 1; i+1 < len(lines); i++ {
		if lines[i] == "420" && isGroupCodeLine(lines[i-1]) && !isGroupCodeLine(lines[i+1]) {
			return true
		}
	}
	return false
}

func isGroupCodeLine(s string) bool {
	s = strings.TrimSpace(s)
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}
