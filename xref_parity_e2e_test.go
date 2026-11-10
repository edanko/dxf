// End-to-end tests for external references: attach/load/detach/bind, missing
// reporting, nested loading and cycle termination.
package dxf_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edanko/dxf"
	"github.com/edanko/dxf/audit"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
)

func writeDrawing(t *testing.T, d *drawing.Drawing, path string) {
	t.Helper()
	var buf bytes.Buffer
	if _, err := d.WriteTo(&buf); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("save %s: %v", path, err)
	}
}

func makeBaseDrawing(t *testing.T, path string) {
	t.Helper()
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddLine(0, 0, 0, 10, 0, 0); err != nil {
		t.Fatal(err)
	}
	writeDrawing(t, d, path)
}

func auditErrorCount(doc *drawing.Drawing) int {
	rep := audit.Run(doc, nil)
	return rep.BySeverity()["error"]
}

func TestAttachLoadDetachXRefE2E(t *testing.T) {
	dir := t.TempDir()
	basePath := filepath.Join(dir, "base.dxf")
	makeBaseDrawing(t, basePath)

	host, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	ins, err := host.AttachXRef(basePath, "BASE", []float64{5, 5, 0}, false)
	if err != nil {
		t.Fatal(err)
	}
	if ins == nil || ins.BlockName() != "BASE" {
		t.Fatalf("attach insert: %v", ins)
	}
	if st := host.XRefStatuses(); len(st) != 1 || st[0].Status != "unresolved" || st[0].Resolved {
		t.Fatalf("status before load: %+v", st)
	}
	hostPath := filepath.Join(dir, "host.dxf")
	host.SetFileName(hostPath)
	writeDrawing(t, host, hostPath)

	doc, err := dxf.FromFileWithXRefs(hostPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	st := doc.XRefStatuses()
	if len(st) != 1 || st[0].Status != "loaded" || !st[0].Resolved {
		t.Fatalf("status after load: %+v", st)
	}
	child, err := doc.GetXRefDrawing("BASE")
	if err != nil {
		t.Fatal(err)
	}
	if len(child.Entities()) != 1 {
		t.Fatalf("child content: %d entities", len(child.Entities()))
	}
	if n := len(doc.Entities()); n != 1 {
		t.Fatalf("host model entities: %d", n)
	}
	if errs := auditErrorCount(doc); errs != 0 {
		rep := audit.Run(doc, nil)
		t.Fatalf("audit errors after load: %d (%v)", errs, rep.Issues)
	}

	if err := doc.DetachXRef("BASE"); err != nil {
		t.Fatal(err)
	}
	if len(doc.Entities()) != 0 {
		t.Fatalf("detach must remove the insert: %d", len(doc.Entities()))
	}
	if st := doc.XRefStatuses(); len(st) != 0 {
		t.Fatalf("detach must remove the reference: %+v", st)
	}
	for _, name := range doc.GetBlockNames() {
		if strings.HasPrefix(name, "*XREF:") || name == "BASE" {
			t.Fatalf("detach left block %q", name)
		}
	}
	if errs := auditErrorCount(doc); errs != 0 {
		rep := audit.Run(doc, nil)
		t.Fatalf("audit errors after detach: %d (%v)", errs, rep.Issues)
	}
}

func TestBindXRefE2E(t *testing.T) {
	dir := t.TempDir()
	basePath := filepath.Join(dir, "base.dxf")
	makeBaseDrawing(t, basePath)

	host, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := host.AttachXRef(basePath, "BASE", nil, false); err != nil {
		t.Fatal(err)
	}
	host.SetFileName(filepath.Join(dir, "host.dxf"))
	if err := host.LoadXRefs(); err != nil {
		t.Fatal(err)
	}
	if err := host.BindXRef("BASE", true); err != nil {
		t.Fatal(err)
	}
	if st := host.XRefStatuses(); len(st) != 0 {
		t.Fatalf("bind must remove the reference: %+v", st)
	}
	blk, err := host.GetBlock("BASE")
	if err != nil {
		t.Fatalf("bound block: %v", err)
	}
	if len(blk.Entities()) != 1 {
		t.Fatalf("bound content: %d entities", len(blk.Entities()))
	}
	var ins *entity.Insert
	for _, e := range host.Entities() {
		if i, ok := e.(*entity.Insert); ok && i.BlockName() == "BASE" {
			ins = i
		}
	}
	if ins == nil {
		t.Fatal("bound insert missing")
	}
	if ins.Flags()&entity.InsertFlagExternalXref != 0 {
		t.Fatal("bound insert must not be external")
	}
	if errs := auditErrorCount(host); errs != 0 {
		rep := audit.Run(host, nil)
		t.Fatalf("audit errors after bind: %d (%v)", errs, rep.Issues)
	}
}

func TestMissingXRefReportedNotFailed(t *testing.T) {
	dir := t.TempDir()
	host, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	ghost := filepath.Join(dir, "ghost.dxf")
	if _, err := host.AttachXRef(ghost, "GHOST", nil, false); err != nil {
		t.Fatal(err)
	}
	host.SetFileName(filepath.Join(dir, "host.dxf"))
	if err := host.LoadXRefs(); err != nil {
		t.Fatalf("missing xref must not fail the load: %v", err)
	}
	st := host.XRefStatuses()
	if len(st) != 1 || st[0].Status != "missing" {
		t.Fatalf("missing status: %+v", st)
	}
	rep := host.RecoveryReport()
	if rep == nil {
		t.Fatal("recovery report expected")
	}
	found := false
	for _, issue := range rep.Issues {
		if strings.Contains(issue, "unresolved external reference") {
			found = true
		}
	}
	if !found {
		t.Fatalf("issues: %v", rep.Issues)
	}
}

func TestNestedXRefsLoad(t *testing.T) {
	dir := t.TempDir()
	leafPath := filepath.Join(dir, "leaf.dxf")
	makeBaseDrawing(t, leafPath)

	mid, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mid.AttachXRef(leafPath, "LEAF", nil, false); err != nil {
		t.Fatal(err)
	}
	midPath := filepath.Join(dir, "mid.dxf")
	mid.SetFileName(midPath)
	writeDrawing(t, mid, midPath)

	host, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := host.AttachXRef(midPath, "MID", nil, false); err != nil {
		t.Fatal(err)
	}
	hostPath := filepath.Join(dir, "host.dxf")
	host.SetFileName(hostPath)
	writeDrawing(t, host, hostPath)

	doc, err := dxf.FromFileWithXRefs(hostPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	st := doc.XRefStatuses()
	if len(st) != 1 || st[0].BlockName != "MID" || st[0].Status != "loaded" {
		t.Fatalf("host status: %+v", st)
	}
	midChild, err := doc.GetXRefDrawing("MID")
	if err != nil {
		t.Fatal(err)
	}
	nested := midChild.XRefStatuses()
	if len(nested) != 1 || nested[0].BlockName != "LEAF" || nested[0].Status != "loaded" {
		t.Fatalf("nested status: %+v", nested)
	}
	leafChild, err := midChild.GetXRefDrawing("LEAF")
	if err != nil {
		t.Fatal(err)
	}
	if len(leafChild.Entities()) != 1 {
		t.Fatalf("leaf content: %d", len(leafChild.Entities()))
	}
}

func TestXRefCycleTerminates(t *testing.T) {
	dir := t.TempDir()
	aPath := filepath.Join(dir, "a.dxf")
	bPath := filepath.Join(dir, "b.dxf")

	a, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.AttachXRef(bPath, "B", nil, false); err != nil {
		t.Fatal(err)
	}
	a.SetFileName(aPath)
	writeDrawing(t, a, aPath)

	b, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.AttachXRef(aPath, "A", nil, false); err != nil {
		t.Fatal(err)
	}
	b.SetFileName(bPath)
	writeDrawing(t, b, bPath)

	doc, err := dxf.FromFileWithXRefs(aPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cycles := doc.XRefCycles(); len(cycles) == 0 {
		t.Fatal("cycle must be reported")
	}
	st := doc.XRefStatuses()
	if len(st) != 1 || st[0].BlockName != "B" || st[0].Status != "loaded" {
		t.Fatalf("host status: %+v", st)
	}
	child, err := doc.GetXRefDrawing("B")
	if err != nil {
		t.Fatal(err)
	}
	nested := child.XRefStatuses()
	if len(nested) != 1 || nested[0].BlockName != "A" || nested[0].Status != "unresolved" {
		t.Fatalf("nested cycle status: %+v", nested)
	}
}
