package geo_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/edanko/dxf/addons/geo"
	"github.com/edanko/dxf/drawing"
)

func TestExportGeoJSON(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddLine(0, 0, 0, 10, 0, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddPoint(5, 5, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddLwPolyline(false, []float64{0, 0, 1, 0, 1, 1}); err != nil {
		t.Fatal(err)
	}

	data, err := geo.ExportGeoJSON(d, nil)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Type     string `json:"type"`
		Features []struct {
			Type     string `json:"type"`
			Geometry struct {
				Type        string          `json:"type"`
				Coordinates json.RawMessage `json:"coordinates"`
			} `json:"geometry"`
		} `json:"features"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Type != "FeatureCollection" || len(doc.Features) != 3 {
		t.Fatalf("features: %d type %s", len(doc.Features), doc.Type)
	}
	types := map[string]bool{}
	for _, f := range doc.Features {
		types[f.Geometry.Type] = true
	}
	if !types["LineString"] || !types["Point"] {
		t.Fatalf("geometry types: %v", types)
	}
	if !strings.Contains(string(data), "\n") {
		t.Fatal("expected indented output")
	}
}

func TestExportGeoJSONCRS(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddPoint(1, 2, 0); err != nil {
		t.Fatal(err)
	}
	data, err := geo.ExportGeoJSON(d, &geo.Options{CRS: func(x, y float64) (float64, float64) { return x * 2, y * 2 }})
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Features []struct {
			Geometry struct {
				Coordinates []float64 `json:"coordinates"`
			} `json:"geometry"`
		} `json:"features"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Features) != 1 || len(doc.Features[0].Geometry.Coordinates) != 2 ||
		doc.Features[0].Geometry.Coordinates[0] != 2 || doc.Features[0].Geometry.Coordinates[1] != 4 {
		t.Fatalf("CRS transform not applied: %s", data)
	}
}
