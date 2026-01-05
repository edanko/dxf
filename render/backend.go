package render

import (
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/math"
)

// RenderBackend defines the interface for all rendering backends (SVG, PDF, PNG, etc.)
type RenderBackend interface {
	// Path operations
	BeginPath(ctx *RenderContext) error
	EndPath(ctx *RenderContext) error
	MoveTo(x, y float64) error
	LineTo(x, y float64) error
	Curve3To(ctrlX, ctrlY, endX, endY float64) error
	Curve4To(ctrl1X, ctrl1Y, ctrl2X, ctrl2Y, endX, endY float64) error
	ClosePath() error
	Fill() error
	Stroke() error

	// Style operations
	SetStrokeColor(c color.ColorNumber) error
	SetFillColor(c color.ColorNumber) error
	SetLineWidth(width float64) error
	SetLineDash(pattern []float64) error
	SetLineCap(cap string) error
	SetLineJoin(join string) error

	// Transformation operations
	SetTransform(matrix *math.Matrix44) error
	PushTransform() error
	PopTransform() error

	// Group operations
	BeginGroup(name string) error
	EndGroup() error

	// Metadata operations
	GetBounds() (float64, float64, float64, float64)
	GetSize() (float64, float64)

	// Output operations
	Flush() error
}
