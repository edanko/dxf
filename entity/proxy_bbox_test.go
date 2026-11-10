package entity

import (
	"testing"

	"github.com/edanko/dxf/lldxf"
)

// TestProxyBBoxFromDecodedGraphics verifies that a proxy without stored
// extents derives its box from decodable proxy geometry, and that a proxy with
// no geometry keeps the empty representation.
func TestProxyBBoxFromDecodedGraphics(t *testing.T) {
	Register("LINE", func(tags lldxf.Tags) (Entity, error) {
		line := NewLine()
		for _, tag := range tags {
			v, _ := tag.Value().(float64)
			switch int(tag.Code()) {
			case 10:
				line.Start[0] = v
			case 20:
				line.Start[1] = v
			case 30:
				line.Start[2] = v
			case 11:
				line.End[0] = v
			case 21:
				line.End[1] = v
			case 31:
				line.End[2] = v
			}
		}
		return line, nil
	})
	payload := EncodeBinaryTags([]BinaryTag{
		{Code: 0, Value: "LINE"},
		{Code: 5, Value: "P1"},
		{Code: 100, Value: "AcDbEntity"},
		{Code: 8, Value: "0"},
		{Code: 10, Value: 2.0},
		{Code: 20, Value: 3.0},
		{Code: 30, Value: 0.0},
		{Code: 11, Value: 6.0},
		{Code: 21, Value: 8.0},
		{Code: 31, Value: 0.0},
	})
	proxy := NewAcadProxyEntity()
	proxy.SetEntityData(payload)
	min, max := proxy.BBox()
	if min[0] != 2 || min[1] != 3 || max[0] != 6 || max[1] != 8 {
		t.Fatalf("decoded proxy bbox = %v %v", min, max)
	}

	empty := NewAcadProxyEntity()
	min, max = empty.BBox()
	if min[0] != 0 || min[1] != 0 || max[0] != 0 || max[1] != 0 {
		t.Fatalf("geometry-less proxy bbox = %v %v", min, max)
	}
}
