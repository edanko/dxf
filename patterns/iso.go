package patterns

import "maps"

// ISO hatch patterns use the ACAD_ISOxxW100 naming convention and metric
// pattern units. They are registered into PredefinedPatterns at init time.
var isoPatterns = map[string]PatternDefinition{
	"ACAD_ISO02W100": isoPattern("ACAD_ISO02W100", "ISO dashed", []PatternLine{
		{Angle: 0, BasePoint: []float64{0, 0}, Offset: []float64{0, 5}, DashLengthItems: []float64{12, -3}},
	}),
	"ACAD_ISO03W100": isoPattern("ACAD_ISO03W100", "ISO dashed wide gap", []PatternLine{
		{Angle: 0, BasePoint: []float64{0, 0}, Offset: []float64{0, 5}, DashLengthItems: []float64{12, -18}},
	}),
	"ACAD_ISO04W100": isoPattern("ACAD_ISO04W100", "ISO dash dot", []PatternLine{
		{Angle: 0, BasePoint: []float64{0, 0}, Offset: []float64{0, 5}, DashLengthItems: []float64{24, -3, 0.5, -3}},
	}),
	"ACAD_ISO05W100": isoPattern("ACAD_ISO05W100", "ISO dash double dot", []PatternLine{
		{Angle: 0, BasePoint: []float64{0, 0}, Offset: []float64{0, 5}, DashLengthItems: []float64{24, -3, 0.5, -3, 0.5, -3}},
	}),
	"ACAD_ISO06W100": isoPattern("ACAD_ISO06W100", "ISO dash triple dot", []PatternLine{
		{Angle: 0, BasePoint: []float64{0, 0}, Offset: []float64{0, 5}, DashLengthItems: []float64{24, -3, 0.5, -3, 0.5, -6.5}},
		{Angle: 0, BasePoint: []float64{0, 0}, Offset: []float64{0, 5}, DashLengthItems: []float64{-34, 0.5, -3}},
	}),
	"ACAD_ISO07W100": isoPattern("ACAD_ISO07W100", "ISO dotted", []PatternLine{
		{Angle: 0, BasePoint: []float64{0, 0}, Offset: []float64{0, 5}, DashLengthItems: []float64{0.5, -3}},
	}),
	"ACAD_ISO08W100": isoPattern("ACAD_ISO08W100", "ISO long dash short dash", []PatternLine{
		{Angle: 0, BasePoint: []float64{0, 0}, Offset: []float64{0, 5}, DashLengthItems: []float64{24, -3, 6, -3}},
	}),
	"ACAD_ISO09W100": isoPattern("ACAD_ISO09W100", "ISO long dash double short dash", []PatternLine{
		{Angle: 0, BasePoint: []float64{0, 0}, Offset: []float64{0, 5}, DashLengthItems: []float64{24, -3, 6, -3, 6, -3}},
	}),
	"ACAD_ISO10W100": isoPattern("ACAD_ISO10W100", "ISO short dash dot", []PatternLine{
		{Angle: 0, BasePoint: []float64{0, 0}, Offset: []float64{0, 5}, DashLengthItems: []float64{12, -3, 0.5, -3}},
	}),
	"ACAD_ISO11W100": isoPattern("ACAD_ISO11W100", "ISO double dash dot", []PatternLine{
		{Angle: 0, BasePoint: []float64{0, 0}, Offset: []float64{0, 5}, DashLengthItems: []float64{12, -3, 12, -3, 0.5, -3}},
	}),
	"ACAD_ISO12W100": isoPattern("ACAD_ISO12W100", "ISO dash double dot", []PatternLine{
		{Angle: 0, BasePoint: []float64{0, 0}, Offset: []float64{0, 5}, DashLengthItems: []float64{12, -3, 0.5, -3, 0.5, -3}},
	}),
	"ACAD_ISO13W100": isoPattern("ACAD_ISO13W100", "ISO double dash double dot", []PatternLine{
		{Angle: 0, BasePoint: []float64{0, 0}, Offset: []float64{0, 5}, DashLengthItems: []float64{12, -3, 12, -3, 0.5, -6.5}},
		{Angle: 0, BasePoint: []float64{0, 0}, Offset: []float64{0, 5}, DashLengthItems: []float64{-33.5, 0.5, -3}},
	}),
	"ACAD_ISO14W100": isoPattern("ACAD_ISO14W100", "ISO dash dot dot", []PatternLine{
		{Angle: 0, BasePoint: []float64{0, 0}, Offset: []float64{0, 5}, DashLengthItems: []float64{12, -3, 0.5, -3, 0.5, -6.5}},
		{Angle: 0, BasePoint: []float64{0, 0}, Offset: []float64{0, 5}, DashLengthItems: []float64{-22, 0.5, -3}},
	}),
	"ACAD_ISO15W100": isoPattern("ACAD_ISO15W100", "ISO double dash dot dot", []PatternLine{
		{Angle: 0, BasePoint: []float64{0, 0}, Offset: []float64{0, 5}, DashLengthItems: []float64{12, -3, 12, -3, 0.5, -10}},
		{Angle: 0, BasePoint: []float64{0, 0}, Offset: []float64{0, 5}, DashLengthItems: []float64{-33.5, 0.5, -3, 0.5, -3}},
	}),
}

func isoPattern(name, description string, lines []PatternLine) PatternDefinition {
	return PatternDefinition{
		Name:        name,
		Description: description,
		Category:    PatternCategoryISO,
		Lines:       lines,
	}
}

func init() {
	maps.Copy(PredefinedPatterns, isoPatterns)
}
