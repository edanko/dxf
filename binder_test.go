package dxf

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/format"
)

// asciiCodeLines renders an entity and returns the group codes of its ASCII
// output. Formatter output is strictly "code\nvalue\n" pairs, so the codes are
// the lines at even positions.
func asciiCodeLines(t *testing.T, e entity.Entity) []int {
	t.Helper()
	f := format.NewASCII()
	e.Format(f)
	lines := strings.Split(f.Output(), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines)%2 != 0 {
		t.Fatalf("entity %s (%T) rendered %d lines, want an even count", e.DXFType(), e, len(lines))
	}
	codes := make([]int, 0, len(lines)/2)
	for i := 0; i < len(lines); i += 2 {
		code, err := strconv.Atoi(strings.TrimSpace(lines[i]))
		if err != nil {
			t.Fatalf("entity %s (%T) line %d is not a group code: %q", e.DXFType(), e, i, lines[i])
		}
		codes = append(codes, code)
	}
	return codes
}

// legacyEmittedCodes reproduces the set attachRawTags used to derive by
// scanning every line of the rendered entity with fmt.Sscanf("%d", ...). That
// scan also matches value lines that begin with digits, such as "3DFACE" or
// "1", so the result is a superset of the codes actually emitted.
func legacyEmittedCodes(e entity.Entity) map[int]bool {
	f := format.NewASCII()
	e.Format(f)
	emitted := make(map[int]bool)
	for line := range strings.SplitSeq(f.Output(), "\n") {
		var code int
		if _, err := fmt.Sscanf(strings.TrimSpace(line), "%d", &code); err == nil {
			emitted[code] = true
		}
	}
	return emitted
}

// TestCodeSetFormatterMatchesLegacy ensures the formatter that replaced the
// ASCII round-trip records exactly the group codes the entity writes, and that
// the old line scan never missed one of them.
func TestCodeSetFormatterMatchesLegacy(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "*.dxf"))
	if err != nil {
		t.Fatal(err)
	}
	checkedEntities := 0
	spurious := 0
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		d, err := FromReaderWithOptions(strings.NewReader(string(data)), &LoadOptions{Recover: true})
		if err != nil {
			continue // corpus contains deliberately malformed fixtures
		}
		for _, e := range d.Entities() {
			if e == nil {
				continue
			}
			checkedEntities++

			want := make(map[int]bool)
			for _, code := range asciiCodeLines(t, e) {
				want[code] = true
			}

			got := newCodeSetFormatter()
			e.Format(got)
			for code := -8; code < maxTrackedGroupCode+8; code++ {
				if got.has(code) != want[code] {
					t.Fatalf("%s: entity %s (%T) code %d: collector=%v, emitted=%v",
						filepath.Base(path), e.DXFType(), e, code, got.has(code), want[code])
				}
			}

			// The old parse must remain a superset: anything the formatter
			// records has to have been visible to it, otherwise raw-tag
			// retention would have changed in the lossy direction.
			legacy := legacyEmittedCodes(e)
			for code := range want {
				if !legacy[code] {
					t.Fatalf("%s: entity %s (%T) code %d was emitted but invisible to the legacy line scan",
						filepath.Base(path), e.DXFType(), e, code)
				}
			}
			spurious += len(legacy) - len(want)
		}
	}
	if checkedEntities == 0 {
		t.Fatal("no entities checked; the equivalence test did not exercise the corpus")
	}
	t.Logf("verified code sets for %d entities (%d spurious legacy marks from numeric value lines)",
		checkedEntities, spurious)
}
