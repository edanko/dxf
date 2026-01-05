package entity

import (
	"math"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
)

// WIPEOUT represents a DXF WIPEOUT entity (polygonal masking area)
type Wipeout struct {
	*entity

	// Basic positioning
	insertPoint dxfmath.Vec3 // 10,20,30 - Insertion point in WCS

	// Wipeout vectors (similar to IMAGE but for masking)
	uVector dxfmath.Vec3 // 11,21,31 - U-pixel vector (X direction in WCS)
	vVector dxfmath.Vec3 // 12,22,32 - V-pixel vector (Y direction in WCS)

	// Image size (concepts borrowed from IMAGE for consistency)
	imageSize dxfmath.Vec2 // 13,23 - Size in pixels

	// Clipping boundary (this defines wipeout shape)
	clippingBoundaryType int                  // 71 - Boundary type (1=rectangular, 2=polygon)
	boundaryPoints       []ImageBoundaryPoint // 14 - Boundary path points in pixel coordinates

	// Display properties
	flags int // 70 - Display flags (similar to IMAGE but simplified)

	// Color
	color int // 62 - Color number
}

// NewWipeout creates a new WIPEOUT entity
func NewWipeout() *Wipeout {
	wipeout := &Wipeout{
		entity:               NewEntity(WIPEOUT),
		insertPoint:          dxfmath.NewVec3(0, 0, 0),
		uVector:              dxfmath.NewVec3(1, 0, 0),
		vVector:              dxfmath.NewVec3(0, 1, 0),
		imageSize:            dxfmath.NewVec2(1, 1),
		clippingBoundaryType: ImageClippingPolygon, // WIPEOUT always uses polygon
		boundaryPoints:       []ImageBoundaryPoint{},
		flags:                ImageFlagShowImage, // Simplified for wipeout
		color:                0,
	}
	return wipeout
}

// IsEntity returns true for WIPEOUT entities
func (w *Wipeout) IsEntity() bool {
	return true
}

// Format writes WIPEOUT entity data to DXF format
func (w *Wipeout) Format(f format.Formatter) {
	w.entity.Format(f)
	f.WriteString(100, "AcDbWipeout")

	// Write insertion point (10,20,30)
	f.WriteFloat(10, w.insertPoint.X())
	f.WriteFloat(20, w.insertPoint.Y())
	f.WriteFloat(30, w.insertPoint.Z())

	// Write U-vector (11,21,31)
	f.WriteFloat(11, w.uVector.X())
	f.WriteFloat(21, w.uVector.Y())
	f.WriteFloat(31, w.uVector.Z())

	// Write V-vector (12,22,32)
	f.WriteFloat(12, w.vVector.X())
	f.WriteFloat(22, w.vVector.Y())
	f.WriteFloat(32, w.vVector.Z())

	// Write image size (13,23)
	f.WriteFloat(13, w.imageSize.X())
	f.WriteFloat(23, w.imageSize.Y())

	// Write flags (70)
	f.WriteInt(70, w.flags)

	// Write clipping boundary type (71)
	f.WriteInt(71, w.clippingBoundaryType)

	// Write boundary points (14)
	if len(w.boundaryPoints) > 0 {
		f.WriteInt(91, len(w.boundaryPoints))
		for _, point := range w.boundaryPoints {
			f.WriteFloat(14, point.X)
			f.WriteFloat(24, point.Y)
		}
	}

	// Write color (62)
	if w.color != 0 {
		f.WriteInt(62, w.color)
	}
}

// SetInsertPoint sets wipeout insertion point
func (w *Wipeout) SetInsertPoint(point dxfmath.Vec3) {
	w.insertPoint = point
}

// SetUVectors sets U and V vectors for wipeout orientation
func (w *Wipeout) SetUVectors(uVector, vVector dxfmath.Vec3) {
	w.uVector = uVector
	w.vVector = vVector
}

// SetSize sets wipeout size
func (w *Wipeout) SetSize(width, height float64) {
	w.imageSize = dxfmath.NewVec2(width, height)
}

// SetRectangularWipeout creates a rectangular wipeout area
func (w *Wipeout) SetRectangularWipeout(xMin, yMin, xMax, yMax float64) {
	w.clippingBoundaryType = ImageClippingRectangular
	w.boundaryPoints = []ImageBoundaryPoint{
		{X: xMin, Y: yMin},
		{X: xMax, Y: yMax},
	}
}

// SetPolygonWipeout creates a polygonal wipeout area
func (w *Wipeout) SetPolygonWipeout(points []ImageBoundaryPoint) {
	if len(points) < 3 {
		return // Need at least 3 points for polygon
	}
	w.clippingBoundaryType = ImageClippingPolygon
	w.boundaryPoints = points

	// Close polygon if not already closed
	if len(points) > 0 {
		first := points[0]
		last := points[len(points)-1]
		if first.X != last.X || first.Y != last.Y {
			w.boundaryPoints = append(w.boundaryPoints, first)
		}
	}
}

// SetColor sets color number
func (w *Wipeout) SetColor(color color.ColorNumber) {
	w.color = int(color)
}

// Getters
func (w *Wipeout) InsertPoint() dxfmath.Vec3            { return w.insertPoint }
func (w *Wipeout) UVector() dxfmath.Vec3                { return w.uVector }
func (w *Wipeout) VVector() dxfmath.Vec3                { return w.vVector }
func (w *Wipeout) ImageSize() dxfmath.Vec2              { return w.imageSize }
func (w *Wipeout) ClippingBoundaryType() int            { return w.clippingBoundaryType }
func (w *Wipeout) BoundaryPoints() []ImageBoundaryPoint { return w.boundaryPoints }
func (w *Wipeout) Flags() int                           { return w.flags }
func (w *Wipeout) Color() int                           { return w.color }

// IsRectangular returns true if using rectangular boundary
func (w *Wipeout) IsRectangular() bool {
	return w.clippingBoundaryType == ImageClippingRectangular
}

// IsPolygon returns true if using polygon boundary
func (w *Wipeout) IsPolygon() bool {
	return w.clippingBoundaryType == ImageClippingPolygon
}

// BBox calculates bounding box of wipeout
func (w *Wipeout) BBox() ([]float64, []float64) {
	if len(w.boundaryPoints) == 0 {
		return []float64{}, []float64{}
	}

	// Transform boundary points to WCS
	minX, minY := w.insertPoint.X(), w.insertPoint.Y()
	maxX, maxY := w.insertPoint.X(), w.insertPoint.Y()

	for _, point := range w.boundaryPoints {
		// Convert pixel coordinates to WCS using U and V vectors
		worldX := w.insertPoint.X() + point.X*w.uVector.X() + point.Y*w.vVector.X()
		worldY := w.insertPoint.Y() + point.X*w.uVector.Y() + point.Y*w.vVector.Y()

		if worldX < minX {
			minX = worldX
		}
		if worldX > maxX {
			maxX = worldX
		}
		if worldY < minY {
			minY = worldY
		}
		if worldY > maxY {
			maxY = worldY
		}
	}

	return []float64{minX, minY, w.insertPoint.Z()}, []float64{maxX, maxY, w.insertPoint.Z()}
}

// CreateRectangularWipeout creates a simple rectangular wipeout
func CreateRectangularWipeout(insertPoint dxfmath.Vec3, width, height float64) *Wipeout {
	wipeout := NewWipeout()
	wipeout.SetInsertPoint(insertPoint)
	wipeout.SetSize(width, height)

	// Set U and V vectors for standard orientation
	wipeout.SetUVectors(
		dxfmath.NewVec3(1, 0, 0), // U vector: width direction
		dxfmath.NewVec3(0, 1, 0), // V vector: height direction
	)

	// Set rectangular boundary from (0,0) to (width,height)
	wipeout.SetRectangularWipeout(0, 0, width, height)
	return wipeout
}

// CreatePolygonalWipeout creates a polygonal wipeout from world coordinates
func CreatePolygonalWipeout(insertPoint dxfmath.Vec3, scale float64, points []dxfmath.Vec2) *Wipeout {
	wipeout := NewWipeout()
	wipeout.SetInsertPoint(insertPoint)

	// Calculate bounding box for size
	minX, maxX := points[0].X(), points[0].X()
	minY, maxY := points[0].Y(), points[0].Y()
	for _, point := range points[1:] {
		if point.X() < minX {
			minX = point.X()
		}
		if point.X() > maxX {
			maxX = point.X()
		}
		if point.Y() < minY {
			minY = point.Y()
		}
		if point.Y() > maxY {
			maxY = point.Y()
		}
	}

	width := maxX - minX
	height := maxY - minY
	wipeout.SetSize(width, height)

	// Set U and V vectors with scale
	wipeout.SetUVectors(
		dxfmath.NewVec3(scale, 0, 0), // U vector: scaled width
		dxfmath.NewVec3(0, scale, 0), // V vector: scaled height
	)

	// Convert world coordinates to pixel coordinates for boundary
	boundaryPoints := make([]ImageBoundaryPoint, len(points))
	for i, point := range points {
		boundaryPoints[i] = ImageBoundaryPoint{
			X: point.X(),
			Y: point.Y(),
		}
	}

	wipeout.SetPolygonWipeout(boundaryPoints)
	return wipeout
}

// CreateCircleWipeout creates a circular wipeout using polygon approximation
func CreateCircleWipeout(insertPoint dxfmath.Vec3, radius, scale float64, segments int) *Wipeout {
	wipeout := NewWipeout()
	wipeout.SetInsertPoint(insertPoint)
	wipeout.SetSize(radius*2, radius*2)

	// Set U and V vectors with scale
	wipeout.SetUVectors(
		dxfmath.NewVec3(scale, 0, 0), // U vector: scaled
		dxfmath.NewVec3(0, scale, 0), // V vector: scaled
	)

	// Generate circle points
	boundaryPoints := make([]ImageBoundaryPoint, segments)
	for i := 0; i < segments; i++ {
		angle := 2.0 * math.Pi * float64(i) / float64(segments)
		x := radius * math.Cos(angle)
		y := radius * math.Sin(angle)
		boundaryPoints[i] = ImageBoundaryPoint{X: x, Y: y}
	}

	wipeout.SetPolygonWipeout(boundaryPoints)
	return wipeout
}
