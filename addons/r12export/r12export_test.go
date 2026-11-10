package r12export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

func TestSplineToPolylineSamplesCurve(t *testing.T) {
	spline := entity.NewSplineFromEnhanced(dxfmath.NewBSpline([]dxfmath.Vec3{
		dxfmath.NewVec3(0, 0, 0),
		dxfmath.NewVec3(5, 10, 0),
		dxfmath.NewVec3(10, 0, 0),
	}, 2))
	spline.SetClosed(true)

	polyline := SplineToPolyline(spline, 8)
	if polyline == nil {
		t.Fatal("expected polyline")
	}
	if len(polyline.Vertices) != 9 {
		t.Fatalf("vertices = %d, want 9", len(polyline.Vertices))
	}
	if polyline.Flag&entity.PolylineFlagClosed == 0 {
		t.Fatal("expected closed polyline")
	}
}

func TestEllipseToPolylineSamplesClosedCurve(t *testing.T) {
	ellipse := entity.NewEllipse()
	ellipse.Center = []float64{0, 0, 0}
	ellipse.Major = 5
	ellipse.Minor = 3
	ellipse.Start = 0
	ellipse.End = 360

	polyline := EllipseToPolyline(ellipse, 12)
	if polyline == nil {
		t.Fatal("expected polyline")
	}
	if len(polyline.Vertices) != 12 {
		t.Fatalf("vertices = %d, want 12", len(polyline.Vertices))
	}
	if polyline.Flag&entity.PolylineFlagClosed == 0 {
		t.Fatal("expected closed polyline")
	}
}

func TestLwPolylineToPolylineCopiesVerticesAndBulges(t *testing.T) {
	lw := entity.NewLwPolyline(3)
	lw.Vertices = [][]float64{{0, 0}, {10, 0}, {10, 10}}
	lw.Bulges = []float64{0, 0.5, 0}
	lw.Closed = true

	polyline := LwPolylineToPolyline(lw)
	if polyline == nil {
		t.Fatal("expected polyline")
	}
	if len(polyline.Vertices) != 3 {
		t.Fatalf("vertices = %d, want 3", len(polyline.Vertices))
	}
	if polyline.Vertices[1].Bulge != 0.5 {
		t.Fatalf("bulge = %v, want 0.5", polyline.Vertices[1].Bulge)
	}
	if polyline.Flag&entity.PolylineFlagClosed == 0 {
		t.Fatal("expected closed polyline")
	}
}

func TestMTextToTextJoinsAdditionalText(t *testing.T) {
	mtext := entity.NewMText()
	mtext.Coord1 = []float64{1, 2, 0}
	mtext.Coord2 = []float64{1, 2, 0}
	mtext.Height = 2.5
	mtext.Rotation = 15
	mtext.Value = "main"
	mtext.AdditionalText = []string{"second", "third"}

	text := MTextToText(mtext)
	if text == nil {
		t.Fatal("expected text")
	}
	if text.Value != "main\\Psecond\\Pthird" {
		t.Fatalf("value = %q", text.Value)
	}
	if text.Height != 2.5 || text.Rotation != 15 {
		t.Fatalf("text formatting lost: height=%v rotation=%v", text.Height, text.Rotation)
	}
}

func TestConvertEntitiesMeshToPolyface(t *testing.T) {
	mesh := entity.NewMesh()
	mesh.AddVertex(dxfmath.NewVec3(0, 0, 0))
	mesh.AddVertex(dxfmath.NewVec3(1, 0, 0))
	mesh.AddVertex(dxfmath.NewVec3(1, 1, 0))
	mesh.AddVertex(dxfmath.NewVec3(0, 1, 0))
	mesh.AddFace([]int{0, 1, 2, 3})

	converted, report := ConvertEntities([]entity.Entity{mesh}, 8)
	if len(converted) != 1 {
		t.Fatalf("converted = %d, want 1", len(converted))
	}
	polyline, ok := converted[0].(*entity.Polyline)
	if !ok {
		t.Fatalf("converted type = %T, want *entity.Polyline", converted[0])
	}
	if polyline.Flag&entity.PolylineFlagPolyface == 0 {
		t.Fatal("expected polyface polyline")
	}
	if len(polyline.MeshVertices()) != 4 {
		t.Fatalf("vertices = %d, want 4", len(polyline.MeshVertices()))
	}
	faces := polyline.FaceRecords()
	if len(faces) != 1 {
		t.Fatalf("face records = %d, want 1", len(faces))
	}
	if faces[0].FaceIndices[0] != 1 || faces[0].FaceIndices[3] != 4 {
		t.Fatalf("face indices = %v, want 1..4", faces[0].FaceIndices)
	}
	if len(report.Converted) != 1 {
		t.Fatalf("converted report = %v", report.Converted)
	}
	entry := report.Converted[0]
	if !strings.Contains(entry, "MESH") || !strings.Contains(entry, "POLYFACE") {
		t.Fatalf("converted report = %v", report.Converted)
	}
}

func TestExportWritesR12StreamAndReport(t *testing.T) {
	doc, err := drawing.New()
	if err != nil {
		t.Fatalf("new drawing: %v", err)
	}
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{10, 0, 0}
	if err := doc.AddEntity(line); err != nil {
		t.Fatalf("add line: %v", err)
	}
	ellipse := entity.NewEllipse()
	ellipse.Center = []float64{5, 5, 0}
	ellipse.Major = 2
	ellipse.Minor = 1
	ellipse.Start = 0
	ellipse.End = 360
	if err := doc.AddEntity(ellipse); err != nil {
		t.Fatalf("add ellipse: %v", err)
	}

	var b strings.Builder
	report, err := Export(&b, doc, Options{Segments: 6})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if b.Len() == 0 || !strings.Contains(b.String(), "SECTION") {
		t.Fatal("expected non-empty DXF stream")
	}
	if len(report.Converted) != 1 {
		t.Fatalf("report = %+v", report)
	}
	entry := report.Converted[0]
	if !strings.Contains(entry, "ELLIPSE") || !strings.Contains(entry, "POLYLINE") {
		t.Fatalf("report = %+v", report)
	}
	if strings.Contains(b.String(), "ELLIPSE") {
		t.Fatal("ellipse should have been converted, not emitted as ELLIPSE")
	}
}

func TestSaveAsWritesFile(t *testing.T) {
	doc, err := drawing.New()
	if err != nil {
		t.Fatalf("new drawing: %v", err)
	}
	path := filepath.Join(t.TempDir(), "out.dxf")
	if _, err := SaveAs(path, doc); err != nil {
		t.Fatalf("save: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestExportRejectsNilInputs(t *testing.T) {
	if _, err := Export(nil, nil); err == nil {
		t.Fatal("expected error for nil writer/drawing")
	}
	var b strings.Builder
	if _, err := Export(&b, nil); err == nil {
		t.Fatal("expected error for nil drawing")
	}
	if _, err := SaveAs("", nil); err == nil {
		t.Fatal("expected error for nil drawing in SaveAs")
	}
}
