package path

import (
	"fmt"
	"math"

	dxfmath "github.com/edanko/dxf/math"
)

// PathCommand represents a command in a 2D path
type PathCommand interface {
	Type() PathCommandType
	EndVertex() dxfmath.Vec2
	String() string
}

// PathCommandType represents different types of path commands
type PathCommandType int

const (
	CommandMoveTo   PathCommandType = iota // Move to point
	CommandLineTo                          // Line to point
	CommandCurve3To                        // Quadratic Bézier curve
	CommandCurve4To                        // Cubic Bézier curve
)

// PathMoveTo represents a MOVE_TO command
type PathMoveTo struct {
	End  dxfmath.Vec2
	user interface{}
}

func NewMoveTo(x, y float64) *PathMoveTo {
	return &PathMoveTo{
		End: dxfmath.NewVec2(x, y),
	}
}

func (m *PathMoveTo) Type() PathCommandType {
	return CommandMoveTo
}

func (m *PathMoveTo) EndVertex() dxfmath.Vec2 {
	return m.End
}

func (m *PathMoveTo) String() string {
	return fmt.Sprintf("MoveTo(%.3f, %.3f)", m.End.X(), m.End.Y())
}

// PathLineTo represents a LINE_TO command
type PathLineTo struct {
	End  dxfmath.Vec2
	user interface{}
}

func NewLineTo(x, y float64) *PathLineTo {
	return &PathLineTo{
		End: dxfmath.NewVec2(x, y),
	}
}

func (l *PathLineTo) Type() PathCommandType {
	return CommandLineTo
}

func (l *PathLineTo) EndVertex() dxfmath.Vec2 {
	return l.End
}

func (l *PathLineTo) String() string {
	return fmt.Sprintf("LineTo(%.3f, %.3f)", l.End.X(), l.End.Y())
}

// PathCurve3To represents a CURVE3_TO command (quadratic Bézier)
type PathCurve3To struct {
	End  dxfmath.Vec2
	Ctrl dxfmath.Vec2
	user interface{}
}

func NewCurve3To(endX, endY, ctrlX, ctrlY float64) *PathCurve3To {
	return &PathCurve3To{
		End:  dxfmath.NewVec2(endX, endY),
		Ctrl: dxfmath.NewVec2(ctrlX, ctrlY),
	}
}

func (c *PathCurve3To) Type() PathCommandType {
	return CommandCurve3To
}

func (c *PathCurve3To) EndVertex() dxfmath.Vec2 {
	return c.End
}

func (c *PathCurve3To) String() string {
	return fmt.Sprintf("Curve3To(%.3f, %.3f, %.3f, %.3f)", c.Ctrl.X(), c.Ctrl.Y(), c.End.X(), c.End.Y())
}

// PathCurve4To represents a CURVE4_TO command (cubic Bézier)
type PathCurve4To struct {
	End   dxfmath.Vec2
	Ctrl1 dxfmath.Vec2
	Ctrl2 dxfmath.Vec2
	user  interface{}
}

func NewCurve4To(endX, endY, ctrl1X, ctrl1Y, ctrl2X, ctrl2Y float64) *PathCurve4To {
	return &PathCurve4To{
		End:   dxfmath.NewVec2(endX, endY),
		Ctrl1: dxfmath.NewVec2(ctrl1X, ctrl1Y),
		Ctrl2: dxfmath.NewVec2(ctrl2X, ctrl2Y),
	}
}

func (c *PathCurve4To) Type() PathCommandType {
	return CommandCurve4To
}

func (c *PathCurve4To) EndVertex() dxfmath.Vec2 {
	return c.End
}

func (c *PathCurve4To) String() string {
	return fmt.Sprintf("Curve4To(%.3f, %.3f, %.3f, %.3f, %.3f, %.3f)",
		c.Ctrl1.X(), c.Ctrl1.Y(), c.Ctrl2.X(), c.Ctrl2.Y(), c.End.X(), c.End.Y())
}

// Path represents a 2D path with command-based construction
type Path struct {
	vertices   []dxfmath.Vec2
	commands   []PathCommand
	startIndex []int
	user       interface{}
}

// NewPath creates a new empty path
func NewPath() *Path {
	return &Path{
		vertices:   make([]dxfmath.Vec2, 0),
		commands:   make([]PathCommand, 0),
		startIndex: make([]int, 0),
	}
}

// addElement adds a command and its vertices to path
func (p *Path) addElement(cmd PathCommand, vertices []dxfmath.Vec2) {
	p.commands = append(p.commands, cmd)

	// Add vertices based on command type
	p.vertices = append(p.vertices, vertices...)

	// Mark if this creates a new sub-path (move command)
	if cmd.Type() == CommandMoveTo {
		p.startIndex = append(p.startIndex, len(p.vertices)-1)
	}
}

// MoveTo moves the current position to a new point
func (p *Path) MoveTo(x, y float64) {
	cmd := NewMoveTo(x, y)
	p.addElement(cmd, []dxfmath.Vec2{cmd.EndVertex()})
}

// LineTo draws a line from the current position to a new point
func (p *Path) LineTo(x, y float64) {
	cmd := NewLineTo(x, y)
	p.addElement(cmd, []dxfmath.Vec2{cmd.EndVertex()})
}

// Curve3To draws a quadratic Bézier curve
func (p *Path) Curve3To(endX, endY, ctrlX, ctrlY float64) {
	cmd := NewCurve3To(endX, endY, ctrlX, ctrlY)
	p.addElement(cmd, []dxfmath.Vec2{cmd.Ctrl, cmd.EndVertex()})
}

// Curve4To draws a cubic Bézier curve
func (p *Path) Curve4To(endX, endY, ctrl1X, ctrl1Y, ctrl2X, ctrl2Y float64) {
	cmd := NewCurve4To(endX, endY, ctrl1X, ctrl1Y, ctrl2X, ctrl2Y)
	p.addElement(cmd, []dxfmath.Vec2{cmd.Ctrl1, cmd.Ctrl2, cmd.EndVertex()})
}

// Close adds a line from the current position back to the start
func (p *Path) Close() {
	if len(p.vertices) == 0 {
		return
	}

	// Get the start position of current sub-path
	var startPoint dxfmath.Vec2
	if len(p.startIndex) > 0 {
		lastStartIndex := p.startIndex[len(p.startIndex)-1]
		startPoint = p.vertices[lastStartIndex]
	} else {
		startPoint = p.vertices[0]
	}

	p.LineTo(startPoint.X(), startPoint.Y())
}

// CurrentPosition returns the current position in the path
func (p *Path) CurrentPosition() dxfmath.Vec2 {
	if len(p.vertices) == 0 {
		return dxfmath.NewVec2(0, 0)
	}
	return p.vertices[len(p.vertices)-1]
}

// BoundingBox returns the bounding box of the path
func (p *Path) BoundingBox() (dxfmath.Vec2, dxfmath.Vec2) {
	if len(p.vertices) == 0 {
		return dxfmath.NewVec2(0, 0), dxfmath.NewVec2(0, 0)
	}

	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64

	for _, vertex := range p.vertices {
		x, y := vertex.X(), vertex.Y()
		minX = math.Min(minX, x)
		minY = math.Min(minY, y)
		maxX = math.Max(maxX, x)
		maxY = math.Max(maxY, y)
	}

	return dxfmath.NewVec2(minX, minY), dxfmath.NewVec2(maxX, maxY)
}

// Length returns the approximate length of the path
func (p *Path) Length() float64 {
	if len(p.vertices) < 2 {
		return 0
	}

	var length float64
	for i := 1; i < len(p.vertices); i++ {
		length += p.vertices[i].Distance(p.vertices[i-1])
	}

	return length
}

// Commands returns a copy of the path commands
func (p *Path) Commands() []PathCommand {
	commands := make([]PathCommand, len(p.commands))
	copy(commands, p.commands)
	return commands
}

// Vertices returns a copy of the path vertices
func (p *Path) Vertices() []dxfmath.Vec2 {
	vertices := make([]dxfmath.Vec2, len(p.vertices))
	copy(vertices, p.vertices)
	return vertices
}

// Clone creates a deep copy of the path
func (p *Path) Clone() *Path {
	return &Path{
		vertices:   p.Vertices(),
		commands:   p.Commands(),
		startIndex: append([]int{}, p.startIndex...),
		user:       p.user,
	}
}

// String returns a string representation of the path
func (p *Path) String() string {
	if len(p.commands) == 0 {
		return "Path(empty)"
	}

	var result string
	for _, cmd := range p.commands {
		result += cmd.String() + " "
	}
	return result
}

// IsClosed returns true if the path start point is close to the end point
func (p *Path) IsClosed() bool {
	if len(p.vertices) < 2 {
		return false
	}
	start := p.vertices[0]
	end := p.vertices[len(p.vertices)-1]
	return start.Distance(end) < 1e-6
}

// HasLines returns true if the path has any line segments
func (p *Path) HasLines() bool {
	for _, cmd := range p.commands {
		if cmd.Type() == CommandLineTo {
			return true
		}
	}
	return false
}

// HasCurves returns true if the path has any curve segments
func (p *Path) HasCurves() bool {
	for _, cmd := range p.commands {
		if cmd.Type() == CommandCurve3To || cmd.Type() == CommandCurve4To {
			return true
		}
	}
	return false
}

// HasSubPaths returns true if the path has multiple sub-paths
func (p *Path) HasSubPaths() bool {
	return len(p.startIndex) > 1
}

// Start returns the start point of the path
func (p *Path) Start() dxfmath.Vec2 {
	if len(p.vertices) == 0 {
		return dxfmath.NewVec2(0, 0)
	}
	return p.vertices[0]
}

// End returns the end point of the path
func (p *Path) End() dxfmath.Vec2 {
	if len(p.vertices) == 0 {
		return dxfmath.NewVec2(0, 0)
	}
	return p.vertices[len(p.vertices)-1]
}

// ControlVertices returns all control vertices in consecutive order
func (p *Path) ControlVertices() []dxfmath.Vec2 {
	vertices := make([]dxfmath.Vec2, len(p.vertices))
	copy(vertices, p.vertices)
	return vertices
}

// Approximate approximates curves with line segments
func (p *Path) Approximate(maxDistance float64, minSegments int) *Path {
	if len(p.commands) == 0 {
		return NewPath()
	}

	result := NewPath()

	// Start with the first vertex
	if len(p.vertices) > 0 {
		result.MoveTo(p.vertices[0].X(), p.vertices[0].Y())
	}

	i := 0 // Current vertex index
	for _, cmd := range p.commands {
		switch cmd.Type() {
		case CommandMoveTo:
			if i < len(p.vertices) {
				result.MoveTo(p.vertices[i].X(), p.vertices[i].Y())
				i++
			}
		case CommandLineTo:
			if i < len(p.vertices) {
				result.LineTo(p.vertices[i].X(), p.vertices[i].Y())
				i++
			}
		case CommandCurve3To:
			if i+1 < len(p.vertices) {
				// Approximate quadratic curve
				start := result.CurrentPosition()
				ctrl := p.vertices[i]
				end := p.vertices[i+1]
				p.approximateQuadraticCurve(start, ctrl, end, maxDistance, minSegments, result)
				i += 2
			}
		case CommandCurve4To:
			if i+2 < len(p.vertices) {
				// Approximate cubic curve
				start := result.CurrentPosition()
				ctrl1 := p.vertices[i]
				ctrl2 := p.vertices[i+1]
				end := p.vertices[i+2]
				p.approximateCubicCurve(start, ctrl1, ctrl2, end, maxDistance, minSegments, result)
				i += 3
			}
		}
	}

	return result
}

// approximateQuadraticCurve approximates a quadratic Bézier curve with line segments
func (p *Path) approximateQuadraticCurve(start, ctrl, end dxfmath.Vec2, maxDistance float64, minSegments int, result *Path) {
	// Simple implementation - subdivide into segments
	segments := max(minSegments, 4)

	for i := 1; i <= segments; i++ {
		t := float64(i) / float64(segments)
		point := p.quadraticBezierPoint(start, ctrl, end, t)
		result.LineTo(point.X(), point.Y())
	}
}

// approximateCubicCurve approximates a cubic Bézier curve with line segments
func (p *Path) approximateCubicCurve(start, ctrl1, ctrl2, end dxfmath.Vec2, maxDistance float64, minSegments int, result *Path) {
	// Simple implementation - subdivide into segments
	segments := max(minSegments, 8)

	for i := 1; i <= segments; i++ {
		t := float64(i) / float64(segments)
		point := p.cubicBezierPoint(start, ctrl1, ctrl2, end, t)
		result.LineTo(point.X(), point.Y())
	}
}

// quadraticBezierPoint calculates a point on a quadratic Bézier curve
func (p *Path) quadraticBezierPoint(start, ctrl, end dxfmath.Vec2, t float64) dxfmath.Vec2 {
	oneMinusT := 1.0 - t
	oneMinusTSquared := oneMinusT * oneMinusT
	tSquared := t * t

	x := oneMinusTSquared*start.X() + 2*oneMinusT*t*ctrl.X() + tSquared*end.X()
	y := oneMinusTSquared*start.Y() + 2*oneMinusT*t*ctrl.Y() + tSquared*end.Y()

	return dxfmath.NewVec2(x, y)
}

// cubicBezierPoint calculates a point on a cubic Bézier curve
func (p *Path) cubicBezierPoint(start, ctrl1, ctrl2, end dxfmath.Vec2, t float64) dxfmath.Vec2 {
	oneMinusT := 1.0 - t
	oneMinusTSquared := oneMinusT * oneMinusT
	oneMinusTCubed := oneMinusTSquared * oneMinusT
	tSquared := t * t
	tCubed := tSquared * t

	x := oneMinusTCubed*start.X() + 3*oneMinusTSquared*t*ctrl1.X() + 3*oneMinusT*tSquared*ctrl2.X() + tCubed*end.X()
	y := oneMinusTCubed*start.Y() + 3*oneMinusTSquared*t*ctrl1.Y() + 3*oneMinusT*tSquared*ctrl2.Y() + tCubed*end.Y()

	return dxfmath.NewVec2(x, y)
}

// SubPaths returns all sub-paths as separate Path objects
func (p *Path) SubPaths() []*Path {
	if !p.HasSubPaths() {
		return []*Path{p.Clone()}
	}

	var subPaths []*Path
	var currentSubPath *Path

	// Process commands sequentially
	vertexIndex := 0

	for _, cmd := range p.commands {
		switch cmd.Type() {
		case CommandMoveTo:
			// Save current sub-path if it exists and has commands
			if currentSubPath != nil && len(currentSubPath.commands) > 0 {
				subPaths = append(subPaths, currentSubPath)
			}
			// Start a new sub-path
			if vertexIndex < len(p.vertices) {
				currentSubPath = NewPath()
				currentSubPath.MoveTo(p.vertices[vertexIndex].X(), p.vertices[vertexIndex].Y())
				vertexIndex++
			}
		case CommandLineTo:
			if currentSubPath != nil && vertexIndex < len(p.vertices) {
				currentSubPath.LineTo(p.vertices[vertexIndex].X(), p.vertices[vertexIndex].Y())
				vertexIndex++
			}
		case CommandCurve3To:
			if currentSubPath != nil && vertexIndex+1 < len(p.vertices) {
				ctrl := p.vertices[vertexIndex]
				end := p.vertices[vertexIndex+1]
				currentSubPath.Curve3To(end.X(), end.Y(), ctrl.X(), ctrl.Y())
				vertexIndex += 2
			}
		case CommandCurve4To:
			if currentSubPath != nil && vertexIndex+2 < len(p.vertices) {
				ctrl1 := p.vertices[vertexIndex]
				ctrl2 := p.vertices[vertexIndex+1]
				end := p.vertices[vertexIndex+2]
				currentSubPath.Curve4To(end.X(), end.Y(), ctrl1.X(), ctrl1.Y(), ctrl2.X(), ctrl2.Y())
				vertexIndex += 3
			}
		}
	}

	// Add the final sub-path
	if currentSubPath != nil && len(currentSubPath.commands) > 0 {
		subPaths = append(subPaths, currentSubPath)
	}

	return subPaths
}

// Transform applies a transformation matrix to the path
func (p *Path) Transform(matrix *dxfmath.Matrix44) *Path {
	result := p.Clone()

	// Transform all vertices
	for i, vertex := range result.vertices {
		// Convert Vec2 to Vec3 for transformation
		v3 := dxfmath.NewVec3(vertex.X(), vertex.Y(), 0)
		transformed := matrix.TransformVector(v3)
		result.vertices[i] = dxfmath.NewVec2(transformed.X(), transformed.Y())
	}

	return result
}
