package entity_test

import (
	"os"
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
)

func loadCubeSAT(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile("../testdata/acis_cube.sat")
	if err != nil {
		t.Fatalf("read SAT fixture: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\r\n"), "\n")
	for i := range lines {
		lines[i] = strings.TrimSuffix(lines[i], "\r")
	}
	return lines
}

func TestDecodeSATMeshCube(t *testing.T) {
	mesh, err := entity.DecodeSATMesh(loadCubeSAT(t))
	if err != nil {
		t.Fatalf("decode SAT: %v", err)
	}
	if got := len(mesh.Vertices()); got != 8 {
		t.Fatalf("vertices = %d, want 8", got)
	}
	if got := len(mesh.Faces()); got != 12 {
		t.Fatalf("triangles = %d, want 12", got)
	}
	for _, face := range mesh.Faces() {
		if len(face) != 3 {
			t.Fatalf("face has %d vertices, want triangle", len(face))
		}
	}
}

func TestDecodeACISMeshBodyEntity(t *testing.T) {
	body := entity.NewBody()
	body.SetSATLines(loadCubeSAT(t))
	mesh, err := entity.DecodeACISMesh(body)
	if err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(mesh.Vertices()) != 8 || len(mesh.Faces()) != 12 {
		t.Fatalf("mesh = %d vertices, %d triangles", len(mesh.Vertices()), len(mesh.Faces()))
	}
}

func TestDecodeACISMeshRejectsSABAndMissingPayload(t *testing.T) {
	body := entity.NewBody()
	body.SetACISBinaryData([]byte("SAB payload"))
	if _, err := entity.DecodeACISMesh(body); err == nil {
		t.Fatal("expected SAB decoding error")
	}
	if _, err := entity.DecodeACISMesh(entity.NewLine()); err == nil {
		t.Fatal("expected non-ACIS error")
	}
}
