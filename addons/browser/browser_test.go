package browser

import (
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

func TestExtractGroupsEntitiesByLayoutAndType(t *testing.T) {
	doc, err := drawing.New()
	if err != nil {
		t.Fatalf("new drawing: %v", err)
	}
	doc.SetSourceVersion("AC1009")
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{10, 0, 0}
	if err := doc.AddEntity(line); err != nil {
		t.Fatalf("add line: %v", err)
	}
	sheet, err := doc.NewLayout("Sheet1")
	if err != nil {
		t.Fatalf("new layout: %v", err)
	}
	circle := entity.NewCircle()
	circle.Center = []float64{5, 5, 0}
	circle.Radius = 2
	if err := sheet.Add(circle); err != nil {
		t.Fatalf("add circle: %v", err)
	}

	data := Extract(doc)
	if data.Version == "" {
		t.Fatal("expected document version")
	}
	if len(data.Layouts) != 2 {
		t.Fatalf("expected two layouts, got %d", len(data.Layouts))
	}

	model := data.Layouts[0]
	if model.Type != "ModelSpace" || model.EntityCount != 1 {
		t.Fatalf("unexpected model layout: %+v", model)
	}
	if len(model.Entities) != 1 || model.Entities[0].Type != "LINE" {
		t.Fatalf("expected model LINE, got %+v", model.Entities)
	}
	if model.Entities[0].Handle == "" {
		t.Fatal("expected entity handle")
	}
	if model.Entities[0].BBox == nil {
		t.Fatal("expected line bounding box")
	}

	paper := data.Layouts[1]
	if paper.Name != "Sheet1" || paper.Type != "PaperSpace" || paper.EntityCount != 1 {
		t.Fatalf("unexpected paper layout: %+v", paper)
	}
	if len(paper.ByType["CIRCLE"]) != 1 {
		t.Fatalf("expected circle grouped under CIRCLE, got %+v", paper.ByType)
	}
}

func TestExtractNilDrawing(t *testing.T) {
	data := Extract(nil)
	if len(data.Layouts) != 0 {
		t.Fatalf("expected no layouts, got %d", len(data.Layouts))
	}
}

func TestEntitiesByType(t *testing.T) {
	doc, err := drawing.New()
	if err != nil {
		t.Fatalf("new drawing: %v", err)
	}
	first := entity.NewLine()
	if err := doc.AddEntity(first); err != nil {
		t.Fatalf("add first line: %v", err)
	}
	second := entity.NewLine()
	if err := doc.AddEntity(second); err != nil {
		t.Fatalf("add second line: %v", err)
	}
	grouped := EntitiesByType(doc)
	if len(grouped["LINE"]) != 2 {
		t.Fatalf("expected two LINE entities, got %+v", grouped)
	}
}
