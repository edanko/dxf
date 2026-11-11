package standards

import (
	"fmt"
	"regexp"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

// DefaultLayerNamePattern rejects characters reserved by DXF and enforces a
// non-empty, length-bounded layer name.
const defaultLayerNamePattern = `^[^<>/\\":;?*|,=]{1,255}$`

// LayerNameRule validates layer table names and entity layer references
// against a configurable regular expression.
type LayerNameRule struct {
	pattern       *regexp.Regexp
	severity      Severity
	checkEntities bool
}

// NewLayerNameRule creates a layer-name rule from a regular expression.
func NewLayerNameRule(pattern string, severity Severity) *LayerNameRule {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return &LayerNameRule{pattern: regexp.MustCompile(defaultLayerNamePattern), severity: severity, checkEntities: true}
	}
	return &LayerNameRule{pattern: compiled, severity: severity, checkEntities: true}
}

// LayerNameRuleCode is the stable rule identifier.
const LayerNameRuleCode = "LAYER_NAME"

// Code returns the rule identifier.
func (r *LayerNameRule) Code() string { return LayerNameRuleCode }

// Description explains the rule.
func (r *LayerNameRule) Description() string { return "layer names match the configured pattern" }

// Severity returns the default violation severity.
func (r *LayerNameRule) Severity() Severity { return r.severity }

// Check validates layer table entries and entity layer references.
func (r *LayerNameRule) Check(d *drawing.Drawing) []Violation {
	if r == nil || r.pattern == nil || d == nil {
		return nil
	}
	var violations []Violation
	layerNames := map[string]bool{}

	for _, t := range d.Tables() {
		if t == nil || t.Name() != "LAYER" {
			continue
		}
		for _, entry := range t.Entries() {
			name := entry.Name()
			layerNames[name] = true
			if r.pattern.MatchString(name) {
				continue
			}
			violations = append(violations, Violation{
				Code:     LayerNameRuleCode,
				Severity: r.severity,
				Path:     "TABLES/LAYER/" + name,
				Handle:   entry.Handle(),
				Message:  fmt.Sprintf("layer name %q does not match %s", name, r.pattern.String()),
			})
		}
	}

	if !r.checkEntities {
		return violations
	}
	for _, e := range d.Entities() {
		name := entityLayerName(e)
		if name == "" || r.pattern.MatchString(name) {
			continue
		}
		if layerNames[name] {
			// Already reported at the layer table entry.
			continue
		}
		violations = append(violations, Violation{
			Code:     LayerNameRuleCode,
			Severity: r.severity,
			Path:     fmt.Sprintf("ENTITIES/%s[%s]/8", e.DXFType(), e.Handle()),
			Handle:   e.Handle(),
			Message:  fmt.Sprintf("entity %s references layer %q which does not match %s", e.DXFType(), name, r.pattern.String()),
		})
	}
	return violations
}

func entityLayerName(e entity.Entity) string {
	if e == nil {
		return ""
	}
	if named, ok := e.(interface{ LayerName() string }); ok {
		if name := named.LayerName(); name != "" {
			return name
		}
	}
	if layer := e.Layer(); layer != nil {
		return layer.Name()
	}
	return ""
}
