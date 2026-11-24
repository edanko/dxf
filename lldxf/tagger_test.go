package lldxf

import (
	"bytes"
	"strings"
	"testing"
)

// collectTags drains a tagger, returning its tags and terminating error.
func collectTags(t Tagger) (Tags, error) {
	var tags Tags
	for t.Next() {
		tags = append(tags, t.Tag())
	}
	return tags, t.Err()
}

func TestASCIITaggerBasic(t *testing.T) {
	tags, err := collectTags(NewASCIITagger(strings.NewReader("0\nSECTION\n2\nENTITIES\n0\nENDSEC\n")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []struct {
		code  GroupCode
		value string
	}{
		{0, "SECTION"},
		{2, "ENTITIES"},
		{0, "ENDSEC"},
	}
	if len(tags) != len(want) {
		t.Fatalf("got %d tags, want %d: %v", len(tags), len(want), tags)
	}
	for i, w := range want {
		if tags[i].Code() != w.code || FormatValue(tags[i]) != w.value {
			t.Errorf("tag %d = (%d, %q), want (%d, %q)", i, tags[i].Code(), FormatValue(tags[i]), w.code, w.value)
		}
	}
}

func TestASCIITaggerLineEndings(t *testing.T) {
	for _, sep := range []string{"\n", "\r\n"} {
		input := strings.ReplaceAll("0\nSECTION\n2\nENTITIES\n", "\n", sep)
		tags, err := collectTags(NewASCIITagger(strings.NewReader(input)))
		if err != nil {
			t.Fatalf("line ending %q: unexpected error: %v", sep, err)
		}
		if len(tags) != 2 || FormatValue(tags[0]) != "SECTION" || FormatValue(tags[1]) != "ENTITIES" {
			t.Fatalf("line ending %q: got %v", sep, tags)
		}
	}
}

func TestASCIITaggerFinalLineWithoutTerminator(t *testing.T) {
	tags, err := collectTags(NewASCIITagger(strings.NewReader("0\nEOF")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 1 || tags[0].Code() != 0 || FormatValue(tags[0]) != "EOF" {
		t.Fatalf("got %v, want one tag (0, EOF)", tags)
	}
}

// TestASCIITaggerLongLine covers lines longer than the read buffer. The
// bufio.Scanner this replaced reported "token too long" for these.
func TestASCIITaggerLongLine(t *testing.T) {
	long := strings.Repeat("A", 200*1024)
	tags, err := collectTags(NewASCIITagger(strings.NewReader("1\n" + long + "\n2\nnext\n")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("got %d tags, want 2", len(tags))
	}
	if got := FormatValue(tags[0]); got != long {
		t.Errorf("long value length = %d, want %d", len(got), len(long))
	}
	if got := FormatValue(tags[1]); got != "next" {
		t.Errorf("tag after long value = %q, want %q", got, "next")
	}
}

func TestASCIITaggerErrors(t *testing.T) {
	t.Run("invalid group code", func(t *testing.T) {
		_, err := collectTags(NewASCIITagger(strings.NewReader("notacode\nvalue\n")))
		if err == nil || !strings.Contains(err.Error(), "invalid group code") {
			t.Fatalf("err = %v, want invalid group code", err)
		}
	})
	t.Run("dangling group code", func(t *testing.T) {
		_, err := collectTags(NewASCIITagger(strings.NewReader("0\n")))
		if err == nil || !strings.Contains(err.Error(), "unexpected EOF") {
			t.Fatalf("err = %v, want unexpected EOF", err)
		}
	})
	t.Run("empty input", func(t *testing.T) {
		tags, err := collectTags(NewASCIITagger(strings.NewReader("")))
		if err != nil || len(tags) != 0 {
			t.Fatalf("tags=%v err=%v, want none and nil", tags, err)
		}
	})
}

func TestASCIITaggerBinaryPayload(t *testing.T) {
	tags, err := collectTags(NewASCIITagger(strings.NewReader("310\nDEADBEEF\n310\nNOTHEX\n")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("got %d tags, want 2", len(tags))
	}
	if got, want := tags[0].Value(), []byte{0xDE, 0xAD, 0xBE, 0xEF}; !bytes.Equal(got.([]byte), want) {
		t.Errorf("hex payload = %v, want %v", got, want)
	}
	if got, want := tags[1].Value(), []byte("NOTHEX"); !bytes.Equal(got.([]byte), want) {
		t.Errorf("non-hex payload = %v, want %v", got, want)
	}
}

// TestASCIITaggerTrimsLikeTrimSpace pins whitespace handling to the
// strings.TrimSpace behavior the scanner replaced, including non-ASCII spaces.
func TestASCIITaggerTrimsLikeTrimSpace(t *testing.T) {
	values := []string{
		"plain",
		"  padded  ",
		"\tpadded\t",
		"\r padded \r",
		"\u00a0nbsp\u00a0",
		"\u2028line separator\u2028",
		"inner  space kept",
		"",
		"0",
	}
	for _, v := range values {
		input := "2\n" + v + "\n"
		tags, err := collectTags(NewASCIITagger(strings.NewReader(input)))
		if err != nil {
			t.Fatalf("value %q: unexpected error: %v", v, err)
		}
		if len(tags) != 1 {
			t.Fatalf("value %q: got %d tags, want 1", v, len(tags))
		}
		if got, want := FormatValue(tags[0]), strings.TrimSpace(v); got != want {
			t.Errorf("value %q trimmed to %q, want %q", v, got, want)
		}
	}
}
