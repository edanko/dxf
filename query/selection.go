// Package query additions: handle-based selection sets, grouping, and
// document-wide query execution.
package query

import (
	"slices"
	"sort"
	"strings"

	"strconv"

	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/math"
)

// EntitySource resolves handles to entities (drawing.Drawing implements it).
type EntitySource interface {
	GetEntity(handle string) (entity.Entity, bool)
}

// Selection is a handle-based selection set: stable across edits, cheap to
// combine, and resolvable against any entity source.
type Selection struct {
	name    string
	handles []string
}

// NewSelection builds a value selection from handles (duplicates removed,
// order preserved).
func NewSelection(handles ...string) *Selection {
	return &Selection{handles: dedupeStrings(handles)}
}

// NamedSelection builds a named selection set.
func NamedSelection(name string, handles ...string) *Selection {
	s := NewSelection(handles...)
	s.name = name
	return s
}

// SelectionFromEntities builds a selection from entities.
func SelectionFromEntities(ents ...entity.Entity) *Selection {
	hs := make([]string, 0, len(ents))
	for _, e := range ents {
		hs = append(hs, e.Handle())
	}
	return NewSelection(hs...)
}

// Name returns the selection name ("" for value selections).
func (s *Selection) Name() string { return s.name }

// SetName renames the selection.
func (s *Selection) SetName(name string) { s.name = name }

// Handles returns the selected handles in order.
func (s *Selection) Handles() []string {
	out := make([]string, len(s.handles))
	copy(out, s.handles)
	return out
}

// Len returns the number of selected handles.
func (s *Selection) Len() int { return len(s.handles) }

// Has reports whether a handle is selected.
func (s *Selection) Has(handle string) bool {
	return slices.Contains(s.handles, handle)
}

// Clone returns an independent copy.
func (s *Selection) Clone() *Selection {
	return &Selection{name: s.name, handles: s.Handles()}
}

// Add appends handles (deduplicated) and returns the selection for chaining.
func (s *Selection) Add(handles ...string) *Selection {
	s.handles = dedupeStrings(append(s.handles, handles...))
	return s
}

// Remove drops handles and returns the selection for chaining.
func (s *Selection) Remove(handles ...string) *Selection {
	drop := map[string]bool{}
	for _, h := range handles {
		drop[h] = true
	}
	out := make([]string, 0, len(s.handles))
	for _, h := range s.handles {
		if !drop[h] {
			out = append(out, h)
		}
	}
	s.handles = out
	return s
}

// Union returns a new selection containing handles from both sets.
func (s *Selection) Union(other *Selection) *Selection {
	return NewSelection(append(s.Handles(), other.Handles()...)...)
}

// Intersect returns handles present in both selections (left order preserved).
func (s *Selection) Intersect(other *Selection) *Selection {
	set := map[string]bool{}
	for _, h := range other.handles {
		set[h] = true
	}
	out := make([]string, 0, len(s.handles))
	for _, h := range s.handles {
		if set[h] {
			out = append(out, h)
		}
	}
	return NewSelection(out...)
}

// Difference returns handles of s that are not in other.
func (s *Selection) Difference(other *Selection) *Selection {
	drop := map[string]bool{}
	for _, h := range other.handles {
		drop[h] = true
	}
	out := make([]string, 0, len(s.handles))
	for _, h := range s.handles {
		if !drop[h] {
			out = append(out, h)
		}
	}
	return NewSelection(out...)
}

// Resolve maps the handles through a source, reporting entities that could
// not be resolved (deleted handles).
func (s *Selection) Resolve(src EntitySource) (ents []entity.Entity, missing []string) {
	for _, h := range s.handles {
		if src == nil {
			missing = append(missing, h)
			continue
		}
		if e, ok := src.GetEntity(h); ok && e != nil {
			ents = append(ents, e)
			continue
		}
		missing = append(missing, h)
	}
	return ents, missing
}

// Entities returns the resolvable entities of the selection.
func (s *Selection) Entities(src EntitySource) []entity.Entity {
	ents, _ := s.Resolve(src)
	return ents
}

// Bounds returns the combined bounding box of the selected entities.
func (s *Selection) Bounds(src EntitySource) (mn, mx math.Vec3, ok bool) {
	var lo, hi math.Vec3
	first := true
	for _, e := range s.Entities(src) {
		c := ContextOf(e)
		if !c.HasBBox {
			continue
		}
		if first {
			lo, hi, first = c.Min, c.Max, false
			continue
		}
		for i := range 3 {
			lo[i] = min(lo[i], c.Min[i])
			hi[i] = max(hi[i], c.Max[i])
		}
	}
	return lo, hi, !first
}

// String renders the selection for diagnostics.
func (s *Selection) String() string {
	if s.name != "" {
		return s.name + ":" + strings.Join(s.handles, ",")
	}
	return strings.Join(s.handles, ",")
}

func dedupeStrings(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

// Execute evaluates an expression over candidate entities and returns the
// matching selection. Entities without bounding boxes pass spatial
// pre-filters conservatively.
func Execute(ents []entity.Entity, expr Expression) (*Selection, error) {
	pred, err := Compile(expr)
	if err != nil {
		return nil, err
	}
	return execute(ents, pred), nil
}

// ExecuteQuery parses, compiles and evaluates an expression.
func ExecuteQuery(ents []entity.Entity, input string) (*Selection, error) {
	pred, err := CompileQuery(input)
	if err != nil {
		return nil, err
	}
	return execute(ents, pred), nil
}

func execute(ents []entity.Entity, pred Predicate) *Selection {
	hs := make([]string, 0, len(ents))
	for _, e := range ents {
		if pred(ContextOf(e)) {
			hs = append(hs, e.Handle())
		}
	}
	return NewSelection(hs...)
}

// GroupByAttribute groups entities by a named attribute: layer, type, color,
// linetype, handle, text, or "attr:TAG" for INSERT attribute values.
func GroupByAttribute(ents []entity.Entity, attr string) map[string][]entity.Entity {
	key := strings.ToLower(attr)
	groups := map[string][]entity.Entity{}
	for _, e := range ents {
		groups[attributeKey(ContextOf(e), key)] = append(groups[attributeKey(ContextOf(e), key)], e)
	}
	return groups
}

// GroupHandles groups selected handles by attribute, resolving through src.
func GroupHandles(sel *Selection, src EntitySource, attr string) map[string]*Selection {
	key := strings.ToLower(attr)
	out := map[string]*Selection{}
	for _, e := range sel.Entities(src) {
		k := attributeKey(ContextOf(e), key)
		if out[k] == nil {
			out[k] = NewSelection()
		}
		out[k].Add(e.Handle())
	}
	return out
}

func attributeKey(c *EntityContext, key string) string {
	if strings.HasPrefix(key, "attr:") {
		return c.Attributes()[strings.ToUpper(key[5:])]
	}
	switch x := c.Property(key).(type) {
	case string:
		return x
	case int:
		return strconv.Itoa(x)
	case nil:
		return ""
	default:
		return ""
	}
}

// SortedGroupKeys returns group keys in lexical order.
func SortedGroupKeys[V any](groups map[string]V) []string {
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
