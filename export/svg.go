package export

import (
	"fmt"
	"math"
	"strings"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

type SVGExporter struct {
	drawing   *drawing.Drawing
	precision int
}

func NewSVGExporter(d *drawing.Drawing) *SVGExporter {
	return &SVGExporter{
		drawing:   d,
		precision: 4,
	}
}

func (e *SVGExporter) SetPrecision(precision int) {
	if precision > 0 && precision <= 10 {
		e.precision = precision
	}
}

func (e *SVGExporter) Export() string {
	var sb strings.Builder

	entities := e.drawing.Entities()
	if len(entities) == 0 {
		return ""
	}

	// Calculate bounding box
	minX, minY := 1e10, 1e10
	maxX, maxY := -1e10, -1e10

	for _, ent := range entities {
		bboxMin, bboxMax := ent.BBox()
		if len(bboxMin) >= 2 {
			if bboxMin[0] < minX {
				minX = bboxMin[0]
			}
			if bboxMin[1] < minY {
				minY = bboxMin[1]
			}
		}
		if len(bboxMax) >= 2 {
			if bboxMax[0] > maxX {
				maxX = bboxMax[0]
			}
			if bboxMax[1] > maxY {
				maxY = bboxMax[1]
			}
		}
	}

	// Add some padding
	width := maxX - minX
	height := maxY - minY
	if width < 100 {
		width = 100
	}
	if height < 100 {
		height = 100
	}

	padding := 10.0
	minX -= padding
	minY -= padding
	width += 2 * padding
	height += 2 * padding

	// Start SVG
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%.2f" height="%.2f" viewBox="0 0 %.2f %.2f">
`, width, height, width, height))
	sb.WriteString(fmt.Sprintf(`<rect width="100%%" height="100%%" fill="white"/>
`))

	// Transform coordinate system
	// DXF Y is up, SVG Y is down
	// DXF origin is at bottom-left of viewBox
	// We need to flip Y and offset

	for _, ent := range entities {
		e.writeEntity(&sb, ent, minX, minY, height)
	}

	sb.WriteString("</svg>")
	return sb.String()
}

func (e *SVGExporter) writeEntity(sb *strings.Builder, ent entity.Entity, offsetX, offsetY, height float64) {
	switch ent := ent.(type) {
	case *entity.Line:
		e.writeLine(sb, ent, offsetX, offsetY, height)
	case *entity.Circle:
		e.writeCircle(sb, ent, offsetX, offsetY, height)
	case *entity.Arc:
		e.writeArc(sb, ent, offsetX, offsetY, height)
	case *entity.LwPolyline:
		e.writeLwPolyline(sb, ent, offsetX, offsetY, height)
	case *entity.Polyline:
		e.writePolyline(sb, ent, offsetX, offsetY, height)
	case *entity.Point:
		e.writePoint(sb, ent, offsetX, offsetY, height)
	}
}

func (e *SVGExporter) formatCoord(coord float64) string {
	return fmt.Sprintf("%.*f", e.precision, coord)
}

func (e *SVGExporter) writeLine(sb *strings.Builder, line *entity.Line, offsetX, offsetY, height float64) {
	if len(line.Start) < 2 || len(line.End) < 2 {
		return
	}

	x1 := line.Start[0] - offsetX
	y1 := height - (line.Start[1] - offsetY)
	x2 := line.End[0] - offsetX
	y2 := height - (line.End[1] - offsetY)

	color := e.getColor(line)
	sb.WriteString(fmt.Sprintf(`<line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s" stroke-width="1"/>
`,
		e.formatCoord(x1), e.formatCoord(y1),
		e.formatCoord(x2), e.formatCoord(y2), color))
}

func (e *SVGExporter) writeCircle(sb *strings.Builder, circle *entity.Circle, offsetX, offsetY, height float64) {
	if len(circle.Center) < 2 {
		return
	}

	cx := circle.Center[0] - offsetX
	cy := height - (circle.Center[1] - offsetY)
	r := circle.Radius
	if r < 0 {
		r = -r
	}

	color := e.getColor(circle)
	sb.WriteString(fmt.Sprintf(`<circle cx="%s" cy="%s" r="%s" fill="none" stroke="%s" stroke-width="1"/>
`,
		e.formatCoord(cx), e.formatCoord(cy), e.formatCoord(r), color))
}

func (e *SVGExporter) writeArc(sb *strings.Builder, arc *entity.Arc, offsetX, offsetY, height float64) {
	if len(arc.Center) < 2 || len(arc.Angle) < 2 {
		return
	}

	cx := arc.Center[0] - offsetX
	cy := height - (arc.Center[1] - offsetY)
	r := arc.Radius
	if r < 0 {
		r = -r
	}

	// Convert angles from degrees to radians
	startAngle := arc.Angle[0]
	endAngle := arc.Angle[1]

	// SVG arc: A rx ry x-axis-rotation large-arc-flag sweep-flag x y
	largeArc := 0
	sweep := 1
	if endAngle < startAngle {
		endAngle += 360
	}
	if endAngle-startAngle > 180 {
		largeArc = 1
	}

	// Calculate end point
	startRad := startAngle * math.Pi / 180

	ex := cx + r*math.Cos(-startRad) // Negative for SVG coordinate system
	ey := cy + r*math.Sin(-startRad)

	color := e.getColor(arc)
	sb.WriteString(fmt.Sprintf(`<path d="M %s %s A %s %s 0 %d %d %s %s" fill="none" stroke="%s" stroke-width="1"/>
`,
		e.formatCoord(ex), e.formatCoord(ey),
		e.formatCoord(r), e.formatCoord(r),
		largeArc, sweep,
		e.formatCoord(ex), e.formatCoord(ey),
		color))
}

func (e *SVGExporter) writeLwPolyline(sb *strings.Builder, lp *entity.LwPolyline, offsetX, offsetY, height float64) {
	if len(lp.Vertices) == 0 {
		return
	}

	var points []string
	for _, v := range lp.Vertices {
		if len(v) >= 2 {
			x := v[0] - offsetX
			y := height - (v[1] - offsetY)
			points = append(points, fmt.Sprintf("%s %s", e.formatCoord(x), e.formatCoord(y)))
		}
	}

	if len(points) == 0 {
		return
	}

	color := e.getColor(lp)
	closed := lp.Closed
	stroke := "stroke"
	fill := "fill"
	if closed {
		fill = e.getColor(lp)
		stroke = "none"
	}

	sb.WriteString(fmt.Sprintf(`<polygon points="%s" %s="%s" %s="%s" stroke-width="1"/>
`,
		strings.Join(points, " "), fill, color, stroke, color))
}

func (e *SVGExporter) writePolyline(sb *strings.Builder, pl *entity.Polyline, offsetX, offsetY, height float64) {
	if len(pl.Vertices) == 0 {
		return
	}

	var points []string
	for _, v := range pl.Vertices {
		if len(v.Coord) >= 2 {
			x := v.Coord[0] - offsetX
			y := height - (v.Coord[1] - offsetY)
			points = append(points, fmt.Sprintf("%s %s", e.formatCoord(x), e.formatCoord(y)))
		}
	}

	if len(points) == 0 {
		return
	}

	color := e.getColor(pl)
	sb.WriteString(fmt.Sprintf(`<polyline points="%s" fill="none" stroke="%s" stroke-width="1"/>
`,
		strings.Join(points, " "), color))
}

func (e *SVGExporter) writePoint(sb *strings.Builder, p *entity.Point, offsetX, offsetY, height float64) {
	if len(p.Coord) < 2 {
		return
	}

	x := p.Coord[0] - offsetX
	y := height - (p.Coord[1] - offsetY)

	color := e.getColor(p)
	sb.WriteString(fmt.Sprintf(`<circle cx="%s" cy="%s" r="2" fill="%s"/>
`,
		e.formatCoord(x), e.formatCoord(y), color))
}

func (e *SVGExporter) getColor(ent entity.Entity) string {
	layer := ent.Layer()
	if layer != nil {
		return "black"
	}
	return "black"
}
