package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/math"
)

// colorNumberToRGB converts ColorNumber to RGB values (simplified)
func colorNumberToRGB(c color.ColorNumber) struct{ R, G, B int } {
	switch c {
	case color.Red:
		return struct{ R, G, B int }{255, 0, 0}
	case color.Yellow:
		return struct{ R, G, B int }{255, 255, 0}
	case color.Green:
		return struct{ R, G, B int }{0, 255, 0}
	case color.Cyan:
		return struct{ R, G, B int }{0, 255, 255}
	case color.Blue:
		return struct{ R, G, B int }{0, 0, 255}
	case color.Magenta:
		return struct{ R, G, B int }{255, 0, 255}
	case color.White:
		return struct{ R, G, B int }{255, 255, 255}
	default:
		return struct{ R, G, B int }{0, 0, 0}
	}
}

// SVGBackend implements RenderBackend for SVG output
type SVGBackend struct {
	writer io.Writer
	width  float64
	height float64
	buffer strings.Builder
}

// NewSVGBackend creates a new SVG backend
func NewSVGBackend(w io.Writer) *SVGBackend {
	return &SVGBackend{
		writer: w,
		buffer: strings.Builder{},
	}
}

// BeginPath starts a new SVG path
func (svg *SVGBackend) BeginPath(ctx *RenderContext) error {
	svg.buffer.WriteString(`<path d="`)
	return nil
}

// EndPath ends the current SVG path
func (svg *SVGBackend) EndPath(ctx *RenderContext) error {
	svg.buffer.WriteString(" />")
	return nil
}

// MoveTo moves to a point without drawing
func (svg *SVGBackend) MoveTo(x, y float64) error {
	svg.buffer.WriteString(fmt.Sprintf("M%.3f,%.3f", x, y))
	return nil
}

// LineTo draws a line to a point
func (svg *SVGBackend) LineTo(x, y float64) error {
	svg.buffer.WriteString(fmt.Sprintf("L%.3f,%.3f", x, y))
	return nil
}

// Curve3To draws a quadratic Bezier curve
func (svg *SVGBackend) Curve3To(ctrlX, ctrlY, endX, endY float64) error {
	svg.buffer.WriteString(fmt.Sprintf("Q%.3f,%.3f %.3f,%.3f", ctrlX, ctrlY, endX, endY))
	return nil
}

// Curve4To draws a cubic Bezier curve
func (svg *SVGBackend) Curve4To(ctrl1X, ctrl1Y, ctrl2X, ctrl2Y, endX, endY float64) error {
	svg.buffer.WriteString(fmt.Sprintf("C%.3f,%.3f %.3f,%.3f %.3f,%.3f", ctrl1X, ctrl1Y, ctrl2X, ctrl2Y, endX, endY))
	return nil
}

// SetStrokeColor sets the stroke color
func (svg *SVGBackend) SetStrokeColor(c color.ColorNumber) error {
	// Convert color to SVG format (simplified)
	rgb := colorNumberToRGB(c)
	svg.buffer.WriteString(fmt.Sprintf(`" stroke="rgb(%d,%d,%d)"`, rgb.R, rgb.G, rgb.B))
	return nil
}

// SetFillColor sets the fill color
func (svg *SVGBackend) SetFillColor(c color.ColorNumber) error {
	rgb := colorNumberToRGB(c)
	svg.buffer.WriteString(fmt.Sprintf("\" fill=\"rgb(%d,%d,%d)\"", rgb.R, rgb.G, rgb.B))
	return nil
}

// SetLineWidth sets the line width
func (svg *SVGBackend) SetLineWidth(width float64) error {
	svg.buffer.WriteString(fmt.Sprintf(` stroke-width="%.6f"`, width))
	return nil
}

// ClosePath closes the current path
func (svg *SVGBackend) ClosePath() error {
	svg.buffer.WriteString("Z")
	return nil
}

// Fill fills the current path
func (svg *SVGBackend) Fill() error {
	// Fill is handled by attributes in SVG
	return nil
}

// Stroke strokes the current path
func (svg *SVGBackend) Stroke() error {
	// Stroke is handled by attributes in SVG
	return nil
}

// BeginGroup starts a new SVG group
func (svg *SVGBackend) BeginGroup(name string) error {
	svg.buffer.WriteString(fmt.Sprintf(`<g id="%s">`, name))
	return nil
}

// EndGroup ends the current SVG group
func (svg *SVGBackend) EndGroup() error {
	svg.buffer.WriteString(`</g>`)
	return nil
}

// SetTransform applies a transformation matrix
func (svg *SVGBackend) SetTransform(matrix *math.Matrix44) error {
	if matrix == nil {
		return nil
	}

	// Extract transform from matrix
	a, b, c, d, e := matrix[0], matrix[1], matrix[2], matrix[3], matrix[4]

	// Create SVG transform string
	transform := fmt.Sprintf("matrix(%g,%g,%g,%g,%g,%g)", a, b, c, d, e, matrix[5])
	svg.buffer.WriteString(fmt.Sprintf(" transform=\"%s\"", transform))
	return nil
}

// GetBounds returns the current drawing bounds
func (svg *SVGBackend) GetBounds() (float64, float64, float64, float64) {
	return 0, 0, svg.width, svg.height
}

// GetSize returns the current drawing size
func (svg *SVGBackend) GetSize() (float64, float64) {
	return svg.width, svg.height
}

// WriteSVGHeader writes SVG header
func (svg *SVGBackend) WriteSVGHeader() error {
	svg.buffer.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%.6f" height="%.6f" viewBox="0 0 %.6f %.6f">`, svg.width, svg.height, svg.width, svg.height))
	return nil
}

// WriteSVGFooter writes SVG footer
func (svg *SVGBackend) WriteSVGFooter() error {
	svg.buffer.WriteString(`</svg>`)
	return nil
}

// Flush writes the SVG content to the underlying writer
func (svg *SVGBackend) Flush() error {
	_, err := svg.writer.Write([]byte(svg.buffer.String()))
	svg.buffer.Reset()
	return err
}
