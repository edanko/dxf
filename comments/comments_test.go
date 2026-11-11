package comments

import (
	"strings"
	"testing"

	"github.com/edanko/dxf/drawing"
)

func TestStripComments(t *testing.T) {
	d, err := drawing.New()
	if err != nil {
		t.Fatal(err)
	}
	d.SetComments(map[string][]string{"s": {"c"}})
	StripComments(d)
	if len(d.Comments()) != 0 {
		t.Fatalf("comments remain: %v", d.Comments())
	}
}

func TestStripCommentsFromBytes(t *testing.T) {
	in := []byte("999\nhello\n0\nSECTION\n2\nHEADER\n")
	out := StripCommentsFromBytes(in)
	if strings.Contains(string(out), "hello") {
		t.Fatalf("comment not stripped: %q", out)
	}
	if !strings.Contains(string(out), "SECTION") {
		t.Fatalf("data lost: %q", out)
	}
}
