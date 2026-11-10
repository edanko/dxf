// Transparency representation: an alpha value in 0..1.
package color

import "fmt"

// Transparency is an alpha transparency value (0 = fully transparent,
// 1 = opaque). DXF stores it as a 440 percentage.
type Transparency struct {
	Alpha float64
}

// NewTransparency validates and returns a Transparency.
func NewTransparency(alpha float64) (Transparency, error) {
	if alpha < 0 || alpha > 1 {
		return Transparency{}, fmt.Errorf("color: transparency alpha %v out of range 0..1", alpha)
	}
	return Transparency{Alpha: alpha}, nil
}

// Percent renders the alpha as the DXF 440 percent value.
func (t Transparency) Percent() int { return int(t.Alpha*100 + 0.5) }
