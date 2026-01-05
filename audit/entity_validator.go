package audit

import (
	"fmt"
	"math"

	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/table"
)

type ValidationResult struct {
	Valid    bool
	Errors   []string
	Warnings []string
}

func (v *ValidationResult) AddError(msg string) {
	v.Valid = false
	v.Errors = append(v.Errors, msg)
}

func (v *ValidationResult) AddWarning(msg string) {
	v.Warnings = append(v.Warnings, msg)
}

type EntityValidator struct {
	validLayers    map[string]bool
	validStyles    map[string]bool
	validLinetypes map[string]bool
}

func NewEntityValidator() *EntityValidator {
	return &EntityValidator{
		validLayers:    make(map[string]bool),
		validStyles:    make(map[string]bool),
		validLinetypes: make(map[string]bool),
	}
}

func (v *EntityValidator) SetValidLayers(layers map[string]*table.Layer) {
	for name := range layers {
		v.validLayers[name] = true
	}
}

func (v *EntityValidator) SetValidStyles(styles map[string]*table.Style) {
	for name := range styles {
		v.validStyles[name] = true
	}
}

func (v *EntityValidator) ValidateEntity(e entity.Entity) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if e == nil {
		result.AddError("entity is nil")
		return result
	}

	v.validateHandle(e, result)
	v.validateLayer(e, result)
	v.validateColor(e, result)
	v.validateGeometry(e, result)

	return result
}

func (v *EntityValidator) validateHandle(e entity.Entity, result *ValidationResult) {
	if e.Handle() == "" {
		result.AddError("entity has no handle")
	}
}

func (v *EntityValidator) validateLayer(e entity.Entity, result *ValidationResult) {
	layer := e.Layer()
	if layer == nil {
		result.AddError("entity has no layer assigned")
		return
	}

	if len(v.validLayers) > 0 {
		if !v.validLayers[layer.Name()] {
			result.AddError(fmt.Sprintf("entity references undefined layer '%s'", layer.Name()))
		}
	}
}

func (v *EntityValidator) validateColor(e entity.Entity, result *ValidationResult) {
}

func (v *EntityValidator) validateGeometry(e entity.Entity, result *ValidationResult) {
	minBbox, maxBbox := e.BBox()

	if len(minBbox) < 2 || len(maxBbox) < 2 {
		result.AddError("entity has invalid bounding box")
		return
	}

	for i := 0; i < 2; i++ {
		if math.IsInf(minBbox[i], 0) || math.IsInf(maxBbox[i], 0) {
			result.AddWarning("entity has infinite bounding box coordinates")
		}
		if math.IsNaN(minBbox[i]) || math.IsNaN(maxBbox[i]) {
			result.AddError("entity has NaN bounding box coordinates")
		}
	}

	switch ent := e.(type) {
	case *entity.Point:
		v.validatePoint(ent, result)
	case *entity.Line:
		v.validateLine(ent, result)
	case *entity.Circle:
		v.validateCircle(ent, result)
	case *entity.Arc:
		v.validateArc(ent, result)
	case *entity.LwPolyline:
		v.validateLwPolyline(ent, result)
	case *entity.Spline:
		v.validateSpline(ent, result)
	}
}

func (v *EntityValidator) validatePoint(p *entity.Point, result *ValidationResult) {
	if len(p.Coord) < 2 {
		result.AddError("point has invalid coordinates")
	}
	for _, c := range p.Coord {
		if math.IsInf(c, 0) || math.IsNaN(c) {
			result.AddError("point has invalid coordinate values")
			break
		}
	}
}

func (v *EntityValidator) validateLine(l *entity.Line, result *ValidationResult) {
	if len(l.Start) < 2 || len(l.End) < 2 {
		result.AddError("line has invalid start or end points")
	}

	for i := 0; i < 3; i++ {
		if i < len(l.Start) && i < len(l.End) {
			if l.Start[i] == l.End[i] && l.Start[i] == 0 {
				continue
			}
			if math.IsInf(l.Start[i], 0) || math.IsInf(l.End[i], 0) ||
				math.IsNaN(l.Start[i]) || math.IsNaN(l.End[i]) {
				result.AddError("line has invalid coordinate values")
				break
			}
		}
	}
}

func (v *EntityValidator) validateCircle(c *entity.Circle, result *ValidationResult) {
	if c.Radius < 0 {
		result.AddError("circle has negative radius")
	}
	if math.IsInf(c.Radius, 0) || math.IsNaN(c.Radius) {
		result.AddError("circle has invalid radius")
	}
	if len(c.Center) < 2 {
		result.AddError("circle has invalid center")
	}
}

func (v *EntityValidator) validateArc(a *entity.Arc, result *ValidationResult) {
	if len(a.Center) < 2 {
		result.AddError("arc has invalid center")
	}
	if a.Radius < 0 {
		result.AddError("arc has negative radius")
	}
	if len(a.Angle) < 2 {
		result.AddError("arc has invalid angles")
	} else {
		if a.Angle[0] == a.Angle[1] {
			result.AddWarning("arc has zero angle span")
		}
	}
}

func (v *EntityValidator) validateLwPolyline(lp *entity.LwPolyline, result *ValidationResult) {
	if lp.Vertices == nil || len(lp.Vertices) == 0 {
		result.AddWarning("LWPOLYLINE has no vertices")
	}

	for i, vtx := range lp.Vertices {
		if len(vtx) < 2 {
			result.AddError(fmt.Sprintf("LWPOLYLINE vertex %d has invalid coordinates", i))
		}
		for _, c := range vtx {
			if math.IsInf(c, 0) || math.IsNaN(c) {
				result.AddError(fmt.Sprintf("LWPOLYLINE vertex %d has invalid coordinates", i))
				break
			}
		}
	}
}

func (v *EntityValidator) validateSpline(s *entity.Spline, result *ValidationResult) {
	if s.Controls == nil || len(s.Controls) == 0 {
		result.AddError("spline has no control points")
	}

	for i, cp := range s.Controls {
		if len(cp) < 2 {
			result.AddError(fmt.Sprintf("spline control point %d has invalid coordinates", i))
		}
		for _, c := range cp {
			if math.IsInf(c, 0) || math.IsNaN(c) {
				result.AddError(fmt.Sprintf("spline control point %d has invalid coordinates", i))
				break
			}
		}
	}

	if s.Degree < 1 || s.Degree > 5 {
		result.AddError(fmt.Sprintf("spline has invalid degree %d", s.Degree))
	}
}

func IsValidLayerName(name string) bool {
	if name == "" {
		return false
	}
	if len(name) > 255 {
		return false
	}

	invalidChars := "<>/\"\\:;?*|,="
	for _, ch := range name {
		for _, invalid := range invalidChars {
			if ch == invalid {
				return false
			}
		}
	}
	return true
}

func IsValidStyleName(name string) bool {
	return name != "" && len(name) <= 255
}

func ValidateEntityHandle(handle string) bool {
	if len(handle) == 0 {
		return false
	}

	for _, c := range handle {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

type DrawingStats struct {
	TotalEntities   int
	EntityCounts    map[string]int
	LayerCounts     map[string]int
	InvalidEntities int
	TotalErrors     int
	TotalWarnings   int
}

func (a *Auditor) GetDrawingStats() *DrawingStats {
	stats := &DrawingStats{
		EntityCounts: make(map[string]int),
		LayerCounts:  make(map[string]int),
	}

	entities := a.doc.Entities()
	stats.TotalEntities = len(entities)

	validator := NewEntityValidator()
	for _, e := range entities {
		if e != nil {
			entityType := e.DXFType()
			stats.EntityCounts[entityType]++

			if layer := e.Layer(); layer != nil {
				stats.LayerCounts[layer.Name()]++
			}

			result := validator.ValidateEntity(e)
			if !result.Valid {
				stats.InvalidEntities++
				stats.TotalErrors += len(result.Errors)
				stats.TotalWarnings += len(result.Warnings)
			}
		}
	}

	return stats
}

func (stats *DrawingStats) String() string {
	return fmt.Sprintf("DrawingStats{total=%d, invalid=%d, errors=%d, warnings=%d}",
		stats.TotalEntities, stats.InvalidEntities, stats.TotalErrors, stats.TotalWarnings)
}
