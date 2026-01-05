package path

import (
	"math"

	dxfmath "github.com/edanko/dxf/math"
)

// Helper functions for Min/Max of multiple values
func min3(a, b, c float64) float64 {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func max3(a, b, c float64) float64 {
	if a > b {
		if a > c {
			return a
		}
		return c
	}
	if b > c {
		return b
	}
	return c
}

func min4(a, b, c, d float64) float64 {
	return math.Min(math.Min(a, b), math.Min(c, d))
}

func max4(a, b, c, d float64) float64 {
	return math.Max(math.Max(a, b), math.Max(c, d))
}

// Tools provides advanced path operations similar to Python ezdxf path/tools.py
type Tools struct {
	path *Path
}

// NewTools creates a new Tools instance for path operations
func NewTools(path *Path) *Tools {
	return &Tools{path: path}
}

// BoundingBox calculates the bounding box of the path
func (t *Tools) BoundingBox() dxfmath.BoundingBox {
	if len(t.path.commands) == 0 {
		return dxfmath.NewBoundingBox(dxfmath.NewVec2(0, 0), dxfmath.NewVec2(0, 0))
	}

	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)

	for _, cmd := range t.path.commands {
		switch c := cmd.(type) {
		case *PathMoveTo:
			minX = math.Min(minX, c.End.X())
			minY = math.Min(minY, c.End.Y())
			maxX = math.Max(maxX, c.End.X())
			maxY = math.Max(maxY, c.End.Y())
		case *PathLineTo:
			minX = math.Min(minX, c.End.X())
			minY = math.Min(minY, c.End.Y())
			maxX = math.Max(maxX, c.End.X())
			maxY = math.Max(maxY, c.End.Y())
		case *PathCurve3To:
			// For quadratic curves, check control point and end point
			points := []dxfmath.Vec2{c.Ctrl, c.End}
			for _, p := range points {
				minX = math.Min(minX, p.X())
				minY = math.Min(minY, p.Y())
				maxX = math.Max(maxX, p.X())
				maxY = math.Max(maxY, p.Y())
			}
		case *PathCurve4To:
			// For cubic curves, check control points and end point
			points := []dxfmath.Vec2{c.Ctrl1, c.Ctrl2, c.End}
			for _, p := range points {
				minX = math.Min(minX, p.X())
				minY = math.Min(minY, p.Y())
				maxX = math.Max(maxX, p.X())
				maxY = math.Max(maxY, p.Y())
			}
		}
	}

	return dxfmath.NewBoundingBox(
		dxfmath.NewVec2(minX, minY),
		dxfmath.NewVec2(maxX, maxY),
	)
}

// Length calculates the approximate length of the path
func (t *Tools) Length() float64 {
	return t.LengthWithSegments(10) // Default 10 segments per curve
}

// LengthWithSegments calculates the approximate length with specified curve segments
func (t *Tools) LengthWithSegments(segments int) float64 {
	if segments < 1 {
		segments = 1
	}

	var totalLength float64
	var currentPos *dxfmath.Vec2
	hasCurrentPos := false

	for _, cmd := range t.path.commands {
		switch c := cmd.(type) {
		case *PathMoveTo:
			currentPos = &c.End
			hasCurrentPos = true
		case *PathLineTo:
			if hasCurrentPos && currentPos != nil {
				totalLength += currentPos.Distance(c.End)
				currentPos = &c.End
			}
		case *PathCurve3To:
			if hasCurrentPos && currentPos != nil {
				// Approximate quadratic Bezier curve length
				length := t.approximateQuadraticLength(*currentPos, c.Ctrl, c.End, segments)
				totalLength += length
				currentPos = &c.End
			}
		case *PathCurve4To:
			if hasCurrentPos && currentPos != nil {
				// Approximate cubic Bezier curve length
				length := t.approximateCubicLength(*currentPos, c.Ctrl1, c.Ctrl2, c.End, segments)
				totalLength += length
				currentPos = &c.End
			}
		}
	}

	return totalLength
}

// IsClosed checks if the path is closed (ends at start point)
func (t *Tools) IsClosed() bool {
	return t.path.IsClosed()
}

// Close ensures the path is closed by adding a line if needed
func (t *Tools) Close() {
	if !t.IsClosed() {
		t.path.Close()
	}
}

// Simplify reduces number of vertices in path while preserving shape
func (t *Tools) Simplify(tolerance float64) *Path {
	if tolerance <= 0 {
		tolerance = 0.1
	}

	// Flatten path to points
	points := t.flatten()
	if len(points) < 3 {
		return t.path.Clone()
	}

	// Douglas-Peucker algorithm
	simplified := t.douglasPeucker(points, tolerance)

	newPath := NewPath()
	if len(simplified) > 0 {
		newPath.MoveTo(simplified[0].X(), simplified[0].Y())
		for i := 1; i < len(simplified); i++ {
			newPath.LineTo(simplified[i].X(), simplified[i].Y())
		}
	}

	return newPath
}

// Transform applies a transformation matrix to the path
func (t *Tools) Transform(matrix *dxfmath.Matrix44) *Path {
	newPath := NewPath()

	for _, cmd := range t.path.commands {
		switch c := cmd.(type) {
		case *PathMoveTo:
			transformed := matrix.MulVec(dxfmath.NewVec3(c.End.X(), c.End.Y(), 0))
			newPath.MoveTo(transformed[0], transformed[1])
		case *PathLineTo:
			transformed := matrix.MulVec(dxfmath.NewVec3(c.End.X(), c.End.Y(), 0))
			newPath.LineTo(transformed[0], transformed[1])
		case *PathCurve3To:
			end := matrix.MulVec(dxfmath.NewVec3(c.End.X(), c.End.Y(), 0))
			ctrl := matrix.MulVec(dxfmath.NewVec3(c.Ctrl.X(), c.Ctrl.Y(), 0))
			newPath.Curve3To(end[0], end[1], ctrl[0], ctrl[1])
		case *PathCurve4To:
			end := matrix.MulVec(dxfmath.NewVec3(c.End.X(), c.End.Y(), 0))
			ctrl1 := matrix.MulVec(dxfmath.NewVec3(c.Ctrl1.X(), c.Ctrl1.Y(), 0))
			ctrl2 := matrix.MulVec(dxfmath.NewVec3(c.Ctrl2.X(), c.Ctrl2.Y(), 0))
			newPath.Curve4To(end[0], end[1], ctrl1[0], ctrl1[1], ctrl2[0], ctrl2[1])
		}
	}

	return newPath
}

// Helper functions

// approximateQuadraticLength approximates the length of a quadratic Bezier curve
func (t *Tools) approximateQuadraticLength(start, ctrl, end dxfmath.Vec2, segments int) float64 {
	var length float64
	prev := start

	for i := 1; i <= segments; i++ {
		s := float64(i) / float64(segments)
		// Quadratic Bezier formula
		x := (1-s)*(1-s)*start.X() + 2*(1-s)*s*ctrl.X() + s*s*end.X()
		y := (1-s)*(1-s)*start.Y() + 2*(1-s)*s*ctrl.Y() + s*s*end.Y()
		curr := dxfmath.NewVec2(x, y)

		length += prev.Distance(curr)
		prev = curr
	}

	return length
}

// approximateCubicLength approximates the length of a cubic Bezier curve
func (t *Tools) approximateCubicLength(start, ctrl1, ctrl2, end dxfmath.Vec2, segments int) float64 {
	var length float64
	prev := start

	for i := 1; i <= segments; i++ {
		s := float64(i) / float64(segments)
		// Cubic Bezier formula
		x := (1-s)*(1-s)*(1-s)*start.X() + 3*(1-s)*(1-s)*s*ctrl1.X() + 3*(1-s)*s*s*ctrl2.X() + s*s*s*end.X()
		y := (1-s)*(1-s)*(1-s)*start.Y() + 3*(1-s)*(1-s)*s*ctrl1.Y() + 3*(1-s)*s*s*ctrl2.Y() + s*s*s*end.Y()
		curr := dxfmath.NewVec2(x, y)

		length += prev.Distance(curr)
		prev = curr
	}

	return length
}

// flatten converts the path to a series of points
func (t *Tools) flatten() []dxfmath.Vec2 {
	var points []dxfmath.Vec2

	for _, cmd := range t.path.commands {
		switch c := cmd.(type) {
		case *PathMoveTo:
			if len(points) == 0 {
				points = append(points, c.End)
			}
		case *PathLineTo:
			points = append(points, c.End)
		case *PathCurve3To:
			// Approximate quadratic curve with line segments
			segments := 10
			if len(points) > 0 {
				start := points[len(points)-1]
				for i := 1; i <= segments; i++ {
					s := float64(i) / float64(segments)
					x := (1-s)*(1-s)*start.X() + 2*(1-s)*s*c.Ctrl.X() + s*s*c.End.X()
					y := (1-s)*(1-s)*start.Y() + 2*(1-s)*s*c.Ctrl.Y() + s*s*c.End.Y()
					points = append(points, dxfmath.NewVec2(x, y))
				}
			}
		case *PathCurve4To:
			// Approximate cubic curve with line segments
			segments := 10
			if len(points) > 0 {
				start := points[len(points)-1]
				for i := 1; i <= segments; i++ {
					s := float64(i) / float64(segments)
					x := (1-s)*(1-s)*(1-s)*start.X() + 3*(1-s)*(1-s)*s*c.Ctrl1.X() + 3*(1-s)*s*s*c.Ctrl2.X() + s*s*s*c.End.X()
					y := (1-s)*(1-s)*(1-s)*start.Y() + 3*(1-s)*(1-s)*s*c.Ctrl1.Y() + 3*(1-s)*s*s*c.Ctrl2.Y() + s*s*s*c.End.Y()
					points = append(points, dxfmath.NewVec2(x, y))
				}
			}
		}
	}

	return points
}

// douglasPeucker implements the Douglas-Peucker line simplification algorithm
func (t *Tools) douglasPeucker(points []dxfmath.Vec2, tolerance float64) []dxfmath.Vec2 {
	if len(points) <= 2 {
		return points
	}

	// Find the point with maximum distance
	maxDist := 0.0
	maxIndex := 0
	start := points[0]
	end := points[len(points)-1]

	for i := 1; i < len(points)-1; i++ {
		dist := t.perpendicularDistance(points[i], start, end)
		if dist > maxDist {
			maxDist = dist
			maxIndex = i
		}
	}

	// If max distance is greater than tolerance, recursively simplify
	if maxDist > tolerance {
		// Recursive call
		left := t.douglasPeucker(points[:maxIndex+1], tolerance)
		right := t.douglasPeucker(points[maxIndex:], tolerance)

		// Build result list
		result := left[:len(left)-1] // Remove duplicate point
		result = append(result, right...)
		return result
	} else {
		// Return simplified line
		return []dxfmath.Vec2{start, end}
	}
}

// perpendicularDistance calculates perpendicular distance from point to line
func (t *Tools) perpendicularDistance(point, lineStart, lineEnd dxfmath.Vec2) float64 {
	// Calculate the distance from point to line segment
	dx := lineEnd.X() - lineStart.X()
	dy := lineEnd.Y() - lineStart.Y()

	if dx == 0 && dy == 0 {
		// Line start and end are the same point
		return point.Distance(lineStart)
	}

	// Calculate the parameter param for the closest point on the line
	param := ((point.X()-lineStart.X())*dx + (point.Y()-lineStart.Y())*dy) / (dx*dx + dy*dy)

	// Clamp param to [0, 1] to stay within the line segment
	if param < 0 {
		param = 0
	} else if param > 1 {
		param = 1
	}

	// Calculate the closest point on the line segment
	closestX := lineStart.X() + param*dx
	closestY := lineStart.Y() + param*dy
	closest := dxfmath.NewVec2(closestX, closestY)

	// Return the distance
	return point.Distance(closest)
}

// HasCurves checks if the path contains any curve commands
func (t *Tools) HasCurves() bool {
	for _, cmd := range t.path.commands {
		switch cmd.(type) {
		case *PathCurve3To, *PathCurve4To:
			return true
		}
	}
	return false
}

// SegmentCount returns the number of path segments
func (t *Tools) SegmentCount() int {
	return len(t.path.commands)
}

// StartPoint returns the starting point of the path
func (t *Tools) StartPoint() dxfmath.Vec2 {
	if len(t.path.commands) == 0 {
		return dxfmath.NewVec2(0, 0)
	}

	switch cmd := t.path.commands[0].(type) {
	case *PathMoveTo:
		return cmd.End
	case *PathLineTo, *PathCurve3To, *PathCurve4To:
		return dxfmath.NewVec2(0, 0)
	}
	return dxfmath.NewVec2(0, 0)
}

// EndPoint returns the ending point of thepath
func (t *Tools) EndPoint() dxfmath.Vec2 {
	if len(t.path.commands) == 0 {
		return dxfmath.NewVec2(0, 0)
	}

	lastCmd := t.path.commands[len(t.path.commands)-1]
	switch cmd := lastCmd.(type) {
	case *PathMoveTo:
		return cmd.End
	case *PathLineTo:
		return cmd.End
	case *PathCurve3To:
		return cmd.End
	case *PathCurve4To:
		return cmd.End
	}
	return dxfmath.NewVec2(0, 0)
}

// Reverse creates a new path with commands in reverse order
func (t *Tools) Reverse() *Path {
	newPath := NewPath()

	// Find the last endpoint to start with MoveTo
	lastEnd := t.EndPoint()
	newPath.MoveTo(lastEnd.X(), lastEnd.Y())

	for i := len(t.path.commands) - 1; i >= 0; i-- {
		switch c := t.path.commands[i].(type) {
		case *PathMoveTo:
			// Skip - we already did MoveTo
		case *PathLineTo:
			newPath.LineTo(c.End.X(), c.End.Y())
		case *PathCurve3To:
			newPath.Curve3To(c.End.X(), c.End.Y(), c.Ctrl.X(), c.Ctrl.Y())
		case *PathCurve4To:
			newPath.Curve4To(c.End.X(), c.End.Y(), c.Ctrl1.X(), c.Ctrl1.Y(), c.Ctrl2.X(), c.Ctrl2.Y())
		}
	}

	return newPath
}

// Scale creates a new path scaled by the given factor
func (t *Tools) Scale(factor float64) *Path {
	newPath := NewPath()

	for _, cmd := range t.path.commands {
		switch c := cmd.(type) {
		case *PathMoveTo:
			newPath.MoveTo(c.End.X()*factor, c.End.Y()*factor)
		case *PathLineTo:
			newPath.LineTo(c.End.X()*factor, c.End.Y()*factor)
		case *PathCurve3To:
			newPath.Curve3To(c.End.X()*factor, c.End.Y()*factor, c.Ctrl.X()*factor, c.Ctrl.Y()*factor)
		case *PathCurve4To:
			newPath.Curve4To(c.End.X()*factor, c.End.Y()*factor,
				c.Ctrl1.X()*factor, c.Ctrl1.Y()*factor,
				c.Ctrl2.X()*factor, c.Ctrl2.Y()*factor)
		}
	}

	return newPath
}

// Translate creates a new path translated by the given offset
func (t *Tools) Translate(dx, dy float64) *Path {
	newPath := NewPath()

	for _, cmd := range t.path.commands {
		switch c := cmd.(type) {
		case *PathMoveTo:
			newPath.MoveTo(c.End.X()+dx, c.End.Y()+dy)
		case *PathLineTo:
			newPath.LineTo(c.End.X()+dx, c.End.Y()+dy)
		case *PathCurve3To:
			newPath.Curve3To(c.End.X()+dx, c.End.Y()+dy, c.Ctrl.X()+dx, c.Ctrl.Y()+dy)
		case *PathCurve4To:
			newPath.Curve4To(c.End.X()+dx, c.End.Y()+dy,
				c.Ctrl1.X()+dx, c.Ctrl1.Y()+dy,
				c.Ctrl2.X()+dx, c.Ctrl2.Y()+dy)
		}
	}

	return newPath
}

// Rotate creates a new path rotated by the given angle (in radians)
func (t *Tools) Rotate(angle float64) *Path {
	sin := math.Sin(angle)
	cos := math.Cos(angle)

	newPath := NewPath()

	for _, cmd := range t.path.commands {
		switch c := cmd.(type) {
		case *PathMoveTo:
			x := c.End.X()*cos - c.End.Y()*sin
			y := c.End.X()*sin + c.End.Y()*cos
			newPath.MoveTo(x, y)
		case *PathLineTo:
			x := c.End.X()*cos - c.End.Y()*sin
			y := c.End.X()*sin + c.End.Y()*cos
			newPath.LineTo(x, y)
		case *PathCurve3To:
			ex := c.End.X()*cos - c.End.Y()*sin
			ey := c.End.X()*sin + c.End.Y()*cos
			cx := c.Ctrl.X()*cos - c.Ctrl.Y()*sin
			cy := c.Ctrl.X()*sin + c.Ctrl.Y()*cos
			newPath.Curve3To(ex, ey, cx, cy)
		case *PathCurve4To:
			ex := c.End.X()*cos - c.End.Y()*sin
			ey := c.End.X()*sin + c.End.Y()*cos
			c1x := c.Ctrl1.X()*cos - c.Ctrl1.Y()*sin
			c1y := c.Ctrl1.X()*sin + c.Ctrl1.Y()*cos
			c2x := c.Ctrl2.X()*cos - c.Ctrl2.Y()*sin
			c2y := c.Ctrl2.X()*sin + c.Ctrl2.Y()*cos
			newPath.Curve4To(ex, ey, c1x, c1y, c2x, c2y)
		}
	}

	return newPath
}

// Offset creates a parallel path at the given distance
func (t *Tools) Offset(distance float64) []*Path {
	if len(t.path.commands) == 0 {
		return []*Path{}
	}

	var offsetPath *Path
	if t.IsClosed() {
		offsetPath = t.offsetClosedPath(distance)
	} else {
		offsetPath = t.offsetOpenPath(distance)
	}

	if offsetPath != nil {
		return []*Path{offsetPath}
	}
	return []*Path{}
}

func (t *Tools) offsetClosedPath(distance float64) *Path {
	newPath := NewPath()
	points := t.flatten()

	if len(points) < 3 {
		return nil
	}

	for i := 0; i < len(points); i++ {
		p0 := points[(i-1+len(points))%len(points)]
		p1 := points[i]
		p2 := points[(i+1)%len(points)]

		// Calculate outward normal
		dx1 := p1.X() - p0.X()
		dy1 := p1.Y() - p0.Y()
		dx2 := p2.X() - p1.X()
		dy2 := p2.Y() - p1.Y()

		// Calculate normals
		len1 := math.Sqrt(dx1*dx1 + dy1*dy1)
		len2 := math.Sqrt(dx2*dx2 + dy2*dy2)

		if len1 == 0 || len2 == 0 {
			continue
		}

		// For outward offset, use right-hand normal (dy, -dx)
		nx1 := dy1 / len1
		ny1 := -dx1 / len1
		nx2 := dy2 / len2
		ny2 := -dx2 / len2

		// Average the normals for corner handling
		ox := distance * (nx1 + nx2)
		oy := distance * (ny1 + ny2)

		if i == 0 {
			newPath.MoveTo(p1.X()+ox, p1.Y()+oy)
		} else {
			newPath.LineTo(p1.X()+ox, p1.Y()+oy)
		}
	}

	newPath.Close()
	return newPath
}

func (t *Tools) offsetOpenPath(distance float64) *Path {
	newPath := NewPath()
	points := t.flatten()

	if len(points) < 2 {
		return nil
	}

	// Offset first segment
	for i := 0; i < len(points); i++ {
		p := points[i]

		dx := 0.0
		dy := 0.0

		if i > 0 && i < len(points)-1 {
			// Interior point - calculate average normal
			pPrev := points[i-1]
			pNext := points[i+1]

			dx1 := p.X() - pPrev.X()
			dy1 := p.Y() - pPrev.Y()
			dx2 := pNext.X() - p.X()
			dy2 := pNext.Y() - p.Y()

			// Perpendicular direction
			nx := -dy1 + dx2
			ny := dx1 + dy2
			len := math.Sqrt(nx*nx + ny*ny)
			if len > 0 {
				dx = distance * nx / len
				dy = distance * ny / len
			}
		} else if i == 0 {
			// First point - use direction of first segment
			pNext := points[1]
			dx = pNext.Y() - p.Y()
			dy := -(pNext.X() - p.X())
			len := math.Sqrt(dx*dx + dy*dy)
			if len > 0 {
				dx = distance * dx / len
				dy = distance * dy / len
			}
		} else {
			// Last point - use direction of last segment
			pPrev := points[i-1]
			dx = p.Y() - pPrev.Y()
			dy := -(p.X() - pPrev.X())
			len := math.Sqrt(dx*dx + dy*dy)
			if len > 0 {
				dx = distance * dx / len
				dy = distance * dy / len
			}
		}

		if i == 0 {
			newPath.MoveTo(p.X()+dx, p.Y()+dy)
		} else {
			newPath.LineTo(p.X()+dx, p.Y()+dy)
		}
	}

	return newPath
}

// Clone creates a copy of the path
func (t *Tools) Clone() *Path {
	return t.path.Clone()
}
