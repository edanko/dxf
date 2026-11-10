package hpgl2_test

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/addons/hpgl2"
	"github.com/edanko/dxf/drawing"
)

func TestExportHPGL2(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddLine(0, 0, 0, 10, 5, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddCircle(5, 5, 0, 2); err != nil {
		t.Fatal(err)
	}
	out := hpgl2.ExportHPGL2(d)
	for _, want := range []string{"IN;", "PD", "CI", "SP0;"} {
		if !strings.Contains(out, want) {
			t.Fatalf("hpgl2 output missing %q: %s", want, out)
		}
	}
}
