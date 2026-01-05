package render

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/edanko/dxf/color"
	dxfmath "github.com/edanko/dxf/math"
)

// PDFBackend implements RenderBackend for PDF output
type PDFBackend struct {
	writer  io.Writer
	width   float64
	height  float64
	pageNum int
	content strings.Builder

	// PDF state
	currentPath strings.Builder
	strokeColor color.ColorNumber
	fillColor   color.ColorNumber
	lineWidth   float64

	// Transform stack
	transformStack   []dxfmath.Matrix44
	currentTransform dxfmath.Matrix44
}

// NewPDFBackend creates a new PDF backend
func NewPDFBackend(w io.Writer, width, height float64) *PDFBackend {
	identity := dxfmath.Identity()
	return &PDFBackend{
		writer:           w,
		width:            width,
		height:           height,
		pageNum:          1,
		content:          strings.Builder{},
		currentTransform: identity,
		transformStack:   make([]dxfmath.Matrix44, 0),
		strokeColor:      1, // color.Black
		fillColor:        1, // color.Black
		lineWidth:        1.0,
	}
}

// BeginPath starts a new PDF path
func (pdf *PDFBackend) BeginPath(ctx *RenderContext) error {
	pdf.currentPath.Reset()
	pdf.currentPath.WriteString("m ")
	return nil
}

// EndPath ends the current PDF path
func (pdf *PDFBackend) EndPath(ctx *RenderContext) error {
	pdf.content.WriteString(pdf.currentPath.String())
	pdf.currentPath.Reset()
	return nil
}

// MoveTo moves to a point without drawing
func (pdf *PDFBackend) MoveTo(x, y float64) error {
	pdf.currentPath.WriteString(fmt.Sprintf("%.3f %.3f m ", x, y))
	return nil
}

// LineTo draws a line to a point
func (pdf *PDFBackend) LineTo(x, y float64) error {
	pdf.currentPath.WriteString(fmt.Sprintf("%.3f %.3f l ", x, y))
	return nil
}

// Curve3To draws a quadratic Bezier curve
func (pdf *PDFBackend) Curve3To(ctrlX, ctrlY, endX, endY float64) error {
	pdf.currentPath.WriteString(fmt.Sprintf("%.3f %.3f %.3f %.3f v ", ctrlX, ctrlY, endX, endY))
	return nil
}

// Curve4To draws a cubic Bezier curve
func (pdf *PDFBackend) Curve4To(ctrl1X, ctrl1Y, ctrl2X, ctrl2Y, endX, endY float64) error {
	pdf.currentPath.WriteString(fmt.Sprintf("%.3f %.3f %.3f %.3f %.3f %.3f c ",
		ctrl1X, ctrl1Y, ctrl2X, ctrl2Y, endX, endY))
	return nil
}

// ClosePath closes the current path
func (pdf *PDFBackend) ClosePath() error {
	pdf.currentPath.WriteString("h ")
	return nil
}

// Fill fills the current path
func (pdf *PDFBackend) Fill() error {
	pdf.content.WriteString("f ")
	return nil
}

// Stroke strokes the current path
func (pdf *PDFBackend) Stroke() error {
	pdf.content.WriteString("S ")
	return nil
}

// SetStrokeColor sets the stroke color
func (pdf *PDFBackend) SetStrokeColor(c color.ColorNumber) error {
	pdf.strokeColor = c
	rgb := colorNumberToRGB(c)
	pdf.content.WriteString(fmt.Sprintf("%.3f %.3f %.3f RG ",
		float64(rgb.R)/255.0, float64(rgb.G)/255.0, float64(rgb.B)/255.0))
	return nil
}

// SetFillColor sets the fill color
func (pdf *PDFBackend) SetFillColor(c color.ColorNumber) error {
	pdf.fillColor = c
	rgb := colorNumberToRGB(c)
	pdf.content.WriteString(fmt.Sprintf("%.3f %.3f %.3f rg ",
		float64(rgb.R)/255.0, float64(rgb.G)/255.0, float64(rgb.B)/255.0))
	return nil
}

// SetLineWidth sets the line width
func (pdf *PDFBackend) SetLineWidth(width float64) error {
	pdf.lineWidth = width
	pdf.content.WriteString(fmt.Sprintf("%.3f w ", width))
	return nil
}

// SetLineDash sets the line dash pattern
func (pdf *PDFBackend) SetLineDash(pattern []float64) error {
	if len(pattern) == 0 {
		pdf.content.WriteString("[] 0 d ")
		return nil
	}

	pdf.content.WriteString("[")
	for i, val := range pattern {
		if i > 0 {
			pdf.content.WriteString(" ")
		}
		pdf.content.WriteString(strconv.FormatFloat(val, 'f', 3, 64))
	}
	pdf.content.WriteString("] 0 d ")
	return nil
}

// SetLineCap sets the line cap style
func (pdf *PDFBackend) SetLineCap(cap string) error {
	var capVal int
	switch cap {
	case "butt":
		capVal = 0
	case "round":
		capVal = 1
	case "square":
		capVal = 2
	default:
		capVal = 0
	}
	pdf.content.WriteString(fmt.Sprintf("%d J ", capVal))
	return nil
}

// SetLineJoin sets the line join style
func (pdf *PDFBackend) SetLineJoin(join string) error {
	var joinVal int
	switch join {
	case "miter":
		joinVal = 0
	case "round":
		joinVal = 1
	case "bevel":
		joinVal = 2
	default:
		joinVal = 0
	}
	pdf.content.WriteString(fmt.Sprintf("%d j ", joinVal))
	return nil
}

// SetTransform applies a transformation matrix
func (pdf *PDFBackend) SetTransform(matrix *dxfmath.Matrix44) error {
	pdf.currentTransform = *matrix

	// Convert to PDF transformation matrix
	// PDF uses [a b 0 c d 0 e f] format for 2D transforms
	// We'll extract the 2D part of the 4x4 matrix
	a := matrix.Get(0, 0)
	b := matrix.Get(0, 1)
	c := matrix.Get(1, 0)
	d := matrix.Get(1, 1)
	e := matrix.Get(0, 3)

	pdf.content.WriteString(fmt.Sprintf("%.6f %.6f %.6f %.6f %.6f cm ", a, b, c, d, e))
	return nil
}

// PushTransform pushes current transform to stack
func (pdf *PDFBackend) PushTransform() error {
	pdf.transformStack = append(pdf.transformStack, pdf.currentTransform)
	return nil
}

// PopTransform pops transform from stack
func (pdf *PDFBackend) PopTransform() error {
	if len(pdf.transformStack) == 0 {
		return fmt.Errorf("transform stack is empty")
	}

	lastIdx := len(pdf.transformStack) - 1
	pdf.currentTransform = pdf.transformStack[lastIdx]
	pdf.transformStack = pdf.transformStack[:lastIdx]

	pdf.SetTransform(&pdf.currentTransform)
	return nil
}

// BeginGroup starts a new PDF group
func (pdf *PDFBackend) BeginGroup(name string) error {
	pdf.content.WriteString("q ")
	return nil
}

// EndGroup ends the current PDF group
func (pdf *PDFBackend) EndGroup() error {
	pdf.content.WriteString("Q ")
	return nil
}

// GetBounds returns the current drawing bounds
func (pdf *PDFBackend) GetBounds() (float64, float64, float64, float64) {
	return 0, 0, pdf.width, pdf.height
}

// GetSize returns the current drawing size
func (pdf *PDFBackend) GetSize() (float64, float64) {
	return pdf.width, pdf.height
}

// Flush writes the PDF content
func (pdf *PDFBackend) Flush() error {
	pdf.writePDFHeader()
	pdf.writePageContent()
	pdf.writePDFFooter()
	return nil
}

// writeString helper function to write to writer
func (pdf *PDFBackend) writeString(s string) {
	if writer, ok := pdf.writer.(io.StringWriter); ok {
		writer.WriteString(s)
		return
	}
	// Fallback for other Writer types
	pdf.writer.Write([]byte(s))
}

// writePDFHeader writes PDF header
func (pdf *PDFBackend) writePDFHeader() {
	pdf.writeString("%PDF-1.4\n")
	pdf.writeString("1 0 obj\n")
	pdf.writeString("<<\n")
	pdf.writeString("/Type /Catalog\n")
	pdf.writeString("/Pages 2 0 R\n")
	pdf.writeString(">>\n")
	pdf.writeString("endobj\n")

	pdf.writeString("2 0 obj\n")
	pdf.writeString("<<\n")
	pdf.writeString("/Type /Pages\n")
	pdf.writeString("/Kids [3 0 R]\n")
	pdf.writeString("/Count 1\n")
	pdf.writeString(">>\n")
	pdf.writeString("endobj\n")
}

// writePageContent writes page content
func (pdf *PDFBackend) writePageContent() {
	content := pdf.content.String()

	// Page object
	pdf.writeString("3 0 obj\n")
	pdf.writeString("<<\n")
	pdf.writeString("/Type /Page\n")
	pdf.writeString("/Parent 2 0 R\n")
	pdf.writeString(fmt.Sprintf("/MediaBox [0 0 %.3f %.3f]\n", pdf.width, pdf.height))
	pdf.writeString("/Contents 4 0 R\n")
	pdf.writeString(">>\n")
	pdf.writeString("endobj\n")

	// Content stream
	pdf.writeString("4 0 obj\n")
	pdf.writeString("<<\n")
	pdf.writeString("/Length " + strconv.Itoa(len(content)) + "\n")
	pdf.writeString(">>\n")
	pdf.writeString("stream\n")
	pdf.writeString(content)
	pdf.writeString("\nendstream\n")
	pdf.writeString("endobj\n")
}

// writePDFFooter writes PDF footer
func (pdf *PDFBackend) writePDFFooter() {
	pdf.writeString("xref\n")
	pdf.writeString("0 5\n")
	pdf.writeString("0000000000 65535 f\n")

	// For simplicity, using approximate offsets
	pdf.writeString("0000000010 00000 n\n")
	pdf.writeString("0000000079 00000 n\n")
	pdf.writeString("0000000173 00000 n\n")
	pdf.writeString("0000000301 00000 n\n")
	pdf.writeString("0000000440 00000 n\n")

	pdf.writeString("trailer\n")
	pdf.writeString("<<\n")
	pdf.writeString("/Size 5\n")
	pdf.writeString("/Root 1 0 R\n")
	pdf.writeString(">>\n")
	pdf.writeString("startxref\n")
	pdf.writeString("550\n")
	pdf.writeString("%%EOF\n")
}
