package math

import (
	"fmt"
	"math"
)

// SanitizePolygon removes consecutive duplicate vertices and vertices that lie
// on a straight run of edges. It returns a polygon that triangulation and
// clipping can process without producing degenerate triangles.
func SanitizePolygon(polygon []Vec2, tolerance float64) []Vec2 {
	if tolerance <= 0 {
		tolerance = DefaultTolerance
	}
	cleaned := make([]Vec2, 0, len(polygon))
	for _, point := range polygon {
		if len(cleaned) > 0 && point.Distance(cleaned[len(cleaned)-1]) <= tolerance {
			continue
		}
		cleaned = append(cleaned, point)
	}
	for len(cleaned) > 1 && cleaned[0].Distance(cleaned[len(cleaned)-1]) <= tolerance {
		cleaned = cleaned[:len(cleaned)-1]
	}

	result := make([]Vec2, 0, len(cleaned))
	for i := range cleaned {
		prev := cleaned[(i+len(cleaned)-1)%len(cleaned)]
		current := cleaned[i]
		next := cleaned[(i+1)%len(cleaned)]
		if math.Abs(crossVec2(prev, current, next)) <= tolerance {
			continue
		}
		result = append(result, current)
	}
	if len(result) < 3 {
		return cleaned
	}
	return result
}

// TriangulationIssueKind classifies a problem found in a triangulation.
type TriangulationIssueKind int

const (
	// IssueInvalidIndex means an index is outside the vertex range.
	IssueInvalidIndex TriangulationIssueKind = iota
	// IssueDegenerateTriangle means a triangle has no area.
	IssueDegenerateTriangle
	// IssueInvertedTriangle means a triangle's winding is opposite the
	// polygon's winding.
	IssueInvertedTriangle
	// IssueDuplicateVertex means a triangle repeats a vertex index.
	IssueDuplicateVertex
	// IssueAreaMismatch means the triangle areas do not add up to the
	// polygon area.
	IssueAreaMismatch
)

// String returns a readable name of the issue kind.
func (k TriangulationIssueKind) String() string {
	switch k {
	case IssueInvalidIndex:
		return "invalid index"
	case IssueDegenerateTriangle:
		return "degenerate triangle"
	case IssueInvertedTriangle:
		return "inverted triangle"
	case IssueDuplicateVertex:
		return "duplicate vertex"
	case IssueAreaMismatch:
		return "area mismatch"
	}
	return "unknown issue"
}

// TriangulationIssue reports one problem in a triangulation. Triangle indexes
// the flat triple (0-based triangle number) when the issue belongs to a
// specific triangle.
type TriangulationIssue struct {
	Kind     TriangulationIssueKind
	Triangle int
	Vertices [3]int
}

// DiagnoseTriangulation checks a flat index triple triangulation of vertices
// and reports every issue it finds. An empty result means the triangulation is
// clean.
func DiagnoseTriangulation(vertices []Point2D, indices []int) []TriangulationIssue {
	issues := make([]TriangulationIssue, 0)
	if len(indices)%3 != 0 {
		for i := 0; i < len(indices); i++ {
			if indices[i] < 0 || indices[i] >= len(vertices) {
				issues = append(issues, TriangulationIssue{
					Kind:     IssueInvalidIndex,
					Triangle: i / 3,
				})
			}
		}
		return issues
	}

	polygonArea := math.Abs(signedPointArea(vertices))
	triAreaSum := 0.0
	for t := 0; t*3+2 < len(indices); t++ {
		idx := [3]int{indices[t*3], indices[t*3+1], indices[t*3+2]}
		bad := false
		for _, index := range idx {
			if index < 0 || index >= len(vertices) {
				issues = append(issues, TriangulationIssue{
					Kind:     IssueInvalidIndex,
					Triangle: t,
					Vertices: idx,
				})
				bad = true
			}
		}
		if bad {
			continue
		}
		if idx[0] == idx[1] || idx[1] == idx[2] || idx[0] == idx[2] {
			issues = append(issues, TriangulationIssue{
				Kind:     IssueDuplicateVertex,
				Triangle: t,
				Vertices: idx,
			})
			continue
		}
		signed := triangleSignedArea(vertices[idx[0]], vertices[idx[1]], vertices[idx[2]])
		triAreaSum += math.Abs(signed) / 2 // triangleSignedArea is twice the area
		if math.Abs(signed) <= DefaultTolerance {
			issues = append(issues, TriangulationIssue{
				Kind:     IssueDegenerateTriangle,
				Triangle: t,
				Vertices: idx,
			})
		} else if signed < 0 {
			issues = append(issues, TriangulationIssue{
				Kind:     IssueInvertedTriangle,
				Triangle: t,
				Vertices: idx,
			})
		}
	}

	if polygonArea > DefaultTolerance &&
		math.Abs(triAreaSum-polygonArea) > math.Max(DefaultTolerance, polygonArea*1e-6) {
		issues = append(issues, TriangulationIssue{Kind: IssueAreaMismatch})
	}
	return issues
}

// TriangulatePolygonSafe sanitizes the exterior ring and holes, triangulates
// them, and returns an error describing the first issue found. Use
// DiagnoseTriangulation for a full report.
func TriangulatePolygonSafe(exterior []Point2D, holes [][]Point2D) ([]int, error) {
	cleanExterior := sanitizePointPolygon(exterior)
	if len(cleanExterior) < 3 {
		// Degenerate exterior: a valid empty result, not an error.
		return nil, nil
	}
	if signedPointArea(cleanExterior) == 0 {
		// Zero-area (fully collinear) exterior is degenerate as well.
		return nil, nil
	}
	cleanHoles := make([][]Point2D, 0, len(holes))
	for _, hole := range holes {
		clean := sanitizePointPolygon(hole)
		if len(clean) < 3 {
			continue
		}
		cleanHoles = append(cleanHoles, clean)
	}

	// Earcut normalises windings on internal copies and returns indices in
	// the caller's vertex order (exterior followed by each hole).
	indices := Earcut(cleanExterior, cleanHoles)
	if len(indices) == 0 {
		return nil, fmt.Errorf("math: triangulation produced no triangles")
	}
	vertices := cleanExterior
	for _, hole := range cleanHoles {
		vertices = append(vertices, hole...)
	}
	issues := DiagnoseTriangulation(vertices, indices)
	for _, issue := range issues {
		if issue.Kind != IssueAreaMismatch {
			return indices, fmt.Errorf("math: triangulation %s at triangle %d %v",
				issue.Kind, issue.Triangle, issue.Vertices)
		}
	}
	return indices, nil
}

func sanitizePointPolygon(polygon []Point2D) []Point2D {
	points := make([]Vec2, len(polygon))
	for i, point := range polygon {
		points[i] = Vec2{point.X, point.Y}
	}
	cleaned := SanitizePolygon(points, DefaultTolerance)
	out := make([]Point2D, len(cleaned))
	for i, point := range cleaned {
		out[i] = Point2D{X: point[0], Y: point[1]}
	}
	return out
}

func triangleSignedArea(a, b, c Point2D) float64 {
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}

func signedPointArea(vertices []Point2D) float64 {
	if len(vertices) < 3 {
		return 0
	}
	var area float64
	for i := range vertices {
		j := (i + 1) % len(vertices)
		area += vertices[i].X*vertices[j].Y - vertices[j].X*vertices[i].Y
	}
	return area / 2
}
