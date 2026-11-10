package entity

import "github.com/edanko/dxf/format"

// writeSATPayload emits retained ACIS text lines. Legacy records use code 10
// with a leading code 90 line count; current SAT records use code 1 for a line
// and code 3 for continuation chunks.
func writeSATPayload(f format.Formatter, lines []string, codes []int) {
	if len(lines) == 0 {
		return
	}
	legacy := true
	for _, code := range codes {
		if code != 10 {
			legacy = false
			break
		}
	}
	if legacy {
		f.WriteInt(90, len(lines))
		for _, line := range lines {
			f.WriteString(10, line)
		}
		return
	}
	for i, line := range lines {
		code := 1
		if i < len(codes) && codes[i] == 10 {
			code = 10
		}
		if code == 10 {
			f.WriteString(10, line)
			continue
		}
		first := true
		for len(line) > 255 {
			if first {
				f.WriteString(1, line[:255])
				first = false
			} else {
				f.WriteString(3, line[:255])
			}
			line = line[255:]
		}
		if first {
			f.WriteString(1, line)
		} else {
			f.WriteString(3, line)
		}
	}
}

// normalizeSATCodes returns one code per line, defaulting to code 10 for
// callers that only provide line text.
func normalizeSATCodes(lines []string, codes []int) []int {
	if len(codes) == len(lines) {
		return append([]int(nil), codes...)
	}
	out := make([]int, len(lines))
	for i := range out {
		out[i] = 10
	}
	return out
}
