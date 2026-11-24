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
	"unicode/utf8"
)

// Tagger represents a DXF tag iterator
type Tagger interface {
	Next() bool
	Tag() Tag
	Err() error
}

// asciiTagger implements Tagger for ASCII DXF files.
//
// Lines are read straight from the buffered reader instead of bufio.Scanner so
// the group-code line never becomes a string: Scanner.Text() allocates a copy
// per line, and half of every DXF file is group-code lines. The value line is
// copied exactly once, into the string the tag retains.
type asciiTagger struct {
	r     *bufio.Reader
	tag   Tag
	err   error
	spill []byte
}

// NewASCIITagger creates a new ASCII DXF tagger
func NewASCIITagger(r io.Reader) Tagger {
	return &asciiTagger{r: bufio.NewReaderSize(r, readerBufferSize(r))}
}

const (
	// smallInputLimit is the size below which a small read buffer is cheaper
	// than allocating a large one: a minimal DXF is a few hundred bytes, and
	// paying for a 64 KiB buffer there costs more than the whole parse.
	smallInputLimit = 8 * 1024
	smallReaderBuf  = 4 * 1024
	largeReaderBuf  = 64 * 1024
	// unknownReaderBuf is used for streams that cannot report their size.
	unknownReaderBuf = 32 * 1024
)

// readerBufferSize picks a read buffer size from the reader's remaining size.
func readerBufferSize(r io.Reader) int {
	size := readerSizeHint(r)
	switch {
	case size <= 0:
		return unknownReaderBuf
	case size <= smallInputLimit:
		return smallReaderBuf
	default:
		return largeReaderBuf
	}
}

// readerSizeHint reports how many bytes the reader has left, or 0 when the
// reader cannot say.
func readerSizeHint(r io.Reader) int {
	switch v := r.(type) {
	case interface{ Len() int }:
		return v.Len()
	case interface{ Stat() (os.FileInfo, error) }:
		if info, err := v.Stat(); err == nil && info.Size() > 0 {
			return int(info.Size())
		}
	}
	return 0
}

// readLine returns the next raw line including its terminator. The returned
// slice is only valid until the following read. ok is false at end of input.
func (t *asciiTagger) readLine() (line []byte, ok bool) {
	line, err := t.r.ReadSlice('\n')
	if err == nil {
		return line, true
	}
	if err == bufio.ErrBufferFull {
		// Line longer than the buffer: accumulate it in owned storage.
		buf := append(t.spill[:0], line...)
		for {
			line, err = t.r.ReadSlice('\n')
			buf = append(buf, line...)
			if err == bufio.ErrBufferFull {
				continue
			}
			t.spill = buf
			if err == nil || err == io.EOF {
				return buf, true
			}
			t.err = err
			return nil, false
		}
	}
	if err == io.EOF {
		if len(line) > 0 {
			return line, true // final line without a terminator
		}
		return nil, false
	}
	t.err = err
	return nil, false
}

// Next advances to the next tag
func (t *asciiTagger) Next() bool {
	if t.err != nil {
		return false
	}

	codeLine, ok := t.readLine()
	if !ok {
		return false
	}
	// The code must be parsed before the next read: the slice returned by
	// readLine aliases the reader's buffer and is invalidated by it.
	code, ok := parseGroupCode(codeLine)
	if !ok {
		t.err = fmt.Errorf("invalid group code: %s", string(trimASCIISpace(codeLine)))
		return false
	}

	valueLine, ok := t.readLine()
	if !ok {
		if t.err == nil {
			t.err = fmt.Errorf("unexpected EOF after group code %d", code)
		}
		return false
	}

	// The value is copied out of the read buffer exactly once: it is retained
	// by the tag, so it must not alias the reader's scratch space.
	value := string(trimASCIISpace(valueLine))

	// Create tag. Binary data codes (260/261, 310-319, 1004) carry hex-encoded
	// payloads in ASCII DXF; decode them to raw bytes so repeated round trips
	// neither grow nor corrupt the payload.
	if isBinaryDataCode(GroupCode(code)) {
		t.tag = BinaryTagFromHex(GroupCode(code), value)
	} else {
		t.tag = NewDXFTag(GroupCode(code), value)
	}
	return true
}

// trimASCIISpace trims the whitespace a DXF line carries around its content.
// The ASCII fast path avoids decoding runes for every line; when a non-ASCII
// byte sits at either end it defers to bytes.TrimSpace so the result stays
// byte-identical to the strings.TrimSpace this reader replaced (which also
// strips characters such as NO-BREAK SPACE).
func trimASCIISpace(b []byte) []byte {
	start := 0
	for start < len(b) && isASCIIWhitespace(b[start]) {
		start++
	}
	end := len(b)
	for end > start && isASCIIWhitespace(b[end-1]) {
		end--
	}
	b = b[start:end]
	if len(b) > 0 && (b[0] >= utf8.RuneSelf || b[len(b)-1] >= utf8.RuneSelf) {
		return bytes.TrimSpace(b)
	}
	return b
}

func isASCIIWhitespace(c byte) bool {
	switch c {
	case ' ', '\t', '\r', '\n', '\v', '\f':
		return true
	}
	return false
}

// maxFastGroupCodeDigits bounds the manual integer parse. Real group codes are
// at most four digits, so anything longer falls back to strconv.Atoi to keep
// its exact overflow behavior.
const maxFastGroupCodeDigits = 6

// parseGroupCode parses a group-code line without allocating.
func parseGroupCode(b []byte) (int, bool) {
	tok := trimASCIISpace(b)
	if len(tok) == 0 {
		return 0, false
	}
	if len(tok) > maxFastGroupCodeDigits {
		code, err := strconv.Atoi(string(tok))
		if err != nil {
			return 0, false
		}
		return code, true
	}
	neg := false
	i := 0
	if tok[0] == '+' || tok[0] == '-' {
		neg = tok[0] == '-'
		i = 1
		if len(tok) == 1 {
			return 0, false
		}
	}
	n := 0
	for ; i < len(tok); i++ {
		c := tok[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	if neg {
		n = -n
	}
	return n, true
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
	tags := make(Tags, 0, tagCapacityHint(r))

	for tagger.Next() {
		tags = append(tags, tagger.Tag())
	}

	if err := tagger.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

// averageTagBytes is the conservative bytes-per-tag ratio used to preallocate
// the tag slice. ASCII DXF stores one tag per two lines, and a small tag such as
// "0\nA\n" occupies four bytes, so the quotient still under-fills in practice.
const averageTagBytes = 8

// maxTagCapacityHint bounds the preallocation. Files dominated by long values
// (binary chunks, MTEXT bodies) yield far fewer tags per byte than the ratio
// above assumes, and the hint must not turn into a large speculative
// allocation; growth past the cap is left to append.
const maxTagCapacityHint = 1 << 21

// tagCapacityHint reports how many tags the reader is likely to yield when it
// can report its remaining size, so LoadTags does not repeatedly regrow and
// copy the slice. Readers of unknown size simply start empty.
func tagCapacityHint(r io.Reader) int {
	n := readerSizeHint(r)
	if n <= 0 {
		return 0
	}
	return min(n/averageTagBytes+1, maxTagCapacityHint)
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
			switch value {
			case "SECTION":
				inSection = true
			case "ENDSEC":
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
