package dxf

import (
	"bytes"
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

func findPolyline(t *testing.T, d *drawing.Drawing) *entity.Polyline {
	t.Helper()
	for _, e := range d.Entities() {
		if p, ok := e.(*entity.Polyline); ok {
			return p
		}
	}
	t.Fatalf("no POLYLINE entity found")
	return nil
}

func findMesh(t *testing.T, d *drawing.Drawing) *entity.Mesh {
	t.Helper()
	for _, e := range d.Entities() {
		if m, ok := e.(*entity.Mesh); ok {
			return m
		}
	}
	t.Fatalf("no MESH entity found")
	return nil
}

func TestPolyfaceMeshTyped(t *testing.T) {
	d, err := FromFile("testdata/polyface_mesh.dxf")
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	p := findPolyline(t, d)
	if !p.IsPolyface() {
		t.Fatalf("polyline flag %d is not a polyface", p.Flag)
	}
	if p.MeshVertexCount != 4 || p.MeshFaceCount != 2 {
		t.Fatalf("mesh counts = %d/%d, want 4/2", p.MeshVertexCount, p.MeshFaceCount)
	}
	if len(p.Vertices) != 6 {
		t.Fatalf("vertex count = %d, want 6", len(p.Vertices))
	}
	if got := len(p.MeshVertices()); got != 4 {
		t.Fatalf("mesh vertices = %d, want 4", got)
	}
	faces := p.FaceRecords()
	if len(faces) != 2 {
		t.Fatalf("face records = %d, want 2", len(faces))
	}
	if !faces[0].IsFaceRecord() {
		t.Fatalf("first face record flag %d missing face bit", faces[0].Flag)
	}
	want := []int{1, 2, 3, -4}
	got := faces[0].FaceVertexIndices()
	if len(got) != len(want) {
		t.Fatalf("face indices = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("face indices = %v, want %v", got, want)
		}
	}
	if got := len(p.Subentities()); got != 6 {
		t.Fatalf("subentities = %d, want 6", got)
	}

	var buf bytes.Buffer
	if _, err := d.WriteTo(&buf); err != nil {
		t.Fatalf("write drawing: %v", err)
	}
	reloaded, err := FromStringData(buf.String())
	if err != nil {
		t.Fatalf("reload drawing: %v", err)
	}
	rp := findPolyline(t, reloaded)
	if !rp.IsPolyface() || rp.MeshVertexCount != 4 || rp.MeshFaceCount != 2 {
		t.Fatalf("reloaded polyline lost polyface metadata: flag=%d counts=%d/%d",
			rp.Flag, rp.MeshVertexCount, rp.MeshFaceCount)
	}
	rfaces := rp.FaceRecords()
	if len(rfaces) != 2 {
		t.Fatalf("reloaded face records = %d, want 2", len(rfaces))
	}
	got = rfaces[0].FaceVertexIndices()
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("reloaded face indices = %v, want %v", got, want)
		}
	}
}

func TestMeshVertexCacheTyped(t *testing.T) {
	d, err := FromFile("testdata/mesh_vertex_cache.dxf")
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	mesh := findMesh(t, d)
	cache := mesh.VertexCache()
	if cache == nil {
		t.Fatalf("mesh vertex cache not linked")
	}
	if cache.VertexCount() != 4 || cache.NormalCount() != 4 ||
		cache.TexCoordCount() != 4 || cache.FaceCount() != 1 {
		t.Fatalf("cache counts = %d/%d/%d/%d, want 4/4/4/1",
			cache.VertexCount(), cache.NormalCount(),
			cache.TexCoordCount(), cache.FaceCount())
	}
	if got := cache.Vertices[2]; got.X() != 1 || got.Y() != 1 || got.Z() != 0 {
		t.Fatalf("cache vertex[2] = %v, want (1,1,0)", got)
	}
	if got := cache.FaceIndices[0]; len(got) != 4 || got[3] != 3 {
		t.Fatalf("cache face = %v, want [0 1 2 3]", got)
	}
	if owner := cache.Owner(); owner == nil || owner.Handle() != "A" {
		t.Fatalf("cache owner = %v, want A", owner)
	}

	var buf bytes.Buffer
	if _, err := d.WriteTo(&buf); err != nil {
		t.Fatalf("write drawing: %v", err)
	}
	reloaded, err := FromStringData(buf.String())
	if err != nil {
		t.Fatalf("reload drawing: %v", err)
	}
	rmesh := findMesh(t, reloaded)
	rcache := rmesh.VertexCache()
	if rcache == nil {
		t.Fatalf("reloaded mesh vertex cache not linked")
	}
	if rcache.VertexCount() != 4 || rcache.FaceCount() != 1 {
		t.Fatalf("reloaded cache counts = %d/%d, want 4/1",
			rcache.VertexCount(), rcache.FaceCount())
	}
	if got := rcache.FaceIndices[0]; len(got) != 4 || got[3] != 3 {
		t.Fatalf("reloaded cache face = %v, want [0 1 2 3]", got)
	}
}
