package path

import (
	"fmt"
	"math"

	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

// Conversion options
type Options struct {
	Segments  int          // Number of segments for approximation
	Distance  float64      // Maximum deviation for flattening
	Extrusion dxfmath.Vec3 // Extrusion vector for 3D
	Precision float64      // Approximation precision
}

type Option func(*Options)

func WithSegments(segments int) Option {
	return func(o *Options) { o.Segments = segments }
}

func WithDistance(distance float64) Option {
	return func(o *Options) { o.Distance = distance }
}

func WithExtrusion(extrusion dxfmath.Vec3) Option {
	return func(o *Options) { o.Extrusion = extrusion }
}

func WithPrecision(precision float64) Option {
	return func(o *Options) { o.Precision = precision }
}

// defaultOptions creates default conversion options
func defaultOptions() *Options {
	return &Options{
		Segments:  16,
		Distance:  0.1,
		Extrusion: dxfmath.NewVec3(0, 0, 0),
		Precision: 1e-6,
	}
}

// MakePath converts an entity to a path representation
func MakePath(ent entity.Entity, opts ...Option) (*Path, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	switch e := ent.(type) {
	case *entity.Line:
		return lineToPath(e, options)
	case *entity.LwPolyline:
		return lwpolylineToPath(e, options)
	case *entity.Polyline:
		return polylineToPath(e, options)
	case *entity.Circle:
		return circleToPath(e, options)
	case *entity.Arc:
		return arcToPath(e, options)
	case *entity.Ellipse:
		return ellipseToPath(e, options)
	case *entity.Spline:
		return splineToPath(e, options)
	case *entity.Hatch:
		return hatchToPath(e, options)
	case *entity.MPolygonEntity:
		return mpolygonToPath(e, options)
	case *entity.Solid:
		return solidToPath(e, options)
	case *entity.Trace:
		return traceToPath(e, options)
	case *entity.ThreeDFace:
		return face3DToPath(e, options)
	case *entity.Image:
		return imageToPath(e, options)
	case *entity.Wipeout:
		return wipeoutToPath(e, options)
	case *entity.Leader:
		return leaderToPath(e, options)
	case *entity.MLeader:
		return mleaderToPath(e, options)
	case *entity.MLine:
		return mlineToPath(e, options)
	case *entity.Point:
		return pointToPath(e, options)
	case *entity.Dimension:
		return dimensionToPath(e, options)
	case infiniteLine:
		return infiniteLineToPath(e, options)
	case *entity.Text, *entity.MText:
		return nil, fmt.Errorf("%T is rendered through the text2path glyph pipeline, not MakePath", ent)
	default:
		return nil, fmt.Errorf("unsupported entity type for path conversion: %T", ent)
	}
}

// lineToPath converts a line entity to a path
func lineToPath(line *entity.Line, _ *Options) (*Path, error) {
	path := NewPath()
	start := dxfmath.NewVec3(
		getCoord(line.Start, 0),
		getCoord(line.Start, 1),
		getCoord(line.Start, 2))
	end := dxfmath.NewVec3(
		getCoord(line.End, 0),
		getCoord(line.End, 1),
		getCoord(line.End, 2))

	path.MoveTo(start.X(), start.Y())
	path.LineTo(end.X(), end.Y())

	return path, nil
}

// lwpolylineToPath converts a lightweight polyline to a path
func lwpolylineToPath(lwpolyline *entity.LwPolyline, options *Options) (*Path, error) {
	path := NewPath()
	vertices := lwpolyline.Vertices

	if len(vertices) == 0 {
		return path, nil
	}

	coords := make([][2]float64, 0, len(vertices))
	for _, v := range vertices {
		coords = append(coords, [2]float64{getCoord(v, 0), getCoord(v, 1)})
	}
	appendBulgeChain(path, coords, lwpolyline.Bulges, lwpolyline.Closed, options)
	return path, nil
}

// polylineToPath converts a 3D polyline to a path
func polylineToPath(polyline *entity.Polyline, options *Options) (*Path, error) {
	path := NewPath()
	vertices := polyline.Vertices

	if len(vertices) == 0 {
		return path, nil
	}

	coords := make([][2]float64, 0, len(vertices))
	bulges := make([]float64, 0, len(vertices))
	for _, v := range vertices {
		coords = append(coords, [2]float64{getCoord(v.Coord, 0), getCoord(v.Coord, 1)})
		bulges = append(bulges, v.Bulge)
	}
	appendBulgeChain(path, coords, bulges, polyline.Flag&1 != 0, options)

	return path, nil
}

// circleToPath converts a circle entity to a path
func circleToPath(circle *entity.Circle, options *Options) (*Path, error) {
	path := NewPath()
	center := dxfmath.NewVec3(
		getCoord(circle.Center, 0),
		getCoord(circle.Center, 1),
		getCoord(circle.Center, 2))
	radius := circle.Radius

	// Approximate circle with segments
	segments := options.Segments
	if segments <= 0 {
		segments = 32 // Default
	}

	// Start point
	startX := center.X() + radius
	startY := center.Y()
	path.MoveTo(startX, startY)

	// Generate circle points
	for i := 1; i <= segments; i++ {
		angle := 2 * math.Pi * float64(i) / float64(segments)
		x := center.X() + radius*math.Cos(angle)
		y := center.Y() + radius*math.Sin(angle)
		path.LineTo(x, y)
	}

	path.Close()
	return path, nil
}

// arcToPath converts an arc entity to a path
func arcToPath(arc *entity.Arc, options *Options) (*Path, error) {
	path := NewPath()
	center := dxfmath.NewVec3(
		getCoord(arc.Center, 0),
		getCoord(arc.Center, 1),
		getCoord(arc.Center, 2))
	radius := arc.Radius

	// Convert angles to radians
	startAngle := arc.Angle[0] * math.Pi / 180.0
	endAngle := arc.Angle[1] * math.Pi / 180.0

	// Approximate arc with segments
	segments := options.Segments
	if segments <= 0 {
		segments = 16 // Default
	}

	startX := center.X() + radius*math.Cos(startAngle)
	startY := center.Y() + radius*math.Sin(startAngle)
	path.MoveTo(startX, startY)

	// Approximate arc
	for i := 1; i < segments; i++ {
		t := float64(i) / float64(segments)
		angle := startAngle + t*(endAngle-startAngle)
		x := center.X() + radius*math.Cos(angle)
		y := center.Y() + radius*math.Sin(angle)
		path.LineTo(x, y)
	}

	return path, nil
}

// ellipseToPath converts an ellipse entity to a path. Start/End are parametric
// angles in degrees; a full ellipse closes the path, an elliptical arc does not.
func ellipseToPath(ellipse *entity.Ellipse, options *Options) (*Path, error) {
	if ellipse.Spline != nil {
		return splineToPath(ellipse.Spline, options)
	}
	path := NewPath()
	cx := getCoord(ellipse.Center, 0)
	cy := getCoord(ellipse.Center, 1)
	majorRadius := ellipse.Major
	minorRadius := ellipse.Minor

	startDeg := ellipse.Start
	endDeg := ellipse.End
	if endDeg < startDeg {
		endDeg += 360
	}
	span := endDeg - startDeg

	segments := min(max(int(math.Ceil(span/7.5)), 16), 256)
	if options.Segments > 0 {
		segments = options.Segments
	}

	startAngle := startDeg * math.Pi / 180.0
	path.MoveTo(cx+majorRadius*math.Cos(startAngle), cy+minorRadius*math.Sin(startAngle))
	for i := 1; i <= segments; i++ {
		angle := (startDeg + span*float64(i)/float64(segments)) * math.Pi / 180.0
		path.LineTo(cx+majorRadius*math.Cos(angle), cy+minorRadius*math.Sin(angle))
	}
	if ellipse.IsFull() {
		path.Close()
	}
	return path, nil
}

// getCoord extracts coordinate from slice with bounds checking
func getCoord(coords []float64, index int) float64 {
	if index >= 0 && index < len(coords) {
		return coords[index]
	}
	return 0.0
}
