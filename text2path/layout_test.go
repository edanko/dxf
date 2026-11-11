package text2path

import (
	"strings"
	"testing"
)

func TestStripCodes(t *testing.T) {
	if got := StripCodes("line1\\Pline2"); got != "line1\nline2" {
		t.Errorf("split: %q", got)
	}
	if got := StripCodes(`{\fArial|b0|i0;Hello \W1.2World}`); !strings.Contains(got, "Hello") ||
		strings.Contains(got, "fArial") || strings.Contains(got, "\\W") {
		t.Errorf("format codes: %q", got)
	}
	if got := StripCodes("A\\~B"); got != "A B" {
		t.Errorf("non-breaking space: %q", got)
	}
	if got := StripCodes("A\\PC"); got != "A\nC" {
		t.Errorf("case-insensitive split: %q", got)
	}
}

func TestLayoutMTextWrap(t *testing.T) {
	text := "alpha beta gamma delta epsilon zeta"
	if lines := LayoutMText(text, 10, 0); len(lines) != 1 {
		t.Fatalf("unconstrained lines %d, want 1: %v", len(lines), lines)
	}
	// narrow box (1.5 words per line): greedy word wrap, no word split
	lines := LayoutMText(text, 10, 40)
	if len(lines) < 3 {
		t.Fatalf("wrapped lines %d, want >= 3: %v", len(lines), lines)
	}
	joined := strings.Join(lines, " ")
	if joined != strings.Join(strings.Fields(text), " ") {
		t.Errorf("word wrapping broke words: %v", lines)
	}
}

func TestLayoutMTextParagraphs(t *testing.T) {
	lines := LayoutMText("first\\Psecond\\Pthird", 8, 0)
	if len(lines) != 3 || lines[0] != "first" || lines[2] != "third" {
		t.Fatalf("paragraphs: %v", lines)
	}
	stripped := LayoutMText(`{\H2;Big}text`, 8, 0)
	if len(stripped) != 1 || !strings.Contains(stripped[0], "text") {
		t.Fatalf("codes stripped: %v", stripped)
	}
}
