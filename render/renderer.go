package render

import (
	"fmt"
	"io"
	"sort"
	"sync"

	dxfcolor "github.com/edanko/dxf/color"
	"github.com/edanko/dxf/math"
	"github.com/edanko/dxf/path"

	"image"
)

// RenderContext carries global rendering parameters for the legacy
// imperative RenderBackend pipeline.
type RenderContext struct {
	BackgroundColor dxfcolor.ColorNumber
	LineWidth       float64
	Scale           float64
	Transform       math.Matrix44
}

// NewRenderContext returns a context with white background, unit scale
// and identity transform.
func NewRenderContext() *RenderContext {
	return &RenderContext{
		BackgroundColor: dxfcolor.White,
		LineWidth:       1.0,
		Scale:           1.0,
		Transform:       math.Identity(),
	}
}

// Renderer draws styled paths on a fixed-size canvas. Canvas coordinates
// have the origin in the top-left corner; the scene builder applies the
// model-to-canvas fit transform (including the Y flip) before drawing.
type Renderer interface {
	// Begin starts a canvas of the given size in canvas units.
	Begin(width, height float64) error
	// SetStyle selects the style used by subsequent DrawPath calls.
	SetStyle(style Style) error
	// DrawPath fills (when Style.Fill is set) and strokes (when
	// Style.Stroke is set) the path under the current clip.
	DrawPath(p *path.Path) error
	// Clip restricts subsequent drawing to the interior of the path.
	// A nil path resets the clip to the full canvas.
	Clip(p *path.Path) error
	// DrawImage blits a raster image at canvas position (x, y) with the
	// given size in canvas units, rotated by rotDeg about its center and
	// faded by opacity (0..1). Backends without raster support may ignore it.
	DrawImage(im image.Image, x, y, w, h, rotDeg, opacity float64) error
	// FillGradient fills the boundary path with a linear or radial
	// gradient. Backends that cannot render a gradient fall back to a
	// solid fill and report it through Diagnostics when supported.
	FillGradient(boundary *path.Path, g *GradientFill) error
	// End finishes the canvas and writes the output.
	End() error
}

// DiagnosticRenderer is implemented by backends that record degradations
// (for example unsupported color spaces or malformed gradients) while
// rendering. Diagnostics are stable for the same drawing and options.
type DiagnosticRenderer interface {
	Renderer
	Diagnostics() []string
}

// StyledPath pairs a canvas-space path with the style it must be drawn with.
type StyledPath struct {
	Path  *path.Path
	Style Style
	// FillHatch, when set, replaces the plain boundary fill with hatch
	// pattern/gradient filling of the boundary.
	// Clip restricts drawing of this path to the given boundary
	// (canvas space). Nil means no clipping.
	Clip      *path.Path
	FillHatch *HatchFill
}

// StyledImage is a raster image placed on the canvas.
type StyledImage struct {
	Image    image.Image
	X, Y     float64 // top-left corner in canvas units
	W, H     float64 // size in canvas units
	Rotation float64 // degrees about the image center
	Opacity  float64 // 0..1
	Clip     *path.Path
}

// HatchFill describes pattern or gradient filling of a hatch boundary.
// Exactly one of Pattern and Gradient may be set.
type HatchFill struct {
	Pattern  *PatternFill
	Gradient *GradientFill
	Scale    float64
	Angle    float64
}

// PatternFill is one line group of a hatch pattern in unit (unscaled) space.
type PatternFill struct {
	Name  string
	Lines []PatternLine
}

// PatternLine is one dashed line of a hatch pattern: direction angle in
// degrees, base point and repeating offset in unit space, and the dash
// cycle (positive dash, negative gap, zero dot).
type PatternLine struct {
	Angle  float64
	Base   [2]float64
	Offset [2]float64
	Dashes []float64
}

// GradientStop is one color stop of a gradient: Offset runs from 0 (start)
// to 1 (end) and Color holds RGB components in 0..1.
type GradientStop struct {
	Offset float64
	Color  [3]float64
}

// GradientFill describes a gradient across the boundary bounding box.
// Rotation is the axis angle in degrees for linear gradients. Type is
// "linear" (default when empty) or "radial"; radial gradients use Center
// (canvas units, default boundary center) and Radius (default half the
// boundary diagonal).
//
// Stops lists all color stops in ascending offset order. When Stops is
// empty the two-color Color1/Color2 form is used. ColorSpace is the color
// space the stops were authored in; empty and "DeviceRGB"/"RGB" are
// supported, other values are reported and rendered as a solid fill.
type GradientFill struct {
	Color1, Color2 [3]float64 // RGB components 0..1
	Rotation       float64
	Centered       bool

	Stops      []GradientStop
	Type       string
	Center     [2]float64
	Radius     float64
	ColorSpace string
}

// GradientStops returns the effective color stops in ascending offset
// order, deriving them from Color1/Color2 when Stops is empty.
func (g *GradientFill) GradientStops() []GradientStop {
	if g == nil {
		return nil
	}
	if len(g.Stops) > 0 {
		return g.Stops
	}
	return []GradientStop{{Offset: 0, Color: g.Color1}, {Offset: 1, Color: g.Color2}}
}

// IsRadial reports whether the gradient is radial.
func (g *GradientFill) IsRadial() bool {
	return g != nil && g.Type == "radial"
}

// SupportedColorSpace reports whether the gradient's color space can be
// rendered directly. Unknown color spaces fall back to a solid fill.
func (g *GradientFill) SupportedColorSpace() bool {
	switch g.ColorSpace {
	case "", "RGB", "DeviceRGB":
		return true
	}
	return false
}

// ValidStops reports whether the effective stops can form a gradient: at
// least two stops with strictly ascending offsets in 0..1.
func (g *GradientFill) ValidStops() bool {
	stops := g.GradientStops()
	if len(stops) < 2 {
		return false
	}
	prev := -1.0
	for _, s := range stops {
		if s.Offset < 0 || s.Offset > 1 || s.Offset <= prev {
			return false
		}
		prev = s.Offset
	}
	return true
}

// BackendFactory creates a renderer writing to w on a width x height canvas.
type BackendFactory func(w io.Writer, width, height float64) (Renderer, error)

var (
	registryMu sync.RWMutex
	registry   = map[string]BackendFactory{}
)

// Register makes a backend available under name, replacing any
// previous registration of the same name.
func Register(name string, factory BackendFactory) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[name] = factory
}

// NewBackend returns the renderer registered under name.
func NewBackend(name string, w io.Writer, width, height float64) (Renderer, error) {
	registryMu.RLock()
	factory, ok := registry[name]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("render: unknown backend %q", name)
	}
	return factory(w, width, height)
}

// BackendNames lists the registered backend names in sorted order.
func BackendNames() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
