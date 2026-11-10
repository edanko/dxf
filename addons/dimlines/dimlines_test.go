package dimlines

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

func TestLinearDimensionEntities(t *testing.T) {
	dim := NewLinearDimension(
		dxfmath.NewVec3(0, 0, 0),
		dxfmath.NewVec3(10, 0, 0),
		2,
	)
	entities, err := dim.Entities()
	if err != nil {
		t.Fatalf("entities: %v", err)
	}
	if len(entities) != 6 {
		t.Fatalf("entities = %d, want 6", len(entities))
	}

	lines, texts := 0, 0
	for _, e := range entities {
		layerName := entityLayerName(e)
		if layerName != "DIMENSIONS" {
			t.Fatalf("layer = %q, want DIMENSIONS", layerName)
		}
		switch e.DXFType() {
		case "LINE":
			lines++
		case "TEXT":
			texts++
			text, ok := e.(*entity.Text)
			if !ok {
				t.Fatalf("TEXT entity has type %T", e)
			}
			if !strings.Contains(text.Value, "10.00") {
				t.Fatalf("text value = %q, want formatted length", text.Value)
			}
		}
	}
	if lines != 5 || texts != 1 {
		t.Fatalf("lines=%d texts=%d, want 5/1", lines, texts)
	}
}

func TestLinearDimensionRejectsCoincidentPoints(t *testing.T) {
	dim := NewLinearDimension(
		dxfmath.NewVec3(1, 2, 0),
		dxfmath.NewVec3(1, 2, 0),
		1,
	)
	if _, err := dim.Entities(); err == nil {
		t.Fatal("expected coincident point error")
	}
}

func TestAngularDimensionEntities(t *testing.T) {
	dim := NewAngularDimension(
		dxfmath.NewVec3(0, 0, 0),
		dxfmath.NewVec3(10, 0, 0),
		dxfmath.NewVec3(0, 10, 0),
		5,
	)
	entities, err := dim.Entities()
	if err != nil {
		t.Fatalf("entities: %v", err)
	}
	if len(entities) != 19 {
		t.Fatalf("entities = %d, want 19", len(entities))
	}
	foundText := false
	for _, e := range entities {
		if e.DXFType() == "TEXT" {
			foundText = true
			text, ok := e.(*entity.Text)
			if !ok {
				t.Fatalf("TEXT entity has type %T", e)
			}
			if !strings.Contains(text.Value, "°") {
				t.Fatalf("text value = %q, want degrees", text.Value)
			}
		}
	}
	if !foundText {
		t.Fatal("expected angular dimension text")
	}
}

func TestAngularDimensionRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name   string
		center dxfmath.Vec3
		start  dxfmath.Vec3
		end    dxfmath.Vec3
		radius float64
	}{
		{"zero angle", dxfmath.NewVec3(0, 0, 0), dxfmath.NewVec3(1, 0, 0), dxfmath.NewVec3(2, 0, 0), 5},
		{"zero radius", dxfmath.NewVec3(0, 0, 0), dxfmath.NewVec3(1, 0, 0), dxfmath.NewVec3(0, 1, 0), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dim := NewAngularDimension(tt.center, tt.start, tt.end, tt.radius)
			if _, err := dim.Entities(); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestAddToDrawing(t *testing.T) {
	doc, err := drawing.New()
	if err != nil {
		t.Fatalf("new drawing: %v", err)
	}
	dim := NewLinearDimension(
		dxfmath.NewVec3(0, 0, 0),
		dxfmath.NewVec3(4, 0, 0),
		1,
	)
	if err := dim.AddTo(doc); err != nil {
		t.Fatalf("add to drawing: %v", err)
	}
	if got := len(doc.Entities()); got != 6 {
		t.Fatalf("drawing entities = %d, want 6", got)
	}
}

func entityLayerName(e entity.Entity) string {
	if named, ok := e.(interface{ LayerName() string }); ok {
		return named.LayerName()
	}
	if layer := e.Layer(); layer != nil {
		return layer.Name()
	}
	return ""
}
