// Fix mode: conservative fixers for safe issues, then re-audit.
package audit

import (
	"github.com/edanko/dxf/block"
	"strings"

	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/table"
)

type ownerSetter interface {
	SetOwner(handle.Handler)
}

type layerNameSetter interface {
	SetLayerName(string)
}

// Options controls audit execution.
type Options struct {
	// Fix applies safe fixers to the document and re-audits.
	Fix bool
	// StrictR12 reports OBJECTS/classes content in R12 files as incompatible.
	StrictR12 bool
}

// Fixer repairs one issue. It returns true when the document changed.
type Fixer func(ctx *Context, issue Issue) bool

var fixers = map[string]Fixer{}

// RegisterFixer installs the safe fixer for a check code. Only codes whose
// repairs are conservative (documented, deterministic) may register fixers;
// everything else stays report-only.
func RegisterFixer(code string, f Fixer) {
	fixers[code] = f
}

func init() {
	RegisterFixer(CodeInvalidOwner, fixInvalidOwner)
	RegisterFixer(CodeMissingTableEntry, fixMissingTableEntry)
	RegisterFixer(CodeMissingLayer, fixMissingLayer)
}

// fixInvalidOwner re-attaches entities to their containing block and table
// entries to their table.
func fixInvalidOwner(ctx *Context, issue Issue) bool {
	if issue.Handle == "" {
		return false
	}
	if blk := ctx.ContainingBlock(issue.Handle); blk != nil {
		obj, ok := ctx.Resolve(issue.Handle)
		if !ok {
			return false
		}
		e, ok := obj.(ownerSetter)
		if !ok {
			return false
		}
		e.SetOwner(modelSpaceOwner(ctx, blk))
		return true
	}
	if ctx.Doc.Tables() == nil {
		return false
	}
	for _, t := range ctx.Doc.Tables() {
		for _, en := range t.Entries() {
			if en.Handle() != issue.Handle {
				continue
			}
			en.SetOwner(t)
			return true
		}
	}
	return false
}

// fixMissingTableEntry creates the required standard entry named by
// "<TABLE>:<name>" in the issue handle.
// modelSpaceOwner prefers the BLOCK_RECORD entry for *Model_Space (the owner
// handle entities carry in files), falling back to the block itself.
func modelSpaceOwner(ctx *Context, blk *block.Block) handle.Handler {
	if t := ctx.TableByName("BLOCK_RECORD"); t != nil {
		if rec, err := t.Contains("*Model_Space"); err == nil && rec != nil {
			if h, ok := rec.(handle.Handler); ok && h.Handle() != "" {
				return h
			}
		}
	}
	return blk
}

func fixMissingTableEntry(ctx *Context, issue Issue) bool {
	parts := strings.SplitN(issue.Handle, ":", 2)
	if len(parts) != 2 {
		return false
	}
	tblName, entry := parts[0], parts[1]
	t := ctx.TableByName(tblName)
	if t == nil {
		return false
	}
	var st table.SymbolTable
	switch strings.ToUpper(tblName) {
	case "LAYER":
		st = table.NewLayer(entry, 7, nil)
	case "LTYPE":
		st = table.NewLineType(entry, entry)
	case "STYLE":
		st = table.NewStyle(entry)
	case "DIMSTYLE":
		st = table.NewDimStyle(entry)
	case "APPID":
		st = table.NewAppID(entry)
	default:
		return false
	}
	st.SetOwner(t)
	t.Add(st)
	return true
}

// fixMissingLayer retags an entity to layer "0" (documented default owner).
func fixMissingLayer(ctx *Context, issue Issue) bool {
	obj, ok := ctx.Resolve(issue.Handle)
	if !ok {
		return false
	}
	if layers := ctx.EntryNames("LAYER"); !layers["0"] {
		return false // fix order guarantees layer 0 exists first
	}
	e, ok := obj.(layerNameSetter)
	if !ok {
		return false
	}
	e.SetLayerName("0")
	return true
}

// collect runs every registered check and returns all issues.
func collect(ctx *Context) []Issue {
	var out []Issue
	for _, c := range checksSnapshot() {
		for _, is := range c.Fn(ctx) {
			// A check function may report several codes; the registry binds
			// each code to one check, so keep only this code's issues.
			if is.Code == c.Code {
				out = append(out, is)
			}
		}
	}
	return out
}

// Run audits the document. With opts.Fix it applies safe fixers, re-audits
// and returns a report in which repaired issues carry Fixed=true and the
// remaining issues come from the re-audit. Unsafe issues are never fixed.
func Run(doc *drawing.Drawing, opts *Options) *Report {
	if opts == nil {
		opts = &Options{}
	}
	rep := &Report{}
	if doc == nil {
		rep.Addf(SeverityCritical, "NO_DOCUMENT", "", "no document to audit")
		return rep
	}
	ctx := NewContext(doc)
	first := collect(ctx)
	if !opts.Fix {
		for _, it := range first {
			rep.Add(it)
		}
		addStrictR12(rep, doc, opts)
		return rep
	}
	for _, it := range first {
		f, ok := fixers[it.Code]
		if !ok {
			continue
		}
		if f(ctx, it) {
			it.Fixed = true
			rep.Add(it)
		}
	}
	for _, it := range collect(NewContext(doc)) {
		rep.Add(it)
	}
	addStrictR12(rep, doc, opts)
	return rep
}

// CodeR12Incompatible marks OBJECTS/classes content in an R12 file.
const CodeR12Incompatible = "R12_INCOMPATIBLE"

// addStrictR12 reports OBJECTS and app-provided-classes content when the
// version profile (R12) does not support those sections.
func addStrictR12(rep *Report, doc *drawing.Drawing, opts *Options) {
	if !opts.StrictR12 {
		return
	}
	h := doc.Header()
	if h == nil {
		return
	}
	profile := format.Profile(h.Version)
	if profile == nil || profile.SupportsObjects {
		return
	}
	var found []string
	if len(doc.Objects()) > 0 {
		found = append(found, "OBJECTS section")
	}
	if len(doc.Classes()) > 0 {
		found = append(found, "CLASSES section")
	}
	if sec := doc.OpaqueSections(); sec != nil {
		for _, name := range []string{"OBJECTS", "CLASSES"} {
			if tags, ok := sec[name]; ok && len(tags) > 0 {
				found = append(found, name+" section (raw)")
			}
		}
	}
	if len(found) == 0 {
		return
	}
	rep.Addf(SeverityError, CodeR12Incompatible, "",
		"version %s does not support %s", h.Version, strings.Join(found, " and "))
}
