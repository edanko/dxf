// Package hpgl2 exports drawings as HPGL/2 plotter commands.
package hpgl2

import (
	"fmt"
	"strings"

	"github.com/edanko/dxf/addons"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

func init() { _ = addons.Register("hpgl2", capability{}) }

type capability struct{}

func (capability) Name() string        { return "hpgl2" }
func (capability) Description() string { return "HPGL/2 plotter export" }

// CommandRunner invokes an external command (injectable for tests).
type CommandRunner interface {
	Run(name string, args []string, stdin []byte) (stdout, stderr []byte, err error)
}

// ExportHPGL2 renders the drawing entities as HPGL/2 commands.
func ExportHPGL2(d *drawing.Drawing) string {
	if d == nil {
		return "IN;SP0;"
	}
	var b strings.Builder
	b.WriteString("IN;SP1;")
	for _, e := range d.Entities() {
		switch t := e.(type) {
		case *entity.Line:
			fmt.Fprintf(&b, "PU%.2f,%.2f;PD%.2f,%.2f;", t.Start[0], t.Start[1], t.End[0], t.End[1])
		case *entity.Circle:
			fmt.Fprintf(&b, "PU%.2f,%.2f;CI%.2f;", t.Center[0], t.Center[1], t.Radius)
		}
	}
	b.WriteString("PU0,0;SP0;")
	return b.String()
}

// PlotWithRunner sends the HPGL/2 stream to a plotter command.
func PlotWithRunner(runner CommandRunner, d *drawing.Drawing) error {
	if runner == nil {
		return fmt.Errorf("hpgl2: no command runner provided")
	}
	_, _, err := runner.Run("plotter", nil, []byte(ExportHPGL2(d)))
	return err
}
