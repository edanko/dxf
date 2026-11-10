package path

import (
	"sort"

	dxfmath "github.com/edanko/dxf/math"
)

// NestingNode is one closed path in a path-nesting tree.
//
// Even depths are outer boundaries, odd depths are holes. A hole may contain
// another outer boundary (an island), which in turn may contain further holes.
type NestingNode struct {
	// Path is the normalized path for this node. Building a nesting tree copies
	// the input paths, so callers can keep using their originals.
	Path *Path

	// Depth is the nesting depth: 0 for a root outer boundary, 1 for a hole,
	// 2 for an island inside a hole, and so on.
	Depth int

	// IsHole reports whether this node is a hole (odd depth).
	IsHole bool

	// Parent is the containing node, or nil for roots.
	Parent *NestingNode

	// Children are the directly nested paths, in deterministic input order.
	Children []*NestingNode
}

// BuildNestingTree classifies closed paths into outer boundaries and nested
// holes.
//
// Containment is determined from the flattened path polygons. Larger paths are
// processed first and each path is attached to the smallest containing path,
// so concentric rings form a chain and disjoint rings form separate roots.
//
// The returned paths are normalized for filling: outer boundaries (even depth)
// are counter-clockwise and holes (odd depth) are clockwise. Input paths are
// not modified. Open or degenerate paths are ignored.
func BuildNestingTree(paths []*Path) []*NestingNode {
	type candidate struct {
		node    *NestingNode
		polygon []dxfmath.Vec2
		area    float64
		order   int
	}

	candidates := make([]candidate, 0, len(paths))
	for index, p := range paths {
		if p == nil || !p.IsClosed() {
			continue
		}
		polygon := pathPolygon(p)
		if len(polygon) < 3 {
			continue
		}
		area := dxfmath.PolygonArea(polygon)
		if area <= dxfmath.DefaultTolerance {
			continue
		}
		candidates = append(candidates, candidate{
			node:    &NestingNode{Path: p.Clone()},
			polygon: polygon,
			area:    area,
			order:   index,
		})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].area != candidates[j].area {
			return candidates[i].area > candidates[j].area
		}
		return candidates[i].order < candidates[j].order
	})

	roots := make([]*NestingNode, 0, len(candidates))
	for i := range candidates {
		child := &candidates[i]

		// Candidates are sorted largest first. Walking backwards visits the
		// smallest containing polygon first, which is the immediate parent.
		var parent *NestingNode
		for j := i - 1; j >= 0; j-- {
			if dxfmath.PointInPolygon(child.polygon[0], candidates[j].polygon) {
				parent = candidates[j].node
				break
			}
		}

		if parent == nil {
			roots = append(roots, child.node)
			continue
		}
		child.node.Parent = parent
		parent.Children = append(parent.Children, child.node)
	}

	for _, root := range roots {
		assignNestingDepth(root, 0)
	}
	for _, root := range roots {
		normalizeNestingNode(root)
	}
	return roots
}

// ClassifyNestedPaths returns the outer boundaries and holes of the given
// closed paths. Both slices use the normalized paths from BuildNestingTree.
func ClassifyNestedPaths(paths []*Path) (outers, holes []*Path) {
	for _, node := range flattenNestingNodes(BuildNestingTree(paths)) {
		if node.IsHole {
			holes = append(holes, node.Path)
		} else {
			outers = append(outers, node.Path)
		}
	}
	return outers, holes
}

func assignNestingDepth(node *NestingNode, depth int) {
	node.Depth = depth
	node.IsHole = depth%2 == 1
	for _, child := range node.Children {
		child.Parent = node
		assignNestingDepth(child, depth+1)
	}
}

func normalizeNestingNode(node *NestingNode) {
	node.Path = orientPath(node.Path, !node.IsHole)
	for _, child := range node.Children {
		normalizeNestingNode(child)
	}
}

// flattenNestingNodes returns a depth-first, pre-order view of the tree.
func flattenNestingNodes(nodes []*NestingNode) []*NestingNode {
	var flat []*NestingNode
	var visit func(*NestingNode)
	visit = func(node *NestingNode) {
		flat = append(flat, node)
		for _, child := range node.Children {
			visit(child)
		}
	}
	for _, node := range nodes {
		visit(node)
	}
	return flat
}

// pathPolygon flattens a path to a polygon ring and removes a duplicated
// closing point.
func pathPolygon(p *Path) []dxfmath.Vec2 {
	points := NewTools(p).flatten()
	polygon := make([]dxfmath.Vec2, 0, len(points))
	for _, point := range points {
		if len(polygon) > 0 && polygon[len(polygon)-1].Distance(point) <= 1e-9 {
			continue
		}
		polygon = append(polygon, point)
	}
	if len(polygon) > 1 && polygon[0].Distance(polygon[len(polygon)-1]) <= 1e-9 {
		polygon = polygon[:len(polygon)-1]
	}
	return polygon
}

// orientPath returns a copy of p with the requested winding. The path is
// assumed to describe a single closed ring.
func orientPath(p *Path, counterClockwise bool) *Path {
	winding := dxfmath.PolygonWinding(pathPolygon(p))
	if winding == dxfmath.PolygonDegenerate {
		return p.Clone()
	}
	want := dxfmath.PolygonCounterClockwise
	if !counterClockwise {
		want = dxfmath.PolygonClockwise
	}
	if winding == want {
		return p.Clone()
	}
	return reversePath(p)
}

// reversePath reverses a single closed ring while preserving curve control
// points. The returned path starts at the original end point.
func reversePath(p *Path) *Path {
	commands := p.Commands()
	if len(commands) == 0 {
		return NewPath()
	}

	result := NewPath()
	last := commands[len(commands)-1].EndVertex()
	result.MoveTo(last.X(), last.Y())

	for i := len(commands) - 1; i >= 1; i-- {
		start := commands[i-1].EndVertex()
		switch command := commands[i].(type) {
		case *PathLineTo:
			result.LineTo(start.X(), start.Y())
		case *PathCurve3To:
			result.Curve3To(start.X(), start.Y(), command.Ctrl.X(), command.Ctrl.Y())
		case *PathCurve4To:
			result.Curve4To(
				start.X(), start.Y(),
				command.Ctrl2.X(), command.Ctrl2.Y(),
				command.Ctrl1.X(), command.Ctrl1.Y(),
			)
		}
	}
	return result
}
