package drawing

import (
	"math"

	"github.com/edanko/dxf/entity"
)

type BatchProcessor struct {
	drawing *Drawing
}

func NewBatchProcessor(d *Drawing) *BatchProcessor {
	return &BatchProcessor{drawing: d}
}

func (bp *BatchProcessor) Translate(dx, dy, dz float64) error {
	entities := bp.drawing.Entities()
	for _, e := range entities {
		bp.translateEntity(e, dx, dy, dz)
	}
	return nil
}

func (bp *BatchProcessor) translateEntity(e entity.Entity, dx, dy, dz float64) {
	switch ent := e.(type) {
	case *entity.Line:
		for i := 0; i < len(ent.Start) && i < 3; i++ {
			switch i {
			case 0:
				ent.Start[i] += dx
			case 1:
				ent.Start[i] += dy
			case 2:
				ent.Start[i] += dz
			}
		}
		for i := 0; i < len(ent.End) && i < 3; i++ {
			switch i {
			case 0:
				ent.End[i] += dx
			case 1:
				ent.End[i] += dy
			case 2:
				ent.End[i] += dz
			}
		}
	case *entity.Circle:
		for i := 0; i < len(ent.Center) && i < 3; i++ {
			switch i {
			case 0:
				ent.Center[i] += dx
			case 1:
				ent.Center[i] += dy
			case 2:
				ent.Center[i] += dz
			}
		}
	case *entity.Point:
		for i := 0; i < len(ent.Coord) && i < 3; i++ {
			switch i {
			case 0:
				ent.Coord[i] += dx
			case 1:
				ent.Coord[i] += dy
			case 2:
				ent.Coord[i] += dz
			}
		}
	}
}

func (bp *BatchProcessor) Scale(factor float64) error {
	entities := bp.drawing.Entities()
	for _, e := range entities {
		bp.scaleEntity(e, factor)
	}
	return nil
}

func (bp *BatchProcessor) scaleEntity(e entity.Entity, factor float64) {
	switch ent := e.(type) {
	case *entity.Line:
		for i := 0; i < len(ent.Start) && i < 3; i++ {
			ent.Start[i] *= factor
		}
		for i := 0; i < len(ent.End) && i < 3; i++ {
			ent.End[i] *= factor
		}
	case *entity.Circle:
		ent.Radius *= factor
		for i := 0; i < len(ent.Center) && i < 3; i++ {
			ent.Center[i] *= factor
		}
	case *entity.Arc:
		ent.Radius *= factor
	}
}

func (bp *BatchProcessor) Rotate(angle float64) error {
	sin := math.Sin(angle)
	cos := math.Cos(angle)
	entities := bp.drawing.Entities()
	for _, e := range entities {
		bp.rotateEntity(e, sin, cos)
	}
	return nil
}

func (bp *BatchProcessor) rotateEntity(e entity.Entity, sin, cos float64) {
	switch ent := e.(type) {
	case *entity.Line:
		bp.rotatePoint(ent.Start, sin, cos)
		bp.rotatePoint(ent.End, sin, cos)
	case *entity.Circle:
		bp.rotatePoint(ent.Center, sin, cos)
	}
}

func (bp *BatchProcessor) rotatePoint(point []float64, sin, cos float64) {
	if len(point) >= 2 {
		x := point[0]
		y := point[1]
		point[0] = x*cos - y*sin
		point[1] = x*sin + y*cos
	}
}

func (bp *BatchProcessor) FilterByLayer(layerName string) []entity.Entity {
	var result []entity.Entity
	entities := bp.drawing.Entities()
	for _, e := range entities {
		if layer := e.Layer(); layer != nil && layer.Name() == layerName {
			result = append(result, e)
		}
	}
	return result
}

func (bp *BatchProcessor) FilterByType(entityType string) []entity.Entity {
	var result []entity.Entity
	entities := bp.drawing.Entities()
	for _, e := range entities {
		if e.DXFType() == entityType {
			result = append(result, e)
		}
	}
	return result
}

func (bp *BatchProcessor) FilterByBoundingBox(minX, minY, maxX, maxY float64) []entity.Entity {
	var result []entity.Entity
	entities := bp.drawing.Entities()
	for _, e := range entities {
		bboxMin, bboxMax := e.BBox()
		if len(bboxMin) >= 2 && len(bboxMax) >= 2 {
			if bboxMin[0] >= minX && bboxMin[1] >= minY &&
				bboxMax[0] <= maxX && bboxMax[1] <= maxY {
				result = append(result, e)
			}
		}
	}
	return result
}

func (bp *BatchProcessor) DeleteByLayer(layerName string) int {
	count := 0
	entities := bp.drawing.Entities()
	for _, e := range entities {
		if layer := e.Layer(); layer != nil && layer.Name() == layerName {
			handle := e.Handle()
			if handle != "" {
				bp.drawing.DeleteEntity(handle)
				count++
			}
		}
	}
	return count
}

func (bp *BatchProcessor) DeleteByType(entityType string) int {
	count := 0
	entities := bp.drawing.Entities()
	for _, e := range entities {
		if e.DXFType() == entityType {
			handle := e.Handle()
			if handle != "" {
				bp.drawing.DeleteEntity(handle)
				count++
			}
		}
	}
	return count
}

func (bp *BatchProcessor) CountByLayer() map[string]int {
	counts := make(map[string]int)
	entities := bp.drawing.Entities()
	for _, e := range entities {
		if layer := e.Layer(); layer != nil {
			counts[layer.Name()]++
		}
	}
	return counts
}

func (bp *BatchProcessor) CountByType() map[string]int {
	counts := make(map[string]int)
	entities := bp.drawing.Entities()
	for _, e := range entities {
		counts[e.DXFType()]++
	}
	return counts
}

func (bp *BatchProcessor) GroupByLayer() map[string][]entity.Entity {
	groups := make(map[string][]entity.Entity)
	entities := bp.drawing.Entities()
	for _, e := range entities {
		layerName := ""
		if layer := e.Layer(); layer != nil {
			layerName = layer.Name()
		}
		groups[layerName] = append(groups[layerName], e)
	}
	return groups
}

func (bp *BatchProcessor) GroupByType() map[string][]entity.Entity {
	groups := make(map[string][]entity.Entity)
	entities := bp.drawing.Entities()
	for _, e := range entities {
		etype := e.DXFType()
		groups[etype] = append(groups[etype], e)
	}
	return groups
}

type BatchStats struct {
	TotalCount  int
	LayerCounts map[string]int
	TypeCounts  map[string]int
}

func (bp *BatchProcessor) GetStats() *BatchStats {
	stats := &BatchStats{
		LayerCounts: make(map[string]int),
		TypeCounts:  make(map[string]int),
	}

	entities := bp.drawing.Entities()
	stats.TotalCount = len(entities)

	for _, e := range entities {
		stats.TypeCounts[e.DXFType()]++
		if layer := e.Layer(); layer != nil {
			stats.LayerCounts[layer.Name()]++
		}
	}

	return stats
}
