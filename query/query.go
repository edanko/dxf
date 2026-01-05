package query

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/math"
)

// EntityQuery represents a query result container for DXF entities
type EntityQuery struct {
	entities []entity.Entity
}

// NewEntityQuery creates a new entity query with the given entities
func NewEntityQuery(entities ...entity.Entity) *EntityQuery {
	return &EntityQuery{
		entities: entities,
	}
}

// Len returns the number of entities in the query
func (q *EntityQuery) Len() int {
	return len(q.entities)
}

// Less returns true if entity at index i should sort before entity at index j
func (q *EntityQuery) Less(i, j int) bool {
	// Simple sorting by handle for now
	return q.entities[i].Handle() < q.entities[j].Handle()
}

// Swap swaps entities at indices i and j
func (q *EntityQuery) Swap(i, j int) {
	q.entities[i], q.entities[j] = q.entities[j], q.entities[i]
}

// Get returns the entity at the given index
func (q *EntityQuery) Get(index int) entity.Entity {
	if index < 0 || index >= len(q.entities) {
		return nil
	}
	return q.entities[index]
}

// getEntityType returns the type name of an entity
func getEntityType(ent entity.Entity) string {
	// Use string conversion on the entity type
	switch ent.(type) {
	case *entity.Line:
		return "LINE"
	case *entity.Circle:
		return "CIRCLE"
	case *entity.Arc:
		return "ARC"
	case *entity.Point:
		return "POINT"
	case *entity.Polyline:
		return "POLYLINE"
	case *entity.Text:
		return "TEXT"
	case *entity.MText:
		return "MTEXT"
	case *entity.Spline:
		return "SPLINE"
	case *entity.Ellipse:
		return "ELLIPSE"
	case *entity.Dimension:
		return "DIMENSION"
	case *entity.Insert:
		return "INSERT"
	case *entity.MLeader:
		return "MLEADER"
	case *entity.MLine:
		return "MLINE"
	// Leader and Xline have different type signatures - skip for now
	case *entity.Ray:
		return "RAY"
	case *entity.ThreeDFace:
		return "3DFACE"
	case *entity.Solid:
		return "SOLID"
	default:
		// Try to get the type through string representation
		return ""
	}
}

// Filter filters entities based on the given predicate
func (q *EntityQuery) Filter(predicate func(entity.Entity) bool) *EntityQuery {
	var filtered []entity.Entity
	for _, ent := range q.entities {
		if predicate(ent) {
			filtered = append(filtered, ent)
		}
	}
	return NewEntityQuery(filtered...)
}

// Select selects entities by type names
func (q *EntityQuery) Select(entityNames ...string) *EntityQuery {
	if len(entityNames) == 0 {
		return NewEntityQuery()
	}

	nameSet := make(map[string]bool)
	all := false
	exclude := make(map[string]bool)

	for _, name := range entityNames {
		name = strings.ToUpper(name)
		if name == "*" {
			all = true
		} else if strings.HasPrefix(name, "!") {
			exclude[strings.TrimPrefix(name, "!")] = true
		} else {
			nameSet[name] = true
		}
	}

	var filtered []entity.Entity
	for _, ent := range q.entities {
		entType := getEntityType(ent)

		// Handle "all" selection with exclusions
		if all {
			if !exclude[entType] {
				filtered = append(filtered, ent)
			}
			continue
		}

		// Handle explicit name selection
		if nameSet[entType] {
			filtered = append(filtered, ent)
		}
	}

	return NewEntityQuery(filtered...)
}

// Where filters entities by attribute conditions
func (q *EntityQuery) Where(conditions string) (*EntityQuery, error) {
	if conditions == "" {
		return NewEntityQuery(q.entities...), nil
	}

	predicate, err := parseAttributeQuery(conditions)
	if err != nil {
		return nil, fmt.Errorf("invalid attribute query: %w", err)
	}

	return q.Filter(predicate), nil
}

// Query executes a complete query string (entity names + attribute conditions)
func (q *EntityQuery) Query(queryStr string) (*EntityQuery, error) {
	if queryStr == "" {
		return NewEntityQuery(), nil
	}

	// Parse query string to separate entity names from attribute conditions
	entityNames, attrConditions, err := parseQueryString(queryStr)
	if err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	// First select by entity type
	result := q.Select(entityNames...)

	// Then apply attribute conditions if any
	if attrConditions != "" {
		return result.Where(attrConditions)
	}

	return result, nil
}

// First returns the first entity in the query
func (q *EntityQuery) First() entity.Entity {
	if len(q.entities) == 0 {
		return nil
	}
	return q.entities[0]
}

// Last returns the last entity in the query
func (q *EntityQuery) Last() entity.Entity {
	if len(q.entities) == 0 {
		return nil
	}
	return q.entities[len(q.entities)-1]
}

// Empty returns true if the query contains no entities
func (q *EntityQuery) Empty() bool {
	return len(q.entities) == 0
}

// ToSlice returns a slice of all entities in the query
func (q *EntityQuery) ToSlice() []entity.Entity {
	result := make([]entity.Entity, len(q.entities))
	copy(result, q.entities)
	return result
}

// Extend adds entities from another query to this query
func (q *EntityQuery) Extend(other *EntityQuery) {
	q.entities = append(q.entities, other.entities...)
}

// Each executes the given function for each entity in the query
func (q *EntityQuery) Each(fn func(entity.Entity)) {
	for _, ent := range q.entities {
		fn(ent)
	}
}

// Map applies a transformation function to each entity and returns the results
func (q *EntityQuery) Map(fn func(entity.Entity) interface{}) []interface{} {
	result := make([]interface{}, len(q.entities))
	for i, ent := range q.entities {
		result[i] = fn(ent)
	}
	return result
}

// GroupBy groups entities by the result of the key function
func (q *EntityQuery) GroupBy(keyFunc func(entity.Entity) string) map[string]*EntityQuery {
	groups := make(map[string]*EntityQuery)
	for _, ent := range q.entities {
		key := keyFunc(ent)
		if groups[key] == nil {
			groups[key] = NewEntityQuery()
		}
		groups[key].Extend(NewEntityQuery(ent))
	}
	return groups
}

// CountBy counts entities by the result of the key function
func (q *EntityQuery) CountBy(keyFunc func(entity.Entity) string) map[string]int {
	counts := make(map[string]int)
	for _, ent := range q.entities {
		key := keyFunc(ent)
		counts[key]++
	}
	return counts
}

// Unique returns a query with unique entities (by handle)
func (q *EntityQuery) Unique() *EntityQuery {
	seen := make(map[string]bool)
	var unique []entity.Entity

	for _, ent := range q.entities {
		handle := ent.Handle()
		if !seen[handle] {
			seen[handle] = true
			unique = append(unique, ent)
		}
	}

	return NewEntityQuery(unique...)
}

// Bounds returns the bounding box of all entities in the query
func (q *EntityQuery) Bounds() (min, max math.Vec3, found bool) {
	if len(q.entities) == 0 {
		return math.Vec3{}, math.Vec3{}, false
	}

	first := true
	for _, ent := range q.entities {
		minPoint, maxPoint := ent.BBox()
		if minPoint == nil || maxPoint == nil {
			continue
		}

		entMin := math.NewVec3(minPoint[0], minPoint[1], minPoint[2])
		entMax := math.NewVec3(maxPoint[0], maxPoint[1], maxPoint[2])

		if first {
			min = entMin
			max = entMax
			first = false
		} else {
			// Manual min/max calculations using methods
			minX := min.X()
			if entMin.X() < minX {
				minX = entMin.X()
			}
			minY := min.Y()
			if entMin.Y() < minY {
				minY = entMin.Y()
			}
			minZ := min.Z()
			if entMin.Z() < minZ {
				minZ = entMin.Z()
			}
			min = math.NewVec3(minX, minY, minZ)

			maxX := max.X()
			if entMax.X() > maxX {
				maxX = entMax.X()
			}
			maxY := max.Y()
			if entMax.Y() > maxY {
				maxY = entMax.Y()
			}
			maxZ := max.Z()
			if entMax.Z() > maxZ {
				maxZ = entMax.Z()
			}
			max = math.NewVec3(maxX, maxY, maxZ)
		}
	}

	return min, max, !first
}

// String returns a string representation of the query
func (q *EntityQuery) String() string {
	return fmt.Sprintf("EntityQuery{count=%d, types=[%s]}",
		len(q.entities), q.getEntityTypes())
}

// getEntityTypes returns a comma-separated list of entity types in the query
func (q *EntityQuery) getEntityTypes() string {
	types := make(map[string]bool)
	for _, ent := range q.entities {
		types[getEntityType(ent)] = true
	}

	var result []string
	for t := range types {
		result = append(result, t)
	}

	return strings.Join(result, ", ")
}

// parseQueryString parses a query string into entity names and attribute conditions
func parseQueryString(queryStr string) (entityNames []string, attrConditions string, err error) {
	// Find attribute conditions in square brackets
	attrStart := strings.Index(queryStr, "[")
	attrEnd := strings.LastIndex(queryStr, "]")

	if attrStart != -1 && attrEnd != -1 && attrEnd > attrStart {
		// Has attribute conditions
		entityPart := strings.TrimSpace(queryStr[:attrStart])
		attrConditions = strings.TrimSpace(queryStr[attrStart+1 : attrEnd])

		if entityPart == "" {
			return nil, "", fmt.Errorf("missing entity names before attribute conditions")
		}

		entityNames = strings.Fields(entityPart)
	} else {
		// No attribute conditions, just entity names
		entityNames = strings.Fields(queryStr)
	}

	if len(entityNames) == 0 {
		return nil, "", fmt.Errorf("no entity names specified")
	}

	return entityNames, attrConditions, nil
}

// parseAttributeQuery parses an attribute query string into a predicate function
func parseAttributeQuery(queryStr string) (func(entity.Entity) bool, error) {
	// Simple implementation for common cases
	// TODO: Implement full query parser with support for complex expressions

	// Handle common patterns like 'layer=="construction"'
	if matched, _ := regexp.MatchString(`^\w+=="[^"]*"$`, queryStr); matched {
		parts := strings.SplitN(queryStr, "==", 2)
		attrName := strings.TrimSpace(parts[0])
		value := strings.Trim(parts[1], `"`)

		return func(ent entity.Entity) bool {
			return getEntityAttribute(ent, attrName) == value
		}, nil
	}

	// Handle layer comparison without quotes (common case)
	if matched, _ := regexp.MatchString(`^layer==\w+$`, queryStr); matched {
		parts := strings.SplitN(queryStr, "==", 2)
		value := strings.TrimSpace(parts[1])

		return func(ent entity.Entity) bool {
			return getEntityAttribute(ent, "layer") == value
		}, nil
	}

	// Handle numeric comparisons
	if matched, _ := regexp.MatchString(`^\w+(<=|>=|<|>)\d+$`, queryStr); matched {
		re := regexp.MustCompile(`^(\w+)(<=|>=|<|>)(\d+)$`)
		matches := re.FindStringSubmatch(queryStr)
		if len(matches) == 4 {
			attrName := matches[1]
			op := matches[2]
			value, _ := strconv.ParseFloat(matches[3], 64)

			return func(ent entity.Entity) bool {
				attrValue := getEntityAttributeAsFloat(ent, attrName)
				switch op {
				case "<=":
					return attrValue <= value
				case ">=":
					return attrValue >= value
				case "<":
					return attrValue < value
				case ">":
					return attrValue > value
				}
				return false
			}, nil
		}
	}

	return nil, fmt.Errorf("unsupported attribute query format: %s", queryStr)
}

// getEntityAttribute returns an entity attribute value as string
func getEntityAttribute(ent entity.Entity, attrName string) string {
	// Common attributes
	switch strings.ToLower(attrName) {
	case "layer":
		if layer := ent.Layer(); layer != nil {
			return layer.Name()
		}
	case "type":
		return getEntityType(ent)
	case "handle":
		return ent.Handle()
	case "color":
		// Need to access color differently since SetColor expects ColorNumber
		if colorable, ok := ent.(interface{ Color() color.ColorNumber }); ok {
			return fmt.Sprintf("%d", colorable.Color())
		}
	}
	return ""
}

// getEntityAttributeAsFloat returns an entity attribute value as float
func getEntityAttributeAsFloat(ent entity.Entity, attrName string) float64 {
	value := getEntityAttribute(ent, attrName)
	if value == "" {
		return 0
	}
	if result, err := strconv.ParseFloat(value, 64); err == nil {
		return result
	}
	return 0
}
