package render

import (
	"fmt"
	"io"
	"sort"
	"sync"

	dxfcolor "github.com/edanko/dxf/color"
	"github.com/edanko/dxf/math"
	"github.com/edanko/dxf/path"
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
	// FillGradient fills the boundary path with a linear gradient.
	// PDF backends fill with the first color (minimal support).
	FillGradient(boundary *path.Path, g *GradientFill) error
	// End finishes the canvas and writes the output.
	End() error
}

// StyledPath pairs a canvas-space path with the style it must be drawn with.
type StyledPath struct {
	Path  *path.Path
	Style Style
	// FillHatch, when set, replaces the plain boundary fill with hatch
	// pattern/gradient filling of the boundary.
	FillHatch *HatchFill
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

// GradientFill maps a linear gradient between two colors across the
// boundary bounding box. Rotation is the gradient axis angle in degrees.
type GradientFill struct {
	Color1, Color2 [3]float64 // RGB components 0..1
	Rotation       float64
	Centered       bool
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
