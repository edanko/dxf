package dxf

import (
	"os"
	"strings"
	"testing"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

func TestJSONEnvelopeRoundTrip(t *testing.T) {
	data, err := os.ReadFile("testdata/columns_R2018.dxf")
	if err != nil {
		t.Fatal(err)
	}
	first, err := FromStringData(string(data))
	if err != nil {
		t.Fatal(err)
	}

	payload, err := first.ToJSON(nil)
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}
	if !strings.Contains(string(payload), `"payload"`) {
		t.Fatalf("envelope does not embed a payload:\n%s", payload)
	}

	var restored drawing.Drawing
	if err := restored.FromJSON(payload); err != nil {
		t.Fatalf("FromJSON: %v", err)
	}

	a, b := inventoryOf(first), inventoryOf(&restored)
	equalStrings(t, "entities", a.Entities, b.Entities)
	equalStrings(t, "tables", a.Tables, b.Tables)
	equalStrings(t, "blocks", a.Blocks, b.Blocks)
	equalStrings(t, "objects", a.Objects, b.Objects)
}

func TestFromJSONMalformedLeavesDrawingUnchanged(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatal(err)
	}
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{1, 1, 0}
	if err := d.AddEntity(line); err != nil {
		t.Fatal(err)
	}

	before := inventoryOf(d)
	if err := d.FromJSON([]byte("{not json")); err == nil {
		t.Fatal("FromJSON accepted malformed JSON")
	}
	after := inventoryOf(d)
	equalStrings(t, "entities", before.Entities, after.Entities)
}

func TestFromJSONUnsupportedShapeNamesField(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatal(err)
	}
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{1, 1, 0}
	if err := d.AddEntity(line); err != nil {
		t.Fatal(err)
	}
	before := inventoryOf(d)

	err = d.FromJSON([]byte(`{"version":1,"format":"dxf-ascii"}`))
	if err == nil {
		t.Fatal("FromJSON accepted an envelope without a payload")
	}
	if !strings.Contains(err.Error(), "payload") {
		t.Fatalf("error does not name the missing field: %v", err)
	}
	after := inventoryOf(d)
	equalStrings(t, "entities", before.Entities, after.Entities)
}

func TestFromJSONUnsupportedVersion(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatal(err)
	}
	line := entity.NewLine()
	line.Start = []float64{0, 0, 0}
	line.End = []float64{1, 1, 0}
	if err := d.AddEntity(line); err != nil {
		t.Fatal(err)
	}
	before := inventoryOf(d)

	err = d.FromJSON([]byte(`{"version":999,"format":"dxf-ascii","payload":"0\nSECTION\n"}`))
	if err == nil {
		t.Fatal("FromJSON accepted an unsupported version")
	}
	if !strings.Contains(err.Error(), "version") {
		t.Fatalf("error does not name the unsupported version: %v", err)
	}
	after := inventoryOf(d)
	equalStrings(t, "entities", before.Entities, after.Entities)
}

func TestToJSONUnsupportedStateErrors(t *testing.T) {
	d, err := NewDrawing()
	if err != nil {
		t.Fatal(err)
	}
	d.SetTargetVersion("AC9999")
	if _, err := d.ToJSON(nil); err == nil {
		t.Fatal("ToJSON accepted an unsupported target version")
	}

	var nilDrawing *drawing.Drawing
	if _, err := nilDrawing.ToJSON(nil); err == nil {
		t.Fatal("ToJSON accepted a nil drawing")
	}
}
