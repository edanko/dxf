// Package r12writer streams an R12 DXF file incrementally: the header and
// tables are written up front, entities are appended one at a time without
// holding the whole document in memory.
package r12writer

import (
	"io"
	"strings"

	"github.com/edanko/dxf/addons"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/lldxf"
)

type capability struct{}

func (capability) Name() string        { return "r12writer" }
func (capability) Description() string { return "incremental R12 DXF writer" }

func init() { _ = addons.Register("r12writer", capability{}) }

// R12Writer writes an R12 DXF stream.
type R12Writer struct {
	w   *lldxf.ASCIITagWriter
	fmt *format.ASCII
}

// New creates a writer for an R12 DXF stream.
func New(w io.Writer) *R12Writer {
	return &R12Writer{
		w:   lldxf.NewASCIITagWriter(w, format.AC1009),
		fmt: format.NewASCIIWithVersion(format.AC1009),
	}
}

// WriteHeader writes the HEADER and TABLES sections and opens the ENTITIES
// section.
func (w *R12Writer) WriteHeader() error {
	if err := lldxf.WriteHeader(w.w, format.AC1009); err != nil {
		return err
	}
	if err := lldxf.WriteTables(w.w); err != nil {
		return err
	}
	if err := w.w.WriteTag(lldxf.NewDXFTag(0, "SECTION")); err != nil {
		return err
	}
	return w.w.WriteTag(lldxf.NewDXFTag(2, "ENTITIES"))
}

// WriteEntity appends one entity record to the ENTITIES section.
func (w *R12Writer) WriteEntity(e entity.Entity) error {
	if e == nil {
		return nil
	}
	w.fmt.Reset()
	e.Format(w.fmt)
	tags, err := lldxf.LoadTags(strings.NewReader(w.fmt.Output()))
	if err != nil {
		return err
	}
	return lldxf.WriteEntity(w.w, tags)
}

// Close closes the ENTITIES section and terminates the file.
func (w *R12Writer) Close() error {
	if err := w.w.WriteTag(lldxf.NewDXFTag(0, "ENDSEC")); err != nil {
		return err
	}
	return lldxf.WriteEOF(w.w)
}
