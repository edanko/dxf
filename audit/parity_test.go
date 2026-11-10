package audit

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edanko/dxf"
	"github.com/edanko/dxf/block"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/handle"
)

// baseDXF is a minimal, structurally complete R2000 document used as the
// fixture for all parity tests. Mutations are applied with string replaces.
const baseDXF = `0
SECTION
2
HEADER
9
$ACADVER
1
AC1015
0
ENDSEC
0
SECTION
2
TABLES
0
TABLE
2
LAYER
5
10
330
0
100
AcDbSymbolTable
70
1
0
LAYER
5
11
330
10
100
AcDbSymbolTable
100
AcDbLayerTableRecord
70
0
2
0
62
7
6
Continuous
370
-3
0
ENDTAB
0
TABLE
2
LTYPE
5
20
330
0
100
AcDbSymbolTable
70
3
0
LTYPE
5
21
330
20
100
AcDbSymbolTable
100
AcDbLinetypeTableRecord
70
0
2
ByLayer
72
65
73
0
40
0.0
0
LTYPE
5
22
330
20
100
AcDbSymbolTable
100
AcDbLinetypeTableRecord
70
0
2
ByBlock
72
65
73
0
40
0.0
0
LTYPE
5
23
330
20
100
AcDbSymbolTable
100
AcDbLinetypeTableRecord
70
0
2
Continuous
72
65
73
0
40
0.0
0
ENDTAB
0
TABLE
2
STYLE
5
30
330
0
100
AcDbSymbolTable
70
1
0
STYLE
5
31
330
30
100
AcDbSymbolTable
100
AcDbTextStyleTableRecord
70
0
2
Standard
40
0.0
41
1.0
50
0.0
71
0
42
2.5
43
0.0
0
ENDTAB
0
TABLE
2
VPORT
5
80
330
0
100
AcDbSymbolTable
70
1
0
VPORT
5
81
330
80
100
AcDbSymbolTable
100
AcDbViewportTableRecord
70
0
2
*
0
ENDTAB
0
TABLE
2
VIEW
5
90
330
0
100
AcDbSymbolTable
70
0
0
ENDTAB
0
TABLE
2
UCS
5
B0
330
0
100
AcDbSymbolTable
70
0
0
ENDTAB
0
TABLE
2
APPID
5
A0
330
0
100
AcDbSymbolTable
70
1
0
APPID
5
A1
330
A0
100
AcDbSymbolTable
100
AcDbRegAppTableRecord
70
0
2
ACAD
0
ENDTAB
0
TABLE
2
DIMSTYLE
5
70
330
0
100
AcDbSymbolTable
70
1
100
AcDbDimStyleTable
0
DIMSTYLE
5
71
330
70
100
AcDbSymbolTable
100
AcDbDimStyleTableRecord
70
0
2
Standard
0
ENDTAB
0
TABLE
2
BLOCK_RECORD
5
D0
330
0
100
AcDbSymbolTable
70
2
0
BLOCK_RECORD
5
E
330
D0
100
AcDbSymbolTable
100
AcDbBlockTableRecord
70
0
280
1
2
*Model_Space
0
BLOCK_RECORD
5
2E
330
D0
100
AcDbSymbolTable
100
AcDbBlockTableRecord
70
0
280
1
2
*Paper_Space
0
ENDTAB
0
ENDSEC
0
SECTION
2
BLOCKS
0
BLOCK
5
F
330
E
100
AcDbBlockBegin
70
0
8
0
2
*Model_Space
10
0.0
20
0.0
30
0.0
1
*Model_Space
3
*Model_Space
0
ENDBLK
5
6F
330
E
100
AcDbBlockEnd
0
BLOCK
5
2F
330
2E
100
AcDbBlockBegin
70
0
8
0
2
*Paper_Space
10
0.0
20
0.0
30
0.0
1
X
3
X
0
ENDBLK
5
7F
330
2E
100
AcDbBlockEnd
0
ENDSEC
0
SECTION
2
ENTITIES
0
LINE
5
41
330
E
100
AcDbEntity
8
0
100
AcDbLine
10
0.0
20
0.0
30
0.0
11
10.0
21
10.0
31
0.0
0
ENDSEC
0
SECTION
2
OBJECTS
0
ACDBDICTIONARYW
5
C0
330
0
100
AcDbObject
100
AcDbDictionary
281
1
3
ACAD_GROUP
350
C1
0
ACDBDICTIONARYW
5
C1
330
C0
100
AcDbObject
100
AcDbDictionary
281
1
0
LAYOUT
5
D1
330
C0
100
AcDbObject
100
AcDbLayout
70
1
1
Layout1
330
E
0
ENDSEC
0
EOF
`

func mustLoad(t *testing.T, src string) *drawing.Drawing {
	t.Helper()
	doc, err := dxf.FromStringData(src)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	return doc
}

func mutate(t *testing.T, src, old, new string) string {
	t.Helper()
	if !strings.Contains(src, old) {
		t.Fatalf("fixture marker not found: %q", old)
	}
	return strings.Replace(src, old, new, 1)
}

func codes(rep *Report) map[string]bool {
	out := map[string]bool{}
	for _, is := range rep.Issues {
		out[is.Code] = true
	}
	return out
}

func hasIssue(rep *Report, code, handle string) bool {
	for _, is := range rep.Issues {
		if is.Code == code && is.Handle == handle {
			return true
		}
	}
	return false
}

func unfixedErrors(rep *Report) []Issue {
	out := []Issue{}
	for _, is := range rep.Issues {
		if is.Severity >= SeverityError && !is.Fixed {
			out = append(out, is)
		}
	}
	return out
}

// ---- 1.1 / 1.3: report, registry, clean document ----

func TestRegistryCodes(t *testing.T) {
	codes := CheckCodes()
	if len(codes) < 15 {
		t.Fatalf("expected >=15 registered checks, got %d: %v", len(codes), codes)
	}
}

func TestCleanDocumentZeroErrors(t *testing.T) {
	doc := mustLoad(t, baseDXF)
	before := entityHandles(doc)
	rep := Run(doc, nil)
	for _, is := range rep.Issues {
		if is.Severity >= SeverityError {
			t.Errorf("clean document reported %s (%s): %s", is.Code, is.Handle, is.Message)
		}
	}
	if rep.HasErrors() {
		t.Errorf("clean document HasErrors()=true: %s", rep)
	}
	after := entityHandles(doc)
	if strings.Join(before, ",") != strings.Join(after, ",") {
		t.Errorf("audit modified the document: %v -> %v", before, after)
	}
}

func entityHandles(doc *drawing.Drawing) []string {
	out := []string{}
	for _, e := range doc.Entities() {
		out = append(out, e.Handle())
	}
	return out
}

func TestReportSerialization(t *testing.T) {
	rep := &Report{}
	rep.Addf(SeverityError, CodeDuplicateHandle, "41", "duplicate handle %q", "41")
	rep.Addf(SeverityWarning, CodeBadExtents, "", "bad extents")
	rep.Add(Issue{Severity: SeverityInfo, Code: "INFO_TEST", Message: "info"})

	js, err := rep.JSON()
	if err != nil {
		t.Fatal(err)
	}
	data := []byte(js)
	var parsed struct {
		Counts    map[string]int `json:"counts"`
		Truncated bool           `json:"truncated"`
		Issues    []Issue        `json:"issues"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Counts["error"] != 1 || parsed.Counts["warning"] != 1 || parsed.Counts["info"] != 1 {
		t.Errorf("counts: %v", parsed.Counts)
	}
	if len(parsed.Issues) != 3 {
		t.Errorf("issues: %d", len(parsed.Issues))
	}
	if got := rep.BySeverity()["error"]; got != 1 {
		t.Errorf("BySeverity error = %d", got)
	}
	if cs := rep.Codes(); cs[CodeDuplicateHandle] != 1 {
		t.Errorf("Codes: %v", cs)
	}
	if !strings.Contains(rep.String(), "3 issues") {
		t.Errorf("String(): %s", rep)
	}

	for i := 0; i < MaxIssues+200; i++ {
		rep.Addf(SeverityWarning, fmt.Sprintf("BULK_%d", i), "", "bulk %d", i)
	}
	if len(rep.Issues) != MaxIssues {
		t.Fatalf("issue cap: len=%d want %d", len(rep.Issues), MaxIssues)
	}
	if !rep.Truncated {
		t.Error("Truncated not set")
	}
}

// ---- 2.1: handles and ownership ----

// beAdapter adapts a plain entity to block.BlockEntity for fixture building.
type beAdapter struct {
	entity.Entity
	raw []block.TagPair
}

func (a beAdapter) RawTags() []block.TagPair      { return a.raw }
func (a *beAdapter) SetRawTags(t []block.TagPair) { a.raw = t }

func TestDuplicateHandleNamesBothObjects(t *testing.T) {
	doc := mustLoad(t, baseDXF)
	l2 := entity.NewLine()
	l2.SetHandleString("41")
	l2.SetLayerName("0")
	for _, blk := range doc.Blocks() {
		if blk.Name == "*Model_Space" {
			blk.AddEntity(&beAdapter{Entity: l2})
		}
	}
	rep := Run(doc, nil)
	if !hasIssue(rep, CodeDuplicateHandle, "41") {
		t.Fatalf("DUPLICATE_HANDLE not reported: %s", rep)
	}
	for _, is := range rep.Issues {
		if is.Code == CodeDuplicateHandle {
			if !strings.Contains(is.Message, "LINE") || !strings.Contains(is.Message, "41") {
				t.Errorf("issue must name both objects: %s", is.Message)
			}
		}
	}
	// Unsafe: fix mode must not touch it and must leave the document unchanged.
	rep2 := Run(doc, &Options{Fix: true})
	for _, is := range rep2.Issues {
		if is.Code == CodeDuplicateHandle && is.Fixed {
			t.Errorf("duplicate handle must not be auto-fixed")
		}
	}
	hs := []string{}
	for _, e := range NewContext(doc).AllEntities() {
		hs = append(hs, EntityHandle(e))
	}
	if len(hs) != 2 || hs[0] != "41" || hs[1] != "41" {
		t.Errorf("fix mode changed document for unsafe issue: %v", hs)
	}
}

func TestMissingOwnerReportedAndFixed(t *testing.T) {
	src := mutate(t, baseDXF, "5\n41\n330\nE\n100\nAcDbEntity", "5\n41\n100\nAcDbEntity")
	doc := mustLoad(t, src)
	rep := Run(doc, nil)
	if !hasIssue(rep, CodeInvalidOwner, "41") {
		t.Fatalf("INVALID_OWNER not reported: %s", rep)
	}
	rep2 := Run(doc, &Options{Fix: true})
	fixed := false
	for _, is := range rep2.Issues {
		if is.Code == CodeInvalidOwner && is.Handle == "41" && is.Fixed {
			fixed = true
		}
	}
	if !fixed {
		t.Fatalf("INVALID_OWNER not marked fixed: %s", rep2)
	}
	if u := unfixedErrors(rep2); len(u) > 0 {
		t.Errorf("re-audit not clean: %v", u)
	}
	var owner handle.Handler
	for _, e := range doc.Entities() {
		if oe, ok := e.(interface{ Owner() handle.Handler }); ok {
			if o := oe.Owner(); o != nil {
				owner = o
			}
		}
	}
	if owner == nil {
		t.Fatal("entity owner not assigned by fixer")
	}
	if owner.Handle() != "E" {
		t.Errorf("owner = %q, want model space record E", owner.Handle())
	}
}

// ---- 3.1: safe fixers: missing default entry + missing layer reference ----

func TestMissingLayerZeroReportedAndFixed(t *testing.T) {
	layerEntry := "0\nLAYER\n5\n11\n330\n10\n100\nAcDbSymbolTable\n100\nAcDbLayerTableRecord\n70\n0\n2\n0\n62\n7\n6\nContinuous\n370\n-3\n"
	src := mutate(t, baseDXF, layerEntry, "")
	doc := mustLoad(t, src)
	rep := Run(doc, nil)
	if !hasIssue(rep, CodeMissingTableEntry, "LAYER:0") {
		t.Fatalf("MISSING_TABLE_ENTRY LAYER:0 not reported: %s", rep)
	}
	rep2 := Run(doc, &Options{Fix: true})
	if u := unfixedErrors(rep2); len(u) > 0 {
		t.Errorf("fix re-audit left errors: %v", u)
	}
	names := NewContext(doc).EntryNames("LAYER")
	if !names["0"] {
		t.Errorf("layer 0 not added by fixer: %v", keysOf(names))
	}
}

func keysOf(m map[string]bool) []string {
	out := []string{}
	for k := range m {
		out = append(out, k)
	}
	return out
}

func TestMissingLayerRefFixed(t *testing.T) {
	src := mutate(t, baseDXF, "8\n0\n100\nAcDbLine", "8\nGHOST\n100\nAcDbLine")
	doc := mustLoad(t, src)
	rep := Run(doc, nil)
	if !hasIssue(rep, CodeMissingLayer, "41") {
		t.Fatalf("MISSING_LAYER not reported: %s", rep)
	}
	rep2 := Run(doc, &Options{Fix: true})
	if u := unfixedErrors(rep2); len(u) > 0 {
		t.Errorf("fix re-audit left errors: %v", u)
	}
	for _, e := range doc.Entities() {
		if ln, ok := e.(interface{ LayerName() string }); ok && ln.LayerName() == "GHOST" {
			t.Fatalf("entity %s still references GHOST after fix", e.Handle())
		}
	}
}

// ---- 2.2: references ----

func TestReferenceIntegrity(t *testing.T) {
	insertRec := "0\nINSERT\n5\n42\n330\nE\n100\nAcDbEntity\n8\n0\n100\nAcDbInsert\n2\nGHOSTB\n10\n0.0\n20\n0.0\n30\n0.0\n"
	textRec := "0\nTEXT\n5\n43\n330\nE\n100\nAcDbEntity\n8\n0\n100\nAcDbText\n10\n0.0\n20\n0.0\n30\n0.0\n40\n2.5\n1\nhello\n7\nGHOSTSTYLE\n"
	dimRec := "0\nDIMENSION\n5\n44\n330\nE\n100\nAcDbEntity\n8\n0\n100\nAcDbDimension\n3\nGHOSTDIM\n10\n0.0\n20\n0.0\n30\n0.0\n"
	endEnt := "0\nENDSEC\n0\nSECTION\n2\nOBJECTS"
	src := mutate(t, baseDXF, endEnt, insertRec+textRec+dimRec+endEnt)
	src = mutate(t, src, "8\n0\n100\nAcDbLine", "8\n0\n6\nGHOSTLT\n100\nAcDbLine")
	src = mutate(t, src, "1\nLayout1\n330\nE", "1\nLayout1\n330\nDEAD")
	rep := Run(mustLoad(t, src), nil)
	cs := codes(rep)
	if !cs[CodeMissingBlock] {
		t.Error("MISSING_BLOCK not reported")
	}
	if !cs[CodeMissingStyle] {
		t.Error("MISSING_STYLE not reported")
	}
	if !cs[CodeMissingLinetype] {
		t.Error("MISSING_LINETYPE not reported")
	}
	if !cs[CodeMissingDimStyle] {
		t.Error("MISSING_DIMSTYLE not reported")
	}
	if !cs[CodeMissingLayoutBlock] {
		t.Error("MISSING_LAYOUT_BLOCK not reported")
	}
}

func TestTableIntegrity(t *testing.T) {
	// Duplicate entry name via file mutation.
	dupEntry := "0\nLAYER\n5\n12\n330\n10\n100\nAcDbSymbolTable\n100\nAcDbLayerTableRecord\n70\n0\n2\n0\n62\n7\n6\nContinuous\n370\n-3\n"
	doc := mustLoad(t, mutate(t, baseDXF, "370\n-3\n0\nENDTAB", "370\n-3\n"+dupEntry+"0\nENDTAB"))
	rep := Run(doc, nil)
	if !hasIssue(rep, CodeDuplicateTableEntry, "12") {
		t.Errorf("DUPLICATE_TABLE_ENTRY not reported: %s", rep)
	}

	// Entry name cleared via API: the loader falls back to the record type
	// name for name-less records, so the empty state is reached by API.
	doc2 := mustLoad(t, baseDXF)
	ents2 := NewContext(doc2).TableByName("LAYER").Entries()
	setter2, ok := ents2[0].(interface{ SetName(string) })
	if !ok {
		t.Fatal("entry has no SetName")
	}
	setter2.SetName("")
	rep2 := Run(doc2, nil)
	if !hasIssue(rep2, CodeEmptyTableName, "11") {
		t.Errorf("EMPTY_TABLE_NAME not reported: %s", rep2)
	}

	// Owner mismatch: set the entry owner link to an unknown handle after
	// load (the loader structurally binds owners while parsing).
	doc3 := mustLoad(t, baseDXF)
	ents3 := NewContext(doc3).TableByName("LAYER").Entries()
	setter3, ok := ents3[0].(interface{ SetOwnerHandle(string) })
	if !ok {
		t.Fatal("entry has no SetOwnerHandle")
	}
	setter3.SetOwnerHandle("FFFF")
	rep3 := Run(doc3, nil)
	if !hasIssue(rep3, CodeOwnerMismatch, "11") {
		t.Errorf("OWNER_MISMATCH not reported: %s", rep3)
	}
}

// ---- 2.3: xdata, reactors, extension dictionaries, dictionaries ----

func TestXDataMissingAppID(t *testing.T) {
	src := mutate(t, baseDXF, "31\n0.0\n0\nENDSEC", "31\n0.0\n1001\nGHOSTAPP\n1000\nxtra\n0\nENDSEC")
	rep := Run(mustLoad(t, src), nil)
	if !hasIssue(rep, CodeXDataMissingAppID, "41") {
		t.Fatalf("XDATA_MISSING_APPID not reported: %s", rep)
	}
}

func TestReactorsAndExtensionDictionaries(t *testing.T) {
	// Object reactor (parsed at load by BaseObject).
	src := mutate(t, baseDXF, "1\nLayout1\n330\nE\n0\nENDSEC",
		"1\nLayout1\n330\nE\n102\n{ACAD_REACTORS\n330\nDEAD\n102\n}\n0\nENDSEC")
	rep := Run(mustLoad(t, src), nil)
	if !hasIssue(rep, CodeReactorUnresolved, "D1") {
		t.Errorf("REACTOR_UNRESOLVED for layout not reported: %s", rep)
	}

	// Table entry extension dictionary (loader discards entity-level 360, so
	// the entry link is the supported path).
	docE := mustLoad(t, baseDXF)
	entsE := NewContext(docE).TableByName("LAYER").Entries()
	setE, ok := entsE[0].(interface{ SetExtDictHandle(string) })
	if !ok {
		t.Fatal("entry has no SetExtDictHandle")
	}
	setE.SetExtDictHandle("DEAD")
	rep2 := Run(docE, nil)
	if !hasIssue(rep2, CodeExtDictUnresolved, "11") {
		t.Errorf("EXTDICT_UNRESOLVED not reported: %s", rep2)
	}

	// Entity reactor group: standalone owner 330 is stripped, the in-group
	// reactor link survives in raw tags.
	srcE := mutate(t, baseDXF, "8\n0\n100\nAcDbLine",
		"8\n0\n102\n{ACAD_REACTORS\n330\nDEAD\n102\n}\n100\nAcDbLine")
	repE := Run(mustLoad(t, srcE), nil)
	if !hasIssue(repE, CodeReactorUnresolved, "41") {
		t.Errorf("REACTOR_UNRESOLVED for entity not reported: %s", repE)
	}

	// Table entry reactor via API (loader does not parse entry reactors).
	doc := mustLoad(t, baseDXF)
	ctx := NewContext(doc)
	tb := ctx.TableByName("LAYER")
	if tb == nil {
		t.Fatal("no LAYER table")
	}
	entries := tb.Entries()
	if len(entries) == 0 {
		t.Fatal("no layer entries")
	}
	setter, ok := entries[0].(interface{ SetReactorHandles([]string) })
	if !ok {
		t.Fatal("entry has no SetReactorHandles")
	}
	setter.SetReactorHandles([]string{"DEAD"})
	rep3 := Run(doc, nil)
	if !hasIssue(rep3, CodeReactorUnresolved, "11") {
		t.Errorf("REACTOR_UNRESOLVED for layer entry not reported: %s", rep3)
	}
}

func TestDictionaryEntryUnresolved(t *testing.T) {
	src := mutate(t, baseDXF, "3\nACAD_GROUP\n350\nC1", "3\nACAD_GROUP\n350\nDEAD")
	rep := Run(mustLoad(t, src), nil)
	if !hasIssue(rep, CodeDictionaryEntryUnresolved, "C0") {
		t.Fatalf("DICTIONARY_ENTRY_UNRESOLVED not reported: %s", rep)
	}
}

// ---- 2.4: extents and entity invariant hooks ----

func TestExtentsAndSplineHook(t *testing.T) {
	src := mutate(t, baseDXF, "1\nAC1015\n0\nENDSEC",
		"1\nAC1015\n9\n$EXTMIN\n10\n100.0\n20\n100.0\n30\n0.0\n9\n$EXTMAX\n10\n-100.0\n20\n-100.0\n30\n0.0\n0\nENDSEC")
	rep := Run(mustLoad(t, src), nil)
	if !hasIssue(rep, CodeBadExtents, "$EXTMIN/$EXTMAX") && !hasIssue(rep, CodeBadExtents, "") {
		t.Errorf("BAD_EXTENTS not reported: %s", rep)
	}

	splineRec := "0\nSPLINE\n5\n45\n330\nE\n100\nAcDbEntity\n8\n0\n100\nAcDbSpline\n71\n1\n72\n2\n73\n3\n74\n0\n10\n0.0\n20\n0.0\n30\n0.0\n10\n1.0\n20\n1.0\n30\n0.0\n10\n2.0\n20\n0.0\n30\n0.0\n40\n0.0\n40\n1.0\n"
	src2 := mutate(t, baseDXF, "0\nENDSEC\n0\nSECTION\n2\nOBJECTS", splineRec+"0\nENDSEC\n0\nSECTION\n2\nOBJECTS")
	rep2 := Run(mustLoad(t, src2), nil)
	if !hasIssue(rep2, CodeEntityInvariants, "45") {
		t.Fatalf("ENTITY_INVARIANTS aggregate not reported: %s", rep2)
	}
	found := false
	for _, is := range rep2.Issues {
		if is.Code == CodeEntityInvariants && strings.Contains(is.Message, CodeSplineInvalid) {
			found = true
		}
	}
	if !found {
		t.Errorf("aggregate must name SPLINE_INVALID: %s", rep2)
	}
}

// ---- 3.3: strict R12 ----

const r12DXF = `0
SECTION
2
HEADER
9
$ACADVER
1
AC1009
0
ENDSEC
0
SECTION
2
TABLES
0
TABLE
2
LAYER
5
10
100
AcDbSymbolTable
70
1
0
LAYER
5
11
330
10
100
AcDbLayerTableRecord
70
0
2
0
62
7
6
Continuous
0
ENDTAB
0
ENDSEC
0
SECTION
2
BLOCKS
0
BLOCK
5
F
100
AcDbBlockBegin
70
0
8
0
2
*Model_Space
10
0.0
20
0.0
30
0.0
1
*Model_Space
3
*Model_Space
0
ENDBLK
5
6F
330
F
100
AcDbBlockEnd
0
ENDSEC
0
SECTION
2
ENTITIES
0
ENDSEC
0
SECTION
2
OBJECTS
0
ACDBDICTIONARYW
5
C0
100
AcDbDictionary
281
1
0
ENDSEC
0
EOF
`

func TestStrictR12RejectsObjects(t *testing.T) {
	doc := mustLoad(t, r12DXF)
	rep := Run(doc, &Options{StrictR12: true})
	if !hasIssue(rep, CodeR12Incompatible, "") {
		t.Fatalf("R12_INCOMPATIBLE not reported: %s", rep)
	}
	rep2 := Run(doc, nil)
	for _, is := range rep2.Issues {
		if is.Code == CodeR12Incompatible {
			t.Errorf("R12_INCOMPATIBLE must only fire in strict mode: %s", is)
		}
	}
}

// ---- 3.2: recovery ----

func TestRecoverTruncatedFile(t *testing.T) {
	idx := strings.Index(baseDXF, "11\n10.0")
	if idx < 0 {
		t.Fatal("marker")
	}
	truncated := baseDXF[:idx]
	doc, rep, err := RecoverReader(strings.NewReader(truncated), nil)
	if err != nil {
		t.Fatalf("recover truncated: %v", err)
	}
	if doc == nil {
		t.Fatal("document not retained")
	}
	if len(entityHandles(doc)) == 0 {
		t.Errorf("entities before the truncation not retained: %s", rep)
	}
	cs := codes(rep)
	if !cs[CodeTruncatedSection] && !cs[CodeLoadIssue] {
		t.Errorf("truncated section not reported: %s", rep)
	}
}

func TestRecoverMalformedEntity(t *testing.T) {
	bad := "0\n\n5\n42\n100\nAcDbEntity\n8\n0\n"
	src := mutate(t, baseDXF, "0\nENDSEC\n0\nSECTION\n2\nOBJECTS",
		bad+"0\nENDSEC\n0\nSECTION\n2\nOBJECTS")
	doc, rep, err := RecoverReader(strings.NewReader(src), nil)
	if err != nil {
		t.Fatalf("recover malformed: %v", err)
	}
	if doc == nil {
		t.Fatal("document nil")
	}
	if !hasIssue(rep, CodeSkippedEntity, "") && !strings.Contains(rep.String(), "skipped") {
		t.Errorf("skipped entity not reported: %s", rep)
	}
	found := false
	for _, h := range entityHandles(doc) {
		if h == "41" {
			found = true
		}
	}
	if !found {
		t.Errorf("good entity dropped: %v", entityHandles(doc))
	}
	for _, h := range entityHandles(doc) {
		if h == "42" {
			t.Errorf("malformed entity was not skipped")
		}
	}
}

func TestRecoverGarbageNoPanic(t *testing.T) {
	for _, data := range []string{"", "not a dxf at all", "\x00\x01\x02", "0\nSECTION\n2\nENTITIES\n0\nLINE\n"} {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("panic on %q: %v", data, r)
				}
			}()
			_, rep, _ := RecoverReader(strings.NewReader(data), &Options{Fix: true})
			if rep != nil && len(rep.Issues) > MaxIssues+16 {
				t.Fatalf("unbounded report on %q: %d", data, len(rep.Issues))
			}
		}()
	}
}

// ---- 4.1: corpus audit + fuzz ----

func TestCorpusAuditNoPanics(t *testing.T) {
	for _, path := range corpusFiles(t) {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("audit panic on %s: %v", path, r)
				}
			}()
			doc, err := dxf.FromFile(path)
			if err != nil {
				t.Logf("%s: load %v", path, err)
				return
			}
			rep := Run(doc, nil)
			if c := rep.Count(SeverityCritical); c > 0 {
				t.Errorf("%s: %d critical issues: %s", path, c, rep)
			}
			if len(rep.Issues) > MaxIssues {
				t.Errorf("%s: unbounded report %d", path, len(rep.Issues))
			}
			t.Logf("%s: %s", path, rep)
		}()
	}
}

func corpusFiles(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob("../testdata/*.dxf")
	if err != nil || len(files) == 0 {
		t.Fatalf("corpus glob: %v (%d files)", err, len(files))
	}
	return files
}

func FuzzAudit(f *testing.F) {
	f.Add(baseDXF)
	f.Add(r12DXF)
	f.Add(baseDXF[:len(baseDXF)/2])
	f.Add("0\nSECTION\n2\nENTITIES\n0\nLINE\n")
	f.Add("0\nEOF\n")
	f.Fuzz(func(t *testing.T, data string) {
		_, rep, _ := RecoverReader(strings.NewReader(data), &Options{Fix: true})
		if rep != nil && len(rep.Issues) > 4*MaxIssues {
			t.Fatalf("unbounded report: %d", len(rep.Issues))
		}
	})
}
