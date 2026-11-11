package edges

import (
	"errors"
	"testing"

	dxfmath "github.com/edanko/dxf/math"
)

func v2(x, y float64) dxfmath.Vec2 { return dxfmath.NewVec2(x, y) }

func rectangle() []Edge {
	return []Edge{
		NewEdge(v2(0, 0), v2(10, 0)),
		NewEdge(v2(10, 0), v2(10, 10)),
		NewEdge(v2(10, 10), v2(0, 10)),
		NewEdge(v2(0, 10), v2(0, 0)),
	}
}

func TestGraphMergesEndpointsWithTolerance(t *testing.T) {
	g := BuildGraph([]Edge{
		NewEdge(v2(0, 0), v2(1, 0)),
		NewEdge(v2(1+1e-7, 0), v2(2, 0)),
		NewEdge(v2(5, 5), v2(6, 5)),
	}, 1e-5)
	if g.VertexCount() != 5 || g.EdgeCount() != 3 {
		t.Fatalf("vertices=%d edges=%d", g.VertexCount(), g.EdgeCount())
	}
	// The shared endpoint has degree two.
	found := false
	for i := 0; i < g.VertexCount(); i++ {
		if g.Degree(i) == 2 {
			found = true
		}
	}
	if !found {
		t.Fatal("shared endpoint not merged")
	}
	if got := len(g.Leaves()); got != 4 {
		t.Fatalf("leaves=%d, want 4", got)
	}
	if got := len(g.Components()); got != 2 {
		t.Fatalf("components=%d, want 2", got)
	}
}

func TestClosedLoopAndOpenChain(t *testing.T) {
	rect := BuildGraph(rectangle(), 1e-6)
	if got := len(rect.Leaves()); got != 0 {
		t.Fatalf("rectangle leaves=%d", got)
	}
	topo := rect.Analyze()
	if len(topo.Loops) != 1 || len(topo.Chains) != 0 {
		t.Fatalf("rectangle topology loops=%d chains=%d", len(topo.Loops), len(topo.Chains))
	}
	loop := topo.Loops[0]
	if len(loop.Vertices) != 4 || len(loop.Steps) != 4 {
		t.Fatalf("loop vertices=%d steps=%d", len(loop.Vertices), len(loop.Steps))
	}
	want := []dxfmath.Vec2{v2(0, 0), v2(10, 0), v2(10, 10), v2(0, 10)}
	for i := range want {
		if loop.Vertices[i].Distance(want[i]) > 1e-9 {
			t.Fatalf("loop[%d]=%v want %v", i, loop.Vertices[i], want[i])
		}
	}
	for _, s := range loop.Steps {
		if s.Reversed {
			t.Fatalf("source orientation not preserved: %+v", loop.Steps)
		}
	}

	open := BuildGraph([]Edge{
		NewEdge(v2(0, 0), v2(1, 0)),
		NewEdge(v2(1, 0), v2(2, 0)),
		NewEdge(v2(2, 0), v2(3, 0)),
	}, 1e-6)
	if got := len(open.Leaves()); got != 2 {
		t.Fatalf("open leaves=%d", got)
	}
	topo = open.Analyze()
	if len(topo.Loops) != 0 || len(topo.Chains) != 1 {
		t.Fatalf("open topology loops=%d chains=%d", len(topo.Loops), len(topo.Chains))
	}
	if got := len(topo.Chains[0].Vertices); got != 4 {
		t.Fatalf("chain vertices=%d", got)
	}
}

func TestMixedLoopAndChain(t *testing.T) {
	es := append(rectangle(), NewEdge(v2(10, 0), v2(15, 0)))
	g := BuildGraph(es, 1e-6)
	topo := g.Analyze()
	if len(topo.Loops) != 1 || len(topo.Chains) != 1 {
		t.Fatalf("mixed topology loops=%d chains=%d", len(topo.Loops), len(topo.Chains))
	}
	if got := len(topo.Chains[0].Steps); got != 1 {
		t.Fatalf("tail steps=%d", got)
	}
}

func TestEntityAndBoundaryCreation(t *testing.T) {
	rect := BuildGraph(rectangle(), 1e-6)
	loop := rect.Analyze().Loops[0]
	lw := loop.ToLwPolyline()
	if !lw.Closed || lw.Num != 4 || lw.Vertices[2][0] != 10 || lw.Vertices[2][1] != 10 {
		t.Fatalf("lwpolyline=%+v", lw)
	}
	pl := loop.ToPolyline()
	if pl.Flag&1 == 0 || len(pl.Vertices) != 4 {
		t.Fatalf("polyline flag=%d vertices=%d", pl.Flag, len(pl.Vertices))
	}
	bp, err := loop.ToHatchBoundaryPath()
	if err != nil || !bp.IsClosed() || len(bp.Vertices()) != 4 {
		t.Fatalf("boundary=%v err=%v", bp, err)
	}

	open := BuildGraph([]Edge{
		NewEdge(v2(0, 0), v2(1, 0)),
		NewEdge(v2(1, 0), v2(2, 0)),
	}, 1e-6).Analyze().Chains[0]
	if _, err := open.ToHatchBoundaryPath(); !errors.Is(err, ErrOpenChain) {
		t.Fatalf("open boundary error=%v", err)
	}
	if _, err := ToHatchBoundaryPath(nil); err == nil {
		t.Fatal("empty loop list must fail")
	}
}

func TestReversalReversesOrderAndOrientation(t *testing.T) {
	loop := BuildGraph(rectangle(), 1e-6).Analyze().Loops[0]
	rev := loop.Reverse()
	if len(rev.Vertices) != len(loop.Vertices) || len(rev.Steps) != len(loop.Steps) {
		t.Fatal("reversal changed length")
	}
	if rev.Vertices[0].Distance(v2(0, 10)) > 1e-9 {
		t.Fatalf("reversed first vertex=%v", rev.Vertices[0])
	}
	if !rev.Steps[0].Reversed {
		t.Fatal("reversed traversal did not flip source orientation")
	}
	again := rev.Reverse()
	for i := range loop.Vertices {
		if again.Vertices[i].Distance(loop.Vertices[i]) > 1e-9 {
			t.Fatalf("double reversal differs at %d", i)
		}
		if again.Steps[i].Reversed != loop.Steps[i].Reversed {
			t.Fatalf("double reversal orientation differs at %d", i)
		}
	}
}

func TestToleranceBoundary(t *testing.T) {
	g := BuildGraph([]Edge{
		NewEdge(v2(0, 0), v2(1, 0)),
		NewEdge(v2(0.999, 0), v2(1, 1)),
	}, 1e-12)
	if g.VertexCount() != 4 {
		t.Fatalf("tight tolerance merged distinct endpoints: %d", g.VertexCount())
	}
}
