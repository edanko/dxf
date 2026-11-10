package gerber_test

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/addons/gerber"
	"github.com/edanko/dxf/drawing"
)

func TestExportGerber(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddLine(0, 0, 0, 10, 10, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddCircle(5, 5, 0, 2); err != nil {
		t.Fatal(err)
	}
	out := gerber.ExportGerber(d)
	for _, want := range []string{"%FSLAX36Y36*%", "G01*", "D02*", "D01*", "G03", "M02*"} {
		if !strings.Contains(out, want) {
			t.Fatalf("gerber output missing %q:\n%s", want, out)
		}
	}
}
