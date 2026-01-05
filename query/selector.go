package query

import (
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/entity"
	dxfmath "github.com/edanko/dxf/math"
)

// SpatialFilter represents spatial query filters
type SpatialFilter interface {
	Contains(point dxfmath.Vec3) bool
	Intersects(bbox [2]dxfmath.Vec3) bool
}

// BoundingBoxFilter implements rectangular spatial filtering
type BoundingBoxFilter struct {
	Min, Max dxfmath.Vec3
}

func NewBoundingBoxFilter(minX, minY, minZ, maxX, maxY, maxZ float64) *BoundingBoxFilter {
	return &BoundingBoxFilter{
		Min: dxfmath.NewVec3(minX, minY, minZ),
		Max: dxfmath.NewVec3(maxX, maxY, maxZ),
	}
}

func (f *BoundingBoxFilter) Contains(point dxfmath.Vec3) bool {
	return point.X() >= f.Min.X() && point.X() <= f.Max.X() &&
		point.Y() >= f.Min.Y() && point.Y() <= f.Max.Y() &&
		point.Z() >= f.Min.Z() && point.Z() <= f.Max.Z()
}

func (f *BoundingBoxFilter) Intersects(bbox [2]dxfmath.Vec3) bool {
	return !(bbox[0].X() > f.Max.X() || bbox[1].X() < f.Min.X() ||
		bbox[0].Y() > f.Max.Y() || bbox[1].Y() < f.Min.Y() ||
		bbox[0].Z() > f.Max.Z() || bbox[1].Z() < f.Min.Z())
}

// PolygonFilter implements polygon-based spatial filtering
type PolygonFilter struct {
	Vertices []dxfmath.Vec2
}

func NewPolygonFilter(vertices []dxfmath.Vec2) *PolygonFilter {
	return &PolygonFilter{Vertices: vertices}
}

func (f *PolygonFilter) Contains(point dxfmath.Vec3) bool {
	// Use 2D point for polygon containment test
	pt := dxfmath.NewVec2(point.X(), point.Y())
	return pointInPolygon(pt, f.Vertices)
}

func (f *PolygonFilter) Intersects(bbox [2]dxfmath.Vec3) bool {
	// Check if any corner of the bbox is inside the polygon
	corners := []dxfmath.Vec2{
		dxfmath.NewVec2(bbox[0].X(), bbox[0].Y()),
		dxfmath.NewVec2(bbox[1].X(), bbox[0].Y()),
		dxfmath.NewVec2(bbox[1].X(), bbox[1].Y()),
		dxfmath.NewVec2(bbox[0].X(), bbox[1].Y()),
	}

	for _, corner := range corners {
		if f.Contains(dxfmath.NewVec3(corner.X(), corner.Y(), 0)) {
			return true
		}
	}

	// Check if any polygon vertex is inside the bbox
	for _, vertex := range f.Vertices {
		if vertex.X() >= bbox[0].X() && vertex.X() <= bbox[1].X() &&
			vertex.Y() >= bbox[0].Y() && vertex.Y() <= bbox[1].Y() {
			return true
		}
	}

	return false
}

// pointInPolygon tests if a point is inside a polygon using ray casting algorithm
func pointInPolygon(point dxfmath.Vec2, polygon []dxfmath.Vec2) bool {
	if len(polygon) < 3 {
		return false
	}

	x, y := point.X(), point.Y()
	n := len(polygon)
	inside := false

	j := n - 1
	for i := 0; i < n; i++ {
		xi, yi := polygon[i].X(), polygon[i].Y()
		xj, yj := polygon[j].X(), polygon[j].Y()

		intersect := ((yi > y) != (yj > y)) &&
			(x < (xj-xi)*(y-yi)/(yj-yi)+xi)

		if intersect {
			inside = !inside
		}
		j = i
	}

	return inside
}

// EntitySelector provides Python ezdxf-style entity selection
type EntitySelector struct {
	query *EntityQuery
}

// NewEntitySelector creates a new entity selector
func NewEntitySelector(query *EntityQuery) *EntitySelector {
	return &EntitySelector{query: query}
}

// Layer selects entities by layer name (Python ezdxf style: query.layer == "MyLayer")
func (s *EntitySelector) Layer(layerName string) *EntityQuery {
	return s.query.Filter(func(ent entity.Entity) bool {
		if layer := ent.Layer(); layer != nil {
			return layer.Name() == layerName
		}
		return false
	})
}

// LayerIn selects entities from any of the specified layers
func (s *EntitySelector) LayerIn(layerNames ...string) *EntityQuery {
	layerSet := make(map[string]bool)
	for _, name := range layerNames {
		layerSet[name] = true
	}

	return s.query.Filter(func(ent entity.Entity) bool {
		if layer := ent.Layer(); layer != nil {
			return layerSet[layer.Name()]
		}
		return false
	})
}

// Color selects entities by color index
func (s *EntitySelector) Color(colorIndex int) *EntityQuery {
	return s.query.Filter(func(ent entity.Entity) bool {
		if colorable, ok := ent.(interface{ Color() color.ColorNumber }); ok {
			return int(colorable.Color()) == colorIndex
		}
		return false
	})
}

// RGB selects entities by RGB color (for entities with RGB support)
func (s *EntitySelector) RGB(r, g, b uint8) *EntityQuery {
	return s.query.Filter(func(ent entity.Entity) bool {
		if rgbEntity, ok := ent.(interface{ RGB() (uint8, uint8, uint8) }); ok {
			er, eg, eb := rgbEntity.RGB()
			return er == r && eg == g && eb == b
		}
		return false
	})
}

// Type selects entities by type name
func (s *EntitySelector) Type(typeName string) *EntityQuery {
	return s.query.Select(typeName)
}

// Types selects entities from any of the specified types
func (s *EntitySelector) Types(typeNames ...string) *EntityQuery {
	return s.query.Select(typeNames...)
}

// Handle selects entities by handle
func (s *EntitySelector) Handle(handle string) *EntityQuery {
	return s.query.Filter(func(ent entity.Entity) bool {
		return ent.Handle() == handle
	})
}

// Point selects entities that contain a specific point
func (s *EntitySelector) Point(x, y, z float64) *EntityQuery {
	return s.query.Filter(func(ent entity.Entity) bool {
		// Check if entity contains the point
		min, max := ent.BBox()
		if len(min) < 3 || len(max) < 3 {
			return false
		}

		return x >= min[0] && x <= max[0] &&
			y >= min[1] && y <= max[1] &&
			z >= min[2] && z <= max[2]
	})
}

// Within selects entities within a bounding box
func (s *EntitySelector) Within(minX, minY, minZ, maxX, maxY, maxZ float64) *EntityQuery {
	return s.query.Filter(func(ent entity.Entity) bool {
		min, max := ent.BBox()
		if len(min) < 3 || len(max) < 3 {
			return false
		}

		// Check if entity bbox intersects with query bbox
		return !(max[0] < minX || min[0] > maxX ||
			max[1] < minY || min[1] > maxY ||
			max[2] < minZ || min[2] > maxZ)
	})
}

// WithinPolygon selects entities within a polygon
func (s *EntitySelector) WithinPolygon(vertices []dxfmath.Vec2) *EntityQuery {
	return s.query.Filter(func(ent entity.Entity) bool {
		min, max := ent.BBox()
		if len(min) < 3 || len(max) < 3 {
			return false
		}

		// Simple bounding box check against polygon
		// This is a simplified approach - full polygon intersection would be more complex
		entMin := dxfmath.NewVec2(min[0], min[1])
		entMax := dxfmath.NewVec2(max[0], max[1])

		// Check if any corner of entity bbox is inside polygon
		corners := []dxfmath.Vec2{
			entMin,
			dxfmath.NewVec2(max[0], min[1]),
			entMax,
			dxfmath.NewVec2(min[0], max[1]),
		}

		for _, corner := range corners {
			if pointInPolygon(corner, vertices) {
				return true
			}
		}

		return false
	})
}

// Distance selects entities within a certain distance from a point
func (s *EntitySelector) Distance(centerX, centerY, centerZ, maxDistance float64) *EntityQuery {
	maxDistSq := maxDistance * maxDistance

	return s.query.Filter(func(ent entity.Entity) bool {
		min, max := ent.BBox()
		if len(min) < 3 || len(max) < 3 {
			return false
		}

		// Simple distance check to closest corner of bbox
		closestX := math.Max(min[0], math.Min(centerX, max[0]))
		closestY := math.Max(min[1], math.Min(centerY, max[1]))
		closestZ := math.Max(min[2], math.Min(centerZ, max[2]))

		dx := centerX - closestX
		dy := centerY - closestY
		dz := centerZ - closestZ
		distanceSq := dx*dx + dy*dy + dz*dz

		return distanceSq <= maxDistSq
	})
}

// Length selects entities with length within a range (for linear entities)
func (s *EntitySelector) Length(minLength, maxLength float64) *EntityQuery {
	return s.query.Filter(func(ent entity.Entity) bool {
		// Calculate entity length
		min, max := ent.BBox()
		if len(min) < 3 || len(max) < 3 {
			return false
		}

		// Simple approximation: use diagonal of bbox
		dx := max[0] - min[0]
		dy := max[1] - min[1]
		dz := max[2] - min[2]
		length := math.Sqrt(dx*dx + dy*dy + dz*dz)

		return length >= minLength && length <= maxLength
	})
}

// Text selects TEXT entities with specific text content
func (s *EntitySelector) Text(text string) *EntityQuery {
	return s.query.Filter(func(ent entity.Entity) bool {
		if textEnt, ok := ent.(*entity.Text); ok {
			return strings.Contains(textEnt.Value, text)
		}
		if mtextEnt, ok := ent.(*entity.MText); ok {
			return strings.Contains(mtextEnt.GetText(), text)
		}
		return false
	})
}

// TextRegex selects TEXT entities matching a regular expression
func (s *EntitySelector) TextRegex(pattern string) *EntityQuery {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return NewEntityQuery() // Return empty query on invalid regex
	}

	return s.query.Filter(func(ent entity.Entity) bool {
		if textEnt, ok := ent.(*entity.Text); ok {
			return re.MatchString(textEnt.Value)
		}
		if mtextEnt, ok := ent.(*entity.MText); ok {
			return re.MatchString(mtextEnt.GetText())
		}
		return false
	})
}

// Attribute provides custom attribute filtering (Python ezdxf style)
func (s *EntitySelector) Attribute(name string, value interface{}) *EntityQuery {
	return s.query.Filter(func(ent entity.Entity) bool {
		attrValue := getEntityAttribute(ent, name)
		if attrValue == "" {
			return false
		}

		switch v := value.(type) {
		case string:
			return attrValue == v
		case int:
			if intVal, err := strconv.Atoi(attrValue); err == nil {
				return intVal == v
			}
		case float64:
			if floatVal, err := strconv.ParseFloat(attrValue, 64); err == nil {
				return math.Abs(floatVal-v) < 1e-6
			}
		}
		return false
	})
}

// HasAttribute selects entities that have a specific attribute
func (s *EntitySelector) HasAttribute(name string) *EntityQuery {
	return s.query.Filter(func(ent entity.Entity) bool {
		return getEntityAttribute(ent, name) != ""
	})
}

// Filter provides custom predicate filtering
func (s *EntitySelector) Filter(predicate func(entity.Entity) bool) *EntityQuery {
	return s.query.Filter(predicate)
}

// Logical operations for combining queries

// Or returns entities that match any of the given selectors
func Or(queries ...*EntityQuery) *EntityQuery {
	if len(queries) == 0 {
		return NewEntityQuery()
	}

	// Collect all entities from all queries
	allEntities := make(map[string]entity.Entity) // key by handle
	for _, query := range queries {
		for _, ent := range query.ToSlice() {
			allEntities[ent.Handle()] = ent
		}
	}

	// Convert back to slice
	result := make([]entity.Entity, 0, len(allEntities))
	for _, ent := range allEntities {
		result = append(result, ent)
	}

	return NewEntityQuery(result...)
}

// And returns entities that match all of the given selectors
func And(queries ...*EntityQuery) *EntityQuery {
	if len(queries) == 0 {
		return NewEntityQuery()
	}

	// Start with first query
	result := queries[0]
	handleCount := make(map[string]int)

	// Count occurrences of each handle
	for _, query := range queries {
		for _, ent := range query.ToSlice() {
			handleCount[ent.Handle()]++
		}
	}

	// Only include entities that appear in all queries
	filtered := make([]entity.Entity, 0)
	for _, ent := range result.ToSlice() {
		if handleCount[ent.Handle()] == len(queries) {
			filtered = append(filtered, ent)
		}
	}

	return NewEntityQuery(filtered...)
}

// Not returns entities from the first query that are not in any of the other queries
func Not(mainQuery *EntityQuery, excludeQueries ...*EntityQuery) *EntityQuery {
	excludeHandles := make(map[string]bool)
	for _, query := range excludeQueries {
		for _, ent := range query.ToSlice() {
			excludeHandles[ent.Handle()] = true
		}
	}

	return mainQuery.Filter(func(ent entity.Entity) bool {
		return !excludeHandles[ent.Handle()]
	})
}

// Sorting operations

// SortByLayer sorts entities by layer name
func (s *EntitySelector) SortByLayer() *EntityQuery {
	result := make([]entity.Entity, len(s.query.entities))
	copy(result, s.query.entities)

	// Simple bubble sort for demonstration
	n := len(result)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			layer1 := getLayerName(result[j])
			layer2 := getLayerName(result[j+1])
			if layer1 > layer2 {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}

	return NewEntityQuery(result...)
}

// SortByHandle sorts entities by handle
func (s *EntitySelector) SortByHandle() *EntityQuery {
	result := make([]entity.Entity, len(s.query.entities))
	copy(result, s.query.entities)

	n := len(result)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if result[j].Handle() > result[j+1].Handle() {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}

	return NewEntityQuery(result...)
}

// SortByColor sorts entities by color
func (s *EntitySelector) SortByColor() *EntityQuery {
	result := make([]entity.Entity, len(s.query.entities))
	copy(result, s.query.entities)

	n := len(result)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			color1 := getEntityColor(result[j])
			color2 := getEntityColor(result[j+1])
			if color1 > color2 {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}

	return NewEntityQuery(result...)
}

// GroupBy groups entities by the result of key function
func (s *EntitySelector) GroupBy(keyFunc func(entity.Entity) string) map[string]*EntityQuery {
	groups := make(map[string]*EntityQuery)
	for _, ent := range s.query.ToSlice() {
		key := keyFunc(ent)
		if groups[key] == nil {
			groups[key] = NewEntityQuery()
		}
		groups[key].Extend(NewEntityQuery(ent))
	}
	return groups
}

// CountBy counts entities by the result of key function
func (s *EntitySelector) CountBy(keyFunc func(entity.Entity) string) map[string]int {
	counts := make(map[string]int)
	for _, ent := range s.query.ToSlice() {
		key := keyFunc(ent)
		counts[key]++
	}
	return counts
}

// Helper functions

func getLayerName(ent entity.Entity) string {
	if layer := ent.Layer(); layer != nil {
		return layer.Name()
	}
	return ""
}

func getEntityColor(ent entity.Entity) int {
	if colorable, ok := ent.(interface{ Color() color.ColorNumber }); ok {
		return int(colorable.Color())
	}
	return 0
}
