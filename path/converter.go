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
	default:
		return nil, fmt.Errorf("unsupported entity type for path conversion: %T", ent)
	}
}

// lineToPath converts a line entity to a path
func lineToPath(line *entity.Line, options *Options) (*Path, error) {
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

	// Convert first vertex
	path.MoveTo(vertices[0][0], vertices[0][1])
	for i := 1; i < len(vertices); i++ {
		path.LineTo(vertices[i][0], vertices[i][1])
	}

	// Close if specified
	if lwpolyline.Closed {
		path.Close()
	}

	return path, nil
}

// polylineToPath converts a 3D polyline to a path
func polylineToPath(polyline *entity.Polyline, options *Options) (*Path, error) {
	path := NewPath()
	vertices := polyline.Vertices

	if len(vertices) == 0 {
		return path, nil
	}

	path.MoveTo(vertices[0].Coord[0], vertices[0].Coord[1])
	for i := 1; i < len(vertices); i++ {
		path.LineTo(vertices[i].Coord[0], vertices[i].Coord[1])
	}

	// Close if specified
	if polyline.Flag&1 != 0 {
		path.Close()
	}

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

// ellipseToPath converts an ellipse entity to a path
func ellipseToPath(ellipse *entity.Ellipse, options *Options) (*Path, error) {
	path := NewPath()
	center := dxfmath.NewVec3(
		getCoord(ellipse.Center, 0),
		getCoord(ellipse.Center, 1),
		getCoord(ellipse.Center, 2))
	majorRadius := ellipse.Major
	minorRadius := ellipse.Minor

	// Approximate ellipse with segments
	segments := options.Segments
	if segments <= 0 {
		segments = 32 // Default
	}

	startX := center.X() + majorRadius
	startY := center.Y()
	path.MoveTo(startX, startY)

	for i := 1; i <= segments; i++ {
		angle := 2.0 * math.Pi * float64(i) / float64(segments)
		x := center.X() + majorRadius*math.Cos(angle)
		y := center.Y() + minorRadius*math.Sin(angle)
		path.LineTo(x, y)
	}

	path.Close()
	return path, nil
}

// vertexToVec3 converts a vertex entity to a Vec3
func vertexToVec3(vertex *entity.Vertex) dxfmath.Vec3 {
	coords := vertex.Coord
	if len(coords) >= 3 {
		return dxfmath.NewVec3(coords[0], coords[1], coords[2])
	}
	return dxfmath.NewVec3(0, 0, 0)
}

// getCoord extracts coordinate from slice with bounds checking
func getCoord(coords []float64, index int) float64 {
	if index >= 0 && index < len(coords) {
		return coords[index]
	}
	return 0.0
}
