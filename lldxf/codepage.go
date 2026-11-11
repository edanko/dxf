package lldxf

import (
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// CodepageName maps a DXF $DWGCODEPAGE value (e.g. "ANSI_1251") to a Go
// encoding name. Unknown names map to cp1252, the AutoCAD default.
func CodepageName(codepage string) string {
	n := strings.ToUpper(strings.TrimSpace(codepage))
	n = strings.TrimPrefix(n, "ANSI_")
	switch n {
	case "", "ANSI":
		return "cp1252"
	case "UTF8", "UTF-8":
		return "utf8"
	default:
		return "cp" + n
	}
}

// DecodeCodepageBytes decodes raw bytes written in the given DXF codepage
// into a Go string. Unmappable byte sequences use U+FFFD.
func DecodeCodepageBytes(data []byte, codepage string) string {
	name := CodepageName(codepage)
	if name == "utf8" {
		return string(data)
	}
	cm, ok := charmapFor(name)
	if !ok {
		return string(data)
	}
	out, _, err := transform.Bytes(cm.NewDecoder(), data)
	if err != nil {
		return string(data)
	}
	return string(out)
}

// EncodeCodepageBytes encodes a Go string into the given DXF codepage.
// Unmappable runes become '?'.
func EncodeCodepageBytes(val string, codepage string) []byte {
	name := CodepageName(codepage)
	if name == "utf8" {
		return []byte(val)
	}
	cm, ok := charmapFor(name)
	if !ok {
		return []byte(val)
	}
	out, _, err := transform.Bytes(cm.NewEncoder(), []byte(val))
	if err != nil {
		return []byte(val)
	}
	return out
}

func charmapFor(name string) (*charmap.Charmap, bool) {
	switch name {
	case "cp1250":
		return charmap.Windows1250, true
	case "cp1251":
		return charmap.Windows1251, true
	case "cp1252":
		return charmap.Windows1252, true
	case "cp1253":
		return charmap.Windows1253, true
	case "cp1254":
		return charmap.Windows1254, true
	case "cp1257":
		return charmap.Windows1257, true
	case "cp850":
		return charmap.CodePage850, true
	case "cp852":
		return charmap.CodePage852, true
	case "cp855":
		return charmap.CodePage855, true
	case "cp860":
		return charmap.CodePage860, true
	case "cp862":
		return charmap.CodePage862, true
	case "cp866":
		return charmap.CodePage866, true
	case "cp10000":
		return charmap.Macintosh, true
	case "koi8r":
		return charmap.KOI8R, true
	case "koi8u":
		return charmap.KOI8U, true
	}
	return nil, false
}
