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

// PlotStyle is one named plot style rule.
type PlotStyle struct {
	Name       string
	Color      int // ACI color index (0 = ByBlock, 256 = ByLayer)
	Lineweight int // lineweight in 1/100 mm
	Screening  int // screening percentage 0..100
}

// PlotStyleTable is a parsed plot style table.
type PlotStyleTable struct {
	Name   string
	Styles []PlotStyle
}

// ParsePlotStyleTable parses the text (STB) dialect: one
// 'Name,Color,Lineweight,Screening' record per line. Binary CTB files are
// rejected with a descriptive error.
func ParsePlotStyleTable(data []byte) (*PlotStyleTable, error) {
	if bytes.IndexByte(data, 0) >= 0 {
		return nil, fmt.Errorf("acadctb: binary CTB files are not supported; convert to STB text")
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
		st := PlotStyle{Name: strings.TrimSpace(parts[0])}
		st.Color, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
		if len(parts) > 2 {
			st.Lineweight, _ = strconv.Atoi(strings.TrimSpace(parts[2]))
		}
		if len(parts) > 3 {
			st.Screening, _ = strconv.Atoi(strings.TrimSpace(parts[3]))
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
	return PlotStyle{}, false
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
