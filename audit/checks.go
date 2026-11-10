// Document-wide audit checks: handles, ownership, tables, references,
// xdata, reactors, extension dictionaries, extents, entity invariants.
package audit

import (
	"fmt"
	"math"
	"strings"

	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/lldxf"
	"github.com/edanko/dxf/object"
)

// Issue codes reported by the built-in checks.
const (
	CodeDuplicateHandle           = "DUPLICATE_HANDLE"
	CodeInvalidOwner              = "INVALID_OWNER"
	CodeEmptyTableName            = "EMPTY_TABLE_NAME"
	CodeDuplicateTableEntry       = "DUPLICATE_TABLE_ENTRY"
	CodeMissingTableEntry         = "MISSING_TABLE_ENTRY"
	CodeOwnerMismatch             = "OWNER_MISMATCH"
	CodeMissingLayer              = "MISSING_LAYER"
	CodeMissingLinetype           = "MISSING_LINETYPE"
	CodeMissingStyle              = "MISSING_STYLE"
	CodeMissingBlock              = "MISSING_BLOCK"
	CodeMissingDimStyle           = "MISSING_DIMSTYLE"
	CodeMissingLayoutBlock        = "MISSING_LAYOUT_BLOCK"
	CodeXDataMissingAppID         = "XDATA_MISSING_APPID"
	CodeReactorUnresolved         = "REACTOR_UNRESOLVED"
	CodeExtDictUnresolved         = "EXTDICT_UNRESOLVED"
	CodeDictionaryEntryUnresolved = "DICTIONARY_ENTRY_UNRESOLVED"
	CodeBadExtents                = "BAD_EXTENTS"
	CodeEntityInvariants          = "ENTITY_INVARIANTS"
	CodeSplineInvalid             = "SPLINE_INVALID"
)

// Required default entries per standard table (ezdxf audit protocol).
var requiredEntries = []struct{ table, entry string }{
	{"LAYER", "0"},
	{"LTYPE", "ByLayer"},
	{"LTYPE", "ByBlock"},
	{"LTYPE", "Continuous"},
	{"STYLE", "Standard"},
	{"DIMSTYLE", "Standard"},
	{"APPID", "ACAD"},
}

func init() {
	RegisterCheck(Check{Code: CodeDuplicateHandle, Severity: SeverityError, Safe: false, Fn: checkDuplicateHandles})
	RegisterCheck(Check{Code: CodeInvalidOwner, Severity: SeverityError, Safe: true, Fn: checkOwners})
	RegisterCheck(Check{Code: CodeEmptyTableName, Severity: SeverityWarning, Safe: false, Fn: checkTableIntegrity})
	RegisterCheck(Check{Code: CodeDuplicateTableEntry, Severity: SeverityError, Safe: false, Fn: checkTableIntegrity})
	RegisterCheck(Check{Code: CodeMissingTableEntry, Severity: SeverityError, Safe: true, Fn: checkTableIntegrity})
	RegisterCheck(Check{Code: CodeOwnerMismatch, Severity: SeverityWarning, Safe: false, Fn: checkTableIntegrity})
	RegisterCheck(Check{Code: CodeMissingLayer, Severity: SeverityError, Safe: true, Fn: checkReferences})
	RegisterCheck(Check{Code: CodeMissingLinetype, Severity: SeverityWarning, Safe: false, Fn: checkReferences})
	RegisterCheck(Check{Code: CodeMissingStyle, Severity: SeverityWarning, Safe: false, Fn: checkReferences})
	RegisterCheck(Check{Code: CodeMissingBlock, Severity: SeverityError, Safe: false, Fn: checkReferences})
	RegisterCheck(Check{Code: CodeMissingDimStyle, Severity: SeverityWarning, Safe: false, Fn: checkReferences})
	RegisterCheck(Check{Code: CodeMissingLayoutBlock, Severity: SeverityWarning, Safe: false, Fn: checkLayouts})
	RegisterCheck(Check{Code: CodeXDataMissingAppID, Severity: SeverityWarning, Safe: false, Fn: checkXData})
	RegisterCheck(Check{Code: CodeReactorUnresolved, Severity: SeverityWarning, Safe: false, Fn: checkReactors})
	RegisterCheck(Check{Code: CodeExtDictUnresolved, Severity: SeverityWarning, Safe: false, Fn: checkExtensionDictionaries})
	RegisterCheck(Check{Code: CodeDictionaryEntryUnresolved, Severity: SeverityWarning, Safe: false, Fn: checkDictionaries})
	RegisterCheck(Check{Code: CodeBadExtents, Severity: SeverityWarning, Safe: false, Fn: checkExtents})
	RegisterCheck(Check{Code: CodeEntityInvariants, Severity: SeverityError, Safe: false, Fn: checkEntityInvariants})
	RegisterEntityHook("SPLINE", checkSpline)
}

// checkDuplicateHandles reports every handle registered by more than one
// object, naming both in the message.
func checkDuplicateHandles(ctx *Context) []Issue {
	var out []Issue
	for _, h := range ctx.Handles() {
		objs := ctx.Objects(h)
		if len(objs) < 2 {
			continue
		}
		names := make([]string, 0, len(objs))
		for _, o := range objs {
			names = append(names, Describe(o))
		}
		out = append(out, Issue{
			Severity: SeverityError,
			Code:     CodeDuplicateHandle,
			Handle:   h,
			Message:  fmt.Sprintf("handle %s used by %d objects: %s", h, len(objs), strings.Join(names, ", ")),
		})
	}
	return out
}

// checkOwnerIssues emits invalid-owner issues for entities (missing or
// unresolvable owner) and objects (owner handle that cannot be resolved).
func checkOwnerIssues(ctx *Context, sevObj Severity) []Issue {
	var out []Issue
	for _, e := range ctx.AllEntities() {
		h := EntityHandle(e)
		owner := EntityOwner(e)
		if owner == nil || owner.Handle() == "" {
			out = append(out, Issue{
				Severity: SeverityError, Code: CodeInvalidOwner, Handle: h,
				Message: fmt.Sprintf("%s has no owner", Describe(e)),
			})
			continue
		}
		if _, ok := ctx.Resolve(owner.Handle()); !ok {
			out = append(out, Issue{
				Severity: SeverityError, Code: CodeInvalidOwner, Handle: h,
				Message: fmt.Sprintf("%s owner handle %s not found", Describe(e), owner.Handle()),
			})
		}
	}
	for _, o := range ctx.Doc.Objects() {
		oh := ObjectOwnerHandle(o)
		if oh == "" || oh == "0" {
			continue
		}
		if _, ok := ctx.Resolve(oh); !ok {
			out = append(out, Issue{
				Severity: sevObj, Code: CodeInvalidOwner, Handle: o.Handle(),
				Message: fmt.Sprintf("%s owner handle %s not found", Describe(o), oh),
			})
		}
	}
	return out
}

func checkOwners(ctx *Context) []Issue { return checkOwnerIssues(ctx, SeverityWarning) }

// checkTableIntegrity verifies entry names (non-empty, unique per table),
// required default entries and entry owner links.
func checkTableIntegrity(ctx *Context) []Issue {
	var out []Issue
	if ctx.Doc.Tables() == nil {
		return out
	}
	for _, t := range ctx.Doc.Tables() {
		seen := map[string]bool{}
		for _, e := range t.Entries() {
			name := e.Name()
			if name == "" {
				out = append(out, Issue{
					Severity: SeverityWarning, Code: CodeEmptyTableName, Handle: e.Handle(),
					Message: fmt.Sprintf("%s entry in table %s has no name", Describe(e), t.Name()),
				})
			} else if seen[name] {
				out = append(out, Issue{
					Severity: SeverityError, Code: CodeDuplicateTableEntry, Handle: e.Handle(),
					Message: fmt.Sprintf("table %s contains duplicate entry %q", t.Name(), name),
				})
			} else {
				seen[name] = true
			}
			owner := EntryOwner(e)
			if owner == nil || owner.Handle() == "" {
				out = append(out, Issue{
					Severity: SeverityError, Code: CodeInvalidOwner, Handle: e.Handle(),
					Message: fmt.Sprintf("%s entry in table %s has no owner", Describe(e), t.Name()),
				})
			} else if owner.Handle() != t.Handle() {
				if _, ok := ctx.Resolve(owner.Handle()); !ok {
					out = append(out, Issue{
						Severity: SeverityWarning, Code: CodeOwnerMismatch, Handle: e.Handle(),
						Message: fmt.Sprintf("%s entry owner handle %s not found (table %s)", Describe(e), owner.Handle(), t.Name()),
					})
				} else {
					out = append(out, Issue{
						Severity: SeverityWarning, Code: CodeOwnerMismatch, Handle: e.Handle(),
						Message: fmt.Sprintf("%s entry owner handle %s does not match table %s (%s)", Describe(e), owner.Handle(), t.Name(), t.Handle()),
					})
				}
			}
		}
	}
	for _, req := range requiredEntries {
		t := ctx.TableByName(req.table)
		if t == nil {
			out = append(out, Issue{
				Severity: SeverityError, Code: CodeMissingTableEntry,
				Handle:  req.table + ":" + req.entry,
				Message: fmt.Sprintf("table %s is missing required entry %q", req.table, req.entry),
			})
			continue
		}
		found := false
		for _, e := range t.Entries() {
			if e.Name() == req.entry {
				found = true
				break
			}
		}
		if !found {
			out = append(out, Issue{
				Severity: SeverityError, Code: CodeMissingTableEntry,
				Handle:  req.table + ":" + req.entry,
				Message: fmt.Sprintf("table %s is missing required entry %q", req.table, req.entry),
			})
		}
	}
	return out
}

// rawTagsOf returns the retained raw tag stream of an entity, nil when the
// entity model does not keep raw tags (e.g. entities created by the API).
func rawTagsOf(e interface{}) []entity.TagPair {
	if r, ok := e.(interface{ RawTags() []entity.TagPair }); ok {
		return r.RawTags()
	}
	return nil
}

// tagValues collects RawTags values with the given group code.
func tagValues(e interface{}, code int) []string {
	var out []string
	for _, t := range rawTagsOf(e) {
		if t.Code == code {
			out = append(out, t.Value)
		}
	}
	return out
}

// entityReactorHandles extracts reactor handles from the {ACAD_REACTORS
// group embedded in an entity's raw tags.
func entityReactorHandles(e interface{}) []string {
	var out []string
	inGroup := false
	for _, t := range rawTagsOf(e) {
		switch {
		case t.Code == 102 && strings.EqualFold(strings.TrimSpace(t.Value), "{ACAD_REACTORS"):
			inGroup = true
		case t.Code == 102 && strings.TrimSpace(t.Value) == "}":
			inGroup = false
		case inGroup && t.Code == 330:
			out = append(out, t.Value)
		}
	}
	return out
}

// checkReferences verifies layer, linetype, text style, block and dimension
// style references of every entity.
func checkReferences(ctx *Context) []Issue {
	var out []Issue
	layers := ctx.EntryNames("LAYER")
	linetypes := ctx.EntryNames("LTYPE")
	styles := ctx.EntryNames("STYLE")
	blocks := ctx.BlockNames()
	for name := range ctx.EntryNames("BLOCK_RECORD") {
		blocks[name] = true
	}
	dimstyles := ctx.EntryNames("DIMSTYLE")
	styleTypes := map[string]bool{"TEXT": true, "MTEXT": true, "ATTRIB": true, "ATTDEF": true, "SHAPE": true}
	for _, e := range ctx.AllEntities() {
		h := EntityHandle(e)
		if ln, ok := e.(interface{ LayerName() string }); ok {
			name := ln.LayerName()
			if name != "" && !layers[name] {
				out = append(out, Issue{
					Severity: SeverityError, Code: CodeMissingLayer, Handle: h,
					Message: fmt.Sprintf("%s references undefined layer %q", Describe(e), name),
				})
			}
		}
		if lt, ok := e.(interface{ Linetype() string }); ok {
			name := lt.Linetype()
			if name != "" && !strings.EqualFold(name, "ByLayer") && !strings.EqualFold(name, "ByBlock") && !linetypes[name] {
				out = append(out, Issue{
					Severity: SeverityWarning, Code: CodeMissingLinetype, Handle: h,
					Message: fmt.Sprintf("%s references undefined linetype %q", Describe(e), name),
				})
			}
		}
		if styleTypes[strings.ToUpper(EntityType(e))] {
			for _, name := range tagValues(e, 7) {
				if name == "" || strings.EqualFold(name, "ByLayer") || strings.EqualFold(name, "ByBlock") {
					continue
				}
				if !styles[name] {
					out = append(out, Issue{
						Severity: SeverityWarning, Code: CodeMissingStyle, Handle: h,
						Message: fmt.Sprintf("%s references undefined text style %q", Describe(e), name),
					})
				}
			}
		}
		if strings.EqualFold(EntityType(e), "INSERT") {
			names := []string{}
			seenName := map[string]bool{}
			if b, ok := e.(interface{ BlockName() string }); ok {
				if n := b.BlockName(); n != "" && !seenName[n] {
					seenName[n] = true
					names = append(names, n)
				}
			}
			for _, n := range tagValues(e, 2) {
				if n != "" && !seenName[n] {
					seenName[n] = true
					names = append(names, n)
				}
			}
			for _, name := range names {
				if !blocks[name] {
					out = append(out, Issue{
						Severity: SeverityError, Code: CodeMissingBlock, Handle: h,
						Message: fmt.Sprintf("%s references undefined block %q", Describe(e), name),
					})
				}
			}
		}
		if strings.EqualFold(EntityType(e), "DIMENSION") {
			names := []string{}
			seenName := map[string]bool{}
			if s, ok := e.(interface{ GetStyle() string }); ok {
				if n := s.GetStyle(); n != "" && !seenName[n] {
					seenName[n] = true
					names = append(names, n)
				}
			}
			for _, n := range tagValues(e, 3) {
				if n != "" && !seenName[n] {
					seenName[n] = true
					names = append(names, n)
				}
			}
			for _, name := range names {
				if !dimstyles[name] {
					out = append(out, Issue{
						Severity: SeverityWarning, Code: CodeMissingDimStyle, Handle: h,
						Message: fmt.Sprintf("%s references undefined dimension style %q", Describe(e), name),
					})
				}
			}
		}
	}
	return out
}

// checkLayouts verifies that every LAYOUT object points at an existing
// BLOCK_RECORD.
func checkLayouts(ctx *Context) []Issue {
	var out []Issue
	for _, o := range ctx.Doc.Objects() {
		br, ok := o.(interface{ BlockRecordHandle() string })
		if !ok {
			continue
		}
		h := br.BlockRecordHandle()
		if h == "" {
			continue
		}
		if _, ok := ctx.Resolve(h); !ok {
			out = append(out, Issue{
				Severity: SeverityWarning, Code: CodeMissingLayoutBlock, Handle: o.Handle(),
				Message: fmt.Sprintf("%s references missing block record handle %s", Describe(o), h),
			})
		}
	}
	return out
}

// checkXData verifies that app IDs referenced by entity XData exist in the
// APPID table.
func checkXData(ctx *Context) []Issue {
	var out []Issue
	appids := ctx.EntryNames("APPID")
	for _, e := range ctx.AllEntities() {
		x, ok := e.(interface{ XDataAppIDs() []string })
		if !ok {
			continue
		}
		for _, id := range x.XDataAppIDs() {
			if !appids[id] {
				out = append(out, Issue{
					Severity: SeverityWarning, Code: CodeXDataMissingAppID, Handle: EntityHandle(e),
					Message: fmt.Sprintf("%s XData references undefined app ID %q", Describe(e), id),
				})
			}
		}
	}
	return out
}

// checkReactors verifies reactor links of entities, table entries and
// objects.
func checkReactors(ctx *Context) []Issue {
	var out []Issue
	for _, e := range ctx.AllEntities() {
		for _, r := range entityReactorHandles(e) {
			if r == "" || r == "0" {
				continue
			}
			if _, ok := ctx.Resolve(r); !ok {
				out = append(out, Issue{
					Severity: SeverityWarning, Code: CodeReactorUnresolved, Handle: EntityHandle(e),
					Message: fmt.Sprintf("%s reactor handle %s not found", Describe(e), r),
				})
			}
		}
	}
	for _, en := range ctx.Entries() {
		r, ok := en.(interface{ ReactorHandles() []string })
		if !ok {
			continue
		}
		for _, rh := range r.ReactorHandles() {
			if rh == "" || rh == "0" {
				continue
			}
			if _, ok := ctx.Resolve(rh); !ok {
				out = append(out, Issue{
					Severity: SeverityWarning, Code: CodeReactorUnresolved, Handle: en.Handle(),
					Message: fmt.Sprintf("%s reactor handle %s not found", Describe(en), rh),
				})
			}
		}
	}
	for _, o := range ctx.Doc.Objects() {
		for _, rh := range ObjectReactors(o) {
			if rh == "" || rh == "0" {
				continue
			}
			if _, ok := ctx.Resolve(rh); !ok {
				out = append(out, Issue{
					Severity: SeverityWarning, Code: CodeReactorUnresolved, Handle: o.Handle(),
					Message: fmt.Sprintf("%s reactor handle %s not found", Describe(o), rh),
				})
			}
		}
	}
	return out
}

// checkExtensionDictionaries verifies extension dictionary links (code 360)
// of entities and table entries.
func checkExtensionDictionaries(ctx *Context) []Issue {
	var out []Issue
	for _, e := range ctx.AllEntities() {
		for _, h := range tagValues(e, 360) {
			if h == "" || h == "0" {
				continue
			}
			if _, ok := ctx.Resolve(h); !ok {
				out = append(out, Issue{
					Severity: SeverityWarning, Code: CodeExtDictUnresolved, Handle: EntityHandle(e),
					Message: fmt.Sprintf("%s extension dictionary handle %s not found", Describe(e), h),
				})
			}
		}
	}
	for _, en := range ctx.Entries() {
		x, ok := en.(interface{ ExtDictHandle() string })
		if !ok {
			continue
		}
		h := x.ExtDictHandle()
		if h == "" || h == "0" {
			continue
		}
		if _, ok := ctx.Resolve(h); !ok {
			out = append(out, Issue{
				Severity: SeverityWarning, Code: CodeExtDictUnresolved, Handle: en.Handle(),
				Message: fmt.Sprintf("%s extension dictionary handle %s not found", Describe(en), h),
			})
		}
	}
	return out
}

// checkDictionaries verifies that every dictionary entry points at an
// existing object.
func checkDictionaries(ctx *Context) []Issue {
	var out []Issue
	for _, o := range ctx.Doc.Objects() {
		if d, ok := o.(*object.Dictionary); ok {
			for _, entry := range d.GetEntries() {
				v := entry.Value
				if v == "" || v == "0" {
					continue
				}
				if _, ok := ctx.Resolve(v); !ok {
					out = append(out, Issue{
						Severity: SeverityWarning, Code: CodeDictionaryEntryUnresolved, Handle: o.Handle(),
						Message: fmt.Sprintf("dictionary entry %q of %s references missing object %s", entry.Key, Describe(o), v),
					})
				}
			}
			continue
		}
		// Raw objects (ACDBDICTIONARYW and friends) carry (3 key, 350 handle)
		// pairs in their retained tags.
		rt, ok := o.(interface{ RawTags() lldxf.Tags })
		if !ok {
			continue
		}
		kind := ""
		if dt, ok := o.(interface{ DXFType() string }); ok {
			kind = strings.ToUpper(dt.DXFType())
		}
		if !strings.Contains(kind, "DICTIONARY") {
			continue
		}
		key := ""
		for _, tag := range rt.RawTags() {
			switch int(tag.Code()) {
			case 3:
				key = fmt.Sprint(tag.Value())
			case 350:
				v := fmt.Sprint(tag.Value())
				if v == "" || v == "0" {
					continue
				}
				if _, ok := ctx.Resolve(v); !ok {
					out = append(out, Issue{
						Severity: SeverityWarning, Code: CodeDictionaryEntryUnresolved, Handle: o.Handle(),
						Message: fmt.Sprintf("dictionary entry %q of %s references missing object %s", key, Describe(o), v),
					})
				}
			}
		}
	}
	return out
}

// safeBBox calls BBox, recovering from panics in exotic entity models.
func safeBBox(bb interface{ BBox() ([]float64, []float64) }) (min, max []float64) {
	defer func() {
		if r := recover(); r != nil {
			min, max = nil, nil
		}
	}()
	return bb.BBox()
}

// validExtents reports whether min/max form a finite, ordered box.
func validExtents(min, max []float64) bool {
	if len(min) < 2 || len(max) < 2 {
		return true // no data, nothing to verify
	}
	sentinel := func(v float64) bool { return math.Abs(v) > 1e19 }
	allZero := true
	for i := range min {
		if min[i] != 0 || max[i] != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		return true // unset extents
	}
	for i := 0; i < len(min) && i < len(max); i++ {
		if math.IsInf(min[i], 0) || math.IsInf(max[i], 0) || math.IsNaN(min[i]) || math.IsNaN(max[i]) {
			return false
		}
		if sentinel(min[i]) || sentinel(max[i]) {
			continue // AutoCAD "unlimited" sentinel values
		}
		if min[i] > max[i] {
			return false
		}
	}
	return true
}

// checkExtents verifies header extents and entity bounding boxes.
func checkExtents(ctx *Context) []Issue {
	var out []Issue
	h := ctx.Doc.Header()
	if h != nil && !validExtents(h.ExtMin, h.ExtMax) {
		out = append(out, Issue{
			Severity: SeverityWarning, Code: CodeBadExtents,
			Message: fmt.Sprintf("header extents %v..%v are not a valid box", h.ExtMin, h.ExtMax),
		})
	}
	for _, e := range ctx.AllEntities() {
		bb, ok := e.(interface{ BBox() ([]float64, []float64) })
		if !ok {
			continue
		}
		min, max := safeBBox(bb)
		if !validExtents(min, max) {
			out = append(out, Issue{
				Severity: SeverityWarning, Code: CodeBadExtents, Handle: EntityHandle(e),
				Message: fmt.Sprintf("%s bounding box %v..%v is not valid", Describe(e), min, max),
			})
		}
	}
	return out
}

// checkEntityInvariants dispatches registered per-type entity invariant
// hooks and structural Audit(*Context) implementors.
func checkEntityInvariants(ctx *Context) []Issue {
	var out []Issue
	for _, e := range ctx.AllEntities() {
		var found []Issue
		if fn, ok := entityHook(EntityType(e)); ok {
			found = append(found, fn(e, ctx)...)
		}
		if a, ok := e.(interface{ Audit(*Context) []Issue }); ok {
			found = append(found, a.Audit(ctx)...)
		}
		for _, is := range found {
			out = append(out, Issue{
				Severity: is.Severity, Code: CodeEntityInvariants, Handle: EntityHandle(e),
				Message: fmt.Sprintf("%s: %s", is.Code, is.Message),
			})
		}
	}
	return out
}

// checkSpline verifies the spline knot count invariant:
// len(Knots) == len(Controls) + Degree + 1 (ezdxf spline validation).
func checkSpline(e interface{}, ctx *Context) []Issue {
	s, ok := e.(*entity.Spline)
	if !ok || len(s.Controls) == 0 || len(s.Knots) == 0 {
		return nil
	}
	if s.Degree < 1 {
		return []Issue{{
			Severity: SeverityError, Code: CodeSplineInvalid, Handle: s.Handle(),
			Message: fmt.Sprintf("spline degree %d is invalid", s.Degree),
		}}
	}
	if want := len(s.Controls) + s.Degree + 1; len(s.Knots) != want {
		return []Issue{{
			Severity: SeverityError, Code: CodeSplineInvalid, Handle: s.Handle(),
			Message: fmt.Sprintf("spline has %d knots for %d control points at degree %d, expected %d",
				len(s.Knots), len(s.Controls), s.Degree, want),
		}}
	}
	return nil
}
