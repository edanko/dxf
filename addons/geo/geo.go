// Package geo exports drawing geometry as GeoJSON.
package geo

import (
	"encoding/json"
	"fmt"

	"github.com/edanko/dxf/addons"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

func init() { _ = addons.Register("geo", capability{}) }

type capability struct{}

func (capability) Name() string        { return "geo" }
func (capability) Description() string { return "geospatial export (GeoJSON)" }

// CRSFunc transforms model coordinates into export coordinates.
type CRSFunc func(x, y float64) (float64, float64)

// Options configures the GeoJSON export.
type Options struct {
	CRS    CRSFunc
	Layers []string
}

// Feature is one GeoJSON feature.
type Feature struct {
	Type       string                 `json:"type"`
	Geometry   map[string]interface{} `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

// ExportGeoJSON renders the drawing entities as a GeoJSON feature collection.
func ExportGeoJSON(d *drawing.Drawing, opts *Options) ([]byte, error) {
	if d == nil {
		return nil, fmt.Errorf("geo: nil drawing")
	}
	var crs CRSFunc
	layerFilter := map[string]bool{}
	if opts != nil {
		crs = opts.CRS
		for _, l := range opts.Layers {
			layerFilter[l] = true
		}
	}
	transform := func(x, y float64) (float64, float64) {
		if crs != nil {
			return crs(x, y)
		}
		return x, y
	}
	feats := []Feature{}
	for _, e := range d.Entities() {
		layer := ""
		if ln, ok := e.(interface{ LayerName() string }); ok {
			layer = ln.LayerName()
		}
		if len(layerFilter) > 0 && !layerFilter[layer] {
			continue
		}
		var geom map[string]interface{}
		switch t := e.(type) {
		case *entity.Line:
			x1, y1 := transform(t.Start[0], t.Start[1])
			x2, y2 := transform(t.End[0], t.End[1])
			geom = map[string]interface{}{"type": "LineString", "coordinates": [][]float64{{x1, y1}, {x2, y2}}}
		case *entity.LwPolyline:
			pts := make([][]float64, 0, len(t.Vertices))
			for _, v := range t.Vertices {
				x, y := transform(v[0], v[1])
				pts = append(pts, []float64{x, y})
			}
			if t.Closed && len(pts) > 0 {
				pts = append(pts, pts[0])
			}
			geom = map[string]interface{}{"type": "LineString", "coordinates": pts}
		case *entity.Point:
			x, y := transform(t.Coord[0], t.Coord[1])
			geom = map[string]interface{}{"type": "Point", "coordinates": []float64{x, y}}
		case *entity.Insert:
			p := t.InsertPoint()
			x, y := transform(p.X(), p.Y())
			geom = map[string]interface{}{"type": "Point", "coordinates": []float64{x, y}}
		default:
			continue
		}
		feats = append(feats, Feature{
			Type:       "Feature",
			Geometry:   geom,
			Properties: map[string]interface{}{"layer": layer, "type": e.DXFType()},
		})
	}
	return json.MarshalIndent(map[string]interface{}{"type": "FeatureCollection", "features": feats}, "", "  ")
}
