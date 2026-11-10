package dxf

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestParityProjectionDetectsFieldDifference verifies the cross-check payload
// includes typed fields, so two drawings with identical inventories but
// different LINE coordinates are not considered equal.
func TestParityProjectionDetectsFieldDifference(t *testing.T) {
	dir := t.TempDir()
	makeFile := func(name string, x float64) string {
		path := filepath.Join(dir, name)
		src := "0\nSECTION\n2\nENTITIES\n0\nLINE\n5\nA1\n100\nAcDbEntity\n8\n0\n100\nAcDbLine\n" +
			"10\n" + ftoa(x) + "\n20\n0\n30\n0\n11\n10\n21\n0\n31\n0\n0\nENDSEC\n0\nEOF\n"
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	left := goParityDump(t, makeFile("left.dxf", 1))
	right := goParityDump(t, makeFile("right.dxf", 2))
	if reflect.DeepEqual(left, right) {
		t.Fatal("typed field difference was not detected")
	}
}

func ftoa(v float64) string {
	if v == 1 {
		return "1"
	}
	return "2"
}
