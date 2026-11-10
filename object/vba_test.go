package object

import (
	"bytes"
	"strings"
	"testing"
)

func TestVBAProjectTyped(t *testing.T) {
	payload := []byte("vba-project-payload")
	dxf := "  0\nVBA_PROJECT\n  5\nVB\n330\nOWNER\n100\nAcDbVbaProject\n 90\n" +
		"19\n310\n" + strings.ToUpper(hexString(payload)) + "\n999\nunknown\n"
	o := parseOne(t, dxf)
	project, ok := o.(*VBAProject)
	if !ok {
		t.Fatalf("type = %T", o)
	}
	if got := TypeName(project); got != "VBA_PROJECT" {
		t.Fatalf("TypeName = %q", got)
	}
	if project.DeclaredSize() != 19 {
		t.Fatalf("declared size = %d", project.DeclaredSize())
	}
	if !bytes.Equal(project.Data(), payload) {
		t.Fatalf("data = %q, want %q", project.Data(), payload)
	}
	out := formatObject(t, project)
	if !strings.Contains(out, "999\nunknown") {
		t.Fatalf("unknown tag lost:\n%s", out)
	}
	reparsed, ok := reparse(t, project).(*VBAProject)
	if !ok {
		t.Fatalf("reparsed type = %T", reparsed)
	}
	if !bytes.Equal(reparsed.Data(), payload) {
		t.Fatalf("reparsed data = %q", reparsed.Data())
	}
}

func hexString(data []byte) string {
	const digits = "0123456789abcdef"
	out := make([]byte, 0, len(data)*2)
	for _, b := range data {
		out = append(out, digits[b>>4], digits[b&0x0f])
	}
	return string(out)
}
