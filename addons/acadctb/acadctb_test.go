package acadctb_test

import (
	"testing"

	"github.com/edanko/dxf/addons/acadctb"
	"github.com/edanko/dxf/entity"
)

func TestParseAndApply(t *testing.T) {
	data := []byte("# plot styles\nNormal,7,13,100\nThin,3,9,50\n")
	tab, err := acadctb.ParsePlotStyleTable(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(tab.Styles) != 2 {
		t.Fatalf("styles: %+v", tab.Styles)
	}
	n := tab.Styles[0]
	if n.Name != "Normal" || n.Color != 7 || n.Lineweight != 13 || n.Screening != 100 {
		t.Fatalf("first style: %+v", n)
	}
	if _, ok := tab.Style("Thin"); !ok {
		t.Fatal("Style lookup failed")
	}

	l := entity.NewLine()
	if err := tab.Apply("Thin", l); err != nil {
		t.Fatal(err)
	}
	if int(l.Color()) != 3 {
		t.Fatalf("applied color: %v", l.Color())
	}

	if _, err := acadctb.ParsePlotStyleTable([]byte{0x00, 0x01, 0x02}); err == nil {
		t.Fatal("binary CTB must be rejected")
	}
	if err := tab.Apply("NoSuchStyle", entity.NewLine()); err == nil {
		t.Fatal("unknown style must error")
	}
}
