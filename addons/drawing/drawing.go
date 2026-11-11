// Package drawing renders documents through the core renderer backends.
package drawing

import (
	"bytes"

	"github.com/edanko/dxf/addons"
	dxfdrawing "github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/render"
)

type capability struct{}

func (capability) Name() string        { return "drawing" }
func (capability) Description() string { return "render drawings via the core renderer backends" }

func init() { _ = addons.Register("drawing", capability{}) }

// RenderSVG renders the drawing to an SVG document.
func RenderSVG(d *dxfdrawing.Drawing, width, height float64, opts *render.SceneOptions) ([]byte, error) {
	return Render(d, width, height, "svg", opts)
}

// Render renders the drawing with the named backend ("svg", "pdf" or "png")
// at the given canvas size, honoring the scene options.
func Render(d *dxfdrawing.Drawing, width, height float64, backend string, opts *render.SceneOptions) ([]byte, error) {
	sc, err := render.BuildScene(d, width, height, opts)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	r, err := render.NewBackend(backend, &buf, width, height)
	if err != nil {
		return nil, err
	}
	if err := r.Begin(width, height); err != nil {
		return nil, err
	}
	for _, sp := range sc.Paths {
		if err := r.Clip(sp.Clip); err != nil {
			return nil, err
		}
		if err := r.SetStyle(sp.Style); err != nil {
			return nil, err
		}
		if sp.FillHatch != nil && sp.FillHatch.Gradient != nil {
			if err := r.FillGradient(sp.Path, sp.FillHatch.Gradient); err != nil {
				return nil, err
			}
			continue
		}
		if err := r.DrawPath(sp.Path); err != nil {
			return nil, err
		}
		if err := r.Clip(nil); err != nil {
			return nil, err
		}
	}
	for _, ri := range sc.Rasters {
		if err := r.Clip(ri.Clip); err != nil {
			return nil, err
		}
		if err := r.DrawImage(ri.Image, ri.X, ri.Y, ri.W, ri.H, ri.Rotation, ri.Opacity); err != nil {
			return nil, err
		}
		if err := r.Clip(nil); err != nil {
			return nil, err
		}
	}
	if err := r.End(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
