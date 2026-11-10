package text2path

import (
	"strings"
	"unicode"
)

// StripCodes removes common MText formatting codes from text,
// converting \P and \p paragraph breaks to newlines. Supported codes:
// \P \p \\ \L \l \K \k \O \o \S (stacked), \A \C \F \H \Q \T \W \p,
// and brace groups. Unknown backslash sequences are dropped.
func StripCodes(text string) string {
	var b strings.Builder
	for i := 0; i < len(text); i++ {
		ch := text[i]
		if ch != '\\' || i+1 >= len(text) {
			if ch == '{' || ch == '}' {
				continue
			}
			b.WriteByte(ch)
			continue
		}
		next := text[i+1]
		switch next {
		case 'P', 'p':
			b.WriteByte('\n')
			i++
		case '\\':
			b.WriteByte('\\')
			i++
		case '~': // non-breaking space
			b.WriteByte(' ')
			i++
		case 'L', 'l', 'K', 'k', 'O', 'o':
			i++
		case 'S': // stacked fraction: \Sabove^below;
			i++
			for i < len(text) && text[i] != ';' {
				c := text[i]
				if c == '^' || c == '#' || c == '/' {
					i++
					b.WriteByte(' ')
					continue
				}
				b.WriteByte(c)
				i++
			}
		case 'A', 'C', 'F', 'H', 'Q', 'T', 'W':
			i++
			for i < len(text) && text[i] != ';' {
				i++
			}
		default:
			i++ // drop unknown escape
		}
	}
	return b.String()
}

// LayoutMText splits MText content into rendered text lines:
// paragraph breaks (\P) become line breaks, and lines wrap greedily at
// the column width using the grid glyph metrics (character width is
// 5/6 of the text height). A non-positive width disables wrapping.
func LayoutMText(text string, height, width float64) []string {
	cleaned := StripCodes(text)
	lines := strings.Split(cleaned, "\n")
	if width <= 0 || height <= 0 {
		return lines
	}
	charWidth := height * 100.0 / 120.0
	maxChars := max(int(width/charWidth), 1)
	var out []string
	for _, line := range lines {
		if len(line) <= maxChars {
			out = append(out, line)
			continue
		}
		current := ""
		words := strings.FieldsSeq(line)
		for w := range words {
			if current == "" {
				current = w
				continue
			}
			if len(current)+1+len(w) <= maxChars {
				current += " " + w
				continue
			}
			out = append(out, current)
			current = w
		}
		if current != "" {
			out = append(out, current)
		}
	}
	return out
}

// IsSpace reports r as layout whitespace (exposed for tests).
func IsSpace(r rune) bool { return unicode.IsSpace(r) }
