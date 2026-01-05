package entity

import (
	"fmt"
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/math"
)

// Leader arrowhead types
const (
	LeaderArrowheadNone        = 0
	LeaderArrowheadClosed      = 1
	LeaderArrowheadDot         = 2
	LeaderArrowheadArch        = 3
	LeaderArrowheadTick        = 4
	LeaderArrowheadOpen        = 5
	LeaderArrowheadOpen90      = 6
	LeaderArrowheadOpen180     = 7
	LeaderArrowheadOrigin      = 8
	LeaderArrowheadDotSmall    = 9
	LeaderArrowheadDotBlank    = 10
	LeaderArrowheadSmall       = 11
	LeaderArrowheadBox         = 12
	LeaderArrowheadBoxBlank    = 13
	LeaderArrowheadClosedBlank = 14
	LeaderArrowheadTriangle    = 15
	LeaderArrowheadBoxFilled   = 16
	LeaderArrowheadDiamond     = 17
	LeaderArrowheadOblique     = 18
)

// Leader represents a LEADER entity (enhanced)
type Leader struct {
	*entity

	// Basic geometry
	points              [][]float64 // 10,20,30 - Leader line points (multiple sets)
	pathType            int         // 72 - Path type (0=straight, 1=spline)
	extrusion           []float64   // 210,220,230 - Extrusion direction
	normalVector        []float64   // 210,220,230 - Normal vector (alias for extrusion)
	horizontalDirection math.Vec3   // 211 - Horizontal direction vector

	// Annotation data
	annotationType   int     // 71 - Annotation type (0=none, 1=text, 2=block, 3=tolerance)
	dimstyle         string  // 3 - Dimension style name
	width            float64 // 40 - Text width (annotation type 1 only)
	height           float64 // 140 - Text height (annotation type 1 only)
	textStyle        string  // 7 - Text style name (annotation type 1 only)
	textRotation     float64 // 51 - Text rotation angle (annotation type 1 only)
	textAlignment    int     // 72 - Text horizontal alignment (annotation type 1 only)
	textColor        int     // 74 - Text color number (annotation type 1 only)
	dimensionType    int     // 340 - Dimension type (if annotation type is dimension)
	blockName        string  // 3 - Block name (if annotation type is block)
	blockColor       int     // 77 - Block color (for BYBLOCK)
	annotationHandle string  // 340 - Handle to annotation entity

	// Attachment data
	hAlign int // 73 - Horizontal attachment type
	vAlign int // 74 - Vertical attachment type

	// Hook line system
	hasHookline                     bool      // 75 - Hook line presence flag
	hooklineDirection               int       // 74 - Hook line direction flag
	leaderOffsetBlockRef            math.Vec3 // 212 - Leader offset for block reference
	leaderOffsetAnnotationPlacement math.Vec3 // 213 - Leader offset for annotation placement

	// Arrowhead system
	hasArrowhead   bool              // 71 - Arrowhead presence flag
	arrowheadType  int               // Arrowhead type
	arrowheadSize  float64           // Arrowhead size
	arrowheadColor color.ColorNumber // Arrowhead color
}

// NewLeader creates a new LEADER entity
func NewLeader() *Leader {
	return &Leader{
		entity:                          NewEntity(LEADER),
		points:                          make([][]float64, 0),
		pathType:                        0,
		extrusion:                       []float64{0.0, 0.0, 1.0},
		normalVector:                    []float64{0.0, 0.0, 1.0},
		horizontalDirection:             math.NewVec3(1, 0, 0),
		annotationType:                  0,
		dimstyle:                        "",
		width:                           1.0,
		height:                          1.0,
		textStyle:                       "",
		textRotation:                    0.0,
		textAlignment:                   0,
		textColor:                       0,
		dimensionType:                   0,
		blockName:                       "",
		blockColor:                      0,
		annotationHandle:                "",
		hAlign:                          0,
		vAlign:                          0,
		hasHookline:                     false,
		hooklineDirection:               0,
		leaderOffsetBlockRef:            math.NewVec3(0, 0, 0),
		leaderOffsetAnnotationPlacement: math.NewVec3(0, 0, 0),
		hasArrowhead:                    false,
		arrowheadType:                   LeaderArrowheadClosed,
		arrowheadSize:                   1.0,
		arrowheadColor:                  0,
	}
}

// NewLeaderWithPoints creates a LEADER with specified points
func NewLeaderWithPoints(points [][]float64) *Leader {
	leader := NewLeader()
	leader.points = points
	return leader
}

// IsEntity is for Entity interface.
func (l *Leader) IsEntity() bool {
	return true
}

// Format writes data to formatter.
func (l *Leader) Format(f format.Formatter) {
	l.entity.Format(f)
	f.WriteString(100, "AcDbLeader")

	for i := 0; i < len(l.points); i++ {
		if len(l.points[i]) >= 3 {
			f.WriteFloat(10, l.points[i][0])
			f.WriteFloat(20, l.points[i][1])
			f.WriteFloat(30, l.points[i][2])
		} else if len(l.points[i]) >= 2 {
			f.WriteFloat(10, l.points[i][0])
			f.WriteFloat(20, l.points[i][1])
			f.WriteFloat(30, 0.0)
		}
	}

	if l.pathType != 0 {
		f.WriteInt(72, l.pathType)
	}

	if l.annotationType != 0 {
		f.WriteInt(71, l.annotationType)

		if l.dimstyle != "" {
			f.WriteString(3, l.dimstyle)
		}

		if l.annotationType == 1 {
			if l.width != 1.0 {
				f.WriteFloat(40, l.width)
			}
			if l.height != 1.0 {
				f.WriteFloat(140, l.height)
			}
			if l.textStyle != "" {
				f.WriteString(7, l.textStyle)
			}
			if l.textRotation != 0.0 {
				f.WriteFloat(51, l.textRotation)
			}
			if l.textAlignment != 0 {
				f.WriteInt(72, l.textAlignment)
			}
			if l.textColor != 0 {
				f.WriteInt(74, l.textColor)
			}
		}

		if l.annotationType == 2 {
			if l.blockName != "" {
				f.WriteString(3, l.blockName)
			}
			if l.blockColor != 0 {
				f.WriteInt(77, l.blockColor)
			}
		}

		if l.annotationType == 3 {
			if l.dimensionType != 0 {
				f.WriteInt(340, l.dimensionType)
			}
		}

		if l.annotationHandle != "" {
			f.WriteString(340, l.annotationHandle)
		}
	}

	if l.hAlign != 0 || l.vAlign != 0 {
		f.WriteInt(73, l.hAlign)
		f.WriteInt(74, l.vAlign)
	}

	if l.hasHookline {
		f.WriteInt(75, 1)
		if l.hooklineDirection != 0 {
			f.WriteInt(74, l.hooklineDirection)
		}
	}

	if l.hasArrowhead {
		f.WriteInt(71, 1)
		if l.arrowheadType != LeaderArrowheadClosed {
			f.WriteInt(71, l.arrowheadType)
		}
		if l.arrowheadSize != 1.0 {
			f.WriteFloat(40, l.arrowheadSize)
		}
		if int(l.arrowheadColor) != 0 {
			f.WriteInt(62, int(l.arrowheadColor))
		}
	}

	if l.horizontalDirection.X() != 1 || l.horizontalDirection.Y() != 0 || l.horizontalDirection.Z() != 0 {
		f.WriteFloat(211, l.horizontalDirection.X())
		f.WriteFloat(221, l.horizontalDirection.Y())
		f.WriteFloat(231, l.horizontalDirection.Z())
	}

	if l.leaderOffsetBlockRef.X() != 0 || l.leaderOffsetBlockRef.Y() != 0 || l.leaderOffsetBlockRef.Z() != 0 {
		f.WriteFloat(212, l.leaderOffsetBlockRef.X())
		f.WriteFloat(222, l.leaderOffsetBlockRef.Y())
		f.WriteFloat(232, l.leaderOffsetBlockRef.Z())
	}

	if l.leaderOffsetAnnotationPlacement.X() != 0 || l.leaderOffsetAnnotationPlacement.Y() != 0 || l.leaderOffsetAnnotationPlacement.Z() != 0 {
		f.WriteFloat(213, l.leaderOffsetAnnotationPlacement.X())
		f.WriteFloat(223, l.leaderOffsetAnnotationPlacement.Y())
		f.WriteFloat(233, l.leaderOffsetAnnotationPlacement.Z())
	}

	if len(l.extrusion) >= 3 && (l.extrusion[0] != 0.0 || l.extrusion[1] != 0.0 || l.extrusion[2] != 1.0) {
		f.WriteFloat(210, l.extrusion[0])
		f.WriteFloat(220, l.extrusion[1])
		f.WriteFloat(230, l.extrusion[2])
	}
}

// BBox returns bounding box
func (l *Leader) BBox() ([]float64, []float64) {
	mins := make([]float64, 3)
	maxs := make([]float64, 3)

	if len(l.points) > 0 {
		if len(l.points[0]) >= 3 {
			copyFloatSlice(mins, l.points[0])
			copyFloatSlice(maxs, l.points[0])
		} else if len(l.points[0]) >= 2 {
			mins[0] = l.points[0][0]
			mins[1] = l.points[0][1]
			mins[2] = 0.0
			maxs[0] = l.points[0][0]
			maxs[1] = l.points[0][1]
			maxs[2] = 0.0
		}

		for _, point := range l.points {
			if len(point) >= 3 {
				for i := 0; i < 3; i++ {
					if point[i] < mins[i] {
						mins[i] = point[i]
					}
					if point[i] > maxs[i] {
						maxs[i] = point[i]
					}
				}
			} else if len(point) >= 2 {
				for i := 0; i < 2; i++ {
					if point[i] < mins[i] {
						mins[i] = point[i]
					}
					if point[i] > maxs[i] {
						maxs[i] = point[i]
					}
				}
			}
		}
	}

	return mins, maxs
}

// AddPoint adds a point to the leader
func (l *Leader) AddPoint(x, y, z float64) {
	l.points = append(l.points, []float64{x, y, z})
}

// SetPoints sets all leader points
func (l *Leader) SetPoints(points [][]float64) {
	l.points = points
}

// SetAnnotationType sets the annotation type
func (l *Leader) SetAnnotationType(annType int) {
	l.annotationType = annType
}

// SetTextProperties sets text annotation properties
func (l *Leader) SetTextProperties(width, height float64, style string, rotation float64, alignment int) {
	l.width = width
	l.height = height
	l.textStyle = style
	l.textRotation = rotation
	l.textAlignment = alignment
	l.annotationType = 1
}

// SetBlockAnnotation sets block annotation properties
func (l *Leader) SetBlockAnnotation(blockName string) {
	l.blockName = blockName
	l.annotationType = 2
}

// SetDimensionAnnotation sets dimension annotation properties
func (l *Leader) SetDimensionAnnotation(dimType int) {
	l.dimensionType = dimType
	l.annotationType = 3
}

// SetAttachment sets the attachment types
func (l *Leader) SetAttachment(hAlign, vAlign int) {
	l.hAlign = hAlign
	l.vAlign = vAlign
}

// SetExtrusion sets the extrusion direction
func (l *Leader) SetExtrusion(x, y, z float64) {
	l.extrusion = []float64{x, y, z}
}

// GetPoints returns all leader points
func (l *Leader) GetPoints() [][]float64 {
	return l.points
}

// GetAnnotationType returns the annotation type
func (l *Leader) GetAnnotationType() int {
	return l.annotationType
}

// Move translates the leader by specified offset
func (l *Leader) Move(dx, dy, dz float64) {
	for _, point := range l.points {
		if len(point) >= 3 {
			point[0] += dx
			point[1] += dy
			point[2] += dz
		} else if len(point) >= 2 {
			point[0] += dx
			point[1] += dy
		}
	}
}

// SetArrowhead sets arrowhead properties
func (l *Leader) SetArrowhead(arrowheadType int, size float64, arrowheadColor color.ColorNumber) {
	l.hasArrowhead = true
	l.arrowheadType = arrowheadType
	l.arrowheadSize = size
	l.arrowheadColor = arrowheadColor
}

// SetHookLine sets hook line properties
func (l *Leader) SetHookLine(hasHookline bool, direction int) {
	l.hasHookline = hasHookline
	l.hooklineDirection = direction
}

// SetPathType sets the path type (straight or spline)
func (l *Leader) SetPathType(pathType int) {
	l.pathType = pathType
}

// SetDimstyle sets the dimension style
func (l *Leader) SetDimstyle(dimstyle string) {
	l.dimstyle = dimstyle
}

// SetAnnotationHandle sets the annotation entity handle
func (l *Leader) SetAnnotationHandle(handle string) {
	l.annotationHandle = handle
}

// SetLeaderOffsets sets the offset vectors
func (l *Leader) SetLeaderOffsets(blockRef, annotationPlacement math.Vec3) {
	l.leaderOffsetBlockRef = blockRef
	l.leaderOffsetAnnotationPlacement = annotationPlacement
}

// SetHorizontalDirection sets the horizontal direction vector
func (l *Leader) SetHorizontalDirection(direction math.Vec3) {
	l.horizontalDirection = direction
}

// CalculateHookLine calculates hook line endpoints based on text properties
func (l *Leader) CalculateHookLine() (math.Vec3, math.Vec3) {
	if !l.hasHookline || len(l.points) == 0 {
		return math.NewVec3(0, 0, 0), math.NewVec3(0, 0, 0)
	}

	var lastPoint math.Vec3
	if len(l.points[len(l.points)-1]) >= 3 {
		lastPoint = math.NewVec3(
			l.points[len(l.points)-1][0],
			l.points[len(l.points)-1][1],
			l.points[len(l.points)-1][2],
		)
	} else if len(l.points[len(l.points)-1]) >= 2 {
		lastPoint = math.NewVec3(
			l.points[len(l.points)-1][0],
			l.points[len(l.points)-1][1],
			0.0,
		)
	}

	hookLength := l.height * 0.5
	var hookDirection math.Vec3

	switch l.hooklineDirection {
	case 0:
		hookDirection = math.NewVec3(-1, 0, 0)
	case 1:
		hookDirection = math.NewVec3(1, 0, 0)
	case 2:
		hookDirection = math.NewVec3(0, 1, 0)
	case 3:
		hookDirection = math.NewVec3(0, -1, 0)
	default:
		hookDirection = math.NewVec3(1, 0, 0)
	}

	hookEnd := math.NewVec3(
		lastPoint.X()+hookDirection.X()*hookLength,
		lastPoint.Y()+hookDirection.Y()*hookLength,
		lastPoint.Z()+hookDirection.Z()*hookLength,
	)
	return lastPoint, hookEnd
}

// GenerateSplinePoints converts leader points to spline control points
func (l *Leader) GenerateSplinePoints() []math.Vec3 {
	if l.pathType != 1 || len(l.points) < 2 {
		splinePoints := make([]math.Vec3, len(l.points))
		for i, point := range l.points {
			if len(point) >= 3 {
				splinePoints[i] = math.NewVec3(point[0], point[1], point[2])
			} else if len(point) >= 2 {
				splinePoints[i] = math.NewVec3(point[0], point[1], 0.0)
			}
		}
		return splinePoints
	}

	splinePoints := make([]math.Vec3, len(l.points))
	for i, point := range l.points {
		if len(point) >= 3 {
			splinePoints[i] = math.NewVec3(point[0], point[1], point[2])
		} else if len(point) >= 2 {
			splinePoints[i] = math.NewVec3(point[0], point[1], 0.0)
		}
	}
	return splinePoints
}

// String returns string representation
func (l *Leader) String() string {
	annType := "None"
	switch l.annotationType {
	case 1:
		annType = "Text"
	case 2:
		annType = "Block"
	case 3:
		annType = "Dimension"
	}

	pathTypeStr := "Straight"
	if l.pathType == 1 {
		pathTypeStr = "Spline"
	}

	var arrowInfo string
	if l.hasArrowhead {
		arrowInfo = fmt.Sprintf(", Arrow:%d", l.arrowheadType)
	}

	var hookInfo string
	if l.hasHookline {
		hookInfo = ", Hook"
	}

	return fmt.Sprintf("Leader{Points: %d, Path: %s%s%s, Annotation: %s}",
		len(l.points), pathTypeStr, arrowInfo, hookInfo, annType)
}

// Helper function to copy slice
func copyFloatSlice(dst, src []float64) {
	if len(dst) >= len(src) {
		for i := 0; i < len(src); i++ {
			dst[i] = src[i]
		}
	}
}

var _ Entity = (*Leader)(nil)
