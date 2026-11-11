package dxf

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
)

// TestRegistryCompleteness checks that every registered loader produces a
// typed entity for its type name: no standard DXF record silently falls back
// to whole-record retention while a loader is registered for it.
func TestRegistryCompleteness(t *testing.T) {
	types := entity.RegisteredTypes()
	if len(types) < 30 {
		t.Fatalf("registered types = %d", len(types))
	}
	fixtureFor := func(name string) string {
		return "0\n" + name + "\n5\nRC1\n100\nAcDbEntity\n8\n0\n"
	}
	for _, name := range types {
		src := "0\nSECTION\n2\nENTITIES\n" + fixtureFor(name) + "0\nENDSEC\n0\nEOF\n"
		d, err := FromStringData(src)
		if err != nil {
			t.Errorf("%s: parse: %v", name, err)
			continue
		}
		es := d.Entities()
		if len(es) != 1 {
			t.Errorf("%s: loaded %d entities", name, len(es))
			continue
		}
		if _, retained := es[0].(*entity.RetainedRecordType); retained {
			t.Errorf("%s: registered loader fell back to retention", name)
		}
		if dt := strings.ToUpper(es[0].DXFType()); dt != name {
			// Loader aliases (MLEADER -> MULTILEADER) keep their canonical
			// wire spelling on re-save; that is intentional.
			t.Logf("%s: record type %q", name, dt)
		}
	}
}
