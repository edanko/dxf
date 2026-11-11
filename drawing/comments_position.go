package drawing

import (
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
)

// writeSectionWithComments emits one section, replaying comments at their
// original record boundaries. Sections without positioned comments keep the
// legacy behavior of writing every comment before the section header.
func (d *Drawing) writeSectionWithComments(f format.Formatter, section string, emit func(format.Formatter)) {
	var comments []lldxf.PositionedComment
	for _, comment := range d.positionedComments {
		if comment.Section == section {
			comments = append(comments, comment)
		}
	}
	if len(comments) == 0 {
		d.writeCommentsTo(f, section)
		emit(f)
		return
	}
	emit(&commentInjector{Formatter: f, comments: comments})
}

// commentInjector writes code 999 comments immediately before the section
// record that followed them in the source.
type commentInjector struct {
	format.Formatter
	comments []lldxf.PositionedComment
	next     int
	records  int
	started  bool
}

func (c *commentInjector) WriteString(code int, value string) {
	if code == 0 {
		switch value {
		case "SECTION":
			c.started = true
		case "ENDSEC":
			c.flushAll()
		default:
			if c.started {
				c.flush(c.records)
				c.records++
			}
		}
	}
	c.Formatter.WriteString(code, value)
}

// flush writes comments anchored at or before the given record index.
func (c *commentInjector) flush(index int) {
	for c.next < len(c.comments) && c.comments[c.next].Index <= index {
		c.Formatter.WriteString(999, c.comments[c.next].Text)
		c.next++
	}
}

func (c *commentInjector) flushAll() {
	c.flush(len(c.comments) + c.records + 1)
}
