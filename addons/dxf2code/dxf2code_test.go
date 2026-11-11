package dxf2code_test

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/edanko/dxf/addons/dxf2code"
	"github.com/edanko/dxf/drawing"
)

func TestGenerateParsesAndInventories(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddLine(0, 0, 0, 10, 10, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddLine(1, 1, 0, 2, 2, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddCircle(5, 5, 0, 3); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Model().AddText(0, 0, 0, 2.5, "hi"); err != nil {
		t.Fatal(err)
	}

	src, err := dxf2code.Generate(d)
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "generated.go", src, 0); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, src)
	}
	if n := strings.Count(src, "AddLine"); n != 2 {
		t.Fatalf("AddLine count: %d", n)
	}
	if n := strings.Count(src, "AddCircle"); n != 1 {
		t.Fatalf("AddCircle count: %d", n)
	}
	if n := strings.Count(src, "AddText"); n != 1 {
		t.Fatalf("AddText count: %d", n)
	}
}
