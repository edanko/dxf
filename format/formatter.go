// DXF Output Formatter
package format

import (
	"io"
)

// Formatter controls output format.
type Formatter interface {
	Reset()
	WriteTo(w io.Writer) (int64, error)
	SetPrecision(p int)
	Output() string
	String(num int, val string) string
	Int(num int, val int) string
	Float(num int, val float64) string
	WriteString(num int, val string)
	WriteInt(num int, val int)
	WriteFloat(num int, val float64)
}
