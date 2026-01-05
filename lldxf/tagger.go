package lldxf

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Tagger represents a DXF tag iterator
type Tagger interface {
	Next() bool
	Tag() Tag
	Err() error
}

// asciiTagger implements Tagger for ASCII DXF files
type asciiTagger struct {
	scanner *bufio.Scanner
	tag     Tag
	err     error
}

// NewASCIITagger creates a new ASCII DXF tagger
func NewASCIITagger(r io.Reader) Tagger {
	scanner := bufio.NewScanner(r)
	return &asciiTagger{scanner: scanner}
}

// Next advances to the next tag
func (t *asciiTagger) Next() bool {
	if t.err != nil {
		return false
	}

	// Read group code
	if !t.scanner.Scan() {
		if t.scanner.Err() != nil {
			t.err = t.scanner.Err()
		}
		return false
	}
	codeStr := strings.TrimSpace(t.scanner.Text())

	// Read value
	if !t.scanner.Scan() {
		t.err = fmt.Errorf("unexpected EOF after group code %s", codeStr)
		return false
	}
	value := strings.TrimSpace(t.scanner.Text())

	// Parse group code
	code, err := strconv.Atoi(codeStr)
	if err != nil {
		t.err = fmt.Errorf("invalid group code: %s", codeStr)
		return false
	}

	// Create tag
	t.tag = NewDXFTag(GroupCode(code), value)
	return true
}

// Tag returns the current tag
func (t *asciiTagger) Tag() Tag {
	return t.tag
}

// Err returns any error that occurred
func (t *asciiTagger) Err() error {
	return t.err
}

// LoadTags loads all tags from a reader
func LoadTags(r io.Reader) (Tags, error) {
	tagger := NewASCIITagger(r)
	var tags Tags

	for tagger.Next() {
		tags = tags.Add(tagger.Tag())
	}

	if err := tagger.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

// LoadTagsFromFile loads all tags from a file
func LoadTagsFromFile(filename string) (Tags, error) {
	// Check if file is gzipped
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Try to detect gzip by reading magic bytes
	buf := make([]byte, 2)
	if _, err := io.ReadFull(file, buf); err != nil {
		return nil, err
	}

	// Reset file position
	if _, err := file.Seek(0, 0); err != nil {
		return nil, err
	}

	var reader io.Reader = file

	// Check for gzip magic number (0x1f, 0x8b)
	if buf[0] == 0x1f && buf[1] == 0x8b {
		gzReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		defer gzReader.Close()
		reader = gzReader
	}

	return LoadTags(reader)
}

// GroupTagsBySection groups tags by DXF sections
func GroupTagsBySection(tags Tags) (map[string]Tags, error) {
	sections := make(map[string]Tags)
	var currentSection string
	var inSection bool

	for _, tag := range tags {
		switch tag.Code() {
		case 0:
			value := fmt.Sprintf("%v", tag.Value())
			if value == "SECTION" {
				inSection = true
			} else if value == "ENDSEC" {
				inSection = false
				currentSection = ""
			}
		case 2:
			if inSection {
				currentSection = fmt.Sprintf("%v", tag.Value())
				sections[currentSection] = NewTags()
			}
		default:
			if currentSection != "" {
				sections[currentSection] = sections[currentSection].Add(tag)
			}
		}
	}

	return sections, nil
}

// ExtractEntities extracts entity tags from the ENTITIES section
func ExtractEntities(sections map[string]Tags) ([]Tags, error) {
	entitiesTags, ok := sections[SectionEntities]
	if !ok {
		return nil, fmt.Errorf("no ENTITIES section found")
	}

	var entities []Tags
	var currentEntity Tags

	for _, tag := range entitiesTags {
		if tag.Code() == 0 {
			// Start of new entity
			if len(currentEntity) > 0 {
				// Save previous entity
				entities = append(entities, currentEntity)
			}
			currentEntity = NewTags().Add(tag)
		} else {
			currentEntity = currentEntity.Add(tag)
		}
	}

	// Add last entity if exists
	if len(currentEntity) > 0 {
		entities = append(entities, currentEntity)
	}

	return entities, nil
}

// ValidateTags performs basic validation on a collection of tags
func ValidateTags(tags Tags) error {
	if len(tags) == 0 {
		return fmt.Errorf("no tags to validate")
	}

	// Check for required section structure
	hasSection := false

	for i, tag := range tags {
		code := tag.Code()
		value := fmt.Sprintf("%v", tag.Value())

		if code == 0 && value == "SECTION" {
			hasSection = true
			// Check for section name
			if i+1 < len(tags) && tags[i+1].Code() == 2 {
				sectionName := fmt.Sprintf("%v", tags[i+1].Value())
				if sectionName == "" {
					return fmt.Errorf("empty section name at position %d", i)
				}
			} else {
				return fmt.Errorf("missing section name after SECTION at position %d", i)
			}
		}
	}

	if !hasSection {
		return fmt.Errorf("no SECTION tag found")
	}

	// Basic structural validation would go here
	return nil
}

// TaggerFromBytes creates a tagger from byte data
func TaggerFromBytes(data []byte) Tagger {
	return NewASCIITagger(bytes.NewReader(data))
}

// TagsToString converts tags back to DXF string format
func TagsToString(tags Tags) string {
	var builder strings.Builder
	for _, tag := range tags {
		builder.WriteString(tag.String())
	}
	return builder.String()
}

// FilterEntities filters entities by type
func FilterEntities(entities []Tags, entityType string) []Tags {
	var filtered []Tags
	for _, entity := range entities {
		if entType, ok := entity.GetFirstValue(0); ok {
			if fmt.Sprintf("%v", entType) == entityType {
				filtered = append(filtered, entity)
			}
		}
	}
	return filtered
}

// GetEntityHandle extracts the handle from an entity
func GetEntityHandle(entity Tags) string {
	if handle, ok := entity.GetFirstValue(5); ok {
		return fmt.Sprintf("%v", handle)
	}
	return ""
}

// GetEntityLayer extracts the layer name from an entity
func GetEntityLayer(entity Tags) string {
	if layer, ok := entity.GetFirstValue(8); ok {
		return fmt.Sprintf("%v", layer)
	}
	return "0"
}

// CountEntitiesByType counts entities by their type
func CountEntitiesByType(entities []Tags) map[string]int {
	counts := make(map[string]int)
	for _, entity := range entities {
		if entType, ok := entity.GetFirstValue(0); ok {
			typeName := fmt.Sprintf("%v", entType)
			counts[typeName]++
		}
	}
	return counts
}
