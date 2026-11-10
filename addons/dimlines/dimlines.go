// Package dimlines builds simple dimension-line geometry from ordinary DXF
// entities. It is a pure-Go alternative to ezdxf's legacy dxfwrite-compatible
// dimension helper: the result is a collection of LINE and TEXT entities on
// the requested layer, not a DIMENSION entity.
package dimlines

import (
	"fmt"
	"math"
	"strconv"

	"github.com/edanko/dxf/addons"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

func init() { _ = addons.Register("dimlines", capability{}) }

type capability struct{}

func (capability) Name() string { return "dimlines" }
func (capability) Description() string {
	return "dimension-line helpers built from LINE and TEXT entities"
}

// Style controls the geometry and text of generated dimension lines.
type Style struct {
	Layer           string
	TextHeight      float64
	ArrowSize       float64
	ExtensionLength float64
	TextGap         float64
	Prefix          string
	Suffix          string
	Decimals        int
}

// DefaultStyle returns a usable style on layer "DIMENSIONS".
func DefaultStyle() Style {
	return Style{
		Layer:           "DIMENSIONS",
		TextHeight:      0.5,
		ArrowSize:       0.25,
		ExtensionLength: 0.25,
		TextGap:         0.1,
		Decimals:        2,
	}
}

func (s Style) normalized() Style {
	if s.Layer == "" {
		s.Layer = "DIMENSIONS"
	}
	if s.TextHeight <= 0 {
		s.TextHeight = 0.5
	}
	if s.ArrowSize <= 0 {
		s.ArrowSize = 0.25
	}
	if s.ExtensionLength < 0 {
		s.ExtensionLength = 0
	}
	if s.TextGap < 0 {
		s.TextGap = 0
	}
	if s.Decimals < 0 {
		s.Decimals = 0
	}
	return s
}

// LinearDimension is a horizontal or vertical measurement between two points,
// offset from the measured line.
type LinearDimension struct {
	Start  dxfmath.Vec3
	End    dxfmath.Vec3
	Offset float64
	Style  Style
}

// NewLinearDimension creates a linear dimension with the default style.
func NewLinearDimension(start, end dxfmath.Vec3, offset float64) *LinearDimension {
	return &LinearDimension{Start: start, End: end, Offset: offset, Style: DefaultStyle()}
}

// Entities returns the dimension line, extension lines, arrow ticks, and text.
func (d *LinearDimension) Entities() ([]entity.Entity, error) {
	style := d.Style.normalized()
	dx := d.End.X() - d.Start.X()
	dy := d.End.Y() - d.Start.Y()
	length := math.Hypot(dx, dy)
	if length <= dxfmath.DefaultTolerance {
		return nil, fmt.Errorf("dimlines: dimension points are coincident")
	}
	ux, uy := dx/length, dy/length
	nx, ny := -uy, ux

	start := dxfmath.NewVec3(d.Start.X()+nx*d.Offset, d.Start.Y()+ny*d.Offset, d.Start.Z())
	end := dxfmath.NewVec3(d.End.X()+nx*d.Offset, d.End.Y()+ny*d.Offset, d.End.Z())
	startExt := dxfmath.NewVec3(start.X()+nx*style.ExtensionLength, start.Y()+ny*style.ExtensionLength, start.Z())
	endExt := dxfmath.NewVec3(end.X()+nx*style.ExtensionLength, end.Y()+ny*style.ExtensionLength, end.Z())

	entities := []entity.Entity{
		line(d.Start, startExt, style.Layer),
		line(d.End, endExt, style.Layer),
		line(start, end, style.Layer),
	}
	entities = append(entities,
		tick(start, ux, uy, style.ArrowSize, style.Layer),
		tick(end, ux, uy, style.ArrowSize, style.Layer),
	)

	mid := dxfmath.NewVec3(
		(start.X()+end.X())/2+nx*style.TextGap,
		(start.Y()+end.Y())/2+ny*style.TextGap,
		(start.Z()+end.Z())/2,
	)
	text := entity.NewText()
	text.Coord1 = []float64{mid.X(), mid.Y(), mid.Z()}
	text.Height = style.TextHeight
	text.Value = style.Prefix + strconv.FormatFloat(length, 'f', style.Decimals, 64) + style.Suffix
	text.SetLayerName(style.Layer)
	entities = append(entities, text)
	return entities, nil
}

// AddTo adds the generated entities to a drawing.
func (d *LinearDimension) AddTo(doc *drawing.Drawing) error {
	entities, err := d.Entities()
	if err != nil {
		return err
	}
	for _, e := range entities {
		if err := doc.AddEntity(e); err != nil {
			return err
		}
	}
	return nil
}

// AngularDimension measures the angle between two rays from a center point.
type AngularDimension struct {
	Center dxfmath.Vec3
	Start  dxfmath.Vec3
	End    dxfmath.Vec3
	Radius float64
	Style  Style
}

// NewAngularDimension creates an angular dimension with the default style.
func NewAngularDimension(center, start, end dxfmath.Vec3, radius float64) *AngularDimension {
	return &AngularDimension{Center: center, Start: start, End: end, Radius: radius, Style: DefaultStyle()}
}

// Entities returns an arc-like polyline between the rays, extension lines, and
// text in degrees.
func (d *AngularDimension) Entities() ([]entity.Entity, error) {
	style := d.Style.normalized()
	startAngle := math.Atan2(d.Start.Y()-d.Center.Y(), d.Start.X()-d.Center.X())
	endAngle := math.Atan2(d.End.Y()-d.Center.Y(), d.End.X()-d.Center.X())
	delta := endAngle - startAngle
	for delta < 0 {
		delta += 2 * math.Pi
	}
	if delta <= dxfmath.DefaultTolerance {
		return nil, fmt.Errorf("dimlines: angular dimension has zero angle")
	}
	if d.Radius <= 0 {
		return nil, fmt.Errorf("dimlines: angular dimension radius must be positive")
	}

	segments := 16
	points := make([]dxfmath.Vec3, 0, segments+1)
	for i := 0; i <= segments; i++ {
		angle := startAngle + delta*float64(i)/float64(segments)
		points = append(points, dxfmath.NewVec3(
			d.Center.X()+d.Radius*math.Cos(angle),
			d.Center.Y()+d.Radius*math.Sin(angle),
			d.Center.Z(),
		))
	}

	var entities []entity.Entity
	for i := 1; i < len(points); i++ {
		entities = append(entities, line(points[i-1], points[i], style.Layer))
	}
	entities = append(entities,
		line(d.Center, points[0], style.Layer),
		line(d.Center, points[len(points)-1], style.Layer),
	)

	midAngle := startAngle + delta/2
	mid := dxfmath.NewVec3(
		d.Center.X()+(d.Radius+style.TextGap)*math.Cos(midAngle),
		d.Center.Y()+(d.Radius+style.TextGap)*math.Sin(midAngle),
		d.Center.Z(),
	)
	text := entity.NewText()
	text.Coord1 = []float64{mid.X(), mid.Y(), mid.Z()}
	text.Height = style.TextHeight
	text.Value = style.Prefix + strconv.FormatFloat(delta*180/math.Pi, 'f', style.Decimals, 64) + "°" + style.Suffix
	text.SetLayerName(style.Layer)
	entities = append(entities, text)
	return entities, nil
}

// AddTo adds the generated entities to a drawing.
func (d *AngularDimension) AddTo(doc *drawing.Drawing) error {
	entities, err := d.Entities()
	if err != nil {
		return err
	}
	for _, e := range entities {
		if err := doc.AddEntity(e); err != nil {
			return err
		}
	}
	return nil
}

func line(start, end dxfmath.Vec3, layer string) *entity.Line {
	l := entity.NewLine()
	l.Start = []float64{start.X(), start.Y(), start.Z()}
	l.End = []float64{end.X(), end.Y(), end.Z()}
	l.SetLayerName(layer)
	return l
}

func tick(point dxfmath.Vec3, ux, uy, size float64, layer string) *entity.Line {
	l := entity.NewLine()
	l.Start = []float64{point.X() - ux*size, point.Y() - uy*size, point.Z()}
	l.End = []float64{point.X() + ux*size, point.Y() + uy*size, point.Z()}
	l.SetLayerName(layer)
	return l
}
