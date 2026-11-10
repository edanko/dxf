// External-reference facade: attach/detach/bind/reload, path resolution and
// cycle-aware nested loading, backed by the xref package and BLOCK_RECORD
// xref flags.
package drawing

import (
	"fmt"
	"io"
	"path/filepath"

	dxfmath "github.com/edanko/dxf/math"

	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/handle"
	"github.com/edanko/dxf/table"
	"github.com/edanko/dxf/xref"
)

// documentLoader is registered by the root package (dxf_main.go) so this
// package can load nested documents without an import cycle.
var documentLoader func(io.Reader) (*Drawing, error)

// RegisterDocumentLoader installs the function used to parse nested documents.
func RegisterDocumentLoader(fn func(io.Reader) (*Drawing, error)) {
	documentLoader = fn
}

// ownerHandle adapts a handle string to the entity owner interface.
type ownerHandle string

func (o ownerHandle) Handle() string                    { return string(o) }
func (o ownerHandle) SetHandle(*handle.HandleGenerator) {}

// AttachXRef attaches a DXF file as an external reference: it creates the
// BLOCK_RECORD (marked as xref with the relative path in code 305), a block
// definition holding the reference content, and a model-space INSERT at
// insertPoint.
func (d *Drawing) AttachXRef(filePath, blockName string, insertPoint []float64, overlay bool) (*entity.Insert, error) {
	if filePath == "" {
		return nil, fmt.Errorf("attach xref: empty file path")
	}
	if blockName == "" {
		return nil, fmt.Errorf("attach xref: empty block name")
	}
	if _, exists := d.xrefManager.Get(blockName); exists {
		return nil, fmt.Errorf("attach xref: %s already attached", blockName)
	}
	relPath := filePath
	if basePath := filepath.Dir(d.fileName); basePath != "" {
		if r, err := filepath.Rel(basePath, filePath); err == nil && r != "." {
			relPath = r
		}
	}
	xrefType := xref.XRefTypeAttachment
	if overlay {
		xrefType = xref.XRefTypeOverlay
	}
	if err := d.xrefManager.AddXRef(relPath, blockName, xrefType); err != nil {
		return nil, err
	}

	blk, err := d.NewBlock(blockName, "external reference")
	if err != nil {
		d.xrefManager.Remove(blockName)
		return nil, err
	}
	t := d.blockRecordTable()
	entry, err := t.Lookup(blockName)
	if err != nil {
		d.detachArtifacts(blockName)
		return nil, err
	}
	br, ok := entry.(*table.BlockRecord)
	if !ok {
		d.detachArtifacts(blockName)
		return nil, fmt.Errorf("attach xref: block record %s is %T", blockName, entry)
	}
	br.SetXRef(true)
	br.SetXRefOverlay(overlay)
	br.SetXRefPath(relPath)
	br.SetXRefResolved(false)

	ins := entity.NewInsert()
	ins.SetBlockName(blockName)
	ins.SetBlockDef(blk)
	ins.SetExternalXref(true)
	var pt dxfmath.Vec3
	if len(insertPoint) > 0 {
		pt[0] = insertPoint[0]
	}
	if len(insertPoint) > 1 {
		pt[1] = insertPoint[1]
	}
	if len(insertPoint) > 2 {
		pt[2] = insertPoint[2]
	}
	ins.SetInsertPoint(pt)
	if mh, ok := d.blockRecordHandle("*Model_Space"); ok && mh != "" {
		ins.SetOwner(ownerHandle(mh))
	}
	if err := d.AddEntity(ins); err != nil {
		return nil, err
	}
	return ins, nil
}

// detachArtifacts removes the block, block record and manager entry of a
// reference (used for cleanup and detach).
func (d *Drawing) detachArtifacts(blockName string) {
	_ = d.DeleteBlock(blockName)
	d.xrefManager.Remove(blockName)
}

// DetachXRef removes an attached reference: its inserts (model space and
// block definitions), its block definition and its BLOCK_RECORD.
func (d *Drawing) DetachXRef(blockName string) error {
	if _, exists := d.xrefManager.Get(blockName); !exists {
		return fmt.Errorf("detach xref: %s not attached", blockName)
	}
	var dead []string
	for _, e := range d.Entities() {
		if ins, ok := e.(*entity.Insert); ok && ins.BlockName() == blockName {
			dead = append(dead, e.Handle())
		}
	}
	for _, h := range dead {
		if err := d.DeleteEntity(h); err != nil {
			return err
		}
	}
	for _, b := range d.Blocks() {
		if b == nil {
			continue
		}
		var kept []block.BlockEntity
		for _, be := range b.Entities() {
			drop := false
			if ee, ok := be.(entity.Entity); ok {
				if ins, ok := ee.(*entity.Insert); ok && ins.BlockName() == blockName {
					drop = true
				}
			}
			if !drop {
				kept = append(kept, be)
			}
		}
		b.SetEntities(kept)
	}
	if err := d.DeleteBlock(blockName); err != nil {
		return err
	}
	d.xrefManager.Remove(blockName)
	return nil
}

// ReloadXRef re-reads the referenced file and refreshes the reference block
// definition. A missing file is reported (recovery issue + missing status)
// instead of failing the host document.
func (d *Drawing) ReloadXRef(blockName string) error {
	x, exists := d.xrefManager.Get(blockName)
	if !exists {
		return fmt.Errorf("xref %s not found", blockName)
	}
	d.xrefLoader.SetBasePath(filepath.Dir(d.fileName))
	path, found := d.xrefLoader.Resolve(x)
	if !found {
		x.SetStatus(xref.XRefStatusMissing)
		d.addRecoveryIssue("unresolved external reference: %s", path)
		return &xref.UnresolvedError{Path: path}
	}
	child, err := LoadDXF(path)
	if err != nil {
		x.SetStatus(xref.XRefStatusUnresolved)
		d.addRecoveryIssue("failed to load xref %s: %v", x.FilePath(), err)
		return err
	}
	d.replaceXRefContent(blockName, child)
	x.SetDrawing(child)
	x.SetLoaded(true)
	d.markXRefResolved(blockName)
	return nil
}

// BindXRef converts an external reference into a local block: the referenced
// content is copied into the (existing) block definition, the xref marks on
// the BLOCK_RECORD and on the referencing inserts are cleared, and the
// external path is removed.
func (d *Drawing) BindXRef(blockName string, insert bool) error {
	x, exists := d.xrefManager.Get(blockName)
	if !exists {
		return fmt.Errorf("xref %s not found", blockName)
	}
	if !x.Loaded() {
		if err := d.ReloadXRef(blockName); err != nil {
			return err
		}
	}
	xdwg, err := d.GetXRefDrawing(blockName)
	if err != nil {
		return err
	}
	blk, err := d.GetBlock(blockName)
	if err != nil {
		return err
	}
	owner := d.recordHandle(blockName)
	blk.SetEntities(nil)
	for _, e := range xdwg.Entities() {
		cp := copyForBlock(e, owner)
		if cp == nil {
			continue
		}
		blk.AddEntity(asBlockEntity(cp))
	}
	if t := d.blockRecordTable(); t != nil {
		if entry, err := t.Lookup(blockName); err == nil {
			if br, ok := entry.(*table.BlockRecord); ok {
				br.SetXRef(false)
				br.SetXRefPath("")
			}
		}
	}
	for _, e := range d.Entities() {
		if ins, ok := e.(*entity.Insert); ok && ins.BlockName() == blockName {
			ins.SetExternalXref(false)
		}
	}
	d.xrefManager.Remove(blockName)
	return nil
}

// copyForBlock deep-copies an entity, clears its handle and re-parents it to
// the owning block record so it can be stored in a block definition.
func copyForBlock(e entity.Entity, owner string) entity.Entity {
	c, ok := e.(entity.Copier)
	if !ok {
		return nil
	}
	cp := c.Copy()
	if cp == nil {
		return nil
	}
	if s, ok := cp.(interface{ SetHandleString(string) }); ok {
		s.SetHandleString("")
	}
	if owner != "" {
		if s, ok := cp.(interface{ SetOwner(handle.Handler) }); ok {
			s.SetOwner(ownerHandle(owner))
		}
	}
	return cp
}

// replaceXRefContent swaps the entities of the reference block with the
// model entities of a freshly loaded child document.
func (d *Drawing) replaceXRefContent(blockName string, child *Drawing) {
	blk, err := d.GetBlock(blockName)
	if err != nil || blk == nil {
		return
	}
	owner := d.recordHandle(blockName)
	blk.SetEntities(nil)
	if child == nil {
		return
	}
	for _, e := range child.Entities() {
		cp := copyForBlock(e, owner)
		if cp == nil {
			continue
		}
		blk.AddEntity(asBlockEntity(cp))
	}
}

// recordHandle returns the BLOCK_RECORD handle of a block name.
func (d *Drawing) recordHandle(blockName string) string {
	t := d.blockRecordTable()
	if t == nil {
		return ""
	}
	entry, err := t.Lookup(blockName)
	if err != nil {
		return ""
	}
	return entry.Handle()
}

func (d *Drawing) markXRefResolved(blockName string) {
	if t := d.blockRecordTable(); t != nil {
		if entry, err := t.Lookup(blockName); err == nil {
			if br, ok := entry.(*table.BlockRecord); ok {
				br.SetXRefResolved(true)
			}
		}
	}
}

// addRecoveryIssue records a load-time problem on the drawing recovery report
// without failing the load.
func (d *Drawing) addRecoveryIssue(format string, args ...any) {
	rep := d.RecoveryReport()
	if rep == nil {
		rep = &RecoveryReport{}
	}
	rep.Issues = append(rep.Issues, fmt.Sprintf(format, args...))
	d.SetRecoveryReport(rep)
}

// XRefStatusInfo describes one external reference.
type XRefStatusInfo struct {
	BlockName string
	FilePath  string
	Resolved  bool
	Status    string
}

// XRefStatuses lists the resolution state of all attached references.
func (d *Drawing) XRefStatuses() []XRefStatusInfo {
	xs := d.xrefManager.GetAll()
	out := make([]XRefStatusInfo, 0, len(xs))
	for _, x := range xs {
		info := XRefStatusInfo{
			BlockName: x.BlockName(),
			FilePath:  x.FilePath(),
			Status:    x.Status().String(),
		}
		if t := d.blockRecordTable(); t != nil {
			if entry, err := t.Lookup(x.BlockName()); err == nil {
				if br, ok := entry.(*table.BlockRecord); ok {
					info.Resolved = br.IsXRefResolved()
				}
			}
		}
		out = append(out, info)
	}
	return out
}

// XRefCycles returns the reference cycles detected during recursive loading.
func (d *Drawing) XRefCycles() [][]string {
	out := make([][]string, len(d.xrefCycles))
	copy(out, d.xrefCycles)
	return out
}

// AddXRefSearchPath appends a directory searched for relative references.
func (d *Drawing) AddXRefSearchPath(p string) {
	d.xrefLoader.AddSearchPath(p)
	d.xrefManager.AddSearchPath(p)
}

// LoadXRefs loads every attached reference, replacing block content. Missing
// files are reported on the recovery report and do not fail the load.
func (d *Drawing) LoadXRefs() error {
	d.xrefLoader.SetBasePath(filepath.Dir(d.fileName))
	for _, x := range d.xrefManager.GetAll() {
		if x.Loaded() {
			continue
		}
		path, found := d.xrefLoader.Resolve(x)
		if !found {
			x.SetStatus(xref.XRefStatusMissing)
			d.addRecoveryIssue("unresolved external reference: %s", path)
			continue
		}
		child, err := LoadDXF(path)
		if err != nil {
			x.SetStatus(xref.XRefStatusUnresolved)
			d.addRecoveryIssue("failed to load xref %s: %v", x.FilePath(), err)
			continue
		}
		d.replaceXRefContent(x.BlockName(), child)
		x.SetDrawing(child)
		x.SetLoaded(true)
		d.markXRefResolved(x.BlockName())
	}
	return nil
}

// LoadXRefsRecursive loads attached references and the references of the
// loaded documents, terminating and reporting cycles.
func (d *Drawing) LoadXRefsRecursive() error {
	d.xrefLoader.SetBasePath(filepath.Dir(d.fileName))
	var visited map[string]bool
	if d.fileName != "" {
		hostPath, err := filepath.Abs(d.fileName)
		if err == nil {
			if resolved, err := filepath.EvalSymlinks(hostPath); err == nil {
				hostPath = resolved
			}
			visited = map[string]bool{hostPath: true}
		}
	}
	return d.loadXRefsRecursive(0, 32, visited)
}

func (d *Drawing) loadXRefsRecursive(depth, maxDepth int, visited map[string]bool) error {
	if depth > maxDepth {
		return fmt.Errorf("xref nesting too deep (max: %d)", maxDepth)
	}
	if visited == nil {
		visited = map[string]bool{}
	}
	for _, x := range d.xrefManager.GetAll() {
		if x.Loaded() {
			continue
		}
		path, found := d.xrefLoader.Resolve(x)
		if !found {
			x.SetStatus(xref.XRefStatusMissing)
			d.addRecoveryIssue("unresolved external reference: %s", path)
			continue
		}
		canonical := path
		if abs, err := filepath.Abs(path); err == nil {
			canonical = abs
		}
		if resolved, err := filepath.EvalSymlinks(canonical); err == nil {
			canonical = resolved
		}
		if visited[canonical] {
			d.addRecoveryIssue("xref cycle at %s", canonical)
			d.xrefCycles = append(d.xrefCycles, []string{canonical})
			continue
		}
		visited[canonical] = true
		child, err := LoadDXF(path)
		if err != nil {
			x.SetStatus(xref.XRefStatusUnresolved)
			d.addRecoveryIssue("failed to load xref %s: %v", x.FilePath(), err)
			continue
		}
		d.replaceXRefContent(x.BlockName(), child)
		x.SetDrawing(child)
		x.SetLoaded(true)
		d.markXRefResolved(x.BlockName())
		child.xrefLoader.SetBasePath(filepath.Dir(path))
		if err := child.loadXRefsRecursive(depth+1, maxDepth, visited); err != nil {
			return err
		}
		d.xrefCycles = append(d.xrefCycles, child.XRefCycles()...)
	}
	return nil
}

// RestoreXRefRecords re-registers external references from the BLOCK_RECORD
// xref flags so documents loaded from disk can resolve them again.
func (d *Drawing) RestoreXRefRecords() {
	t := d.blockRecordTable()
	if t == nil {
		return
	}
	for _, entry := range t.Entries() {
		br, ok := entry.(*table.BlockRecord)
		if !ok || !br.IsXRef() {
			continue
		}
		if _, exists := d.xrefManager.Get(br.Name()); exists {
			continue
		}
		xt := xref.XRefTypeAttachment
		if br.IsXRefOverlay() {
			xt = xref.XRefTypeOverlay
		}
		_ = d.xrefManager.AddXRef(br.XRefPath(), br.Name(), xt)
	}
}
