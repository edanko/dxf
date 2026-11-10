// Package xplayer exports animation frames as SVG renders.
package xplayer

import (
	"bytes"
	"fmt"
	"math"

	"github.com/edanko/dxf/addons"
	"github.com/edanko/dxf/drawing"
	dxfmath "github.com/edanko/dxf/math"
	"github.com/edanko/dxf/render"
)

type capability struct{}

func (capability) Name() string        { return "xplayer" }
func (capability) Description() string { return "animation frame export (SVG)" }

func init() { _ = addons.Register("xplayer", capability{}) }

// ExportAnimationSVG renders one SVG frame per rotation step about the
// scene center. frames must be positive.
func ExportAnimationSVG(d *drawing.Drawing, frames int, width, height float64, opts *render.SceneOptions) ([]string, error) {
	if d == nil {
		return nil, fmt.Errorf("xplayer: nil drawing")
	}
	if frames <= 0 {
		return nil, fmt.Errorf("xplayer: frames must be positive")
	}
	sc, err := render.BuildScene(d, width, height, opts)
	if err != nil {
		return nil, err
	}
	cx := (sc.ExtMinX + sc.ExtMaxX) / 2
	cy := (sc.ExtMinY + sc.ExtMaxY) / 2
	out := make([]string, 0, frames)
	for i := range frames {
		ang := 2 * math.Pi * float64(i) / float64(frames)
		m := dxfmath.Translation(cx, cy, 0).
			Mul(dxfmath.RotationZ(ang)).
			Mul(dxfmath.Translation(-cx, -cy, 0))
		var buf bytes.Buffer
		r, err := render.NewBackend("svg", &buf, width, height)
		if err != nil {
			return nil, err
		}
		if err := r.Begin(width, height); err != nil {
			return nil, err
		}
		for _, sp := range sc.Paths {
			p := sp.Path
			if i > 0 {
				p = sp.Path.Clone()
				p.Transform(&m)
			}
			if err := r.Clip(sp.Clip); err != nil {
				return nil, err
			}
			if err := r.SetStyle(sp.Style); err != nil {
				return nil, err
			}
			if err := r.DrawPath(p); err != nil {
				return nil, err
			}
			if err := r.Clip(nil); err != nil {
				return nil, err
			}
		}
		if err := r.End(); err != nil {
			return nil, err
		}
		out = append(out, buf.String())
	}
	return out, nil
}
