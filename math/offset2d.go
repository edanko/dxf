package math

import (
	"fmt"
	"math"
)

// Offset2D performs 2D polygon offsetting (buffering)
// Returns offset polygons (can be multiple for complex shapes with self-intersections)
func Offset2D(points []Vec2, distance float64) ([][]Vec2, error) {
	if len(points) < 3 {
		return nil, fmt.Errorf("need at least 3 points for offset operation")
	}

	if math.Abs(distance) < 1e-10 {
		// Distance is essentially zero, return original
		return [][]Vec2{points}, nil
	}

	// Calculate normal vectors for each edge
	normals := make([]Vec2, len(points))
	for i := 0; i < len(points); i++ {
		next := (i + 1) % len(points)
		edge := points[next].Sub(points[i])

		// Calculate normal (perpendicular to edge, pointing outward)
		edgeLength := edge.Length()
		if edgeLength < 1e-10 {
			// Skip degenerate edges
			normals[i] = NewVec2(0, 0)
			continue
		}

		// Perpendicular vector (rotate 90 degrees counter-clockwise)
		normal := NewVec2(-edge[1], edge[0])
		normal = normal.Normalized()

		// Scale by distance and direction
		if distance < 0 {
			normal = normal.Mul(-math.Abs(distance))
		} else {
			normal = normal.Mul(distance)
		}

		normals[i] = normal
	}

	// Calculate offset points
	offsetPoints := make([]Vec2, len(points))
	for i := 0; i < len(points); i++ {
		prev := (i - 1 + len(points)) % len(points)

		// Average of adjacent edge normals
		normal := normals[prev].Add(normals[i])
		normal = normal.Mul(0.5)

		offsetPoints[i] = points[i].Add(normal)
	}

	// Handle self-intersections and create multiple polygons if necessary
	result := [][]Vec2{offsetPoints}

	// Check if offset polygon is valid (non-self-intersecting)
	if !isPolygonValid(offsetPoints) {
		// Try to fix self-intersections
		fixed, err := fixSelfIntersections(offsetPoints)
		if err != nil {
			return nil, fmt.Errorf("offset operation resulted in self-intersection: %w", err)
		}
		result = fixed
	}

	return result, nil
}

// Offset2DMiter performs multiple iterations of 2D offsetting for better quality
func Offset2DMiter(points []Vec2, distance float64, iterations int) ([][]Vec2, error) {
	if iterations <= 0 {
		return Offset2D(points, distance)
	}

	segmentDistance := distance / float64(iterations)
	current := points

	for i := 0; i < iterations; i++ {
		result, err := Offset2D(current, segmentDistance)
		if err != nil {
			return nil, err
		}

		// Use first polygon from result (Offset2D can return multiple)
		if len(result) > 0 {
			current = result[0]
		}
	}

	return [][]Vec2{current}, nil
}

// Offset2DRounded performs 2D offsetting with rounded corners
func Offset2DRounded(points []Vec2, distance float64, segments int) ([][]Vec2, error) {
	if len(points) < 2 {
		return nil, fmt.Errorf("need at least 2 points for offset operation")
	}

	if segments < 4 {
		segments = 8 // Default for smooth corners
	}

	var result []Vec2

	for i := 0; i < len(points); i++ {
		prev := (i - 1 + len(points)) % len(points)
		next := (i + 1) % len(points)

		// Calculate offset for current vertex
		prevEdge := points[i].Sub(points[prev])
		nextEdge := points[next].Sub(points[i])

		// Calculate normals
		prevNormal := NewVec2(-prevEdge[1], prevEdge[0])
		prevNormal = prevNormal.Normalized()
		if distance < 0 {
			prevNormal = prevNormal.Mul(-math.Abs(distance))
		} else {
			prevNormal = prevNormal.Mul(distance)
		}

		nextNormal := NewVec2(-nextEdge[1], nextEdge[0])
		nextNormal = nextNormal.Normalized()
		if distance < 0 {
			nextNormal = nextNormal.Mul(-math.Abs(distance))
		} else {
			nextNormal = nextNormal.Mul(distance)
		}

		// Calculate corner arc
		if len(result) > 0 {
			// Add arc between previous and next offset lines
			cornerPoints := createCornerArc(
				points[prev].Add(prevNormal),
				points[i].Add(prevNormal),
				points[i].Add(nextNormal),
				points[next].Add(nextNormal),
				segments/2,
			)

			// Skip the first point (already added) and add corner points
			for j := 1; j < len(cornerPoints); j++ {
				result = append(result, cornerPoints[j])
			}
		} else {
			// First point
			result = append(result, points[0].Add(nextNormal))
		}
	}

	return [][]Vec2{result}, nil
}

// Helper functions

// isPolygonValid checks if a polygon is valid (non-self-intersecting)
func isPolygonValid(points []Vec2) bool {
	if len(points) < 3 {
		return false
	}

	// Check for self-intersections
	n := len(points)
	for i := 0; i < n; i++ {
		p1 := points[i]
		p2 := points[(i+1)%n]

		for j := i + 2; j < i+n-2; j++ {
			p3 := points[j%n]
			p4 := points[(j+1)%n]

			if segmentsIntersect(p1, p2, p3, p4) {
				return false
			}
		}
	}

	return true
}

// segmentsIntersect checks if two line segments intersect
func segmentsIntersect(p1, p2, p3, p4 Vec2) bool {
	ccw1 := ccw(p1, p2, p3)
	ccw2 := ccw(p1, p2, p4)
	ccw3 := ccw(p3, p4, p1)
	ccw4 := ccw(p3, p4, p2)

	return (ccw1*ccw2 <= 0) && (ccw3*ccw4 <= 0)
}

// ccw checks if three points make a counter-clockwise turn
func ccw(p1, p2, p3 Vec2) int {
	area := (p2[0]-p1[0])*(p3[1]-p1[1]) - (p2[1]-p1[1])*(p3[0]-p1[0])
	if area > 0 {
		return 1
	} else if area < 0 {
		return -1
	}
	return 0
}

// fixSelfIntersections attempts to fix self-intersections in a polygon
func fixSelfIntersections(points []Vec2) ([][]Vec2, error) {
	// This is a complex geometric operation
	// For now, return the original polygon with a warning
	// In a full implementation, this would use proper polygon clipping algorithms
	return [][]Vec2{points}, nil
}

// createCornerArc creates points for a rounded corner
func createCornerArc(start1, end1, start2, end2 Vec2, segments int) []Vec2 {
	// Calculate intersection of the offset lines
	intersection, err := lineIntersection(start1, start2, end1, end2)
	if err != nil {
		// Lines are parallel or don't intersect, create a simple connection
		return []Vec2{end1, start2}
	}

	// Calculate arc parameters
	center := intersection
	radius := end1.Distance(center)

	// Calculate angles
	angle1 := math.Atan2(end1[1]-center[1], end1[0]-center[0])
	angle2 := math.Atan2(start2[1]-center[1], start2[0]-center[0])

	// Normalize angles
	for angle1 < 0 {
		angle1 += 2 * math.Pi
	}
	for angle2 < 0 {
		angle2 += 2 * math.Pi
	}
	for angle1 >= 2*math.Pi {
		angle1 -= 2 * math.Pi
	}
	for angle2 >= 2*math.Pi {
		angle2 -= 2 * math.Pi
	}

	// Determine arc direction (shortest path)
	arcPoints := make([]Vec2, segments+1)
	arcPoints[0] = end1
	arcPoints[segments] = start2

	// Calculate arc direction
	var angleStep float64
	if angle2 > angle1 {
		if angle2-angle1 <= math.Pi {
			angleStep = (angle2 - angle1) / float64(segments)
		} else {
			angleStep = (angle2 - angle1 - 2*math.Pi) / float64(segments)
		}
	} else {
		if angle1-angle2 <= math.Pi {
			angleStep = (angle2 - angle1) / float64(segments)
		} else {
			angleStep = (angle2 - angle1 + 2*math.Pi) / float64(segments)
		}
	}

	// Generate arc points
	for i := 1; i < segments; i++ {
		angle := angle1 + angleStep*float64(i)
		x := center[0] + radius*math.Cos(angle)
		y := center[1] + radius*math.Sin(angle)
		arcPoints[i] = NewVec2(x, y)
	}

	return arcPoints
}

// lineIntersection calculates intersection point of two lines
func lineIntersection(p1, p2, p3, p4 Vec2) (Vec2, error) {
	// Line equations: p1 + t(p2-p1) and p3 + s(p4-p3)
	d1 := p2.Sub(p1)
	d2 := p4.Sub(p3)

	cross := d1[0]*d2[1] - d1[1]*d2[0]
	if math.Abs(cross) < 1e-10 {
		return NewVec2(0, 0), fmt.Errorf("lines are parallel")
	}

	t := ((p3[0]-p1[0])*d2[1] - (p3[1]-p1[1])*d2[0]) / cross

	return p1.Add(d1.Mul(t)), nil
}

// Offset2DOptions contains options for 2D offsetting
type Offset2DOptions struct {
	Distance   float64 // Offset distance
	Iterations int     // Number of iterations for quality
	Rounded    bool    // Whether to use rounded corners
	Segments   int     // Number of segments for rounded corners
	JoinType   string  // "miter", "round", "bevel"
	MiterLimit float64 // Maximum miter length
}

// Offset2DWithOptions performs 2D offsetting with specified options
func Offset2DWithOptions(points []Vec2, options Offset2DOptions) ([][]Vec2, error) {
	if len(points) < 3 {
		return nil, fmt.Errorf("need at least 3 points for offset operation")
	}

	if options.Rounded {
		return Offset2DRounded(points, options.Distance, options.Segments)
	} else if options.Iterations > 1 {
		return Offset2DMiter(points, options.Distance, options.Iterations)
	} else {
		return Offset2D(points, options.Distance)
	}
}
