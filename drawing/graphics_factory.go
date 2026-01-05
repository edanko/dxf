package drawing

import (
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

// GraphicsFactory provides high-level entity creation methods
type GraphicsFactory struct {
	drawing *Drawing
}

// NewGraphicsFactory creates a new graphics factory for given drawing
func NewGraphicsFactory(d *Drawing) *GraphicsFactory {
	return &GraphicsFactory{
		drawing: d,
	}
}

// Point creates a point entity at specified coordinates
func (gf *GraphicsFactory) Point(x, y, z float64) *entity.Point {
	point := entity.NewPoint()
	point.Coord = []float64{x, y, z}
	gf.drawing.AddEntity(point)
	return point
}

// Line creates a line entity between two points
func (gf *GraphicsFactory) Line(x1, y1, z1, x2, y2, z2 float64) *entity.Line {
	line := entity.NewLine()
	line.Start = []float64{x1, y1, z1}
	line.End = []float64{x2, y2, z2}
	gf.drawing.AddEntity(line)
	return line
}

// Circle creates a circle entity with center and radius
func (gf *GraphicsFactory) Circle(cx, cy, cz, radius float64) *entity.Circle {
	circle := entity.NewCircle()
	circle.Center = []float64{cx, cy, cz}
	circle.Radius = radius
	gf.drawing.AddEntity(circle)
	return circle
}

// Arc creates an arc entity with center, radius, and angles
func (gf *GraphicsFactory) Arc(cx, cy, cz, radius, startAngle, endAngle float64) *entity.Arc {
	circle := entity.NewCircle()
	circle.Center = []float64{cx, cy, cz}
	circle.Radius = radius

	arc := entity.NewArc(circle)
	arc.Angle = []float64{startAngle, endAngle}
	gf.drawing.AddEntity(arc)
	return arc
}

// Text creates a text entity at specified position
func (gf *GraphicsFactory) Text(x, y, z, height float64, text string) *entity.Text {
	textEnt := entity.NewText()
	textEnt.Coord1 = []float64{x, y, z}
	textEnt.Coord2 = []float64{0, 0, 0}
	textEnt.Height = height
	textEnt.Value = text
	gf.drawing.AddEntity(textEnt)
	return textEnt
}

// MText creates a multiline text entity
func (gf *GraphicsFactory) MText(x, y, z, width, height float64, text string) *entity.MText {
	mtext := entity.NewMText()
	mtext.Coord1 = []float64{x, y, z}
	mtext.RectWidth = width
	mtext.Height = height
	mtext.Value = text
	gf.drawing.AddEntity(mtext)
	return mtext
}

// LwPolyline creates a lightweight polyline entity
func (gf *GraphicsFactory) LwPolyline(closed bool, vertices ...float64) *entity.LwPolyline {
	vertexCount := len(vertices) / 3
	lwPolyline := entity.NewLwPolyline(vertexCount)

	for i := 0; i < vertexCount && i*3+2 < len(vertices); i++ {
		if i < len(lwPolyline.Vertices) {
			lwPolyline.Vertices[i] = []float64{
				vertices[i*3],
				vertices[i*3+1],
				0,
			}
		}
	}

	if closed {
		lwPolyline.Close()
	}

	gf.drawing.AddEntity(lwPolyline)
	return lwPolyline
}

// Ellipse creates an ellipse entity
func (gf *GraphicsFactory) Ellipse(cx, cy, cz, majorAxis, minorAxis, rotation float64) *entity.Ellipse {
	ellipse := entity.NewEllipse()
	ellipse.Center = []float64{cx, cy, cz}
	ellipse.Major = majorAxis
	ellipse.Minor = minorAxis
	ellipse.Start = rotation
	ellipse.End = 360.0 + rotation
	gf.drawing.AddEntity(ellipse)
	return ellipse
}

// Spline creates a spline entity with control points
func (gf *GraphicsFactory) Spline(controlPoints ...float64) *entity.Spline {
	pointCount := len(controlPoints) / 3
	spline := entity.NewSpline(3)

	for i := 0; i < pointCount; i++ {
		x := controlPoints[i*3]
		y := controlPoints[i*3+1]
		z := controlPoints[i*3+2]
		spline.Controls = append(spline.Controls, []float64{x, y, z})
	}

	gf.drawing.AddEntity(spline)
	return spline
}

// Face3d creates a 3D face entity
func (gf *GraphicsFactory) Face3d(x1, y1, z1, x2, y2, z2, x3, y3, z3, x4, y4, z4 float64) *entity.ThreeDFace {
	face := entity.New3DFace()
	face.Points = [][]float64{
		{x1, y1, z1},
		{x2, y2, z2},
		{x3, y3, z3},
		{x4, y4, z4},
	}
	gf.drawing.AddEntity(face)
	return face
}

// Solid creates a solid entity
func (gf *GraphicsFactory) Solid(x1, y1, z1, x2, y2, z2, x3, y3, z3, x4, y4, z4 float64) *entity.Solid {
	solid := entity.NewSolid()
	solid.FirstPoint = []float64{x1, y1, z1}
	solid.SecondPoint = []float64{x2, y2, z2}
	solid.ThirdPoint = []float64{x3, y3, z3}
	if x4 != 0 || y4 != 0 || z4 != 0 {
		solid.FourthPoint = []float64{x4, y4, z4}
	} else {
		solid.FourthPoint = solid.FirstPoint
	}
	gf.drawing.AddEntity(solid)
	return solid
}

// Trace creates a trace entity (filled polygon)
func (gf *GraphicsFactory) Trace(x1, y1, z1, x2, y2, z2, x3, y3, z3, x4, y4, z4 float64) *entity.Trace {
	points := []float64{x1, y1, z1, x2, y2, z2, x3, y3, z3, x4, y4, z4}
	trace := entity.NewTrace(points, 0, 0)
	gf.drawing.AddEntity(trace)
	return trace
}

// Mesh creates a mesh entity
func (gf *GraphicsFactory) Mesh() *entity.Mesh {
	mesh := entity.NewMesh()
	gf.drawing.AddEntity(mesh)
	return mesh
}

// Helix creates a helix entity
func (gf *GraphicsFactory) Helix(cx, cy, cz, radius, height, turns, pitch float64) *entity.Helix {
	helix := entity.NewHelix()
	helix.SetStartPoint(dxfmath.NewVec3(cx, cy, cz))
	helix.SetRadius(radius)
	helix.SetTurns(turns)
	helix.SetTurnHeight(pitch)
	gf.drawing.AddEntity(helix)
	return helix
}

// Ray creates a ray entity
func (gf *GraphicsFactory) Ray(x, y, z, dirX, dirY, dirZ float64) *entity.Ray {
	ray := entity.NewRayFromPointDirection(
		dxfmath.NewVec3(x, y, z),
		dxfmath.NewVec3(dirX, dirY, dirZ),
	)
	gf.drawing.AddEntity(ray)
	return ray
}

// XLine creates a construction line entity
func (gf *GraphicsFactory) XLine(x1, y1, z1, x2, y2, z2 float64) *entity.XLine {
	xline := entity.NewXLineFromPointDirection(
		dxfmath.NewVec3(x1, y1, z1),
		dxfmath.NewVec3(x2-x1, y2-y1, z2-z1),
	)
	gf.drawing.AddEntity(xline)
	return xline
}

// Insert creates a block insert entity
func (gf *GraphicsFactory) Insert(blockName string, x, y, z float64) *entity.Insert {
	insert := entity.NewInsert()
	insert.SetInsertPoint(dxfmath.NewVec3(x, y, z))
	insert.SetBlockName(blockName)
	gf.drawing.AddEntity(insert)
	return insert
}

// Polyline2D creates a 2D polyline
func (gf *GraphicsFactory) Polyline2D(closed bool, vertices ...float64) *entity.Polyline {
	polyline := entity.NewPolyline()
	vertexCount := len(vertices) / 2
	for i := 0; i < vertexCount; i++ {
		x := vertices[i*2]
		y := vertices[i*2+1]
		polyline.AddVertex(x, y, 0)
	}
	if closed {
		polyline.Close()
	}
	gf.drawing.AddEntity(polyline)
	return polyline
}

// Polyline3D creates a 3D polyline
func (gf *GraphicsFactory) Polyline3D(closed bool, vertices ...float64) *entity.Polyline {
	polyline := entity.NewPolyline()
	vertexCount := len(vertices) / 3
	for i := 0; i < vertexCount; i++ {
		x := vertices[i*3]
		y := vertices[i*3+1]
		z := vertices[i*3+2]
		polyline.AddVertex(x, y, z)
	}
	if closed {
		polyline.Close()
	}
	gf.drawing.AddEntity(polyline)
	return polyline
}

// Rectangle creates a rectangle using a polyline
func (gf *GraphicsFactory) Rectangle(x, y, width, height float64) *entity.Polyline {
	x2 := x + width
	y2 := y + height
	polyline := entity.NewPolyline()
	polyline.AddVertex(x, y, 0)
	polyline.AddVertex(x2, y, 0)
	polyline.AddVertex(x2, y2, 0)
	polyline.AddVertex(x, y2, 0)
	polyline.Close()
	gf.drawing.AddEntity(polyline)
	return polyline
}

// Grid creates a grid of lines
func (gf *GraphicsFactory) Grid(x, y, width, height float64, rows, cols int) []*entity.Line {
	var lines []*entity.Line
	for i := 0; i <= rows; i++ {
		yPos := y + float64(i)*height/float64(rows)
		line := gf.Line(x, yPos, 0, x+width, yPos, 0)
		lines = append(lines, line)
	}
	for j := 0; j <= cols; j++ {
		xPos := x + float64(j)*width/float64(cols)
		line := gf.Line(xPos, y, 0, xPos, y+height, 0)
		lines = append(lines, line)
	}
	return lines
}

// Block creates a block record
func (gf *GraphicsFactory) Block(name string) *entity.BlockRecord {
	record := entity.NewBlockRecordWithName(name)
	gf.drawing.AddEntity(record)
	return record
}

// GetDrawing returns the associated drawing
func (gf *GraphicsFactory) GetDrawing() *Drawing {
	return gf.drawing
}

// Dimension creation methods

// LinearDimension creates a linear (rotated) dimension
func (gf *GraphicsFactory) LinearDimension(x1, y1, x2, y2, textX, textY float64) *entity.Dimension {
	dim := entity.NewDimension()
	dim.SetDimensionType(entity.DimensionTypeRotated)
	dim.SetDefinitionPoint([]float64{x1, y1, 0})
	dim.SetUserLocation(dxfmath.NewVec3(textX, textY, 0))
	gf.drawing.AddEntity(dim)
	return dim
}

// AlignedDimension creates an aligned linear dimension
func (gf *GraphicsFactory) AlignedDimension(x1, y1, x2, y2, textX, textY float64) *entity.Dimension {
	dim := entity.NewDimension()
	dim.SetDimensionType(entity.DimensionTypeAligned)
	dim.SetDefinitionPoint([]float64{x1, y1, 0})
	dim.SetUserLocation(dxfmath.NewVec3(textX, textY, 0))
	gf.drawing.AddEntity(dim)
	return dim
}

// AngularDimension creates an angular dimension
func (gf *GraphicsFactory) AngularDimension(cx, cy, startX, startY, endX, endY, textX, textY float64) *entity.Dimension {
	dim := entity.NewDimension()
	dim.SetDimensionType(entity.DimensionTypeAngular)
	dim.SetDefinitionPoint([]float64{cx, cy, 0})
	dim.SetUserLocation(dxfmath.NewVec3(textX, textY, 0))
	gf.drawing.AddEntity(dim)
	return dim
}

// Angular3PDimension creates an angular dimension from 3 points
func (gf *GraphicsFactory) Angular3PDimension(cx, cy, x1, y1, x2, y2, textX, textY float64) *entity.Dimension {
	dim := entity.NewDimension()
	dim.SetDimensionType(entity.DimensionTypeAngular3P)
	dim.SetDefinitionPoint([]float64{cx, cy, 0})
	dim.SetUserLocation(dxfmath.NewVec3(textX, textY, 0))
	gf.drawing.AddEntity(dim)
	return dim
}

// DiameterDimension creates a diameter dimension
func (gf *GraphicsFactory) DiameterDimension(cx, cy, radius float64) *entity.Dimension {
	dim := entity.NewDimension()
	dim.SetDimensionType(entity.DimensionTypeDiameter)
	dim.SetDefinitionPoint([]float64{cx, cy, 0})
	dim.SetUserLocation(dxfmath.NewVec3(cx+radius, cy, 0))
	gf.drawing.AddEntity(dim)
	return dim
}

// RadiusDimension creates a radius dimension
func (gf *GraphicsFactory) RadiusDimension(cx, cy, radius float64) *entity.Dimension {
	dim := entity.NewDimension()
	dim.SetDimensionType(entity.DimensionTypeRadius)
	dim.SetDefinitionPoint([]float64{cx, cy, 0})
	dim.SetUserLocation(dxfmath.NewVec3(cx+radius, cy, 0))
	gf.drawing.AddEntity(dim)
	return dim
}

// OrdinateDimension creates an ordinate dimension
func (gf *GraphicsFactory) OrdinateDimension(x, y, length float64, isXAxis bool) *entity.Dimension {
	dim := entity.NewDimension()
	dim.SetDimensionType(entity.DimensionTypeOrdinate)
	dim.SetDefinitionPoint([]float64{x, y, 0})
	if isXAxis {
		dim.SetUserLocation(dxfmath.NewVec3(x+length, y, 0))
	} else {
		dim.SetUserLocation(dxfmath.NewVec3(x, y+length, 0))
	}
	gf.drawing.AddEntity(dim)
	return dim
}

// ArcLengthDimension creates an arc length dimension
func (gf *GraphicsFactory) ArcLengthDimension(cx, cy, radius, startAngle, endAngle float64) *entity.Dimension {
	dim := entity.NewDimension()
	dim.SetDimensionType(entity.DimensionTypeArc)
	dim.SetDefinitionPoint([]float64{cx, cy, 0})
	gf.drawing.AddEntity(dim)
	return dim
}
