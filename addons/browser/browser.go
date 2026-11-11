// Package browser exposes GUI-free inspection data for DXF drawings.
//
// The data structures are intentionally plain Go values so CLI, web, or GUI
// front ends can render them without depending on a particular toolkit.
package browser

import (
	"sort"

	"github.com/edanko/dxf/addons"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

func init() {
	_ = addons.Register("browser", Capability{})
}

// Capability describes the browser add-on in the add-on registry.
type Capability struct{}

// Name implements addons.Capability.
func (Capability) Name() string { return "browser" }

// Description implements addons.Capability.
func (Capability) Description() string {
	return "GUI-free DXF document and entity inspection data"
}

// BoundingBox is a three-dimensional bounding box in world coordinates.
type BoundingBox struct {
	Min []float64 `json:"min"`
	Max []float64 `json:"max"`
}

// Entity is one browsable drawing entity.
type Entity struct {
	Handle string       `json:"handle,omitempty"`
	Type   string       `json:"type"`
	Layer  string       `json:"layer,omitempty"`
	BBox   *BoundingBox `json:"bbox,omitempty"`
}

// Layout groups the entities of one model or paper-space layout.
type Layout struct {
	Name        string              `json:"name"`
	Type        string              `json:"type"`
	TabOrder    int                 `json:"tab_order"`
	Entities    []Entity            `json:"entities"`
	ByType      map[string][]Entity `json:"by_type"`
	EntityCount int                 `json:"entity_count"`
}

// Document is the top-level browser inspection result.
type Document struct {
	Version string   `json:"version,omitempty"`
	Layouts []Layout `json:"layouts"`
}

// Extract builds GUI-free inspection data for every layout in d.
func Extract(d *drawing.Drawing) Document {
	result := Document{Layouts: []Layout{}}
	if d == nil {
		return result
	}
	result.Version = d.SourceVersion()

	layouts := d.Layouts()
	result.Layouts = make([]Layout, 0, len(layouts))
	for _, info := range layouts {
		layout := Layout{
			Name:     info.Name,
			Type:     info.Type,
			TabOrder: info.TabOrder,
			ByType:   map[string][]Entity{},
		}
		for _, e := range d.LayoutEntities(info.Name) {
			item := Entity{
				Handle: e.Handle(),
				Type:   e.DXFType(),
				Layer:  entityLayerName(e),
				BBox:   entityBBox(e),
			}
			layout.Entities = append(layout.Entities, item)
			layout.ByType[item.Type] = append(layout.ByType[item.Type], item)
		}
		layout.EntityCount = len(layout.Entities)
		sort.SliceStable(layout.Entities, func(i, j int) bool {
			if layout.Entities[i].Type != layout.Entities[j].Type {
				return layout.Entities[i].Type < layout.Entities[j].Type
			}
			return layout.Entities[i].Handle < layout.Entities[j].Handle
		})
		result.Layouts = append(result.Layouts, layout)
	}
	return result
}

// EntitiesByType groups all entities in d by entity type, preserving layout and
// entity order within each group.
func EntitiesByType(d *drawing.Drawing) map[string][]Entity {
	groups := map[string][]Entity{}
	for _, layout := range Extract(d).Layouts {
		for _, item := range layout.Entities {
			groups[item.Type] = append(groups[item.Type], item)
		}
	}
	return groups
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

func entityBBox(e entity.Entity) *BoundingBox {
	mins, maxs := e.BBox()
	if len(mins) < 2 || len(maxs) < 2 {
		return nil
	}
	return &BoundingBox{Min: mins, Max: maxs}
}
