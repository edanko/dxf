// Typed Add<Type> factory methods on Space, mirroring ezdxf's
// graphicsfactory (msp.add_line, msp.add_text, ...). Each method builds the
// entity, routes it through Space.Add (handle allocation, layout
// assignment), and applies the attribute options.

package drawing

import (
	"fmt"

	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

// AddPoint adds a POINT at (x, y, z).
func (s *Space) AddPoint(x, y, z float64, opts ...EntityOption) (*entity.Point, error) {
	e := entity.NewPoint()
	e.Coord = []float64{x, y, z}
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddLine adds a LINE from (x1, y1, z1) to (x2, y2, z2).
func (s *Space) AddLine(x1, y1, z1, x2, y2, z2 float64, opts ...EntityOption) (*entity.Line, error) {
	e := entity.NewLine()
	e.Start = []float64{x1, y1, z1}
	e.End = []float64{x2, y2, z2}
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddCircle adds a CIRCLE with the given center and radius.
func (s *Space) AddCircle(cx, cy, cz, radius float64, opts ...EntityOption) (*entity.Circle, error) {
	e := entity.NewCircle()
	e.Center = []float64{cx, cy, cz}
	e.Radius = radius
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddArc adds an ARC with center, radius, and start/end angles in degrees.
func (s *Space) AddArc(cx, cy, cz, radius, startAngle, endAngle float64, opts ...EntityOption) (*entity.Arc, error) {
	base := entity.NewCircle()
	base.Center = []float64{cx, cy, cz}
	base.Radius = radius
	e := entity.NewArc(base)
	e.Angle = []float64{startAngle, endAngle}
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddText adds a TEXT entity at (x, y, z).
func (s *Space) AddText(x, y, z, height float64, text string, opts ...EntityOption) (*entity.Text, error) {
	e := entity.NewText()
	e.Coord1 = []float64{x, y, z}
	e.Coord2 = []float64{0, 0, 0}
	e.Height = height
	e.Value = text
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddMText adds an MTEXT entity at (x, y, z) with a fixed rectangle width.
func (s *Space) AddMText(x, y, z, width, height float64, text string, opts ...EntityOption) (*entity.MText, error) {
	e := entity.NewMText()
	e.Coord1 = []float64{x, y, z}
	e.RectWidth = width
	e.SetHeight(height)
	e.Value = text
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddEllipse adds an ELLIPSE.
func (s *Space) AddEllipse(cx, cy, cz, majorAxis, minorAxis, rotation float64, opts ...EntityOption) (*entity.Ellipse, error) {
	e := entity.NewEllipse()
	e.Center = []float64{cx, cy, cz}
	e.Major = majorAxis
	e.Minor = minorAxis
	e.Start = rotation
	e.End = 360.0 + rotation
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddSpline adds a SPLINE through control point triples.
func (s *Space) AddSpline(controlPoints []float64, opts ...EntityOption) (*entity.Spline, error) {
	pointCount := len(controlPoints) / 3
	e := entity.NewSpline(3)
	for i := 0; i < pointCount; i++ {
		e.Controls = append(e.Controls, []float64{controlPoints[i*3], controlPoints[i*3+1], controlPoints[i*3+2]})
	}
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddFace3D adds a 3DFACE from four point triples.
func (s *Space) AddFace3D(x1, y1, z1, x2, y2, z2, x3, y3, z3, x4, y4, z4 float64, opts ...EntityOption) (*entity.ThreeDFace, error) {
	e := entity.New3DFace()
	e.Points = [][]float64{
		{x1, y1, z1}, {x2, y2, z2}, {x3, y3, z3}, {x4, y4, z4},
	}
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddSolid adds a SOLID from four point triples.
func (s *Space) AddSolid(x1, y1, z1, x2, y2, z2, x3, y3, z3, x4, y4, z4 float64, opts ...EntityOption) (*entity.Solid, error) {
	e := entity.NewSolid()
	e.FirstPoint = []float64{x1, y1, z1}
	e.SecondPoint = []float64{x2, y2, z2}
	e.ThirdPoint = []float64{x3, y3, z3}
	if x4 != 0 || y4 != 0 || z4 != 0 {
		e.FourthPoint = []float64{x4, y4, z4}
	} else {
		e.FourthPoint = e.FirstPoint
	}
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddTrace adds a TRACE from four point triples.
func (s *Space) AddTrace(x1, y1, z1, x2, y2, z2, x3, y3, z3, x4, y4, z4 float64, opts ...EntityOption) (*entity.Trace, error) {
	e := entity.NewTrace([]float64{x1, y1, z1, x2, y2, z2, x3, y3, z3, x4, y4, z4}, 0, 0)
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddMesh adds an empty MESH entity.
func (s *Space) AddMesh(opts ...EntityOption) (*entity.Mesh, error) {
	e := entity.NewMesh()
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddHelix adds a HELIX.
func (s *Space) AddHelix(cx, cy, cz, radius, turns, pitch float64, opts ...EntityOption) (*entity.Helix, error) {
	e := entity.NewHelix()
	e.SetStartPoint(dxfmath.NewVec3(cx, cy, cz))
	e.SetRadius(radius)
	e.SetTurns(turns)
	e.SetTurnHeight(pitch)
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddRay adds a RAY from a base point along a direction.
func (s *Space) AddRay(x, y, z, dirX, dirY, dirZ float64, opts ...EntityOption) (*entity.Ray, error) {
	e := entity.NewRayFromPointDirection(dxfmath.NewVec3(x, y, z), dxfmath.NewVec3(dirX, dirY, dirZ))
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddXLine adds a construction XLINE from a base point along a direction.
func (s *Space) AddXLine(x, y, z, dirX, dirY, dirZ float64, opts ...EntityOption) (*entity.XLine, error) {
	e := entity.NewXLineFromPointDirection(dxfmath.NewVec3(x, y, z), dxfmath.NewVec3(dirX, dirY, dirZ))
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddInsert adds an INSERT of the named block at (x, y, z). Use
// s.WithAttribs to attach attribute values.
func (s *Space) AddInsert(blockName string, x, y, z float64, opts ...EntityOption) (*entity.Insert, error) {
	e := entity.NewInsert()
	e.SetInsertPoint(dxfmath.NewVec3(x, y, z))
	e.SetBlockName(blockName)
	if b, err := s.drawing.GetBlock(blockName); err == nil {
		e.SetBlockDef(b)
	}
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddPolyline2D adds a 2D POLYLINE from x, y pairs.
func (s *Space) AddPolyline2D(closed bool, vertices []float64, opts ...EntityOption) (*entity.Polyline, error) {
	e := entity.NewPolyline()
	for i := 0; i+1 < len(vertices); i += 2 {
		e.AddVertex(vertices[i], vertices[i+1], 0)
	}
	if closed {
		e.Close()
	}
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddPolyline3D adds a 3D POLYLINE from x, y, z triples.
func (s *Space) AddPolyline3D(closed bool, vertices []float64, opts ...EntityOption) (*entity.Polyline, error) {
	e := entity.NewPolyline()
	for i := 0; i+2 < len(vertices); i += 3 {
		e.AddVertex(vertices[i], vertices[i+1], vertices[i+2])
	}
	if closed {
		e.Close()
	}
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddRectangle adds a closed rectangular POLYLINE.
func (s *Space) AddRectangle(x, y, width, height float64, opts ...EntityOption) (*entity.Polyline, error) {
	e, err := s.AddPolyline2D(true, []float64{x, y, x + width, y, x + width, y + height, x, y + height}, opts...)
	return e, err
}

// AddLwPolyline adds an LWPOLYLINE from x, y pairs.
func (s *Space) AddLwPolyline(closed bool, vertices []float64, opts ...EntityOption) (*entity.LwPolyline, error) {
	count := len(vertices) / 2
	e := entity.NewLwPolyline(count)
	for i := 0; i < count && i < len(e.Vertices); i++ {
		e.Vertices[i] = []float64{vertices[i*2], vertices[i*2+1], 0}
	}
	if closed {
		e.Close()
	}
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddGrid adds rows x cols grid of LINES spanning (x, y) to (x+width,
// y+height) and returns them.
func (s *Space) AddGrid(x, y, width, height float64, rows, cols int, opts ...EntityOption) ([]*entity.Line, error) {
	if rows <= 0 || cols <= 0 {
		return nil, fmt.Errorf("grid: rows and cols must be positive, got %d, %d", rows, cols)
	}
	lines := make([]*entity.Line, 0, rows+cols+2)
	for i := 0; i <= rows; i++ {
		yPos := y + float64(i)*height/float64(rows)
		line, err := s.AddLine(x, yPos, 0, x+width, yPos, 0, opts...)
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	for j := 0; j <= cols; j++ {
		xPos := x + float64(j)*width/float64(cols)
		line, err := s.AddLine(xPos, y, 0, xPos, y+height, 0, opts...)
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	return lines, nil
}

// AddLinearDimension adds a rotated linear DIMENSION.
func (s *Space) AddLinearDimension(x1, y1, x2, y2, textX, textY float64, opts ...EntityOption) (*entity.Dimension, error) {
	return s.addDimension(entity.DimensionTypeRotated, []float64{x1, y1, 0}, dxfmath.NewVec3(textX, textY, 0), opts...)
}

// AddAlignedDimension adds an aligned linear DIMENSION.
func (s *Space) AddAlignedDimension(x1, y1, x2, y2, textX, textY float64, opts ...EntityOption) (*entity.Dimension, error) {
	return s.addDimension(entity.DimensionTypeAligned, []float64{x1, y1, 0}, dxfmath.NewVec3(textX, textY, 0), opts...)
}

// AddAngularDimension adds a 2-line angular DIMENSION.
func (s *Space) AddAngularDimension(cx, cy, textX, textY float64, opts ...EntityOption) (*entity.Dimension, error) {
	return s.addDimension(entity.DimensionTypeAngular, []float64{cx, cy, 0}, dxfmath.NewVec3(textX, textY, 0), opts...)
}

// AddAngular3PDimension adds a 3-point angular DIMENSION.
func (s *Space) AddAngular3PDimension(cx, cy, textX, textY float64, opts ...EntityOption) (*entity.Dimension, error) {
	return s.addDimension(entity.DimensionTypeAngular3P, []float64{cx, cy, 0}, dxfmath.NewVec3(textX, textY, 0), opts...)
}

// AddDiameterDimension adds a diameter DIMENSION.
func (s *Space) AddDiameterDimension(cx, cy, radius float64, opts ...EntityOption) (*entity.Dimension, error) {
	return s.addDimension(entity.DimensionTypeDiameter, []float64{cx, cy, 0}, dxfmath.NewVec3(cx+radius, cy, 0), opts...)
}

// AddRadiusDimension adds a radius DIMENSION.
func (s *Space) AddRadiusDimension(cx, cy, radius float64, opts ...EntityOption) (*entity.Dimension, error) {
	return s.addDimension(entity.DimensionTypeRadius, []float64{cx, cy, 0}, dxfmath.NewVec3(cx+radius, cy, 0), opts...)
}

// AddOrdinateDimension adds an ordinate DIMENSION at (x, y).
func (s *Space) AddOrdinateDimension(x, y, length float64, isXAxis bool, opts ...EntityOption) (*entity.Dimension, error) {
	loc := dxfmath.NewVec3(x, y+length, 0)
	if isXAxis {
		loc = dxfmath.NewVec3(x+length, y, 0)
	}
	return s.addDimension(entity.DimensionTypeOrdinate, []float64{x, y, 0}, loc, opts...)
}

// AddArcLengthDimension adds an arc-length DIMENSION.
func (s *Space) AddArcLengthDimension(cx, cy float64, opts ...EntityOption) (*entity.Dimension, error) {
	return s.addDimension(entity.DimensionTypeArc, []float64{cx, cy, 0}, dxfmath.NewVec3(cx, cy, 0), opts...)
}

func (s *Space) addDimension(t int, defPoint []float64, loc dxfmath.Vec3, opts ...EntityOption) (*entity.Dimension, error) {
	e := entity.NewDimension()
	e.SetDimensionType(t)
	e.SetDefinitionPoint(defPoint)
	e.SetUserLocation(loc)
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddAttdef adds an ATTDEF definition.
func (s *Space) AddAttdef(tag, prompt, defaultValue string, x, y, z, height float64, opts ...EntityOption) (*entity.Attdef, error) {
	e := entity.NewAttdefWithValues(tag, prompt, defaultValue, x, y, z)
	e.SetHeight(height)
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddAttrib adds a standalone ATTRIB.
func (s *Space) AddAttrib(tag, text string, x, y, z, height float64, opts ...EntityOption) (*entity.Attrib, error) {
	e := entity.NewAttribWithValues(tag, text, x, y, z)
	e.SetHeight(height)
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddLeader adds a LEADER through point triples.
func (s *Space) AddLeader(points [][]float64, opts ...EntityOption) (*entity.Leader, error) {
	e := entity.NewLeaderWithPoints(points)
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddMLine adds an MLINE with the given style, justification, and scale.
func (s *Space) AddMLine(styleName string, justification int, scale float64, opts ...EntityOption) (*entity.MLine, error) {
	e := entity.NewMLineWithStyle(styleName, justification, scale)
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// Add3DSolid adds a 3DSOLID entity (BREP/SAT payload filled by the caller).
func (s *Space) Add3DSolid(opts ...EntityOption) (*entity.Solid3d, error) {
	e := entity.NewSolid3d()
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddShape adds a SHAPE with a name, size, and insertion point.
func (s *Space) AddShape(name string, size float64, x, y, z float64, opts ...EntityOption) (*entity.Shape, error) {
	e := entity.NewShape()
	e.SetShapeName(name)
	e.SetSize(size)
	e.SetInsertPoint(dxfmath.NewVec3(x, y, z))
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddTolerance adds a GEOMETRIC_TOLERANCE (TOLERANCE) entity.
func (s *Space) AddTolerance(text string, x, y, z float64, opts ...EntityOption) (*entity.Tolerance, error) {
	e := entity.NewTolerance()
	e.SetContent(text)
	e.SetInsertPoint(dxfmath.NewVec3(x, y, z))
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddWipeout adds a WIPEOUT entity (boundary filled by the caller).
func (s *Space) AddWipeout(opts ...EntityOption) (*entity.Wipeout, error) {
	e := entity.NewWipeout()
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddImage adds an IMAGE (definition reactor filled by the caller).
func (s *Space) AddImage(opts ...EntityOption) (*entity.Image, error) {
	e := entity.NewImage()
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddUnderlay adds a PDF/DWF/DGN UNDERLAY from its definition entity.
func (s *Space) AddUnderlay(def *entity.UnderlayDefinition, opts ...EntityOption) (*entity.Underlay, error) {
	e := entity.NewPdfUnderlay(def)
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddHatch adds an enhanced HATCH entity (boundary paths filled by the
// caller).
func (s *Space) AddHatch(opts ...EntityOption) (*entity.HatchEnhanced, error) {
	e := entity.NewHatchEnhanced()
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}

// AddViewport adds a VPORT entity (frozen/off layers set by the caller).
func (s *Space) AddViewport(opts ...EntityOption) (*entity.Viewport, error) {
	e := entity.NewViewport()
	if err := s.Add(e, opts...); err != nil {
		return nil, err
	}
	return e, nil
}
