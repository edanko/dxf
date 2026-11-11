package importer_test

import (
	"testing"

	"github.com/edanko/dxf/addons/importer"
	"github.com/edanko/dxf/entity"
)

const testSVG = `<?xml version="1.0"?>
<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
  <line x1="0" y1="0" x2="10" y2="10"/>
  <rect x="1" y="1" width="5" height="3"/>
  <circle cx="7" cy="7" r="2"/>
  <polyline points="0,0 2,2 4,0"/>
  <path d="M 0 5 L 3 5 L 3 8 Z"/>
</svg>`

func TestImportSVG(t *testing.T) {
	d, err := importer.ImportSVG([]byte(testSVG))
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, e := range d.Entities() {
		counts[e.DXFType()]++
	}
	if counts["LINE"] != 1 {
		t.Fatalf("lines: %d", counts["LINE"])
	}
	if counts["CIRCLE"] != 1 {
		t.Fatalf("circles: %d", counts["CIRCLE"])
	}
	if counts["LWPOLYLINE"] != 3 {
		t.Fatalf("polylines: %d", counts["LWPOLYLINE"])
	}
	var circle *entity.Circle
	for _, e := range d.Entities() {
		if c, ok := e.(*entity.Circle); ok {
			circle = c
		}
	}
	if circle == nil || circle.Radius != 2 {
		t.Fatalf("circle: %v", circle)
	}
}

func TestImportSVGMalformed(t *testing.T) {
	if _, err := importer.ImportSVG([]byte("<svg><line")); err == nil {
		t.Fatal("malformed SVG must error")
	}
}
