// Parity tests for change 13: boundary validation of color values.
package dxf_test

import (
	"strings"
	"testing"

	dxf "github.com/edanko/dxf"
	"github.com/edanko/dxf/color"
	"github.com/edanko/dxf/entity"
)

const invalidACIDXF = `0
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
ENTITIES
0
LINE
5
41
62
999
10
0.0
20
0.0
30
0.0
11
1.0
21
1.0
31
0.0
0
ENDSEC
0
EOF
`

func TestInvalidACIReportedAndPreserved(t *testing.T) {
	doc, err := dxf.FromStringData(invalidACIDXF)
	if err != nil {
		t.Fatal(err)
	}
	rep := doc.RecoveryReport()
	if rep == nil {
		t.Fatal("no recovery report")
	}
	found := false
	for _, is := range rep.Issues {
		if strings.Contains(is, "invalid ACI") {
			found = true
		}
	}
	if !found {
		t.Fatalf("issues: %v", rep.Issues)
	}
	var line *entity.Line
	for _, e := range doc.Entities() {
		if l, ok := e.(*entity.Line); ok {
			line = l
			break
		}
	}
	if line == nil {
		t.Fatal("LINE entity missing")
	}
	c, ok := interface{}(line).(interface{ Color() color.ColorNumber })
	if !ok {
		t.Fatal("entity has no Color getter")
	}
	if int(c.Color()) != 999 {
		t.Fatalf("raw color not preserved: %d", c.Color())
	}
}
