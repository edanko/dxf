// Tagged color value: ACI index, true color, BYLAYER or BYBLOCK.
package color

import "fmt"

// Kind identifies how a Value selects a color.
type Kind int

const (
	KindACI Kind = iota
	KindTrueColor
	KindByLayer
	KindByBlock
)

// String returns the kind name.
func (k Kind) String() string {
	switch k {
	case KindTrueColor:
		return "truecolor"
	case KindByLayer:
		return "bylayer"
	case KindByBlock:
		return "byblock"
	default:
		return "aci"
	}
}

// Value is a tagged DXF color selection.
type Value struct {
	Kind Kind
	ACI  ColorNumber
	RGB  uint32 // 0x00RRGGBB
}

// NewACI returns an ACI-indexed color value.
func NewACI(index ColorNumber) Value { return Value{Kind: KindACI, ACI: index} }

// NewTrueColor returns a 24-bit RGB color value.
func NewTrueColor(rgb uint32) Value { return Value{Kind: KindTrueColor, RGB: rgb & 0xffffff} }

// NewByLayer returns the BYLAYER value.
func NewByLayer() Value { return Value{Kind: KindByLayer, ACI: ByLayer} }

// NewByBlock returns the BYBLOCK value.
func NewByBlock() Value { return Value{Kind: KindByBlock, ACI: 0} }

// KindOf reports the value kind.
func (v Value) KindOf() Kind { return v.Kind }

// IsByLayer reports whether the value is BYLAYER.
func (v Value) IsByLayer() bool { return v.Kind == KindByLayer }

// IsByBlock reports whether the value is BYBLOCK.
func (v Value) IsByBlock() bool { return v.Kind == KindByBlock }

// RGB8 returns the concrete RGB bytes for ACI and true-color values.
func (v Value) RGB8() (r, g, b uint8, ok bool) {
	switch v.Kind {
	case KindTrueColor:
		return uint8(v.RGB >> 16), uint8(v.RGB >> 8), uint8(v.RGB), true
	case KindACI:
		rgb := IndexColor(uint8(v.ACI))
		if len(rgb) < 3 {
			return 0, 0, 0, false
		}
		return rgb[0], rgb[1], rgb[2], true
	default:
		return 0, 0, 0, false
	}
}

// NearestIndex returns the ACI index of the closest documented color to the
// given RGB bytes.
func NearestIndex(r, g, b uint8) ColorNumber {
	return ColorIndex([]int{int(r), int(g), int(b)})
}

// ValidateACI reports whether n is a usable ACI index: 1..255, plus the
// documented 0/BYLAYER sentinels.
func ValidateACI(n ColorNumber) error {
	if n >= 1 && n <= 255 {
		return nil
	}
	if n == 0 || n == ByLayer {
		return nil
	}
	return fmt.Errorf("color: invalid ACI index %d (valid range 1..255)", n)
}
