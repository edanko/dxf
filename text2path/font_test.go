package text2path

import (
	"testing"

	"github.com/edanko/dxf/path"
)

func TestTextToPathsTTF(t *testing.T) {
	ps, name, err := TextToPathsTTF("AB", 10)
	if err != nil {
		t.Skipf("no system TrueType font: %v", err)
	}
	if name == "" {
		t.Fatal("empty font name")
	}
	if len(ps) != 1 {
		t.Fatalf("paths %d, want 1 merged", len(ps))
	}
	p := ps[0]
	lo, hi := p.BoundingBox()
	if hi.X() <= 5 {
		t.Errorf("advance width %v, want > 5", hi.X())
	}
	if lo.Y() < -0.4*10 || hi.Y() > 1.3*10 {
		t.Errorf("bbox y [%v, %v] outside baseline band", lo.Y(), hi.Y())
	}
	if len(p.Vertices()) < 40 {
		t.Errorf("vertices %d, want dense contours", len(p.Vertices()))
	}
	movements := 0
	for _, c := range p.Commands() {
		if _, ok := c.(*path.PathMoveTo); ok {
			movements++
		}
	}
	if movements < 3 {
		t.Errorf("closed subpaths (MoveTo) %d, want >= 3 for A+B glyphs", movements)
	}
}

func TestTextToPathsTTFCaching(t *testing.T) {
	_, n1, err1 := TextToPathsTTF("A", 4)
	if err1 != nil {
		t.Skipf("no system TrueType font: %v", err1)
	}
	_, n2, err2 := TextToPathsTTF("B", 6)
	if err2 != nil || n1 != n2 {
		t.Fatalf("cache inconsistency: %q/%v vs %q/%v", n1, err1, n2, err2)
	}
}

func TestTextToPathsTTFEmpty(t *testing.T) {
	ps, _, err := TextToPathsTTF("", 10)
	if err != nil || ps != nil {
		t.Fatalf("empty text: %v %v", ps, err)
	}
	ps, _, err = TextToPathsTTF("A", 0)
	if err != nil || ps != nil {
		t.Fatalf("zero height: %v %v", ps, err)
	}
}
