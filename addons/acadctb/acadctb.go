// Package acadctb parses AutoCAD plot style tables (STB text dialect) and
// applies them to entities.
package acadctb

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/edanko/dxf/addons"
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/entity"
)

type capability struct{}

func (capability) Name() string        { return "acadctb" }
func (capability) Description() string { return "plot style table parsing and application" }

func init() { _ = addons.Register("acadctb", capability{}) }

// PlotStyle is one named plot style rule shared by STB and CTB tables.
type PlotStyle struct {
	Name          string
	Index         int // zero-based CTB style index; -1 for named styles
	Description   string
	Color         int     // ACI color index (0 = ByBlock, 256 = ByLayer)
	TrueColor     uint32  // 0x00BBGGRR true color; valid when HasTrueColor
	HasTrueColor  bool    // TrueColor is present
	Lineweight    int     // lineweight in 1/100 mm
	LineType      string  // plot linetype name
	LinetypeScale float64 // linetype scale
	Screening     int     // screening percentage 0..100
	ScreeningSet  bool    // Screening was explicitly configured
	FillStyle     int     // solid fill style (CTB/STB extension)
	EndStyle      int     // line end style
	Flags         int     // parser-specific style flags
}

// PlotStyleTable is a parsed STB or CTB plot style table.
type PlotStyleTable struct {
	Name        string
	Description string
	Styles      []PlotStyle
	ByColor     bool // CTB color-dependent table
}

// PlotOverrides is a render-time style override resolved from a plot style.
// Unset fields are represented by the Has* flags.
type PlotOverrides struct {
	Color         int
	HasColor      bool
	TrueColor     uint32
	HasTrueColor  bool
	Lineweight    int
	HasLineweight bool
	Screening     int
	HasScreening  bool
	LineType      string
}

// ParsePlotStyleTable parses both text (STB CSV dialect) and compressed binary
// CTB/STB files. Binary tables are detected by the PIAFILEVERSION header.
func ParsePlotStyleTable(data []byte) (*PlotStyleTable, error) {
	if bytes.HasPrefix(data, []byte("PIAFILEVERSION")) {
		return parseBinaryPlotStyleTable(data)
	}
	if bytes.IndexByte(data, 0) >= 0 {
		return nil, fmt.Errorf("acadctb: unrecognized binary plot style table")
	}
	t := &PlotStyleTable{}
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			return nil, fmt.Errorf("acadctb: malformed plot style record %q", line)
		}
		st := PlotStyle{Name: strings.TrimSpace(parts[0]), Index: -1}
		st.Color, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
		if len(parts) > 2 {
			st.Lineweight, _ = strconv.Atoi(strings.TrimSpace(parts[2]))
		}
		if len(parts) > 3 {
			st.Screening, _ = strconv.Atoi(strings.TrimSpace(parts[3]))
			st.ScreeningSet = true
		}
		t.Styles = append(t.Styles, st)
	}
	if len(t.Styles) == 0 {
		return nil, fmt.Errorf("acadctb: no plot styles found")
	}
	return t, nil
}

// Style returns the named plot style.
func (t *PlotStyleTable) Style(name string) (PlotStyle, bool) {
	for _, s := range t.Styles {
		if s.Name == name {
			return s, true
		}
	}
	for _, s := range t.Styles {
		if strings.EqualFold(s.Name, name) {
			return s, true
		}
	}
	return PlotStyle{}, false
}

// StyleByColor returns the CTB style configured for an ACI color. Color
// names may also be numeric strings in STB-converted color tables.
func (t *PlotStyleTable) StyleByColor(aci int) (PlotStyle, bool) {
	for _, s := range t.Styles {
		if s.Index == aci-1 {
			return s, true
		}
	}
	name := strconv.Itoa(aci)
	for _, s := range t.Styles {
		if s.Name == name {
			return s, true
		}
	}
	return PlotStyle{}, false
}

// Overrides resolves a named plot style to render-time overrides. Color 0
// (ByBlock) and 256 (ByLayer) are intentionally left unset so rendering can
// keep the object's resolved color.
func (t *PlotStyleTable) Overrides(name string) (PlotOverrides, bool) {
	s, ok := t.Style(name)
	if !ok {
		return PlotOverrides{}, false
	}
	return styleOverrides(s), true
}

// ColorOverrides resolves a CTB color-dependent plot style for an ACI color.
func (t *PlotStyleTable) ColorOverrides(aci int) (PlotOverrides, bool) {
	s, ok := t.StyleByColor(aci)
	if !ok {
		return PlotOverrides{}, false
	}
	return styleOverrides(s), true
}

func styleOverrides(s PlotStyle) PlotOverrides {
	o := PlotOverrides{LineType: s.LineType}
	if s.Color > 0 && s.Color < 256 {
		o.Color = s.Color
		o.HasColor = true
	}
	if s.HasTrueColor {
		o.TrueColor = s.TrueColor
		o.HasTrueColor = true
	}
	if s.Lineweight != 0 {
		o.Lineweight = s.Lineweight
		o.HasLineweight = true
	}
	if s.ScreeningSet || s.Screening != 0 {
		o.Screening = s.Screening
		o.HasScreening = true
	}
	return o
}

// Apply applies a named plot style to an entity: ACI color and lineweight
// are written to the entity when it supports them. Screening is a
// plot-time-only property and is not stored on entities.
func (t *PlotStyleTable) Apply(styleName string, e entity.Entity) error {
	s, ok := t.Style(styleName)
	if !ok {
		return fmt.Errorf("acadctb: plot style %q not found", styleName)
	}
	if s.Color > 0 {
		if c, ok := e.(entity.ColorSetter); ok {
			c.SetColor(color.ColorNumber(s.Color))
		}
	}
	if s.Lineweight != 0 {
		if lw, ok := e.(entity.LineWeightSetter); ok {
			lw.SetLineWeight(s.Lineweight)
		}
	}
	return nil
}
