package drawing_test

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/addons/drawing"
	dxfdrawing "github.com/edanko/dxf/drawing"
)

func TestRenderSVG(t *testing.T) {
	d, err := dxfdrawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddLine(0, 0, 0, 10, 10, 0); err != nil {
		t.Fatal(err)
	}
	out, err := drawing.RenderSVG(d, 800, 600, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "<path") {
		t.Fatalf("svg output has no path:\n%s", out)
	}
}
