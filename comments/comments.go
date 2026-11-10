// Package comments provides comment stripping helpers for DXF documents.
package comments

import (
	"strings"

	"github.com/edanko/dxf/drawing"
)

// StripComments removes all comments stored on the document.
func StripComments(d *drawing.Drawing) {
	d.SetComments(map[string][]string{})
}

// StripCommentsFromBytes removes 999 comment tag pairs from an ASCII DXF
// byte stream, including the writer's own header comment.
func StripCommentsFromBytes(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	out := make([]string, 0, len(lines))
	for i := 0; i+1 < len(lines); i += 2 {
		if strings.TrimSpace(lines[i]) == "999" {
			continue
		}
		out = append(out, lines[i], lines[i+1])
	}
	if len(lines)%2 == 1 {
		out = append(out, lines[len(lines)-1])
	}
	return []byte(strings.Join(out, "\n"))
}
