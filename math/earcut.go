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

	for i := range n {
		j := (i + 1) % n
		area += (points[j].X - points[i].X) * (points[j].Y + points[i].Y)
	}

	// The shoelace sum Σ(xj−xi)(yj+yi) equals −2A where A is the signed
	// (counter-clockwise positive) area, so positive means clockwise.
	return area > 0
}

// Node represents a linked list node for triangulation
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

// Earcut triangulates a simple polygon with holes using ear clipping with
// needle bridges (the mapbox earcut approach: each hole is merged into the
// exterior ring through a visibility bridge, then convex ears are clipped
// from the resulting simple ring).
//
// The returned indices refer to the concatenation of exterior followed by
// each hole in order. Exterior rings are triangulated counter-clockwise;
// windings are normalised on the internal copies (input slices are not
// modified). Degenerate inputs produce an empty result.
//
// Hole bridges target visible exterior vertices, closest first. The
// resulting triangle areas are checked against the exterior-minus-holes
// area; when the closest bridge choice falls short (for example a centered
// hole whose nearest exterior vertex produces a pinched merged ring), the
// merge is retried with farther visible bridge targets until the area
// matches or the deterministic candidate budget is exhausted.
func Earcut(exterior []Point2D, holes [][]Point2D) []int {
	if len(exterior) < 3 {
		return []int{}
	}

	// Build the concatenated vertex list and per-ring index lists.
	vertices := make([]Point2D, 0, len(exterior))
	vertices = append(vertices, exterior...)
	rings := make([][]int, 0, 1+len(holes))
	extRing := make([]int, 0, len(exterior))
	for i := range exterior {
		extRing = append(extRing, i)
	}
	rings = append(rings, extRing)
	for _, hole := range holes {
		base := len(vertices)
		holeRing := make([]int, 0, len(hole))
		for i, p := range hole {
			vertices = append(vertices, p)
			holeRing = append(holeRing, base+i)
		}
		if len(holeRing) >= 3 {
			// Normalise hole windings on the internal copy to clockwise:
			// the needle-bridge merge below traverses a hole from the bridge
			// entry along its Next chain, and the triangulated region must
			// lie to the left of that traversal (i.e. outside the hole).
			if ringArea(vertices, holeRing) > 0 {
				reverseInts(holeRing)
			}
			rings = append(rings, holeRing)
		}
	}

	exteriorRing := rings[0]
	if ringArea(vertices, exteriorRing) < 0 {
		reverseInts(exteriorRing)
	}
	outer := ringList(vertices, exteriorRing)
	if outer == nil {
		return []int{}
	}
	exteriorPoly := ringPoints(outer)

	// Original hole polygons for visibility checks and ear-clipping guards.
	holePolys := make([][]Point2D, 0, len(holes))
	for _, hole := range rings[1:] {
		poly := make([]Point2D, 0, len(hole))
		for _, idx := range hole {
			poly = append(poly, vertices[idx])
		}
		holePolys = append(holePolys, poly)
	}

	// Expected region area: exterior minus holes (the hole windings were
	// normalised on the internal index rings above).
	expectedArea := math.Abs(ringArea(vertices, exteriorRing)) / 2
	for i := 1; i < len(rings); i++ {
		expectedArea -= math.Abs(ringArea(vertices, rings[i])) / 2
	}
	if expectedArea < 0 {
		expectedArea = 0
	}

	if len(rings) == 1 {
		return earClip(outer, holePolys)
	}

	// Candidate bridge targets per hole: visible exterior vertices ordered
	// by distance from the hole's leftmost vertex (deterministic ties).
	candidates := make([][]Point2D, len(rings)-1)
	for i := 1; i < len(rings); i++ {
		holeRing := ringList(vertices, rings[i])
		if holeRing == nil {
			continue
		}
		candidates[i-1] = bridgeCandidates(holeRing, exteriorPoly, holePolys)
	}

	attempts := 1
	for _, list := range candidates {
		if len(list) == 0 {
			continue
		}
		attempts *= len(list)
		if attempts > maxBridgeAttempts {
			attempts = maxBridgeAttempts
			break
		}
	}

	var best []int
	bestDiff := math.Inf(1)
	for attempt := 0; attempt < attempts; attempt++ {
		ranks := bridgeRanks(attempt, candidates)
		outer = ringList(vertices, exteriorRing)
		ok := true
		for i := 1; i < len(rings); i++ {
			holeRing := ringList(vertices, rings[i])
			if holeRing == nil {
				continue
			}
			holeNode := leftmostNode(holeRing)
			var target *Node
			if len(candidates[i-1]) > 0 {
				target = findNodeByPoint(outer, candidates[i-1][ranks[i-1]])
			}
			if target == nil {
				target = findBridgeTarget(outer, holeNode, holePolys)
			}
			if target == nil {
				ok = false
				break
			}
			outer = splitPolygon(target, holeNode)
		}
		if !ok {
			continue
		}
		triangles := earClip(outer, holePolys)
		diff := math.Abs(triangulationArea(vertices, triangles) - expectedArea)
		if diff <= areaTolerance(expectedArea) {
			return triangles
		}
		if len(best) == 0 || diff < bestDiff {
			best = triangles
			bestDiff = diff
		}
	}
	return best
}

// maxBridgeAttempts bounds the deterministic bridge-target search when
// several holes each have many visible exterior vertices.
const maxBridgeAttempts = 64

// bridgeCandidates returns the visible exterior vertices for a hole's
// leftmost node, ordered by distance (then X, then Y for deterministic
// ties). The closest candidate is tried first; farther candidates are
// fallbacks when the merged ring fails the area check.
func bridgeCandidates(hole *Node, exteriorPoly []Point2D, holePolys [][]Point2D) []Point2D {
	start := leftmostNode(hole).Point
	candidates := make([]Point2D, 0, len(exteriorPoly))
	for _, p := range exteriorPoly {
		if p == start || !visible(start, p, exteriorPoly, holePolys) {
			continue
		}
		candidates = append(candidates, p)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		di := dist2(start, candidates[i])
		dj := dist2(start, candidates[j])
		if di != dj {
			return di < dj
		}
		if candidates[i].X != candidates[j].X {
			return candidates[i].X < candidates[j].X
		}
		return candidates[i].Y < candidates[j].Y
	})
	return candidates
}

// bridgeRanks decodes an attempt counter as mixed-radix candidate ranks,
// one rank per hole.
func bridgeRanks(attempt int, candidates [][]Point2D) []int {
	ranks := make([]int, len(candidates))
	for i, list := range candidates {
		if len(list) == 0 {
			continue
		}
		ranks[i] = attempt % len(list)
		attempt /= len(list)
	}
	return ranks
}

// findNodeByPoint returns the first ring node at p in traversal order.
func findNodeByPoint(head *Node, p Point2D) *Node {
	if head == nil {
		return nil
	}
	for cur := head; ; cur = cur.Next {
		if cur.Point == p {
			return cur
		}
		if cur.Next == head {
			return nil
		}
	}
}

// triangulationArea returns the sum of absolute triangle areas.
func triangulationArea(vertices []Point2D, indices []int) float64 {
	total := 0.0
	for i := 0; i+2 < len(indices); i += 3 {
		total += math.Abs(Area(vertices[indices[i]], vertices[indices[i+1]], vertices[indices[i+2]])) / 2
	}
	return total
}

// areaTolerance is the documented tolerance for comparing a triangulated
// area with the polygon-with-holes area.
func areaTolerance(area float64) float64 {
	return math.Max(DefaultTolerance, area*1e-6)
}

// ringArea returns the doubled signed area of an index ring (positive for
// counter-clockwise).
func ringArea(vertices []Point2D, ring []int) float64 {
	area := 0.0
	for i, idx := range ring {
		j := ring[(i+1)%len(ring)]
		area += vertices[idx].X*vertices[j].Y - vertices[j].X*vertices[idx].Y
	}
	return area
}

func reverseInts(ring []int) {
	for i, j := 0, len(ring)-1; i < j; i, j = i+1, j-1 {
		ring[i], ring[j] = ring[j], ring[i]
	}
}

// ringList builds a circular doubly linked list from an index ring,
// dropping consecutive duplicate points and collinear middles.
func ringList(vertices []Point2D, ring []int) *Node {
	var head, tail *Node
	for _, idx := range ring {
		p := vertices[idx]
		if tail != nil && tail.Point == p {
			continue
		}
		node := NewNode(idx, p)
		if head == nil {
			head = node
			node.Prev = node
			node.Next = node
		} else {
			node.Prev = tail
			node.Next = head
			tail.Next = node
			head.Prev = node
		}
		tail = node
	}
	if head == nil || head.Next == head {
		return nil
	}
	for {
		if nodeCount(head) < 3 {
			return nil
		}
		changed := false
		cur := head
		for i := 0; i < nodeCount(head); i++ {
			next, prev := cur.Next, cur.Prev
			if next == prev && cur.Point == prev.Point {
				RemoveNode(cur)
				head = prev
				changed = true
				cur = prev
				continue
			}
			if cur.Point != prev.Point && cur.Point != next.Point && Area(prev.Point, cur.Point, next.Point) == 0 {
				RemoveNode(cur)
				head = next
				changed = true
				cur = next
				continue
			}
			cur = next
		}
		if !changed {
			return head
		}
	}
}

func nodeCount(node *Node) int {
	if node == nil {
		return 0
	}
	count := 0
	for cur := node; ; cur = cur.Next {
		count++
		if cur.Next == node {
			return count
		}
	}
}

// reverseRingList reverses a circular linked list in place and returns the
// new head (the old tail).
func reverseRingList(head *Node) *Node {
	if head == nil {
		return nil
	}
	var tail *Node
	cur := head
	for {
		cur.Prev, cur.Next = cur.Next, cur.Prev
		tail = cur
		if cur.Next == head {
			break
		}
		cur = cur.Next
	}
	return tail
}

// bridgeHoleIntoRing merges the hole ring into the outer ring with a
// visibility bridge. holesPolys are the original hole polygons (including
// hole itself) used for visibility checks.
func bridgeHoleIntoRing(outer, hole *Node, holesPolys [][]Point2D) *Node {
	holeNode := leftmostNode(hole)
	if holeNode == nil || outer == nil {
		return outer
	}
	bridge := findBridgeTarget(outer, holeNode, holesPolys)
	if bridge == nil {
		// No visible vertex: leave the hole unmerged (degenerate input).
		return outer
	}
	return splitPolygon(bridge, holeNode)
}

func leftmostNode(head *Node) *Node {
	best := head
	for cur := head.Next; cur != head; cur = cur.Next {
		if cur.Point.X < best.Point.X || (cur.Point.X == best.Point.X && cur.Point.Y < best.Point.Y) {
			best = cur
		}
	}
	return best
}

// findBridgeTarget returns the outer ring vertex to bridge the hole's
// leftmost node to. The selection follows the mapbox earcut findHoleBridge
// rule: candidates lie at or left of the hole's leftmost vertex and must be
// visible; the winner maximizes the slope ratio (q.y - p.y) / (q.x - p.x)
// (horizontal bridges score highest, targets below-left beat above-right),
// which keeps the splitPolygon connector segments (hole-chain tail ->
// bridge start and bridge end -> outer continuation) clear of the hole.
// Ties break toward the closer vertex. Falls back to the closest visible
// vertex when no left candidate exists.
func findBridgeTarget(outer, holeNode *Node, holesPolys [][]Point2D) *Node {
	outerPoly := ringPoints(outer)
	q := holeNode.Point

	var best *Node
	bestTx := math.Inf(-1)
	bestDist := math.Inf(1)
	for cur := outer; ; cur = cur.Next {
		if cur.Point.X <= q.X && visible(q, cur.Point, outerPoly, holesPolys) {
			var tx float64
			switch {
			case cur.Point.Y == q.Y:
				d := cur.Point.Y - q.Y
				tx = 1 / (d*d + 1e-5)
			case cur.Point.X == q.X:
				tx = 1e10
			default:
				tx = (q.Y - cur.Point.Y) / (q.X - cur.Point.X)
			}
			d := dist2(q, cur.Point)
			if tx > bestTx || (tx == bestTx && d < bestDist) {
				bestTx = tx
				bestDist = d
				best = cur
			}
		}
		if cur.Next == outer {
			break
		}
	}
	if best != nil {
		return best
	}

	// Fallback: closest visible vertex (holes flush against the left edge).
	for cur := outer; ; cur = cur.Next {
		if visible(q, cur.Point, outerPoly, holesPolys) {
			d := dist2(q, cur.Point)
			if d < bestDist {
				bestDist = d
				best = cur
			}
		}
		if cur.Next == outer {
			break
		}
	}
	return best
}

func ringPoints(head *Node) []Point2D {
	pts := make([]Point2D, 0, nodeCount(head))
	for cur := head; ; cur = cur.Next {
		pts = append(pts, cur.Point)
		if cur.Next == head {
			return pts
		}
	}
}

func dist2(a, b Point2D) float64 {
	return (a.X-b.X)*(a.X-b.X) + (a.Y-b.Y)*(a.Y-b.Y)
}

// visible reports whether the segment a-b is a valid bridge: it properly
// crosses no edge of the outer polygon or any hole, and its midpoint lies
// strictly inside the outer polygon and strictly outside every hole.
func visible(a, b Point2D, outerPoly []Point2D, holesPolys [][]Point2D) bool {
	if a == b {
		return false
	}
	if !edgesClear(a, b, outerPoly) {
		return false
	}
	for _, hole := range holesPolys {
		if !edgesClear(a, b, hole) {
			return false
		}
	}
	mid := Point2D{X: (a.X + b.X) / 2, Y: (a.Y + b.Y) / 2}
	if !pointInPolygonStrict(mid, outerPoly) {
		return false
	}
	for _, hole := range holesPolys {
		if pointInPolygonOnBoundary(mid, hole) {
			return false
		}
	}
	return true
}

// edgesClear reports whether segment a-b properly crosses (or collinearly
// overlaps) any edge of poly, ignoring edges that share an endpoint with a
// or b (touching is allowed for bridges).
func edgesClear(a, b Point2D, poly []Point2D) bool {
	n := len(poly)
	for i := range poly {
		p := poly[i]
		q := poly[(i+1)%n]
		if p == a || q == a || p == b || q == b {
			continue
		}
		if bridgeSegmentsIntersect(a, b, p, q) {
			return false
		}
	}
	return true
}

// bridgeSegmentsIntersect reports proper crossings and collinear overlaps.
func bridgeSegmentsIntersect(a1, a2, b1, b2 Point2D) bool {
	o1 := Area(a1, a2, b1)
	o2 := Area(a1, a2, b2)
	o3 := Area(b1, b2, a1)
	o4 := Area(b1, b2, a2)
	if (o1 > 0) != (o2 > 0) && (o3 > 0) != (o4 > 0) && o1 != 0 && o2 != 0 && o3 != 0 && o4 != 0 {
		return true
	}
	// Collinear cases: point on segment (inclusive).
	if o1 == 0 && onSegment(a1, a2, b1) {
		return true
	}
	if o2 == 0 && onSegment(a1, a2, b2) {
		return true
	}
	if o3 == 0 && onSegment(b1, b2, a1) {
		return true
	}
	if o4 == 0 && onSegment(b1, b2, a2) {
		return true
	}
	return false
}

func onSegment(a, b, p Point2D) bool {
	return p.X >= math.Min(a.X, b.X) && p.X <= math.Max(a.X, b.X) &&
		p.Y >= math.Min(a.Y, b.Y) && p.Y <= math.Max(a.Y, b.Y)
}

// pointInPolygonStrict uses ray casting; points on the boundary (edges or
// vertices) count as outside — strict containment is required for bridge
// midpoints and for the ear-clipping hole-overlap predicates.
func pointInPolygonStrict(p Point2D, poly []Point2D) bool {
	n := len(poly)
	for i, j := 0, n-1; i < n; j, i = i, i+1 {
		if Area(poly[j], poly[i], p) == 0 && onSegment(poly[j], poly[i], p) {
			return false
		}
	}
	inside := false
	for i, j := 0, n-1; i < n; j, i = i, i+1 {
		a, b := poly[j], poly[i]
		if (a.Y > p.Y) != (b.Y > p.Y) {
			xInt := a.X + (p.Y-a.Y)/(b.Y-a.Y)*(b.X-a.X)
			if xInt > p.X {
				inside = !inside
			}
		}
	}
	return inside
}

// pointInPolygonOnBoundary treats boundary points as inside.
func pointInPolygonOnBoundary(p Point2D, poly []Point2D) bool {
	n := len(poly)
	for i, j := 0, n-1; i < n; j, i = i, i+1 {
		a, b := poly[j], poly[i]
		if Area(a, b, p) == 0 && onSegment(a, b, p) {
			return true
		}
	}
	return pointInPolygonStrict(p, poly)
}

// splitPolygon splices nodes a and b of one circular list into a single
// list traversed a→b→…→bPrev→b2→a2→aNext→…→a (the mapbox splitPolygon with
// duplicated bridge nodes). The copies b2 and a2 reproduce b and a so the
// connector b2→a2 is the reverse of the bridge a→b and the bridge is a
// zero-width needle instead of a real notch between bPrev and a.
func splitPolygon(a, b *Node) *Node {
	a2 := NewNode(a.Index, a.Point)
	b2 := NewNode(b.Index, b.Point)
	aNext := a.Next
	bPrev := b.Prev

	a.Next = b
	b.Prev = a

	a2.Next = aNext
	aNext.Prev = a2

	b2.Next = a2
	a2.Prev = b2

	bPrev.Next = b2
	b2.Prev = bPrev

	return a
}

// earClip clips convex ears from the simple ring until fewer than three
// nodes remain. The mapbox fallback (split along the diagonal of the
// smallest convex vertex) guarantees progress when no ear exists. holes
// are the original hole polygons: an ear must never swallow one.
func earClip(ring *Node, holes [][]Point2D) []int {
	triangles := make([]int, 0)
	if ring == nil {
		return triangles
	}
	n := nodeCount(ring)
	guard := 0
	limit := 8*n + 16
	for n >= 3 && guard < limit {
		guard++
		earFound := false
		cur := ring
		for i := 0; i < n; i++ {
			if isEar(cur, holes) {
				triangles = append(triangles, cur.Index, cur.Next.Index, cur.Prev.Index)
				next := cur.Next
				RemoveNode(cur)
				ring = next
				n--
				earFound = true
				break
			}
			cur = cur.Next
		}
		if !earFound {
			split := smallestConvex(ring, n)
			if split == nil {
				break
			}
			ring = splitPolygon(split, split.Next.Next)
			n += 2
		}
	}
	return triangles
}

// isEar reports whether node is a convex corner whose triangle contains no
// other ring vertex strictly inside. When holes are provided, the triangle
// must also not overlap any hole interior (holes are the original hole
// polygons; the merged needle ring contains hole boundary edges as needle
// edges, so the strict predicates of holeOverlapsTriangle are required —
// the inclusive bridge intersection tests would reject every ear).
func isEar(node *Node, holes [][]Point2D) bool {
	prev, next := node.Prev, node.Next
	if prev == node || next == node || prev == next {
		return false
	}
	if Area(prev.Point, node.Point, next.Point) <= 0 {
		return false
	}
	// Scan every other ring node (mapbox pointsInside), skipping the three
	// triangle corners by pointer; duplicated bridge points share the same
	// coordinates and are filtered as non-strict (on-edge) cases.
	for p := node.Next.Next; ; p = p.Next {
		if p != node && p != prev && p != next &&
			pointInTriangleStrict(p.Point, prev.Point, node.Point, next.Point) {
			return false
		}
		if p == node {
			break
		}
	}
	if len(holes) > 0 {
		tri := [3]Point2D{prev.Point, node.Point, next.Point}
		for _, hole := range holes {
			if holeOverlapsTriangle(hole, tri) {
				return false
			}
		}
	}
	return true
}

// holeOverlapsTriangle reports whether the interior of the hole polygon
// strictly overlaps the interior of the CCW triangle. Only strict
// predicates are used — shared bridge needles and tangent boundary touches
// must never reject an ear. Detection: a hole vertex strictly inside the
// triangle, the triangle centroid strictly inside the hole, the hole
// centroid (when inside the hole) strictly inside the triangle, a triangle
// vertex strictly inside the hole, an edge midpoint of one polygon
// strictly inside the other (catches tangential overlap where every vertex
// sits on the other's boundary), or a proper edge crossing.
func holeOverlapsTriangle(hole []Point2D, tri [3]Point2D) bool {
	a, b, c := tri[0], tri[1], tri[2]
	centroid := Point2D{X: (a.X + b.X + c.X) / 3, Y: (a.Y + b.Y + c.Y) / 3}
	if pointInPolygonStrict(centroid, hole) {
		return true
	}
	var hx, hy float64
	for _, p := range hole {
		hx += p.X
		hy += p.Y
		if pointInTriangleStrict(p, a, b, c) {
			return true
		}
	}
	holeCentroid := Point2D{X: hx / float64(len(hole)), Y: hy / float64(len(hole))}
	if pointInPolygonStrict(holeCentroid, hole) && pointInTriangleStrict(holeCentroid, a, b, c) {
		return true
	}
	for _, p := range tri {
		if pointInPolygonStrict(p, hole) {
			return true
		}
	}
	for i := range hole {
		p, q := hole[i], hole[(i+1)%len(hole)]
		mid := Point2D{X: (p.X + q.X) / 2, Y: (p.Y + q.Y) / 2}
		if pointInTriangleStrict(mid, a, b, c) {
			return true
		}
		for j := range tri {
			if segmentsProperlyCross(p, q, tri[j], tri[(j+1)%3]) {
				return true
			}
		}
	}
	for j := range tri {
		mid := Point2D{X: (tri[j].X + tri[(j+1)%3].X) / 2, Y: (tri[j].Y + tri[(j+1)%3].Y) / 2}
		if pointInPolygonStrict(mid, hole) {
			return true
		}
	}
	return false
}

// segmentsProperlyCross reports a strict transverse crossing of segments
// ab and cd (a point lying on the other segment's endpoint does not count;
// collinear overlap is handled by the strict interior tests above).
func segmentsProperlyCross(a, b, c, d Point2D) bool {
	o1 := Area(a, b, c)
	o2 := Area(a, b, d)
	o3 := Area(c, d, a)
	o4 := Area(c, d, b)
	return o1 != 0 && o2 != 0 && o3 != 0 && o4 != 0 &&
		(o1 > 0) != (o2 > 0) && (o3 > 0) != (o4 > 0)
}

// pointInTriangleStrict tests strict containment (CCW triangle a,b,c).
func pointInTriangleStrict(p, a, b, c Point2D) bool {
	return Area(a, b, p) > 0 && Area(b, c, p) > 0 && Area(c, a, p) > 0
}

// smallestConvex returns the convex node with the smallest corner area
// (the mapbox ear-queue fallback vertex).
func smallestConvex(ring *Node, n int) *Node {
	var best *Node
	bestArea := math.Inf(1)
	cur := ring
	for range n {
		area := Area(cur.Prev.Point, cur.Point, cur.Next.Point)
		if area > 0 && area < bestArea {
			bestArea = area
			best = cur
		}
		cur = cur.Next
	}
	return best
}

// TriangulatePolygon triangulates a simple polygon (counter-clockwise
// winding is enforced in place on the input slice).
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

// TriangulatePolygonWithHoles triangulates a polygon with holes (exterior
// counter-clockwise, holes clockwise; windings are enforced in place on the
// input slices).
func TriangulatePolygonWithHoles(exterior []Point2D, holes [][]Point2D) []int {
	if len(exterior) < 3 {
		return []int{}
	}

	// Ensure exterior is in counter-clockwise order
	if IsClockwise(exterior) {
		// Reverse the exterior
		for i, j := 0, len(exterior)-1; i < j; i, j = i+1, j-1 {
			exterior[i], exterior[j] = exterior[j], exterior[i]
		}
	}

	// Ensure holes are in clockwise order
	for i, hole := range holes {
		if !IsClockwise(hole) {
			// Reverse the hole
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

// ValidateTriangulation checks that every index is in range and that no
// two triangles overlap in their interiors (shared edges and vertices are
// allowed).
func ValidateTriangulation(vertices []Point2D, indices []int) bool {
	if len(indices)%3 != 0 {
		return false
	}
	for _, index := range indices {
		if index < 0 || index >= len(vertices) {
			return false
		}
	}

	count := len(indices) / 3
	for i := range count {
		tri1 := trianglePoints(vertices, indices, i)
		for j := i + 1; j < count; j++ {
			tri2 := trianglePoints(vertices, indices, j)
			if trianglesOverlap(tri1, tri2) {
				return false
			}
		}
	}

	return true
}

func trianglePoints(vertices []Point2D, indices []int, t int) [3]Point2D {
	return [3]Point2D{
		vertices[indices[t*3]],
		vertices[indices[t*3+1]],
		vertices[indices[t*3+2]],
	}
}

// trianglesOverlap reports strict interior overlap: a strict segment
// crossing of two edges, or a vertex strictly inside the other triangle.
func trianglesOverlap(tri1, tri2 [3]Point2D) bool {
	for i := range 3 {
		a1, a2 := tri1[i], tri1[(i+1)%3]
		for j := range 3 {
			b1, b2 := tri2[j], tri2[(j+1)%3]
			if properCross(a1, a2, b1, b2) {
				return true
			}
		}
	}
	for i := range 3 {
		if pointInTriangleStrict(tri1[i], tri2[0], tri2[1], tri2[2]) {
			return true
		}
		if pointInTriangleStrict(tri2[i], tri1[0], tri1[1], tri1[2]) {
			return true
		}
	}
	return false
}

// properCross reports a proper transversal crossing of the two segments
// (no shared endpoints, no touching).
func properCross(a1, a2, b1, b2 Point2D) bool {
	if a1 == b1 || a1 == b2 || a2 == b1 || a2 == b2 {
		return false
	}
	o1 := Area(a1, a2, b1)
	o2 := Area(a1, a2, b2)
	o3 := Area(b1, b2, a1)
	o4 := Area(b1, b2, a2)
	return o1 != 0 && o2 != 0 && o3 != 0 && o4 != 0 &&
		(o1 > 0) != (o2 > 0) && (o3 > 0) != (o4 > 0)
}
