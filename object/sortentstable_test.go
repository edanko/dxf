// SORTENTSTABLE construction round-trip.
package object

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
)

func TestSortEntsTableRoundTrip(t *testing.T) {
	frontToBack := []string{"A1", "B2", "C3"}
	st := NewSortEntsTable("FA", "E", frontToBack)
	if st.DXFType() != "SORTENTSTABLE" {
		t.Fatalf("type: %s", st.DXFType())
	}
	if st.BlockRecordHandle() != "E" {
		t.Fatalf("block record: %s", st.BlockRecordHandle())
	}
	want := [][2]string{{"A1", "-1"}, {"B2", "C3"}, {"C3", "-2"}}
	if got := st.SortEntries(); got[1][1] != "C3" || got[0][1] != "-1" || got[2][1] != "-2" {
		t.Fatalf("sort mapping: %v", got)
	}

	f := format.NewASCII()
	st.Format(f)
	out := f.Output()
	if !strings.Contains(out, "AcDbSortentsTable") {
		t.Fatalf("marker missing:\n%s", out)
	}

	tags, err := tagsFromASCII(out)
	if err != nil {
		t.Fatal(err)
	}
	o := ParseObject(tags)
	rt, ok := o.(*SortEntsTable)
	if !ok {
		t.Fatalf("reload type %T", o)
	}
	if rt.Handle() != "FA" {
		t.Fatalf("handle: %s", rt.Handle())
	}
	if rt.BlockRecordHandle() != "E" {
		t.Fatalf("reloaded block record: %s", rt.BlockRecordHandle())
	}
	entries := rt.SortEntries()
	if len(entries) != 3 {
		t.Fatalf("reloaded entries: %v", entries)
	}
	for i, w := range want {
		if entries[i] != w {
			t.Fatalf("entry %d: %v want %v", i, entries[i], w)
		}
	}

	// Second save must be byte-identical (retained tags round-trip).
	f2 := format.NewASCII()
	rt.Format(f2)
	if f2.Output() != out {
		t.Fatalf("second save differs:\n%s\n---\n%s", out, f2.Output())
	}
}

// tagsFromASCII parses an ASCII record stream into lldxf tags.
func tagsFromASCII(out string) (lldxf.Tags, error) {
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	tags := lldxf.NewTags()
	for i := 0; i+1 < len(lines); i += 2 {
		code, err := strconv.Atoi(strings.TrimSpace(lines[i]))
		if err != nil {
			return nil, fmt.Errorf("tag line %d: %w", i, err)
		}
		tags = tags.Add(lldxf.NewDXFTag(lldxf.GroupCode(code), strings.TrimSpace(lines[i+1])))
	}
	return tags, nil
}
