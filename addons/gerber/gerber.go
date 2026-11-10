// Package gerber exports Gerber RS-274X photoplotter output.
package gerber

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/edanko/dxf/addons"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

type capability struct{}

func (capability) Name() string        { return "gerber" }
func (capability) Description() string { return "Gerber RS-274X photoplotter output" }

func init() { _ = addons.Register("gerber", capability{}) }

// ExportGerber renders the drawing as RS-274X with a 3.6 coordinate format
// (inches). LINE entities become draw segments; CIRCLE entities become full
// arcs.
func ExportGerber(d *drawing.Drawing) string {
	var b strings.Builder
	b.WriteString("%FSLAX36Y36*%\n")
	b.WriteString("%MOIN*%\n")
	b.WriteString("G01*\n")
	for _, e := range d.Entities() {
		switch t := e.(type) {
		case *entity.Line:
			fmt.Fprintf(&b, "G01X%sY%sD02*\n", coord(t.Start[0]), coord(t.Start[1]))
			fmt.Fprintf(&b, "G01X%sY%sD01*\n", coord(t.End[0]), coord(t.End[1]))
		case *entity.Circle:
			if len(t.Center) < 2 {
				continue
			}
			sx := t.Center[0] + t.Radius
			fmt.Fprintf(&b, "G75*\n")
			fmt.Fprintf(&b, "G03X%sY%sI%sJ%sD01*\n",
				coord(sx), coord(t.Center[1]), coord(-t.Radius), coord(0))
		}
	}
	b.WriteString("M02*\n")
	return b.String()
}

// coord formats a drawing unit as a 3.6 Gerber integer (micro-inches).
func coord(v float64) string {
	return strconv.FormatInt(int64(math.Round(v*1e6)), 10)
}
