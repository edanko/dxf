package path

import (
	"testing"

	dxfmath "github.com/edanko/dxf/math"
)

func squarePath(x0, y0, x1, y1 float64, clockwise bool) *Path {
	p := NewPath()
	p.MoveTo(x0, y0)
	if clockwise {
		p.LineTo(x0, y1)
		p.LineTo(x1, y1)
		p.LineTo(x1, y0)
	} else {
		p.LineTo(x1, y0)
		p.LineTo(x1, y1)
		p.LineTo(x0, y1)
	}
	p.Close()
	return p
}

func TestNestingConcentricRings(t *testing.T) {
	outer := squarePath(0, 0, 10, 10, false)
	hole := squarePath(3, 3, 7, 7, false)
	holeInputWinding := dxfmath.PolygonWinding(pathPolygon(hole))

	roots := BuildNestingTree([]*Path{hole, outer})
	if len(roots) != 1 {
		t.Fatalf("expected one outer boundary, got %d", len(roots))
	}

	root := roots[0]
	if root.Depth != 0 || root.IsHole {
		t.Fatalf("expected root outer boundary, got depth=%d hole=%v", root.Depth, root.IsHole)
	}
	if len(root.Children) != 1 {
		t.Fatalf("expected one hole, got %d", len(root.Children))
	}

	holeNode := root.Children[0]
	if holeNode.Depth != 1 || !holeNode.IsHole {
		t.Fatalf("expected hole at depth 1, got depth=%d hole=%v", holeNode.Depth, holeNode.IsHole)
	}
	if len(holeNode.Children) != 0 {
		t.Fatalf("expected hole without nested paths, got %d", len(holeNode.Children))
	}

	if got := dxfmath.PolygonWinding(pathPolygon(root.Path)); got != dxfmath.PolygonCounterClockwise {
		t.Fatalf("outer boundary is not counter-clockwise: %v", got)
	}
	if got := dxfmath.PolygonWinding(pathPolygon(holeNode.Path)); got != dxfmath.PolygonClockwise {
		t.Fatalf("hole is not clockwise: %v", got)
	}
	if got := dxfmath.PolygonWinding(pathPolygon(hole)); got != holeInputWinding {
		t.Fatal("BuildNestingTree modified the input path")
	}

	outers, holes := ClassifyNestedPaths([]*Path{outer, hole})
	if len(outers) != 1 || len(holes) != 1 {
		t.Fatalf("expected 1 outer and 1 hole, got %d and %d", len(outers), len(holes))
	}
}

func TestNestingDisjointRings(t *testing.T) {
	first := squarePath(0, 0, 4, 4, true)
	second := squarePath(10, 0, 14, 4, false)

	roots := BuildNestingTree([]*Path{first, second})
	if len(roots) != 2 {
		t.Fatalf("expected two outer boundaries, got %d", len(roots))
	}
	for _, root := range roots {
		if root.Depth != 0 || root.IsHole {
			t.Fatalf("expected outer boundary, got depth=%d hole=%v", root.Depth, root.IsHole)
		}
		if len(root.Children) != 0 {
			t.Fatalf("disjoint ring should not have children, got %d", len(root.Children))
		}
		if got := dxfmath.PolygonWinding(pathPolygon(root.Path)); got != dxfmath.PolygonCounterClockwise {
			t.Fatalf("outer boundary is not counter-clockwise: %v", got)
		}
	}
}

func TestNestingIslandInsideHole(t *testing.T) {
	outer := squarePath(0, 0, 20, 20, false)
	hole := squarePath(2, 2, 18, 18, false)
	island := squarePath(4, 4, 16, 16, true)
	innerHole := squarePath(6, 6, 14, 14, false)

	roots := BuildNestingTree([]*Path{innerHole, island, hole, outer})
	if len(roots) != 1 {
		t.Fatalf("expected one root, got %d", len(roots))
	}

	root := roots[0]
	if root.Depth != 0 || root.IsHole {
		t.Fatalf("expected outer root, got depth=%d hole=%v", root.Depth, root.IsHole)
	}
	if len(root.Children) != 1 {
		t.Fatalf("expected one hole, got %d", len(root.Children))
	}
	holeNode := root.Children[0]
	if holeNode.Depth != 1 || !holeNode.IsHole {
		t.Fatalf("expected hole, got depth=%d hole=%v", holeNode.Depth, holeNode.IsHole)
	}
	if len(holeNode.Children) != 1 {
		t.Fatalf("expected one island, got %d", len(holeNode.Children))
	}
	islandNode := holeNode.Children[0]
	if islandNode.Depth != 2 || islandNode.IsHole {
		t.Fatalf("expected island, got depth=%d hole=%v", islandNode.Depth, islandNode.IsHole)
	}
	if len(islandNode.Children) != 1 {
		t.Fatalf("expected one inner hole, got %d", len(islandNode.Children))
	}
	innerNode := islandNode.Children[0]
	if innerNode.Depth != 3 || !innerNode.IsHole {
		t.Fatalf("expected inner hole, got depth=%d hole=%v", innerNode.Depth, innerNode.IsHole)
	}

	want := []dxfmath.PolygonOrientation{
		dxfmath.PolygonCounterClockwise,
		dxfmath.PolygonClockwise,
		dxfmath.PolygonCounterClockwise,
		dxfmath.PolygonClockwise,
	}
	nodes := []*NestingNode{root, holeNode, islandNode, innerNode}
	for i, node := range nodes {
		if got := dxfmath.PolygonWinding(pathPolygon(node.Path)); got != want[i] {
			t.Fatalf("depth %d winding = %v, want %v", node.Depth, got, want[i])
		}
	}
}
