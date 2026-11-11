// Package r12export writes drawings in the R12 (AC1009) format while
// converting modern entities to R12-compatible forms.
package r12export

import (
	"fmt"
	"io"
	"strings"

	"github.com/edanko/dxf/addons"
	"github.com/edanko/dxf/addons/r12writer"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/filemanagement"
	dxfmath "github.com/edanko/dxf/math"
)

func init() { _ = addons.Register("r12export", capability{}) }

type capability struct{}

func (capability) Name() string { return "r12export" }
func (capability) Description() string {
	return "R12 DXF export with modern-entity conversion"
}

// Options controls entity conversion.
type Options struct {
	// Segments is the number of line segments used to approximate curves.
	// Values below 4 are replaced by the default of 16.
	Segments int
}

// Report describes conversions performed during an export.
type Report struct {
	Converted []string `json:"converted"`
	Dropped   []string `json:"dropped"`
}

// Export writes d as an R12 (AC1009) DXF stream to w.
//
// Supported modern entities are converted:
//   - SPLINE and ELLIPSE -> POLYLINE approximation
//   - LWPOLYLINE -> POLYLINE
//   - MTEXT -> TEXT
//   - MESH -> POLYFACE POLYLINE
//
// Other entity types are written unchanged and left to the R12 writer.
func Export(w io.Writer, d *drawing.Drawing, options ...Options) (Report, error) {
	var report Report
	if w == nil {
		return report, fmt.Errorf("r12export: nil writer")
	}
	if d == nil {
		return report, fmt.Errorf("r12export: nil drawing")
	}
	opts := Options{Segments: 16}
	if len(options) > 0 && options[0].Segments != 0 {
		opts = options[0]
	}
	if opts.Segments < 4 {
		opts.Segments = 16
	}

	writer := r12writer.New(w)
	writer.WriteHeader()
	for _, e := range d.Entities() {
		converted, kind := convertEntity(e, opts.Segments)
		if converted == nil {
			report.Dropped = append(report.Dropped, entityName(e))
			continue
		}
		if kind != "" {
			report.Converted = append(report.Converted, fmt.Sprintf("%s -> %s", entityName(e), kind))
		}
		writer.WriteEntity(converted)
	}
	writer.Close()
	return report, nil
}

// SaveAs writes d to path as an R12 DXF file using an atomic file replacement.
func SaveAs(path string, d *drawing.Drawing, options ...Options) (Report, error) {
	var buf strings.Builder
	report, err := Export(&buf, d, options...)
	if err != nil {
		return report, err
	}
	if err := filemanagement.AtomicWrite(path, []byte(buf.String()), false); err != nil {
		return report, fmt.Errorf("r12export: write %s: %w", path, err)
	}
	return report, nil
}

// ConvertEntities returns a converted copy of entities suitable for R12.
func ConvertEntities(entities []entity.Entity, segments int) ([]entity.Entity, Report) {
	if segments < 4 {
		segments = 16
	}
	var report Report
	result := make([]entity.Entity, 0, len(entities))
	for _, e := range entities {
		converted, kind := convertEntity(e, segments)
		if converted == nil {
			report.Dropped = append(report.Dropped, entityName(e))
			continue
		}
		if kind != "" {
			report.Converted = append(report.Converted, fmt.Sprintf("%s -> %s", entityName(e), kind))
		}
		result = append(result, converted)
	}
	return result, report
}

func convertEntity(e entity.Entity, segments int) (entity.Entity, string) {
	switch source := e.(type) {
	case *entity.Spline:
		return splineToPolyline(source, segments), "POLYLINE"
	case *entity.Ellipse:
		return ellipseToPolyline(source, segments), "POLYLINE"
	case *entity.LwPolyline:
		return lwPolylineToPolyline(source), "POLYLINE"
	case *entity.MText:
		return mtextToText(source), "TEXT"
	case *entity.Mesh:
		return meshToPolyface(source), "POLYFACE"
	default:
		if mesh, ok := e.(interface {
			Vertices() []dxfmath.Vec3
			Faces() [][]int
		}); ok {
			if polyline := meshSourceToPolyface(mesh); polyline != nil {
				copyLayer(e, polyline)
				return polyline, "POLYFACE"
			}
		}
		return e, ""
	}
}

// SplineToPolyline approximates s with a POLYLINE.
func SplineToPolyline(s *entity.Spline, segments int) *entity.Polyline {
	if segments < 4 {
		segments = 16
	}
	polyline := entity.NewPolyline()
	polyline.Flag = 0
	for i := 0; i <= segments; i++ {
		point := s.PointAt(float64(i) / float64(segments))
		polyline.AddVertex(point.X(), point.Y(), point.Z())
	}
	if s.IsClosed() {
		polyline.Flag = entity.PolylineFlagClosed
	}
	copyLayer(s, polyline)
	return polyline
}

// EllipseToPolyline approximates the ellipse arc with a POLYLINE.
func EllipseToPolyline(e *entity.Ellipse, segments int) *entity.Polyline {
	if segments < 4 {
		segments = 16
	}
	polyline := entity.NewPolyline()
	polyline.Flag = 0
	closed := e.Start == 0 && e.End == 360
	limit := segments
	if closed {
		limit = segments - 1
	}
	for i := 0; i <= limit; i++ {
		point := e.PointAt(float64(i) / float64(segments))
		polyline.AddVertex(point.X(), point.Y(), point.Z())
	}
	if closed {
		polyline.Flag = entity.PolylineFlagClosed
	}
	copyLayer(e, polyline)
	return polyline
}

// LwPolylineToPolyline converts a lightweight polyline to a classic POLYLINE.
func LwPolylineToPolyline(source *entity.LwPolyline) *entity.Polyline {
	polyline := entity.NewPolyline()
	polyline.Flag = 0
	closed := source.Closed
	for i, vertex := range source.Vertices {
		x, y := 0.0, 0.0
		if len(vertex) > 0 {
			x = vertex[0]
		}
		if len(vertex) > 1 {
			y = vertex[1]
		}
		v := polyline.AddVertex(x, y, 0)
		if i < len(source.Bulges) {
			v.Bulge = source.Bulges[i]
		}
	}
	if closed {
		polyline.Flag = entity.PolylineFlagClosed
	}
	copyLayer(source, polyline)
	return polyline
}

// MTextToText converts MTEXT content to a TEXT entity.
func MTextToText(source *entity.MText) *entity.Text {
	text := entity.NewText()
	text.Coord1 = append([]float64{}, source.Coord1...)
	text.Coord2 = append([]float64{}, source.Coord2...)
	text.Height = source.Height
	text.Rotation = source.Rotation
	text.Value = mtextValue(source)
	copyLayer(source, text)
	return text
}

func splineToPolyline(s *entity.Spline, segments int) *entity.Polyline {
	return SplineToPolyline(s, segments)
}

func ellipseToPolyline(e *entity.Ellipse, segments int) *entity.Polyline {
	return EllipseToPolyline(e, segments)
}

func lwPolylineToPolyline(l *entity.LwPolyline) *entity.Polyline {
	return LwPolylineToPolyline(l)
}

func mtextToText(m *entity.MText) *entity.Text {
	return MTextToText(m)
}

func meshToPolyface(m *entity.Mesh) *entity.Polyline {
	polyline := meshSourceToPolyface(m)
	if polyline != nil {
		copyLayer(m, polyline)
	}
	return polyline
}

func meshSourceToPolyface(mesh interface {
	Vertices() []dxfmath.Vec3
	Faces() [][]int
}) *entity.Polyline {
	vertices := mesh.Vertices()
	faces := mesh.Faces()
	if len(vertices) == 0 {
		return nil
	}
	polyline := entity.NewPolyline()
	polyline.Flag = entity.PolylineFlagPolyface
	for _, vertex := range vertices {
		polyline.AddPolyfaceVertex(vertex.X(), vertex.Y(), vertex.Z())
	}
	for _, face := range faces {
		if len(face) < 3 {
			continue
		}
		indices := make([]int, len(face))
		for i, index := range face {
			indices[i] = index + 1 // Polyline.AddFace expects 1-based indices.
		}
		polyline.AddFace(indices...)
	}
	return polyline
}

func mtextValue(m *entity.MText) string {
	value := m.Value
	if len(m.AdditionalText) > 0 {
		if value != "" {
			value += "\\P"
		}
		value += strings.Join(m.AdditionalText, "\\P")
	}
	return value
}

func copyLayer(source, target entity.Entity) {
	if layer := source.Layer(); layer != nil {
		target.SetLayer(layer)
	}
}

func entityName(e entity.Entity) string {
	handle := e.Handle()
	if handle == "" {
		return e.DXFType()
	}
	return e.DXFType() + "[" + handle + "]"
}
