package acadctb_test

import (
	"testing"

	"github.com/edanko/dxf/addons/acadctb"
)

func TestSTBNamedStyleOverrides(t *testing.T) {
	table, err := acadctb.ParsePlotStyleTable([]byte("Normal,0,0,100\nBlack,7,50,50\nThin,3,13,100\n"))
	if err != nil {
		t.Fatal(err)
	}
	if table.ByColor {
		t.Fatal("text STB must be a named table")
	}
	style, ok := table.Style("black")
	if !ok || style.Index != -1 {
		t.Fatalf("case-insensitive lookup = %+v, ok=%v", style, ok)
	}
	overrides, ok := table.Overrides("Black")
	if !ok {
		t.Fatal("missing Black overrides")
	}
	if !overrides.HasColor || overrides.Color != 7 {
		t.Fatalf("color override = %+v", overrides)
	}
	if !overrides.HasLineweight || overrides.Lineweight != 50 {
		t.Fatalf("lineweight override = %+v", overrides)
	}
	if !overrides.HasScreening || overrides.Screening != 50 {
		t.Fatalf("screening override = %+v", overrides)
	}
	if _, ok := table.Overrides("Normal"); !ok {
		t.Fatal("Normal style must resolve")
	}
	// Normal has no plot overrides.
	normal, _ := table.Style("Normal")
	no := acadctb.PlotOverrides{}
	if no.HasColor || no.HasLineweight || no.HasScreening {
		t.Fatal("zero overrides unexpectedly set")
	}
	if normal.Color != 0 || normal.Lineweight != 0 {
		t.Fatalf("Normal = %+v", normal)
	}
}
