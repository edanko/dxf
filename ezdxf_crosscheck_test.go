package dxf

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/object"
)

// parityDump is the canonical cross-check payload. Keep the field order and
// names in sync with scripts/ezdxf_dump.py.
type parityDump struct {
	Version  string                        `json:"version"`
	Entities []parityEntity                `json:"entities"`
	Blocks   []parityBlock                 `json:"blocks"`
	Tables   map[string][]parityTableEntry `json:"tables"`
	Objects  []parityObject                `json:"objects"`
}

type parityEntity struct {
	Type   string `json:"type"`
	Handle string `json:"handle"`
	Layer  string `json:"layer"`
}

type parityBlock struct {
	Name     string         `json:"name"`
	Entities []parityEntity `json:"entities"`
}

type parityTableEntry struct {
	Name   string `json:"name"`
	Handle string `json:"handle"`
}

type parityObject struct {
	Type   string `json:"type"`
	Handle string `json:"handle"`
}

// TestEzdxfCrossCheck compares the Go loader against an ezdxf dump for every
// testdata DXF file. It is opt-in because it creates a throwaway Python
// virtual environment and may need to install ezdxf:
//
//	DXF_EZDXF_CROSSCHECK=1 go test -run TestEzdxfCrossCheck .
//
// Set DXF_EZDXF_SOURCE to a local ezdxf checkout (for example,
// /home/egor/src/ezdxf) to match the reference version exactly. Set
// DXF_EZDXF_PYTHON to reuse an existing interpreter instead of creating a venv.
func TestEzdxfCrossCheck(t *testing.T) {
	if os.Getenv("DXF_EZDXF_CROSSCHECK") != "1" {
		t.Skip("set DXF_EZDXF_CROSSCHECK=1 to run the ezdxf cross-check")
	}

	files, err := filepath.Glob("testdata/*.dxf")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no testdata fixtures")
	}
	sort.Strings(files)

	python := ezdxfPython(t)
	args := append([]string{"scripts/ezdxf_dump.py"}, files...)
	cmd := exec.Command(python, args...)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("ezdxf dump: %v", err)
	}

	var expected map[string]parityDump
	if err := json.Unmarshal(out, &expected); err != nil {
		t.Fatalf("decode ezdxf dump: %v", err)
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			want, ok := expected[file]
			if !ok {
				t.Fatalf("ezdxf dump missing %s", file)
			}
			got := goParityDump(t, file)
			if want.Tables == nil {
				// No source TABLES section: the loader synthesizes defaults.
				want.Tables = got.Tables
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("semantic mismatch for %s:\n%s", file, parityMismatch(want, got))
			}
		})
	}
}

func goParityDump(t *testing.T, path string) parityDump {
	t.Helper()
	d, err := FromFile(path)
	if err != nil {
		t.Fatalf("load %s: %v", path, err)
	}

	version := d.Header().Version
	dump := parityDump{
		Version:  version,
		Entities: make([]parityEntity, 0, len(d.Entities())),
		Blocks:   make([]parityBlock, 0, len(d.Blocks())),
		Tables:   make(map[string][]parityTableEntry, len(d.Tables())),
		Objects:  make([]parityObject, 0, len(d.Objects())),
	}

	// ezdxf exposes graphical entities from the entity database, which includes
	// block content. Mirror that inventory by unioning ENTITIES and block
	// definitions, deduplicating by handle where handles exist.
	seenHandles := make(map[string]bool)
	addEntity := func(e parityEntity) {
		if e.Type == "SEQEND" {
			return
		}
		if e.Handle != "" {
			if seenHandles[e.Handle] {
				return
			}
			seenHandles[e.Handle] = true
		}
		dump.Entities = append(dump.Entities, e)
	}

	for _, e := range d.Entities() {
		addEntityTree(e, addEntity, version)
	}

	for _, b := range d.Blocks() {
		columnRefs := make(map[string]bool)
		for _, tag := range b.RawTags() {
			if tag.Code == 1005 {
				columnRefs[tag.Value] = true
			}
		}
		block := parityBlock{
			Name:     b.Name,
			Entities: make([]parityEntity, 0, len(b.Entities())),
		}
		for _, e := range b.Entities() {
			pe := newParityEntity(e.DXFType(), e.Handle(), entityLayerName(e), version)
			// ezdxf detaches MTEXT column fragments from the block that
			// physically contains them; the union inventory still includes them.
			if !isSubentityType(pe.Type) && !(pe.Type == "MTEXT" && columnRefs[pe.Handle]) {
				block.Entities = append(block.Entities, pe)
			}
			addBlockEntityTree(e, addEntity, version)
		}
		sort.Slice(block.Entities, func(i, j int) bool {
			return lessEntity(block.Entities[i], block.Entities[j])
		})
		if strings.HasPrefix(b.Name, "*") {
			continue
		}
		dump.Blocks = append(dump.Blocks, block)
	}
	sort.Slice(dump.Entities, func(i, j int) bool {
		return lessEntity(dump.Entities[i], dump.Entities[j])
	})
	sort.Slice(dump.Blocks, func(i, j int) bool {
		return dump.Blocks[i].Name < dump.Blocks[j].Name
	})

	for _, tbl := range d.Tables() {
		if tbl == nil {
			continue
		}
		name := tbl.Name()
		entries := make([]parityTableEntry, 0, len(tbl.Entries()))
		for _, e := range tbl.Entries() {
			entries = append(entries, parityTableEntry{
				Name:   e.Name(),
				Handle: normalizeHandle(version, e.Handle()),
			})
		}
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].Name != entries[j].Name {
				return entries[i].Name < entries[j].Name
			}
			return entries[i].Handle < entries[j].Handle
		})
		dump.Tables[name] = entries
	}

	for _, o := range d.Objects() {
		dump.Objects = append(dump.Objects, parityObject{
			Type:   object.TypeName(o),
			Handle: normalizeHandle(version, o.Handle()),
		})
	}
	sort.Slice(dump.Objects, func(i, j int) bool {
		if dump.Objects[i].Type != dump.Objects[j].Type {
			return dump.Objects[i].Type < dump.Objects[j].Type
		}
		return dump.Objects[i].Handle < dump.Objects[j].Handle
	})

	return dump
}

func newParityEntity(dxfType, handle, layer, version string) parityEntity {
	return parityEntity{
		Type:   dxfType,
		Handle: normalizeHandle(version, handle),
		Layer:  layer,
	}
}

// addEntityTree adds an entity and its owned subentities (POLYLINE vertices,
// INSERT attributes, and similar composite records) to the inventory.
func addEntityTree(e entity.Entity, add func(parityEntity), version string) {
	add(newParityEntity(e.DXFType(), e.Handle(), entityLayerName(e), version))
	owner, ok := e.(entity.SubentityOwner)
	if !ok {
		return
	}
	for _, child := range owner.Subentities() {
		addEntityTree(child, add, version)
	}
}

// addBlockEntityTree mirrors addEntityTree for block entities, unwrapping the
// block adapter when an external implementation did not expose the entity.
func addBlockEntityTree(e interface {
	DXFType() string
	Handle() string
}, add func(parityEntity), version string) {
	add(newParityEntity(e.DXFType(), e.Handle(), entityLayerName(e), version))

	var underlying entity.Entity
	switch value := e.(type) {
	case entity.Entity:
		underlying = value
	case interface{ Unwrap() entity.Entity }:
		underlying = value.Unwrap()
	}
	if underlying == nil {
		return
	}
	owner, ok := underlying.(entity.SubentityOwner)
	if !ok {
		return
	}
	for _, child := range owner.Subentities() {
		addEntityTree(child, add, version)
	}
}

func normalizeHandle(version, handle string) string {
	if version == "AC1006" || version == "AC1009" {
		return ""
	}
	return handle
}

func entityLayerName(v any) string {
	if namer, ok := v.(entity.LayerNamer); ok {
		if name := namer.LayerName(); name != "" {
			return name
		}
	}
	return "0"
}

// isSubentityType reports records that belong to a preceding composite entity.
// Block inventories expose only the top-level composite, matching ezdxf's
// block iteration.
func isSubentityType(dxfType string) bool {
	switch dxfType {
	case "ATTRIB", "VERTEX", "SEQEND":
		return true
	default:
		return false
	}
}

func lessEntity(a, b parityEntity) bool {
	if a.Type != b.Type {
		return a.Type < b.Type
	}
	if a.Handle != b.Handle {
		return a.Handle < b.Handle
	}
	return a.Layer < b.Layer
}

// parityMismatch reports the first differing JSON line with a little context so
// a failure does not dump both full inventories.
func parityMismatch(want, got parityDump) string {
	wantJSON, _ := json.MarshalIndent(want, "", "  ")
	gotJSON, _ := json.MarshalIndent(got, "", "  ")
	wantLines := strings.Split(string(wantJSON), "\n")
	gotLines := strings.Split(string(gotJSON), "\n")

	limit := len(wantLines)
	if len(gotLines) < limit {
		limit = len(gotLines)
	}
	for i := 0; i < limit; i++ {
		if wantLines[i] == gotLines[i] {
			continue
		}
		start := i - 4
		if start < 0 {
			start = 0
		}
		end := i + 6
		if end > limit {
			end = limit
		}
		return fmt.Sprintf("first difference at line %d\nezdxf:\n%s\nGo:\n%s",
			i+1,
			strings.Join(wantLines[start:end], "\n"),
			strings.Join(gotLines[start:end], "\n"))
	}
	return fmt.Sprintf("inventories differ in length: ezdxf %d lines, Go %d lines", len(wantLines), len(gotLines))
}

// ezdxfPython returns an interpreter with ezdxf installed. It reuses
// DXF_EZDXF_PYTHON when set, otherwise creates a temporary virtual environment
// with uv (preferred) or python3 -m venv.
func ezdxfPython(t *testing.T) string {
	t.Helper()
	if python := os.Getenv("DXF_EZDXF_PYTHON"); python != "" {
		return python
	}

	source := os.Getenv("DXF_EZDXF_SOURCE")
	if source == "" {
		source = "ezdxf"
	}

	venv := filepath.Join(t.TempDir(), "ezdxf-venv")
	if uv, err := exec.LookPath("uv"); err == nil {
		runCommand(t, uv, "venv", venv)
		python := venvPython(venv)
		runCommand(t, uv, "pip", "install", "--python", python, source)
		return python
	}

	python3, err := exec.LookPath("python3")
	if err != nil {
		t.Fatalf("neither uv nor python3 found; set DXF_EZDXF_PYTHON")
	}
	runCommand(t, python3, "-m", "venv", venv)
	python := venvPython(venv)
	runCommand(t, python, "-m", "pip", "install", source)
	return python
}

func venvPython(venv string) string {
	if filepath.Separator == '\\' {
		return filepath.Join(venv, "Scripts", "python.exe")
	}
	return filepath.Join(venv, "bin", "python")
}

func runCommand(t *testing.T, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s %s: %v", name, strings.Join(args, " "), err)
	}
}
