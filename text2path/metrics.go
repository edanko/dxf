package text2path

import (
	"fmt"

	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// MeasureTTF measures text using the system TrueType/OpenType face at the
// requested height. The returned metrics are in drawing units: Width is the
// summed advance, Ascent and Descent come from the face metrics.
func MeasureTTF(text string, height float64) (TextMetrics, string, error) {
	if height <= 0 {
		return TextMetrics{}, "", fmt.Errorf("text2path: text height must be positive")
	}
	f, name, err := loadFont()
	if err != nil {
		return TextMetrics{}, "", err
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: height, DPI: 72})
	if err != nil {
		return TextMetrics{}, name, err
	}
	m := face.Metrics()
	var pen fixed.Int26_6
	for _, r := range text {
		a, ok := face.GlyphAdvance(r)
		if !ok {
			a, _ = face.GlyphAdvance(' ')
		}
		pen += a
	}
	return TextMetrics{
		Width:   float64(pen) / 64,
		Height:  height,
		Ascent:  float64(m.Ascent) / 64,
		Descent: float64(m.Descent) / 64,
	}, name, nil
}

// TextMetricsForHeight returns metrics for the requested text using the
// supplied SHX font when present, otherwise the system TTF face.
func TextMetricsForHeight(shx *SHXFont, text string, height float64) (TextMetrics, string, error) {
	if shx != nil {
		return shx.Measure(text, height), shx.Name, nil
	}
	return MeasureTTF(text, height)
}
