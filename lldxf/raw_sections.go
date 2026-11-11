package lldxf

import (
	"bytes"
	"fmt"
)

// RawSections holds ordered section bodies exactly as they were tagged, before
// any entity, table, block, or object binding.
type RawSections struct {
	Order    []string
	Sections map[string]Tags
}

// Tags returns the ordered tags of a named section.
func (r *RawSections) Tags(name string) (Tags, bool) {
	if r == nil || r.Sections == nil {
		return nil, false
	}
	tags, ok := r.Sections[name]
	return tags, ok
}

// ParseRawSections reads ASCII or binary DXF data into named section tag
// streams without binding the document. On truncated input it returns every
// tag read so far together with an error naming the open section.
func ParseRawSections(data []byte) (*RawSections, error) {
	var tagger Tagger
	if IsBinaryDXF(data) {
		tagger = NewBinaryTaggerFromBytes(data)
	} else {
		tagger = NewASCIITagger(bytes.NewReader(data))
	}

	rs := &RawSections{Sections: make(map[string]Tags)}
	var current string
	var inSection bool
	var openAt int
	count := 0
	for tagger.Next() {
		count++
		tag := tagger.Tag()
		if tag == nil {
			continue
		}
		code := int(tag.Code())
		value := fmt.Sprintf("%v", tag.Value())
		switch {
		case code == 0 && value == "SECTION":
			if inSection {
				return rs, fmt.Errorf("raw sections: SECTION inside open section %q at tag %d", current, count)
			}
			inSection = true
			current = ""
			openAt = count
		case code == 2 && inSection && current == "":
			current = value
			if _, ok := rs.Sections[current]; !ok {
				rs.Order = append(rs.Order, current)
			}
		case code == 0 && value == "ENDSEC":
			if !inSection {
				return rs, fmt.Errorf("raw sections: ENDSEC without SECTION at tag %d", count)
			}
			if current == "" {
				return rs, fmt.Errorf("raw sections: SECTION at tag %d has no name", openAt)
			}
			inSection = false
			current = ""
		case inSection:
			if current == "" {
				return rs, fmt.Errorf("raw sections: SECTION at tag %d has no name", openAt)
			}
			rs.Sections[current] = append(rs.Sections[current], tag)
		}
	}

	if err := tagger.Err(); err != nil {
		if inSection && current != "" {
			return rs, fmt.Errorf("raw sections: truncated input in section %q: %w", current, err)
		}
		return rs, fmt.Errorf("raw sections: %w", err)
	}
	if inSection {
		if current == "" {
			return rs, fmt.Errorf("raw sections: truncated input in unnamed section")
		}
		return rs, fmt.Errorf("raw sections: truncated input in section %q", current)
	}
	return rs, nil
}
