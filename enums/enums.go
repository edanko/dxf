// Typed DXF enumerations with name/numeric conversion and unknown retention.
package enums

import "fmt"

// Set is a bidirectional name<->number enumeration. Values that are not
// registered stay representable as plain ints and are reported as unknown.
type Set struct {
	names   map[int]string
	numbers map[string]int
}

// NewSet creates an empty enumeration set.
func NewSet() *Set {
	return &Set{names: map[int]string{}, numbers: map[string]int{}}
}

// Register adds a name/number pair.
func (s *Set) Register(n int, name string) {
	s.names[n] = name
	s.numbers[name] = n
}

// Name returns the name of a numeric value.
func (s *Set) Name(n int) (string, bool) {
	name, ok := s.names[n]
	return name, ok
}

// Number returns the numeric value of a name.
func (s *Set) Number(name string) (int, bool) {
	n, ok := s.numbers[name]
	return n, ok
}

// Valid reports whether n is a registered value.
func (s *Set) Valid(n int) bool {
	_, ok := s.names[n]
	return ok
}

// Unknown returns a diagnostic for an unregistered value, or "" when valid.
func (s *Set) Unknown(n int) string {
	if s.Valid(n) {
		return ""
	}
	return fmt.Sprintf("unknown enum value %d", n)
}

// TextAlign enumerates horizontal text justification (DXF 72/73).
var TextAlign = func() *Set {
	s := NewSet()
	s.Register(1, "left")
	s.Register(2, "center")
	s.Register(3, "right")
	s.Register(4, "aligned")
	s.Register(5, "middle")
	s.Register(6, "fit")
	return s
}()

// MTextAttachment enumerates MTEXT attachment corners (DXF 71).
var MTextAttachment = func() *Set {
	s := NewSet()
	s.Register(1, "top-left")
	s.Register(2, "top-center")
	s.Register(3, "top-right")
	s.Register(4, "middle-left")
	s.Register(5, "middle-center")
	s.Register(6, "middle-right")
	s.Register(7, "bottom-left")
	s.Register(8, "bottom-center")
	s.Register(9, "bottom-right")
	return s
}()

// HatchPatternType enumerates hatch pattern sources (DXF 75).
var HatchPatternType = func() *Set {
	s := NewSet()
	s.Register(0, "predefined")
	s.Register(1, "user")
	s.Register(2, "custom")
	return s
}()

// DimensionType enumerates dimension types (DXF 70 low bits).
var DimensionType = func() *Set {
	s := NewSet()
	s.Register(0, "rotated")
	s.Register(1, "aligned")
	s.Register(2, "angular")
	s.Register(3, "diameter")
	s.Register(4, "radius")
	s.Register(5, "angular-3-point")
	s.Register(6, "ordinate")
	return s
}()

// Documented drawing defaults (DXF reference values).
const (
	DefaultLineWeight = -3 // ByLayer
	DefaultLinetype   = "ByLayer"
	DefaultTextStyle  = "Standard"
)

// ValidateLineWeight reports whether w is a valid DXF lineweight (-3..211).
func ValidateLineWeight(w int) bool { return w >= -3 && w <= 211 }

// ValidateLinetypeName reports whether name is a usable linetype name.
func ValidateLinetypeName(name string) bool {
	return len(name) > 0 && len(name) <= 255
}
