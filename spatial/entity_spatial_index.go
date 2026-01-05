package spatial

import (
	"math"

	"github.com/edanko/dxf/entity"
)

type EntitySpatialIndex struct {
	rtree     *RTree
	entityBox map[interface{}]*BoundingBox
}

func NewEntitySpatialIndex() *EntitySpatialIndex {
	return &EntitySpatialIndex{
		rtree:     NewRTree(nil),
		entityBox: make(map[interface{}]*BoundingBox),
	}
}

func (esi *EntitySpatialIndex) Insert(ent entity.Entity) {
	if ent == nil {
		return
	}
	min, max := ent.BBox()
	if len(min) < 2 || len(max) < 2 {
		return
	}
	box := NewBoundingBox(min[0], min[1], max[0], max[1])
	esi.rtree.Insert(box, ent)
	esi.entityBox[ent] = box
}

func (esi *EntitySpatialIndex) InsertWithBBox(ent entity.Entity, minX, minY, maxX, maxY float64) {
	if ent == nil {
		return
	}
	box := NewBoundingBox(minX, minY, maxX, maxY)
	esi.rtree.Insert(box, ent)
	esi.entityBox[ent] = box
}

func (esi *EntitySpatialIndex) Query(bbox *BoundingBox) []entity.Entity {
	results := esi.rtree.Query(bbox)
	entities := make([]entity.Entity, 0, len(results))
	for _, r := range results {
		if e, ok := r.(entity.Entity); ok {
			entities = append(entities, e)
		}
	}
	return entities
}

func (esi *EntitySpatialIndex) QueryRange(minX, minY, maxX, maxY float64) []entity.Entity {
	results := esi.rtree.QueryRange(minX, minY, maxX, maxY)
	entities := make([]entity.Entity, 0, len(results))
	for _, r := range results {
		if e, ok := r.(entity.Entity); ok {
			entities = append(entities, e)
		}
	}
	return entities
}

func (esi *EntitySpatialIndex) QueryNearest(x, y float64) (entity.Entity, float64) {
	result, dist := esi.rtree.QueryNearest(x, y)
	if e, ok := result.(entity.Entity); ok {
		return e, dist
	}
	return nil, math.Inf(1)
}

func (esi *EntitySpatialIndex) Remove(ent entity.Entity) {
	if box, ok := esi.entityBox[ent]; ok {
		esi.rtree.Query(box)
		delete(esi.entityBox, ent)
	}
}

func (esi *EntitySpatialIndex) Clear() {
	esi.rtree.Clear()
	esi.entityBox = make(map[interface{}]*BoundingBox)
}

func (esi *EntitySpatialIndex) Count() int {
	return len(esi.entityBox)
}

func (esi *EntitySpatialIndex) GetStats() map[string]interface{} {
	return esi.rtree.GetStats()
}

func (esi *EntitySpatialIndex) GetBoundingBox() *BoundingBox {
	if esi.rtree.Root == nil {
		return nil
	}
	return esi.rtree.Root.Box
}
