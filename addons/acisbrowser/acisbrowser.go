// Package acisbrowser exposes ACIS body data for inspection without a GUI.
package acisbrowser

import (
	"sort"

	"github.com/edanko/dxf/addons"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

// Capability registers the ACIS browser inspection add-on.
type Capability struct{}

func (Capability) Name() string { return "acisbrowser" }

func (Capability) Description() string {
	return "GUI-free ACIS text and binary body inspection data"
}

func init() {
	_ = addons.Register(Capability{}.Name(), Capability{})
}

// Entity describes the ACIS payload carried by one drawing entity.
type Entity struct {
	Name     string              `json:"name"`
	Handle   string              `json:"handle,omitempty"`
	Type     string              `json:"type"`
	Format   string              `json:"format"`
	Lines    []string            `json:"lines,omitempty"`
	Binary   []byte              `json:"binary,omitempty"`
	Metadata entity.ACISMetadata `json:"metadata"`
}

// Extract returns ACIS payloads from all ACIS-bearing entities in layout and
// document order. Entities without ACIS data are skipped.
func Extract(d *drawing.Drawing) []Entity {
	if d == nil {
		return []Entity{}
	}
	var result []Entity
	for _, layout := range d.Layouts() {
		for _, e := range d.LayoutEntities(layout.Name) {
			if data, ok := ExtractEntity(e); ok {
				result = append(result, data)
			}
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Handle < result[j].Handle
	})
	return result
}

// ExtractEntity extracts ACIS data from Region, Body and Solid3d entities.
func ExtractEntity(e entity.Entity) (Entity, bool) {
	if e == nil {
		return Entity{}, false
	}
	var (
		info   entity.ACISMetadata
		lines  []string
		binary []byte
	)
	switch value := e.(type) {
	case *entity.Region:
		info = value.ACISInfo()
		lines = value.SATLines()
		binary = value.ACISBinaryData()
	case *entity.Body:
		info = value.ACISInfo()
		lines = value.SATLines()
		binary = value.ACISBinaryData()
	case *entity.Solid3d:
		info = value.ACISInfo()
		lines = value.SATLines()
		binary = value.ACISBinaryData()
	default:
		return Entity{}, false
	}
	if len(lines) == 0 && len(binary) == 0 && !info.HasPayload() {
		return Entity{}, false
	}
	format := info.Format.String()
	if info.Format == entity.ACISFormatUnknown {
		if len(binary) > 0 {
			format = entity.ACISFormatSAB.String()
		} else {
			format = entity.ACISFormatSAT.String()
		}
	}
	return Entity{
		Name:     e.DXFType() + "[" + e.Handle() + "]",
		Handle:   e.Handle(),
		Type:     e.DXFType(),
		Format:   format,
		Lines:    append([]string(nil), lines...),
		Binary:   append([]byte(nil), binary...),
		Metadata: info,
	}, true
}
