package text2path

import (
	"math"
	"testing"
)

func TestMeasureTTF(t *testing.T) {
	m, name, err := MeasureTTF("ABC", 10)
	if err != nil {
		t.Skipf("no system TTF font: %v", err)
	}
	if name == "" {
		t.Fatal("expected font name")
	}
	if m.Height != 10 || m.Width <= 0 || m.Ascent <= 0 {
		t.Fatalf("metrics = %+v", m)
	}
	wide, _, err := MeasureTTF("ABCABC", 10)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(wide.Width-2*m.Width) > 1e-6 {
		t.Fatalf("width scaling: %v vs %v", wide.Width, 2*m.Width)
	}
}

func TestMeasureTTFRejectsInvalidHeight(t *testing.T) {
	if _, _, err := MeasureTTF("A", 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestTextMetricsForSHX(t *testing.T) {
	font := &SHXFont{Name: "test.shx", Above: 20, Glyphs: map[int]*SHXGlyph{
		'A': {Number: 65, Name: "A", Codes: []int{160, 0}},
	}}
	m, name, err := TextMetricsForHeight(font, "AA", 10)
	if err != nil {
		t.Fatal(err)
	}
	if name != "test.shx" || m.Width <= 0 || m.Height != 10 {
		t.Fatalf("metrics = %+v name=%q", m, name)
	}
}
