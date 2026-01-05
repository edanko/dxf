package math

import (
	"math"
	"sort"
)

// Point2D represents a 2D point for triangulation
type Point2D struct {
	X, Y float64
}

// NewPoint2D creates a new 2D point
func NewPoint2D(x, y float64) Point2D {
	return Point2D{X: x, Y: y}
}

// Equal checks if two points are equal within tolerance
func (p Point2D) Equal(other Point2D, tolerance float64) bool {
	return math.Abs(p.X-other.X) < tolerance && math.Abs(p.Y-other.Y) < tolerance
}

// Sub subtracts two points
func (p Point2D) Sub(other Point2D) Point2D {
	return Point2D{X: p.X - other.X, Y: p.Y - other.Y}
}

// Add adds two points
func (p Point2D) Add(other Point2D) Point2D {
	return Point2D{X: p.X + other.X, Y: p.Y + other.Y}
}

// Mul multiplies point by scalar
func (p Point2D) Mul(scalar float64) Point2D {
	return Point2D{X: p.X * scalar, Y: p.Y * scalar}
}

// Area calculates signed area of triangle
func Area(a, b, c Point2D) float64 {
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}

// IsClockwise checks if polygon is in clockwise order
func IsClockwise(points []Point2D) bool {
	area := 0.0
	n := len(points)

	for i := 0; i < n; i++ {
		j := (i + 1) % n
		area += (points[j].X - points[i].X) * (points[j].Y + points[i].Y)
	}

	return area < 0
}

// Node represents a linked list node for earcut algorithm
type Node struct {
	Index   int
	Point   Point2D
	Prev    *Node
	Next    *Node
	Z       int
	PrevZ   *Node
	NextZ   *Node
	Steiner bool
}

// NewNode creates a new node
func NewNode(index int, point Point2D) *Node {
	return &Node{
		Index:   index,
		Point:   point,
		Z:       0,
		Steiner: false,
	}
}

// InsertNode inserts a node between two existing nodes
func InsertNode(a, b *Node, index int, point Point2D) *Node {
	node := NewNode(index, point)
	node.Prev = a
	node.Next = b
	a.Next = node
	b.Prev = node
	return node
}

// RemoveNode removes a node from the linked list
func RemoveNode(node *Node) {
	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev
}

// Triangulation represents the result of earcut triangulation
type Triangulation struct {
	Vertices []Point2D
	Indices  []int
}

// NewTriangulation creates an empty triangulation
func NewTriangulation() *Triangulation {
	return &Triangulation{
		Vertices: make([]Point2D, 0),
		Indices:  make([]int, 0),
	}
}

// Earcut implements Mapbox earcut algorithm for polygon triangulation with holes
func Earcut(exterior []Point2D, holes [][]Point2D) []int {
	points := make([]Point2D, 0, len(exterior)+len(holes)*3)

	// Add exterior points
	for _, p := range exterior {
		points = append(points, p)
	}

	// Add hole points in reverse order (clockwise)
	holesIndices := make([]int, 0, len(holes))
	for _, hole := range holes {
		holesIndices = append(holesIndices, len(points))
		for _, p := range hole {
			points = append(points, p)
		}
	}

	// Find proper hole that fits in each hole if there are nested holes
	processedHoles := make([][]int, 0)
	for i := range holes {
		processedHoles = append(processedHoles, []int{i})
	}

	processedHoles = eliminateHoles(points, processedHoles, len(exterior))

	// Filter out empty hole lists
	var finalHoles [][]int
	for _, holeGroup := range processedHoles {
		if len(holeGroup) > 0 {
			finalHoles = append(finalHoles, holeGroup)
		}
	}

	// Triangulate
	indices := earcutLinked(points, finalHoles, len(exterior))

	return indices
}

// eliminateHoles finds the proper hole that fits in each group of holes
func eliminateHoles(points []Point2D, holes [][]int, exteriorLen int) [][]int {
	if len(holes) == 0 {
		return holes
	}

	// Sort holes by x coordinate of the first point
	sort.Slice(holes, func(i, j int) bool {
		if len(holes[i]) == 0 || len(holes[j]) == 0 {
			return false
		}
		return points[holes[i][0]].X < points[holes[j][0]].X
	})

	// Process holes to eliminate nested ones
	var result [][]int
	processed := make([]bool, len(holes))

	for i := range holes {
		if processed[i] {
			continue
		}

		currentHole := holes[i]
		minDist := math.Inf(1)
		minHole := i

		// Find hole with minimum distance
		for j := range holes {
			if processed[j] || i == j {
				continue
			}

			dist := minHoleDistance(points, currentHole, holes[j])
			if dist < minDist {
				minDist = dist
				minHole = j
			}
		}

		// If no smaller hole found, this is a valid outer hole
		if minHole == i {
			result = append(result, currentHole)
			processed[i] = true
		}
	}

	return result
}

// minHoleDistance finds minimum distance between two holes
func minHoleDistance(points []Point2D, holeA, holeB []int) float64 {
	minDist := math.Inf(1)

	// Check distance from each point in holeA to each point in holeB
	for _, i := range holeA {
		for _, j := range holeB {
			dist := math.Sqrt(
				math.Pow(points[i].X-points[j].X, 2) +
					math.Pow(points[i].Y-points[j].Y, 2),
			)
			if dist < minDist {
				minDist = dist
			}
		}
	}

	return minDist
}

// earcutLinked performs the actual earcut triangulation on linked list
func earcutLinked(points []Point2D, holes [][]int, exteriorLen int) []int {
	hasHoles := len(holes) > 0

	// Create linked lists
	nodes := make([]*Node, len(points))
	outerNode := linkedList(points, 0, exteriorLen, nodes)

	if hasHoles {
		minX, minY := points[0].X, points[0].Y
		for _, p := range points {
			if p.X < minX {
				minX = p.X
			}
			if p.Y < minY {
				minY = p.Y
			}
		}

		for _, hole := range holes {
			if len(hole) == 0 {
				continue
			}
			holeNode := linkedList(points, hole[0], len(hole), nodes)

			// Find the bridge between hole and outer polygon
			bridgeNode := findHoleBridge(holeNode, outerNode, minX, minY, points)

			if bridgeNode != nil {
				splitPolygon(bridgeNode, holeNode, nodes, points)
			}
		}
	}

	// Filter out eliminated nodes
	var filteredNodes []*Node
	for _, node := range nodes {
		if node != nil {
			filteredNodes = append(filteredNodes, node)
		}
	}

	// Triangulate the resulting polygon
	triangles := filterPoints(filteredNodes, points)

	return triangles
}

// linkedList creates a circular doubly linked list from a range of nodes
func linkedList(points []Point2D, start, end int, nodes []*Node) *Node {
	var lastNode *Node

	for i := start; i < end; i++ {
		node := NewNode(i, points[i])
		nodes[i] = node

		if lastNode == nil {
			lastNode = node
		} else {
			lastNode.Next = node
			node.Prev = lastNode
		}

		lastNode = node
	}

	if lastNode != nil {
		// Make it circular
		lastNode.Next = nodes[start]
		nodes[start].Prev = lastNode
	}

	return nodes[start]
}

// findHoleBridge finds the bridge between a hole and the outer polygon
func findHoleBridge(holeNode, outerNode *Node, minX, minY float64, points []Point2D) *Node {
	var bridge *Node

	// Find the node in the outer polygon that can connect to the hole
	p := holeNode
	for {
		// Try to find a valid bridge
		bridge = findBridge(p, outerNode, minX, minY, points)
		if bridge != nil {
			break
		}

		p = p.Next

		// If we've gone all the way around the hole
		if p == holeNode {
			break
		}
	}

	return bridge
}

// findBridge finds a valid bridge between two nodes
func findBridge(leftNode, rightNode *Node, minX, minY float64, points []Point2D) *Node {
	var p *Node = leftNode

	for {
		// Find intersection of ray with polygon
		if isPointInPolygon(points, p.Point, minX, minY) {
			return p
		}

		p = p.Next

		if p == leftNode {
			break
		}
	}

	return nil
}

// isPointInPolygon checks if a point is inside a polygon using ray casting
func isPointInPolygon(points []Point2D, testPoint Point2D, minX, minY float64) bool {
	inside := false
	p1 := points[0]

	for i := 1; i <= len(points); i++ {
		var p2 Point2D
		if i == len(points) {
			p2 = points[0]
		} else {
			p2 = points[i]
		}

		if testPoint.Y > math.Min(p1.Y, p2.Y) {
			if testPoint.Y <= math.Max(p1.Y, p2.Y) {
				if testPoint.X <= math.Max(p1.X, p2.X) {
					if p1.Y != p2.Y {
						xIntersection := (testPoint.Y-p1.Y)*(p2.X-p1.X)/(p2.Y-p1.Y) + p1.X
						if p1.Y == p2.Y || testPoint.X <= xIntersection {
							inside = !inside
						}
					}
				}
			}
		}
		p1 = p2
	}

	return inside
}

// splitPolygon splits a polygon by connecting a hole to it
func splitPolygon(bridgeNode, holeNode *Node, nodes []*Node, points []Point2D) {
	// Find the nodes to split
	bridgeNext := bridgeNode.Next

	// Create new nodes for the split
	bridgeIndex := len(nodes)
	holeIndex := bridgeIndex + 1

	bridgeNew := NewNode(bridgeIndex, bridgeNode.Point)
	holeNew := NewNode(holeIndex, holeNode.Point)

	nodes = append(nodes, bridgeNew, holeNew)

	// Reconnect the linked lists
	bridgeNode.Next = holeNew
	holeNew.Prev = bridgeNode

	holeNode.Next = bridgeNext
	bridgeNext.Prev = holeNode

	holeNew.Next = bridgeNew
	bridgeNew.Prev = holeNew
}

// filterPoints performs ear clipping and returns triangle indices
func filterPoints(nodes []*Node, points []Point2D) []int {
	var triangles []int

	// While there are at least 3 nodes left
	for len(nodes) >= 3 {
		earNode := findEar(nodes, points)

		if earNode == nil {
			// No ear found, something went wrong
			break
		}

		// Add triangle indices
		triangles = append(triangles, earNode.Prev.Index, earNode.Index, earNode.Next.Index)

		// Remove the ear
		RemoveNode(earNode)

		// Rebuild nodes array without the removed node
		var newNodes []*Node
		current := earNode.Next
		for current != earNode {
			newNodes = append(newNodes, current)
			current = current.Next
		}
		nodes = newNodes
	}

	return triangles
}

// findEar finds an ear in the polygon
func findEar(nodes []*Node, points []Point2D) *Node {
	for _, node := range nodes {
		if isEar(nodes, node, points) {
			return node
		}
	}
	return nil
}

// isEar checks if a node forms an ear
func isEar(nodes []*Node, node *Node, points []Point2D) bool {
	a := node.Prev.Point
	b := node.Point
	c := node.Next.Point

	// Check if triangle is clockwise
	if Area(a, b, c) >= 0 {
		return false
	}

	// Check if any other point is inside the triangle
	for _, otherNode := range nodes {
		if otherNode == node || otherNode == node.Prev || otherNode == node.Next {
			continue
		}

		if isPointInTriangle(otherNode.Point, a, b, c) {
			return false
		}
	}

	return true
}

// isPointInTriangle checks if a point is inside a triangle
func isPointInTriangle(p, a, b, c Point2D) bool {
	// Check if point is on same side of all edges
	sign1 := signArea(p, a, b)
	sign2 := signArea(p, b, c)
	sign3 := signArea(p, c, a)

	hasNeg := (sign1 < 0) || (sign2 < 0) || (sign3 < 0)
	hasPos := (sign1 > 0) || (sign2 > 0) || (sign3 > 0)

	return !(hasNeg && hasPos)
}

// signArea helps determine which side of a line a point is on
func signArea(p, a, b Point2D) float64 {
	return (p.X-b.X)*(a.Y-b.Y) - (a.X-b.X)*(p.Y-b.Y)
}

// TriangulatePolygon triangulates a simple polygon (no holes)
func TriangulatePolygon(polygon []Point2D) []int {
	if len(polygon) < 3 {
		return []int{}
	}

	// Ensure polygon is in counter-clockwise order
	if IsClockwise(polygon) {
		// Reverse the polygon
		for i, j := 0, len(polygon)-1; i < j; i, j = i+1, j-1 {
			polygon[i], polygon[j] = polygon[j], polygon[i]
		}
	}

	return Earcut(polygon, nil)
}

// TriangulatePolygonWithHoles triangulates a polygon with holes
func TriangulatePolygonWithHoles(exterior []Point2D, holes [][]Point2D) []int {
	if len(exterior) < 3 {
		return []int{}
	}

	// Ensure exterior is in counter-clockwise order
	if IsClockwise(exterior) {
		// Reverse the exterior polygon
		for i, j := 0, len(exterior)-1; i < j; i, j = i+1, j-1 {
			exterior[i], exterior[j] = exterior[j], exterior[i]
		}
	}

	// Ensure holes are in clockwise order
	for i, hole := range holes {
		if !IsClockwise(hole) {
			// Reverse the hole polygon
			for j, k := 0, len(hole)-1; j < k; j, k = j+1, k-1 {
				hole[j], hole[k] = hole[k], hole[j]
			}
			holes[i] = hole
		}
	}

	return Earcut(exterior, holes)
}

// TriangulateConvexPolygon triangulates a convex polygon efficiently
func TriangulateConvexPolygon(polygon []Point2D) []int {
	n := len(polygon)
	if n < 3 {
		return []int{}
	}

	var indices []int

	// Fan triangulation from first vertex
	for i := 1; i < n-1; i++ {
		indices = append(indices, 0, i, i+1)
	}

	return indices
}

// ValidateTriangulation checks if a triangulation is valid
func ValidateTriangulation(vertices []Point2D, indices []int) bool {
	// Check that all indices are valid
	for _, index := range indices {
		if index < 0 || index >= len(vertices) {
			return false
		}
	}

	// Check that triangles are formed correctly
	if len(indices)%3 != 0 {
		return false
	}

	// Check that triangles don't overlap (simplified check)
	for i := 0; i < len(indices); i += 3 {
		tri1 := [3]Point2D{vertices[indices[i]], vertices[indices[i+1]], vertices[indices[i+2]]}

		for j := i + 3; j < len(indices); j += 3 {
			tri2 := [3]Point2D{vertices[indices[j]], vertices[indices[j+1]], vertices[indices[j+2]]}

			if trianglesOverlap(tri1, tri2) {
				return false
			}
		}
	}

	return true
}

// trianglesOverlap checks if two triangles overlap (simplified)
func trianglesOverlap(tri1, tri2 [3]Point2D) bool {
	// This is a simplified check - full implementation would be more complex
	// For now, just check if they share any vertices
	for _, v1 := range tri1 {
		for _, v2 := range tri2 {
			if v1.Equal(v2, 1e-6) {
				return true // They share a vertex, not an overlap but adjacent
			}
		}
	}

	return false
}
