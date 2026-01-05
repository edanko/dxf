package drawing

import (
	"fmt"
	"strings"

	"github.com/edanko/dxf/format"
	"github.com/edanko/dxf/header"
)

// Section represents a generic DXF section
type Section interface {
	Type() SectionType
	WriteHeader(f format.Formatter) error
	WriteContent(f format.Formatter) error
	WriteFooter(f format.Formatter) error
}

// BaseSection provides common functionality for all sections
type BaseSection struct {
	sectionType SectionType
	name        string
}

// Type returns the section type
func (bs *BaseSection) Type() SectionType {
	return bs.sectionType
}

// WriteHeader writes the standard DXF section header
func (bs *BaseSection) WriteHeader(f format.Formatter) error {
	f.WriteString(0, bs.name)
	return nil
}

// WriteFooter writes the standard DXF section footer
func (bs *BaseSection) WriteFooter(f format.Formatter) error {
	f.WriteString(0, "ENDSEC")
	return nil
}

// NewBaseSection creates a new base section
func NewBaseSection(sectionType SectionType, name string) *BaseSection {
	return &BaseSection{
		sectionType: sectionType,
		name:        name,
	}
}

// HeaderSection represents the DXF HEADER section
type HeaderSection struct {
	*BaseSection
	header *header.Header
}

// NewHeaderSection creates a new header section
func NewHeaderSection(hdr *header.Header) *HeaderSection {
	return &HeaderSection{
		BaseSection: NewBaseSection(HEADER, "HEADER"),
		header:      hdr,
	}
}

// WriteHeader writes the header section start
func (hs *HeaderSection) WriteHeader(f format.Formatter) error {
	if err := hs.BaseSection.WriteHeader(f); err != nil {
		return err
	}
	return nil
}

// WriteContent writes the header content
func (hs *HeaderSection) WriteContent(f format.Formatter) error {
	if hs.header == nil {
		return fmt.Errorf("header is nil")
	}
	hs.header.Format(f)
	return nil
}

// ClassesSection represents the DXF CLASSES section
type ClassesSection struct {
	*BaseSection
	classes []string
}

// NewClassesSection creates a new classes section
func NewClassesSection() *ClassesSection {
	return &ClassesSection{
		BaseSection: NewBaseSection(CLASSES, "CLASSES"),
		classes:     make([]string, 0),
	}
}

// AddClass adds a class to the section
func (cs *ClassesSection) AddClass(classDef string) {
	cs.classes = append(cs.classes, classDef)
}

// WriteContent writes the classes content
func (cs *ClassesSection) WriteContent(f format.Formatter) error {
	for _, classDef := range cs.classes {
		f.WriteString(0, classDef)
	}
	return nil
}

// TablesSection represents the DXF TABLES section
type TablesSection struct {
	*BaseSection
	tables map[string]TableWriter
}

// TableWriter interface for different table types
type TableWriter interface {
	Write(f format.Formatter) error
	Name() string
}

// NewTablesSection creates a new tables section
func NewTablesSection() *TablesSection {
	return &TablesSection{
		BaseSection: NewBaseSection(TABLES, "TABLES"),
		tables:      make(map[string]TableWriter),
	}
}

// AddTable adds a table to the section
func (ts *TablesSection) AddTable(name string, table TableWriter) {
	ts.tables[name] = table
}

// WriteContent writes the tables content
func (ts *TablesSection) WriteContent(f format.Formatter) error {
	// Write each table
	for name, table := range ts.tables {
		f.WriteString(0, fmt.Sprintf("  %s", strings.ToUpper(name)))
		if err := table.Write(f); err != nil {
			return err
		}
		f.WriteString(0, "ENDTAB")
	}
	return nil
}

// BlocksSection represents the DXF BLOCKS section
type BlocksSection struct {
	*BaseSection
	blocks map[string]BlockWriter
}

// BlockWriter interface for block definitions
type BlockWriter interface {
	Write(f format.Formatter) error
	Name() string
	Handle() string
}

// NewBlocksSection creates a new blocks section
func NewBlocksSection() *BlocksSection {
	return &BlocksSection{
		BaseSection: NewBaseSection(BLOCKS, "BLOCKS"),
		blocks:      make(map[string]BlockWriter),
	}
}

// AddBlock adds a block to the section
func (bs *BlocksSection) AddBlock(name string, block BlockWriter) {
	bs.blocks[name] = block
}

// WriteContent writes the blocks content
func (bs *BlocksSection) WriteContent(f format.Formatter) error {
	for name, block := range bs.blocks {
		f.WriteString(0, fmt.Sprintf("  %s", strings.ToUpper(name)))
		if err := block.Write(f); err != nil {
			return err
		}
		f.WriteString(0, "ENDBLK")
	}
	return nil
}

// ObjectsSection represents the DXF OBJECTS section
type ObjectsSection struct {
	*BaseSection
	objects map[string]ObjectWriter
}

// ObjectWriter interface for object definitions
type ObjectWriter interface {
	Write(f format.Formatter) error
	Name() string
	Handle() string
}

// NewObjectsSection creates a new objects section
func NewObjectsSection() *ObjectsSection {
	return &ObjectsSection{
		BaseSection: NewBaseSection(OBJECTS, "OBJECTS"),
		objects:     make(map[string]ObjectWriter),
	}
}

// AddObject adds an object to the section
func (os *ObjectsSection) AddObject(name string, object ObjectWriter) {
	os.objects[name] = object
}

// WriteContent writes the objects content
func (os *ObjectsSection) WriteContent(f format.Formatter) error {
	for name, object := range os.objects {
		f.WriteString(0, fmt.Sprintf("  %s", strings.ToUpper(name)))
		if err := object.Write(f); err != nil {
			return err
		}
		f.WriteString(0, "ENDOBJ")
	}
	return nil
}

// SectionWriter manages writing multiple sections to DXF format
type SectionWriter struct {
	formatter format.Formatter
	sections  []Section
}

// NewSectionWriter creates a new section writer
func NewSectionWriter(f format.Formatter) *SectionWriter {
	return &SectionWriter{
		formatter: f,
		sections:  make([]Section, 0),
	}
}

// AddSection adds a section to the writer
func (sw *SectionWriter) AddSection(section Section) error {
	sw.sections = append(sw.sections, section)
	return nil
}

// WriteAll writes all sections to the formatter
func (sw *SectionWriter) WriteAll() error {
	for _, section := range sw.sections {
		// Write section header
		if err := section.WriteHeader(sw.formatter); err != nil {
			return err
		}

		// Write section content
		if err := section.WriteContent(sw.formatter); err != nil {
			return err
		}

		// Write section footer
		if err := section.WriteFooter(sw.formatter); err != nil {
			return err
		}
	}
	return nil
}

// WriteDXFHeader writes the standard DXF file header
func (sw *SectionWriter) WriteDXFHeader() error {
	sw.formatter.WriteString(999, "DXF file created by Go DXF library")
	sw.formatter.WriteString(0, "SECTION")
	return nil
}

// WriteDXFFooter writes the standard DXF file footer
func (sw *SectionWriter) WriteDXFFooter() error {
	sw.formatter.WriteString(0, "ENDSEC")
	sw.formatter.WriteString(0, "EOF")
	return nil
}
