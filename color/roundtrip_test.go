package color_test

import (
	"bytes"
	"testing"

	dxf "github.com/edanko/dxf"
	dxfcolor "github.com/edanko/dxf/color"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

func TestTransparencyRoundTrip(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	ln, err := d.Model().AddLine(0, 0, 0, 1, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	ln.SetTransparency(0.5)
	var buf bytes.Buffer
	if _, err := d.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	doc, err := dxf.FromStringData(buf.String())
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range doc.Entities() {
		if l, ok := e.(*entity.Line); ok {
			if l.Transparency() != 0.5 {
				t.Fatalf("transparency: %v", l.Transparency())
			}
			return
		}
	}
	t.Fatal("line not reloaded")
}

func TestTransparencyValue(t *testing.T) {
	tr, err := dxfcolor.NewTransparency(0.25)
	if err != nil {
		t.Fatal(err)
	}
	if tr.Percent() != 25 {
		t.Fatalf("percent: %d", tr.Percent())
	}
}
