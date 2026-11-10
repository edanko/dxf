// Document-wide audit context: handle map, object resolution, check registry.
package audit

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/table"
)

type stringOwner interface {
	Owner() string
}

type dxfTypeReader interface {
	DXFType() string
}

type objectNameReader interface {
	Name() string
}

type handleReader interface {
	Handle() string
}

// Context carries the document under audit plus a document-wide handle map
// built once per run (duplicates and dangling references are detected against
// it in a single pass).
type Context struct {
	Doc *drawing.Drawing

	objs     map[string][]any
	entOwner map[string]*block.Block // entity handle -> containing block (for fixes)
	blocks   []*block.Block
	seen     map[any]bool
}

// NewContext builds the handle map for the whole document: table entries,
// blocks, block entities and objects.
func NewContext(doc *drawing.Drawing) *Context {
	c := &Context{
		Doc:      doc,
		objs:     map[string][]any{},
		entOwner: map[string]*block.Block{},
		seen:     map[any]bool{},
	}
	if doc == nil {
		return c
	}
	// Tables.
	if doc.Tables() != nil {
		for _, t := range doc.Tables() {
			add(c, t.Handle(), t)
			for _, e := range t.Entries() {
				add(c, e.Handle(), e)
			}
		}
	}
	// Blocks and their entities.
	for _, b := range doc.Blocks() {
		add(c, b.Handle(), b)
	}
	c.blocks = doc.Blocks()
	for _, b := range c.blocks {
		for _, e := range b.Entities() {
			if c.seen[e] {
				continue
			}
			c.seen[e] = true
			add(c, e.Handle(), e)
			if h := e.Handle(); h != "" {
				c.entOwner[h] = b
			}
		}
	}
	// Loose model-space entities (not inside any block record): ownership
	// default is the *Model_Space block.
	mb := c.ModelBlock()
	for _, e := range doc.Entities() {
		if c.seen[e] {
			continue
		}
		c.seen[e] = true
		add(c, e.Handle(), e)
		if mb != nil {
			if h := e.Handle(); h != "" {
				c.entOwner[h] = mb
			}
		}
	}
	// Objects.
	for _, o := range doc.Objects() {
		add(c, o.Handle(), o)
	}
	return c
}

func add(c *Context, h string, obj any) {
	if h == "" {
		return
	}
	c.objs[h] = append(c.objs[h], obj)
}

// Resolve returns the single object registered for handle h.
func (c *Context) Resolve(h string) (any, bool) {
	objs, ok := c.objs[h]
	if !ok || len(objs) == 0 {
		return nil, false
	}
	return objs[0], true
}

// Objects returns every object registered under handle h (len > 1 means the
// handle is duplicated).
func (c *Context) Objects(h string) []any {
	return c.objs[h]
}

// Handles returns all registered handles, sorted.
func (c *Context) Handles() []string {
	hs := make([]string, 0, len(c.objs))
	for h := range c.objs {
		hs = append(hs, h)
	}
	sort.Strings(hs)
	return hs
}

// Entries iterates every entry of every table.
func (c *Context) Entries() []table.SymbolTable {
	var out []table.SymbolTable
	if c.Doc == nil || c.Doc.Tables() == nil {
		return out
	}
	for _, t := range c.Doc.Tables() {
		out = append(out, t.Entries()...)
	}
	return out
}

// TableByName returns the table with the given name (VPORT, LTYPE, LAYER,
// STYLE, VIEW, UCS, APPID, DIMSTYLE, BLOCK_RECORD).
func (c *Context) TableByName(name string) *table.Table {
	if c.Doc == nil || c.Doc.Tables() == nil {
		return nil
	}
	for _, t := range c.Doc.Tables() {
		if strings.EqualFold(t.Name(), name) {
			return t
		}
	}
	return nil
}

// EntryNames returns the set of entry names in a table.
func (c *Context) EntryNames(table string) map[string]bool {
	m := map[string]bool{}
	t := c.TableByName(table)
	if t == nil {
		return m
	}
	for _, e := range t.Entries() {
		m[e.Name()] = true
	}
	return m
}

// Blocks returns every block of the document.
func (c *Context) Blocks() []*block.Block { return c.blocks }

// BlockNames returns the set of defined block names.
func (c *Context) BlockNames() map[string]bool {
	m := map[string]bool{}
	for _, b := range c.blocks {
		m[b.Name] = true
	}
	return m
}

// ModelBlock returns the *Model_Space block, nil when absent.
func (c *Context) ModelBlock() *block.Block {
	for _, b := range c.blocks {
		if b.Name == "*Model_Space" {
			return b
		}
	}
	return nil
}

// AllEntities returns every entity of the document, deduplicated: entities
// reached through blocks plus loose model-space entities. The dynamic types
// are concrete entity models; access capabilities via structural assertions.
func (c *Context) AllEntities() []any {
	var out []any
	done := map[any]bool{}
	for _, b := range c.blocks {
		for _, be := range b.Entities() {
			// Adapters (raw-tag wrappers) hide the entity method set; audit
			// the unwrapped model so owner/layer/bbox checks see it.
			e := any(be)
			if u, ok := be.(entity.Unwrapper); ok {
				e = u.Unwrap()
			}
			if done[e] {
				continue
			}
			done[e] = true
			out = append(out, e)
		}
	}
	for _, e := range c.Doc.Entities() {
		if done[e] {
			continue
		}
		done[e] = true
		out = append(out, e)
	}
	return out
}

// ContainingBlock returns the block that holds the entity with the given
// handle (nil when the entity is loose).
func (c *Context) ContainingBlock(entityHandle string) *block.Block {
	return c.entOwner[entityHandle]
}

// EntryOwner returns the owner handler of a table entry (nil when the entry
// does not track ownership).
func EntryOwner(e table.SymbolTable) handle.Handler {
	if o, ok := e.(entity.OwnerGetter); ok {
		return o.Owner()
	}
	return nil
}

// EntityOwner returns the owner handler of an entity.
func EntityOwner(e any) handle.Handler {
	if o, ok := e.(entity.OwnerGetter); ok {
		return o.Owner()
	}
	return nil
}

// ObjectOwnerHandle returns the owner handle recorded on an object.
func ObjectOwnerHandle(o any) string {
	if s, ok := o.(stringOwner); ok {
		return s.Owner()
	}
	return ""
}

// ObjectReactors returns parsed reactor links of an object, nil when the
// object model does not track them.
func ObjectReactors(o any) []string {
	if r, ok := o.(entity.ReactorReader); ok {
		return r.ReactorHandles()
	}
	return nil
}

// Describe renders a short human-readable identity for an audited object.
func Describe(obj any) string {
	if obj == nil {
		return "<nil>"
	}
	typeStr := ""
	if d, ok := obj.(dxfTypeReader); ok {
		typeStr = d.DXFType()
	}
	nameStr := ""
	if n, ok := obj.(objectNameReader); ok {
		nameStr = n.Name()
	}
	hStr := ""
	if h, ok := obj.(handleReader); ok {
		hStr = h.Handle()
	}
	if typeStr == "" {
		typeStr = fmt.Sprintf("%T", obj)
	}
	switch {
	case nameStr != "" && hStr != "":
		return fmt.Sprintf("%s %q (handle %s)", typeStr, nameStr, hStr)
	case nameStr != "":
		return fmt.Sprintf("%s %q", typeStr, nameStr)
	case hStr != "":
		return fmt.Sprintf("%s (handle %s)", typeStr, hStr)
	}
	return typeStr
}

// Check is a named document check. Safe fixers are applied in fix mode;
// unsafe issues are only reported.
type Check struct {
	Code     string
	Severity Severity
	Safe     bool
	Fn       func(*Context) []Issue
}

var (
	registryMu sync.RWMutex
	registry   = map[string]Check{}
	order      []string
)

// RegisterCheck registers (or replaces) the check for a code, preserving the
// registration order of new codes.
func RegisterCheck(c Check) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := registry[c.Code]; !exists {
		order = append(order, c.Code)
	}
	registry[c.Code] = c
}

// CheckCodes returns the registered check codes in registration order.
func CheckCodes() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]string, len(order))
	copy(out, order)
	return out
}

func checksSnapshot() []Check {
	registryMu.RLock()
	defer registryMu.RUnlock()
	out := make([]Check, 0, len(order))
	for _, code := range order {
		out = append(out, registry[code])
	}
	return out
}

// EntityHook validates one entity of a given DXF type.
type EntityHook func(e any, ctx *Context) []Issue

var (
	hooksMu sync.RWMutex
	hooks   = map[string]EntityHook{}
)

// RegisterEntityHook installs a per-type entity invariant hook. Hooks are
// registered here (never in the entity package) to avoid import cycles; the
// document audit invokes them for every entity whose DXFType matches.
func RegisterEntityHook(dxfType string, fn EntityHook) {
	hooksMu.Lock()
	defer hooksMu.Unlock()
	hooks[strings.ToUpper(dxfType)] = fn
}

func entityHook(dxfType string) (EntityHook, bool) {
	hooksMu.RLock()
	defer hooksMu.RUnlock()
	fn, ok := hooks[strings.ToUpper(dxfType)]
	return fn, ok
}

// EntityHandle returns the handle of an entity of any model.
func EntityHandle(e any) string {
	if h, ok := e.(handleReader); ok {
		return h.Handle()
	}
	return ""
}

// EntityType returns the DXF type of an entity of any model.
func EntityType(e any) string {
	if t, ok := e.(dxfTypeReader); ok {
		return t.DXFType()
	}
	return ""
}
