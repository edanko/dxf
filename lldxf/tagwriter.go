package lldxf

import (
	"fmt"
	"io"
	"strconv"
)

// TagWriter interface for writing DXF tags
type TagWriter interface {
	WriteTag(tag Tag) error
	WriteTags(tags Tags) error
	Close() error
}

// ASCIITagWriter writes tags in ASCII DXF format
type ASCIITagWriter struct {
	writer  io.Writer
	version string
}

// NewASCIITagWriter creates a new ASCII tag writer
func NewASCIITagWriter(w io.Writer, version string) *ASCIITagWriter {
	if version == "" {
		version = VersionR2000
	}
	return &ASCIITagWriter{
		writer:  w,
		version: version,
	}
}

// WriteTag writes a single tag
func (w *ASCIITagWriter) WriteTag(tag Tag) error {
	_, err := io.WriteString(w.writer, tag.String())
	return err
}

// WriteTags writes multiple tags
func (w *ASCIITagWriter) WriteTags(tags Tags) error {
	for _, tag := range tags {
		if err := w.WriteTag(tag); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the writer
func (w *ASCIITagWriter) Close() error {
	// No-op for basic writer
	return nil
}

// TagCollector collects tags in memory (useful for testing)
type TagCollector struct {
	tags Tags
}

// NewTagCollector creates a new tag collector
func NewTagCollector() *TagCollector {
	return &TagCollector{
		tags: NewTags(),
	}
}

// WriteTag writes a single tag to collection
func (c *TagCollector) WriteTag(tag Tag) error {
	c.tags = c.tags.Add(tag)
	return nil
}

// WriteTags writes multiple tags to collection
func (c *TagCollector) WriteTags(tags Tags) error {
	c.tags = append(c.tags, tags...)
	return nil
}

// Close does nothing for collector
func (c *TagCollector) Close() error {
	return nil
}

// Tags returns the collected tags
func (c *TagCollector) Tags() Tags {
	return c.tags
}

// String returns the DXF string representation
func (c *TagCollector) String() string {
	return TagsToString(c.tags)
}

// Reset clears the collected tags
func (c *TagCollector) Reset() {
	c.tags = NewTags()
}

// WriteSection writes a complete DXF section
func WriteSection(writer TagWriter, name string, tags Tags) error {
	// Write section start
	sectionStart := []Tag{
		NewDXFTag(0, "SECTION"),
		NewDXFTag(2, name),
	}

	if err := writer.WriteTags(sectionStart); err != nil {
		return err
	}

	// Write section content
	if err := writer.WriteTags(tags); err != nil {
		return err
	}

	// Write section end
	if err := writer.WriteTag(NewDXFTag(0, "ENDSEC")); err != nil {
		return err
	}

	return nil
}

// WriteEntity writes a complete entity with proper structure
func WriteEntity(writer TagWriter, entity Tags) error {
	if len(entity) == 0 {
		return fmt.Errorf("empty entity")
	}

	// Ensure entity starts with type
	if entity[0].Code() != 0 {
		return fmt.Errorf("entity must start with group code 0")
	}

	return writer.WriteTags(entity)
}

// WriteHeader writes header section with basic structure
func WriteHeader(writer TagWriter, version string) error {
	headerTags := []Tag{
		NewDXFTag(9, "$ACADVER"), NewDXFTag(1, version),
		NewDXFTag(9, "$HANDSEED"), NewDXFTag(5, "FFFF"),
		NewDXFTag(9, "$INSUNITS"), NewDXFTag(70, 0),
		NewDXFTag(9, "$LUNITS"), NewDXFTag(70, 2),
		NewDXFTag(9, "$LTSCALE"), NewDXFTag(40, 1.0),
	}

	return WriteSection(writer, SectionHeader, headerTags)
}

// WriteTables writes basic tables section
func WriteTables(writer TagWriter) error {
	tableTags := NewTags()

	// Layer table
	layerTable := []Tag{
		NewDXFTag(0, "TABLE"),
		NewDXFTag(2, "LAYER"),
		NewDXFTag(70, 1), // Max entries
		NewDXFTag(0, "LAYER"),
		NewDXFTag(2, "0"),          // Layer name
		NewDXFTag(70, 0),           // Flags
		NewDXFTag(62, 7),           // Color (white)
		NewDXFTag(6, "Continuous"), // Linetype
		NewDXFTag(0, "ENDTAB"),
	}

	tableTags = append(tableTags, layerTable...)

	return WriteSection(writer, SectionTables, tableTags)
}

// WriteBlocks writes basic blocks section
func WriteBlocks(writer TagWriter) error {
	blockTags := NewTags()

	// Model space block
	modelSpace := []Tag{
		NewDXFTag(0, "BLOCK"),
		NewDXFTag(8, "0"), // Layer
		NewDXFTag(2, "*Model_Space"),
		NewDXFTag(70, 0),   // Flags
		NewDXFTag(10, 0.0), // Base point X
		NewDXFTag(20, 0.0), // Base point Y
		NewDXFTag(30, 0.0), // Base point Z
		NewDXFTag(3, "*Model_Space"),
		NewDXFTag(1, ""), // Xref path
		NewDXFTag(0, "ENDBLK"),
	}

	blockTags = append(blockTags, modelSpace...)

	// Paper space block
	paperSpace := []Tag{
		NewDXFTag(0, "BLOCK"),
		NewDXFTag(8, "0"), // Layer
		NewDXFTag(2, "*Paper_Space"),
		NewDXFTag(70, 0),   // Flags
		NewDXFTag(10, 0.0), // Base point X
		NewDXFTag(20, 0.0), // Base point Y
		NewDXFTag(30, 0.0), // Base point Z
		NewDXFTag(3, "*Paper_Space"),
		NewDXFTag(1, ""), // Xref path
		NewDXFTag(0, "ENDBLK"),
	}

	blockTags = append(blockTags, paperSpace...)

	return WriteSection(writer, SectionBlocks, blockTags)
}

// WriteEntities writes entities section
func WriteEntities(writer TagWriter, entities []Tags) error {
	// Write section start
	if err := writer.WriteTag(NewDXFTag(0, "SECTION")); err != nil {
		return err
	}
	if err := writer.WriteTag(NewDXFTag(2, "ENTITIES")); err != nil {
		return err
	}

	// Write all entities
	for _, entity := range entities {
		if err := WriteEntity(writer, entity); err != nil {
			return err
		}
	}

	// Write section end
	return writer.WriteTag(NewDXFTag(0, "ENDSEC"))
}

// WriteEOF writes end of file marker
func WriteEOF(writer TagWriter) error {
	return writer.WriteTag(NewDXFTag(0, "EOF"))
}

// CreateBasicDXF creates a minimal valid DXF file
func CreateBasicDXF(writer TagWriter, version string) error {
	if version == "" {
		version = VersionR2000
	}

	// Header
	if err := WriteHeader(writer, version); err != nil {
		return err
	}

	// Tables
	if err := WriteTables(writer); err != nil {
		return err
	}

	// Blocks
	if err := WriteBlocks(writer); err != nil {
		return err
	}

	// Empty entities section
	if err := WriteSection(writer, SectionEntities, NewTags()); err != nil {
		return err
	}

	// EOF
	return WriteEOF(writer)
}

// FormatTagValue formats a tag value for output
func FormatTagValue(value interface{}) string {
	switch v := value.(type) {
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 64)
	case int, int32, int64:
		return fmt.Sprintf("%d", v)
	case string:
		return v
	case []byte:
		// Binary data gets special handling in real implementation
		return string(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ValidateEntity performs basic validation on an entity
func ValidateEntity(entity Tags) error {
	if len(entity) == 0 {
		return fmt.Errorf("empty entity")
	}

	// Check entity type
	if entity[0].Code() != 0 {
		return fmt.Errorf("entity must start with group code 0 (entity type)")
	}

	entityType := fmt.Sprintf("%v", entity[0].Value())
	if entityType == "" {
		return fmt.Errorf("entity type cannot be empty")
	}

	// Entity-specific validation would go here
	return nil
}
