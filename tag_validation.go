package dxf

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/edanko/dxf/lldxf"
)

// tagValidationIssue records one deterministic tag-stream defect. Code, Value
// and Handle identify the offending group, its wire value, and the record
// handle so a malformed fixture can be reported precisely.
type tagValidationIssue struct {
	Code     int
	Value    string
	Handle   string
	Message  string
	Repaired bool
}

func (i tagValidationIssue) String() string {
	return fmt.Sprintf("tag validation: code %d value %q handle %q: %s", i.Code, i.Value, i.Handle, i.Message)
}

// validateTagStream checks section terminators, required handles, numeric
// ranges, duplicate handles, and missing block terminators before typed
// binding. When repair is true only deterministic safe defects are fixed:
// terminators are inserted and duplicate/missing handles are assigned.
func validateTagStream(tags lldxf.Tags, repair bool) (lldxf.Tags, []tagValidationIssue) {
	var issues []tagValidationIssue
	seen := make(map[string]bool)
	maxHandle := 0
	for _, tag := range tags {
		code := int(tag.Code())
		if code != 5 && code != 105 {
			continue
		}
		h := tagStringValue(tag)
		if h != "" {
			if v, err := strconv.ParseUint(h, 16, 64); err == nil && int(v) > maxHandle {
				maxHandle = int(v)
			}
		}
	}
	fresh := func() string {
		for {
			maxHandle++
			h := strings.ToUpper(strconv.FormatInt(int64(maxHandle), 16))
			if h != "" && !seen[h] {
				return h
			}
		}
	}

	out := make(lldxf.Tags, 0, len(tags))
	sectionOpen := false
	blockOpen := false
	sectionName := ""

	insert := func(name string) lldxf.Tag {
		return lldxf.NewDXFTag(0, name)
	}
	report := func(code int, value, handle, msg string, fixed bool) {
		issues = append(issues, tagValidationIssue{Code: code, Value: value, Handle: handle, Message: msg, Repaired: fixed})
	}

	for i := 0; i < len(tags); {
		tag := tags[i]
		code := int(tag.Code())
		if code != 0 {
			out = append(out, tag)
			i++
			continue
		}

		name := strings.ToUpper(tagStringValue(tag))
		handle := ""
		// Find the record's handle before writing the code-0 marker.
		if name != "SECTION" && name != "ENDSEC" && name != "EOF" && name != "TABLE" &&
			name != "ENDTAB" && name != "BLOCK" && name != "ENDBLK" && name != "CLASS" {
			for j := i + 1; j < len(tags) && int(tags[j].Code()) != 0; j++ {
				c := int(tags[j].Code())
				if c == 5 || c == 105 {
					handle = tagStringValue(tags[j])
					break
				}
			}
		}

		switch name {
		case "SECTION":
			if sectionOpen {
				report(0, name, "", "SECTION opened before previous ENDSEC", false)
			}
			sectionOpen = true
			blockOpen = false
			sectionName = ""
		case "ENDSEC":
			if blockOpen {
				report(0, "ENDBLK", "", "block missing ENDBLK before ENDSEC", repair)
				if repair {
					out = append(out, insert("ENDBLK"))
					blockOpen = false
				}
			}
			if !sectionOpen {
				report(0, name, "", "ENDSEC without SECTION", false)
			}
			sectionOpen = false
			sectionName = ""
		case "EOF":
			if blockOpen {
				report(0, "ENDBLK", "", "block missing ENDBLK before EOF", repair)
				if repair {
					out = append(out, insert("ENDBLK"))
					blockOpen = false
				}
			}
			if sectionOpen {
				report(0, "ENDSEC", "", "section missing ENDSEC before EOF", repair)
				if repair {
					out = append(out, insert("ENDSEC"))
					sectionOpen = false
				}
			}
		case "BLOCK":
			if blockOpen {
				report(0, name, "", "nested BLOCK before ENDBLK", false)
			}
			blockOpen = true
		case "ENDBLK":
			if !blockOpen {
				report(0, name, handle, "ENDBLK without BLOCK", false)
			}
			blockOpen = false
		}

		// Copy the record, validating and repairing its handle.
		out = append(out, tag)
		i++
		record := make(lldxf.Tags, 0, 8)
		for i < len(tags) && int(tags[i].Code()) != 0 {
			record = append(record, tags[i])
			i++
		}
		if sectionOpen && sectionName == "" && len(record) > 0 && int(record[0].Code()) == 2 {
			sectionName = strings.ToUpper(tagStringValue(record[0]))
		}

		needsHandle := name != "SECTION" && name != "ENDSEC" && name != "EOF" && name != "TABLE" &&
			name != "ENDTAB" && name != "BLOCK" && name != "ENDBLK" && name != "CLASS"
		handleIndex := -1
		for j, t := range record {
			c := int(t.Code())
			if c == 5 || c == 105 {
				handleIndex = j
				handle = tagStringValue(t)
				break
			}
		}
		if needsHandle && handleIndex < 0 {
			report(0, name, "", "record missing required handle (code 5)", repair)
			if repair {
				handle = fresh()
				seen[handle] = true
				record = append([]lldxf.Tag{lldxf.NewDXFTag(5, handle)}, record...)
				handleIndex = 0
			}
		}
		if handleIndex >= 0 && handle != "" {
			if seen[handle] {
				report(5, handle, handle, "duplicate handle", repair)
				if repair {
					next := fresh()
					if setter, ok := record[handleIndex].(interface{ SetValue(any) }); ok {
						setter.SetValue(next)
					}
					handle = next
				}
			}
			seen[handle] = true
		}

		for _, t := range record {
			c := int(t.Code())
			value := tagStringValue(t)
			if msg := numericRangeError(c, value); msg != "" {
				report(c, value, handle, msg, false)
			}
		}
		out = append(out, record...)
	}

	if sectionOpen {
		report(0, "ENDSEC", "", "section missing ENDSEC at end of stream", repair)
		if repair {
			out = append(out, insert("ENDSEC"))
		}
	}
	if blockOpen {
		report(0, "ENDBLK", "", "block missing ENDBLK at end of stream", repair)
		if repair {
			out = append(out, insert("ENDBLK"))
		}
	}
	return out, issues
}

// numericRangeError validates parseable numeric group codes and obvious range
// constraints. Empty string means the value is acceptable.
func numericRangeError(code int, value string) string {
	if !isNumericTagCode(code) {
		return ""
	}
	n, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return "numeric group code has a non-numeric value"
	}
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return "numeric value is not finite"
	}
	switch code {
	case 62:
		if n < 0 || n > 256 {
			return "color index out of range 0..256"
		}
	case 60:
		if n != 0 && n != 1 {
			return "invisible flag must be 0 or 1"
		}
	case 290, 291:
		if n != 0 && n != 1 {
			return "boolean flag must be 0 or 1"
		}
	case 70, 71, 72, 73, 74, 75, 76, 77, 78:
		if n < 0 || n > math.MaxInt16 {
			return "integer flag out of range"
		}
	}
	return ""
}

func isNumericTagCode(code int) bool {
	switch {
	case code >= 10 && code <= 59:
		return true
	case code >= 60 && code <= 79:
		return true
	case code >= 90 && code <= 99:
		return true
	case code >= 110 && code <= 149:
		return true
	case code >= 170 && code <= 179:
		return true
	case code >= 210 && code <= 239:
		return true
	case code >= 270 && code <= 289:
		return true
	case code >= 290 && code <= 299:
		return true
	case code >= 370 && code <= 389:
		return true
	case code >= 400 && code <= 409:
		return true
	case code == 420:
		return true
	case code == 440:
		return true
	case code >= 450 && code <= 459:
		return true
	case code >= 460 && code <= 469:
		return true
	case code >= 470 && code <= 479:
		return true
	}
	return false
}

// applyTagValidation validates a source tag stream and optionally repairs it,
// returning messages suitable for a RecoveryReport.
func applyTagValidation(tags lldxf.Tags, repair bool) (lldxf.Tags, []string) {
	out, issues := validateTagStream(tags, repair)
	if len(issues) == 0 {
		return out, nil
	}
	msgs := make([]string, 0, len(issues))
	for _, issue := range issues {
		msgs = append(msgs, issue.String())
	}
	return out, msgs
}
