package entity

import (
	"math"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	dxfmath "github.com/edanko/dxf/math"
)

// IMAGE entity flags
const (
	ImageFlagShowImage           = 1 // Show image
	ImageFlagShowWhenNotAligned  = 2 // Show image when not aligned with UCS
	ImageFlagUseClippingBoundary = 4 // Use clipping boundary
	ImageFlagUseTransparency     = 8 // Use transparency
)

// IMAGE clipping boundary types
const (
	ImageClippingRectangular = 1 // Rectangular clipping
	ImageClippingPolygon     = 2 // Polygon clipping
)

// IMAGE clip modes (DXF R2010+)
const (
	ImageClipModeOutside = 0 // Clip outside boundary
	ImageClipModeInside  = 1 // Clip inside boundary
)

// ImageBoundaryPoint represents a boundary point in image pixel coordinates
type ImageBoundaryPoint struct {
	X float64 // 14 - X coordinate in pixels
	Y float64 // 14 - Y coordinate in pixels
}

// IMAGE represents a DXF IMAGE entity
type Image struct {
	*entity

	// Basic positioning
	insertPoint dxfmath.Vec3 // 10,20,30 - Insertion point in WCS

	// Image vectors (pixel size and orientation)
	uVector dxfmath.Vec3 // 11,21,31 - U-pixel vector (X direction in WCS, represents image width)
	vVector dxfmath.Vec3 // 12,22,32 - V-pixel vector (Y direction in WCS, represents image height)

	// Image size and references
	imageSize      dxfmath.Vec2 // 13,23 - Image size in pixels (width, height)
	imageDefHandle string       // 340 - Handle to ImageDef object

	// Display properties
	flags      int     // 70 - Display flags
	clipping   int     // 280 - Clipping state (0=off, 1=on)
	brightness float64 // 281 - Brightness (-100 to 100, default=0)
	contrast   float64 // 282 - Contrast (-100 to 100, default=0)
	fade       float64 // 283 // Fade (0 to 100, default=0)

	// Clipping boundary
	clippingBoundaryType int                  // 71 - Clipping boundary type (1=rectangular, 2=polygon)
	boundaryPoints       []ImageBoundaryPoint // 14 - Boundary path points in pixel coordinates
	clipMode             int                  // 290 - Clip mode (DXF2010+, 0=outside, 1=inside)

	// Reactor handle
	imageDefReactorHandle string // 360 - Handle to ImageDefReactor object

	// Color
	color int // 62 - Color number
}

// NewImage creates a new IMAGE entity
func NewImage() *Image {
	img := &Image{
		entity:                NewEntity(IMAGE),
		insertPoint:           dxfmath.NewVec3(0, 0, 0),
		uVector:               dxfmath.NewVec3(1, 0, 0),
		vVector:               dxfmath.NewVec3(0, 1, 0),
		imageSize:             dxfmath.NewVec2(1, 1),
		imageDefHandle:        "",
		flags:                 ImageFlagShowImage | ImageFlagShowWhenNotAligned,
		clipping:              0,
		brightness:            0.0,
		contrast:              0.0,
		fade:                  0.0,
		clippingBoundaryType:  ImageClippingRectangular,
		boundaryPoints:        []ImageBoundaryPoint{},
		clipMode:              ImageClipModeOutside,
		imageDefReactorHandle: "",
		color:                 0,
	}
	return img
}

// IsEntity returns true for IMAGE entities
func (i *Image) IsEntity() bool {
	return true
}

// Format writes IMAGE entity data to DXF format
func (i *Image) Format(f format.Formatter) {
	i.entity.Format(f)
	f.WriteString(100, "AcDbImage")

	// Write insertion point (10,20,30)
	f.WriteFloat(10, i.insertPoint.X())
	f.WriteFloat(20, i.insertPoint.Y())
	f.WriteFloat(30, i.insertPoint.Z())

	// Write U-vector (11,21,31) - represents image width direction and scale
	f.WriteFloat(11, i.uVector.X())
	f.WriteFloat(21, i.uVector.Y())
	f.WriteFloat(31, i.uVector.Z())

	// Write V-vector (12,22,32) - represents image height direction and scale
	f.WriteFloat(12, i.vVector.X())
	f.WriteFloat(22, i.vVector.Y())
	f.WriteFloat(32, i.vVector.Z())

	// Write image size (13,23)
	f.WriteFloat(13, i.imageSize.X())
	f.WriteFloat(23, i.imageSize.Y())

	// Write ImageDef handle (340)
	if i.imageDefHandle != "" {
		f.WriteString(340, i.imageDefHandle)
	}

	// Write flags (70)
	f.WriteInt(70, i.flags)

	// Write clipping state (280)
	f.WriteInt(280, i.clipping)

	// Write brightness (281)
	if i.brightness != 0.0 {
		f.WriteFloat(281, i.brightness)
	}

	// Write contrast (282)
	if i.contrast != 0.0 {
		f.WriteFloat(282, i.contrast)
	}

	// Write fade (283)
	if i.fade != 0.0 {
		f.WriteFloat(283, i.fade)
	}

	// Write ImageDefReactor handle (360)
	if i.imageDefReactorHandle != "" {
		f.WriteString(360, i.imageDefReactorHandle)
	}

	// Write clipping boundary type (71)
	f.WriteInt(71, i.clippingBoundaryType)

	// Write boundary points (14)
	if len(i.boundaryPoints) > 0 {
		f.WriteInt(91, len(i.boundaryPoints))
		for _, point := range i.boundaryPoints {
			f.WriteFloat(14, point.X)
			f.WriteFloat(24, point.Y)
		}
	}

	// Write clip mode (290) - DXF2010+
	if i.clipMode != ImageClipModeOutside {
		f.WriteInt(290, i.clipMode)
	}

	// Write color (62)
	if i.color != 0 {
		f.WriteInt(62, i.color)
	}
}

// SetInsertPoint sets image insertion point
func (i *Image) SetInsertPoint(point dxfmath.Vec3) {
	i.insertPoint = point
}

// SetUVectors sets U and V vectors for image orientation and scale
func (i *Image) SetUVectors(uVector, vVector dxfmath.Vec3) {
	i.uVector = uVector
	i.vVector = vVector
}

// SetImageSize sets image size in pixels
func (i *Image) SetImageSize(width, height float64) {
	i.imageSize = dxfmath.NewVec2(width, height)
}

// SetImageDefHandle sets ImageDef object handle
func (i *Image) SetImageDefHandle(handle string) {
	i.imageDefHandle = handle
}

// SetDisplayProperties sets brightness, contrast, and fade
func (i *Image) SetDisplayProperties(brightness, contrast, fade float64) {
	i.brightness = brightness
	i.contrast = contrast
	i.fade = fade
}

// SetClipping enables or disables clipping
func (i *Image) SetClipping(enabled bool) {
	if enabled {
		i.clipping = 1
		i.flags |= ImageFlagUseClippingBoundary
	} else {
		i.clipping = 0
		i.flags &^= ImageFlagUseClippingBoundary
	}
}

// SetRectangularClipping sets rectangular clipping boundary
func (i *Image) SetRectangularClipping(xMin, yMin, xMax, yMax float64) {
	i.clippingBoundaryType = ImageClippingRectangular
	i.boundaryPoints = []ImageBoundaryPoint{
		{X: xMin, Y: yMin},
		{X: xMax, Y: yMax},
	}
	i.SetClipping(true)
}

// SetPolygonClipping sets polygon clipping boundary
func (i *Image) SetPolygonClipping(points []ImageBoundaryPoint) {
	if len(points) < 3 {
		return // Need at least 3 points for polygon
	}
	i.clippingBoundaryType = ImageClippingPolygon
	i.boundaryPoints = points
	i.SetClipping(true)
}

// ResetToDefaultBoundary resets clipping to show full image
func (i *Image) ResetToDefaultBoundary() {
	i.clippingBoundaryType = ImageClippingRectangular
	// Default boundary: (-0.5, -0.5) to (width-0.5, height-0.5)
	i.boundaryPoints = []ImageBoundaryPoint{
		{X: -0.5, Y: -0.5},
		{X: i.imageSize.X() - 0.5, Y: i.imageSize.Y() - 0.5},
	}
	i.SetClipping(false)
}

// SetColor sets color number
func (i *Image) SetColor(color color.ColorNumber) {
	i.color = int(color)
}

// Getters
func (i *Image) InsertPoint() dxfmath.Vec3            { return i.insertPoint }
func (i *Image) UVector() dxfmath.Vec3                { return i.uVector }
func (i *Image) VVector() dxfmath.Vec3                { return i.vVector }
func (i *Image) ImageSize() dxfmath.Vec2              { return i.imageSize }
func (i *Image) ImageDefHandle() string               { return i.imageDefHandle }
func (i *Image) Flags() int                           { return i.flags }
func (i *Image) Clipping() int                        { return i.clipping }
func (i *Image) Brightness() float64                  { return i.brightness }
func (i *Image) Contrast() float64                    { return i.contrast }
func (i *Image) Fade() float64                        { return i.fade }
func (i *Image) ClippingBoundaryType() int            { return i.clippingBoundaryType }
func (i *Image) BoundaryPoints() []ImageBoundaryPoint { return i.boundaryPoints }
func (i *Image) ClipMode() int                        { return i.clipMode }
func (i *Image) ImageDefReactorHandle() string        { return i.imageDefReactorHandle }
func (i *Image) Color() int                           { return i.color }

// IsUsingClipping returns true if clipping is enabled
func (i *Image) IsUsingClipping() bool {
	return i.clipping == 1
}

// IsRectangularClipping returns true if using rectangular clipping
func (i *Image) IsRectangularClipping() bool {
	return i.clippingBoundaryType == ImageClippingRectangular
}

// IsPolygonClipping returns true if using polygon clipping
func (i *Image) IsPolygonClipping() bool {
	return i.clippingBoundaryType == ImageClippingPolygon
}

// BBox calculates bounding box of image
func (i *Image) BBox() ([]float64, []float64) {
	if len(i.boundaryPoints) == 0 {
		return []float64{}, []float64{}
	}

	// Transform boundary points to WCS
	minX, minY := i.insertPoint.X(), i.insertPoint.Y()
	maxX, maxY := i.insertPoint.X(), i.insertPoint.Y()

	for _, point := range i.boundaryPoints {
		// Convert pixel coordinates to WCS using U and V vectors
		worldX := i.insertPoint.X() + point.X*i.uVector.X() + point.Y*i.vVector.X()
		worldY := i.insertPoint.Y() + point.X*i.uVector.Y() + point.Y*i.vVector.Y()
		_ = i.insertPoint.Z() + point.X*i.uVector.Z() + point.Y*i.vVector.Z() // Calculate Z for completeness

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

	return []float64{minX, minY, i.insertPoint.Z()}, []float64{maxX, maxY, i.insertPoint.Z()}
}

// CreateSimpleImage creates a basic image with default rectangular boundary
func CreateSimpleImage(insertPoint dxfmath.Vec3, width, height float64, imageDefHandle string) *Image {
	image := NewImage()
	image.SetInsertPoint(insertPoint)
	image.SetImageSize(width, height)
	image.SetImageDefHandle(imageDefHandle)

	// Set U and V vectors for standard orientation (1 unit per pixel)
	image.SetUVectors(
		dxfmath.NewVec3(1, 0, 0), // U vector: image width direction
		dxfmath.NewVec3(0, 1, 0), // V vector: image height direction
	)

	image.ResetToDefaultBoundary()
	return image
}

// CreateScaledImage creates an image with custom scale (world units per pixel)
func CreateScaledImage(insertPoint dxfmath.Vec3, widthPixels, heightPixels, scale float64, imageDefHandle string) *Image {
	image := NewImage()
	image.SetInsertPoint(insertPoint)
	image.SetImageSize(widthPixels, heightPixels)
	image.SetImageDefHandle(imageDefHandle)

	// Set U and V vectors with custom scale
	image.SetUVectors(
		dxfmath.NewVec3(scale, 0, 0), // U vector: scaled width
		dxfmath.NewVec3(0, scale, 0), // V vector: scaled height
	)

	image.ResetToDefaultBoundary()
	return image
}

// CreateRotatedImage creates an image with rotation angle
func CreateRotatedImage(insertPoint dxfmath.Vec3, widthPixels, heightPixels, scale, angle float64, imageDefHandle string) *Image {
	image := NewImage()
	image.SetInsertPoint(insertPoint)
	image.SetImageSize(widthPixels, heightPixels)
	image.SetImageDefHandle(imageDefHandle)

	// Calculate rotated U and V vectors
	cosAngle := math.Cos(angle)
	sinAngle := math.Sin(angle)

	image.SetUVectors(
		dxfmath.NewVec3(scale*cosAngle, scale*sinAngle, 0),  // Rotated U vector
		dxfmath.NewVec3(-scale*sinAngle, scale*cosAngle, 0), // Rotated V vector (perpendicular)
	)

	image.ResetToDefaultBoundary()
	return image
}
