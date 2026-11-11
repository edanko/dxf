package acadctb_test

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"testing"

	"github.com/edanko/dxf/addons/acadctb"
)

func TestParseBinaryCTB(t *testing.T) {
	body := `description="Test table
aci_table_available=TRUE
scale_factor=1.0
apply_factor=FALSE
custom_lineweight_display_units=0
aci_table{
 0="Color_1"
 1="Color_2"
}
plot_style{
 0{
  name="Color_1
  localized_name="Color_1
  description="
  color=-1
  color_policy=1
  physical_pen_number=0
  virtual_pen_number=0
  screen=50
  linepattern_size=0.5
  linetype=31
  adaptive_linetype=TRUE
  lineweight=13
  fill_style=73
  end_style=4
  join_style=5
 }
 1{
  name="Color_2
  localized_name="Color_2
  description="
  color=1
  mode_color=3271557121
  color_policy=1
  physical_pen_number=0
  virtual_pen_number=0
  screen=100
  linepattern_size=0.5
  linetype=31
  adaptive_linetype=TRUE
  lineweight=0
  fill_style=73
  end_style=4
  join_style=5
 }
}
custom_lineweight_table{
 0=0.00
 13=0.50
}
`
	data := makeBinaryPlotStyle(t, "CTBVER1", body)
	table, err := acadctb.ParsePlotStyleTable(data)
	if err != nil {
		t.Fatal(err)
	}
	if !table.ByColor {
		t.Fatal("expected a color-dependent table")
	}
	if len(table.Styles) != 2 {
		t.Fatalf("styles = %d", len(table.Styles))
	}
	first := table.Styles[0]
	if first.Index != 0 || first.Name != "1" || first.Screening != 50 || !first.ScreeningSet {
		t.Fatalf("first style = %+v", first)
	}
	if first.Lineweight != 50 {
		t.Fatalf("lineweight = %d", first.Lineweight)
	}
	second := table.Styles[1]
	if second.Color != 1 {
		t.Fatalf("second color = %d", second.Color)
	}
	overrides, ok := table.ColorOverrides(1)
	if !ok || !overrides.HasScreening || overrides.Screening != 50 || overrides.Lineweight != 50 {
		t.Fatalf("overrides = %+v, ok=%v", overrides, ok)
	}
}

func TestParseBinarySTBNamedStyles(t *testing.T) {
	body := `description="Named
plot_style{
 0{
  name="Normal
  description="
  color=-1
  screen=100
  lineweight=0
 }
 1{
  name="Black
  description="
  color=7
  screen=100
  lineweight=13
 }
}
`
	table, err := acadctb.ParsePlotStyleTable(makeBinaryPlotStyle(t, "STBVER1", body))
	if err != nil {
		t.Fatal(err)
	}
	if table.ByColor {
		t.Fatal("named table reported as color-dependent")
	}
	if _, ok := table.Style("Black"); !ok {
		t.Fatalf("styles = %+v", table.Styles)
	}
}

func makeBinaryPlotStyle(t *testing.T, kind, body string) []byte {
	t.Helper()
	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	if _, err := zw.Write([]byte(body + "\x00")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	out.WriteString("PIAFILEVERSION_2.0," + kind + ",compress\r\npmzlibcodec")
	var meta [12]byte
	binary.LittleEndian.PutUint32(meta[0:4], 0)
	binary.LittleEndian.PutUint32(meta[4:8], uint32(len(body)+1))
	binary.LittleEndian.PutUint32(meta[8:12], uint32(compressed.Len()))
	out.Write(meta[:])
	out.Write(compressed.Bytes())
	return out.Bytes()
}
