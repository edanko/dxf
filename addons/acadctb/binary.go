package acadctb

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
)

// defaultLineweights is the standard AutoCAD plot lineweight table in mm.
var defaultLineweights = []float64{
	0.00, 0.05, 0.09, 0.10, 0.13, 0.15, 0.18, 0.20, 0.25, 0.30,
	0.35, 0.40, 0.45, 0.50, 0.53, 0.60, 0.65, 0.70, 0.80, 0.90,
	1.00, 1.06, 1.20, 1.40, 1.58, 2.00, 2.11,
}

const plotStyleHeaderLen = 60

// parseBinaryPlotStyleTable parses a compressed CTB or STB file. The binary
// container is a fixed header followed by a zlib-compressed textual body.
func parseBinaryPlotStyleTable(data []byte) (*PlotStyleTable, error) {
	if len(data) < plotStyleHeaderLen {
		return nil, fmt.Errorf("acadctb: truncated binary plot style table")
	}
	header := string(data[:plotStyleHeaderLen])
	byColor := !strings.Contains(header, "STBVER1")

	zr, err := zlib.NewReader(bytes.NewReader(data[plotStyleHeaderLen:]))
	if err != nil {
		return nil, fmt.Errorf("acadctb: invalid compressed plot style table: %w", err)
	}
	body, err := io.ReadAll(zr)
	closeErr := zr.Close()
	if err != nil {
		return nil, fmt.Errorf("acadctb: decompress plot style table: %w", err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("acadctb: decompress plot style table: %w", closeErr)
	}
	text := strings.TrimRight(string(body), "\x00\n\r")
	parsed := parsePlotStyleText(text)

	table := &PlotStyleTable{
		Name:        strings.TrimSpace(stringValue(parsed, "description")),
		Description: strings.TrimSpace(stringValue(parsed, "description")),
		ByColor:     byColor,
	}
	lineweights := parseLineweightTable(parsed)

	styles, _ := parsed["plot_style"].(map[string]any)
	keys := make([]string, 0, len(styles))
	for key := range styles {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, _ := strconv.Atoi(keys[i])
		b, _ := strconv.Atoi(keys[j])
		return a < b
	})
	for _, key := range keys {
		record, ok := styles[key].(map[string]any)
		if !ok {
			continue
		}
		index, err := strconv.Atoi(strings.TrimSpace(key))
		if err != nil {
			continue
		}
		table.Styles = append(table.Styles, parseCTBStyle(index, record, lineweights, byColor))
	}
	if len(table.Styles) == 0 {
		return nil, fmt.Errorf("acadctb: no plot styles found in binary table")
	}
	return table, nil
}

func parseCTBStyle(index int, record map[string]any, lineweights map[int]float64, byColor bool) PlotStyle {
	style := PlotStyle{
		Name:          strings.TrimSpace(stringValue(record, "name")),
		Index:         index,
		Description:   strings.TrimSpace(stringValue(record, "description")),
		Color:         intValue(record, "color", 0),
		Screening:     intValue(record, "screen", 100),
		ScreeningSet:  true,
		FillStyle:     intValue(record, "fill_style", 0),
		EndStyle:      intValue(record, "end_style", 0),
		LinetypeScale: floatValue(record, "linepattern_size", 0),
	}
	if style.Name == "" {
		style.Name = fmt.Sprintf("Color_%d", index+1)
	}
	if byColor {
		// CTB style index 0 is ACI 1; keep the numeric alias used by callers
		// that build color tables from text.
		if style.Name == fmt.Sprintf("Color_%d", index+1) {
			style.Name = strconv.Itoa(index + 1)
		}
	}

	lineweightIndex := intValue(record, "lineweight", 0)
	if mm, ok := lineweights[lineweightIndex]; ok && mm > 0 {
		style.Lineweight = int(math.Round(mm * 100))
	}

	if mode, ok := record["mode_color"]; ok {
		if raw, err := strconv.ParseInt(strings.TrimSpace(asString(mode)), 10, 64); err == nil {
			if aci, rgb, hasRGB, valid := decodeModeColor(raw); valid {
				if hasRGB {
					style.TrueColor = rgb
					style.HasTrueColor = true
				} else if aci > 0 {
					style.Color = aci
				}
			}
		}
	}
	if style.Color < 0 {
		style.Color = 0
	}
	if linetype := intValue(record, "linetype", 31); linetype != 31 {
		style.LineType = strconv.Itoa(linetype)
	}
	return style
}

// decodeModeColor decodes an AutoCAD packed mode_color value.
func decodeModeColor(value int64) (aci int, rgb uint32, hasRGB bool, valid bool) {
	u := uint32(value)
	kind := u >> 24
	switch kind {
	case 0xC0, 0xC1: // ByLayer / ByBlock
		return 0, 0, false, true
	case 0xC2: // RGB
		return 0, ((u >> 16) & 0xFF << 16) | ((u >> 8) & 0xFF << 8) | (u & 0xFF), true, true
	case 0xC3: // ACI
		return int(u & 0xFF), 0, false, true
	default:
		return 0, 0, false, false
	}
}

func parseLineweightTable(parsed map[string]any) map[int]float64 {
	result := make(map[int]float64, len(defaultLineweights))
	for i, mm := range defaultLineweights {
		result[i] = mm
	}
	table, _ := parsed["custom_lineweight_table"].(map[string]any)
	for key, value := range table {
		index, err := strconv.Atoi(strings.TrimSpace(key))
		if err != nil {
			continue
		}
		if mm, err := strconv.ParseFloat(strings.TrimSpace(asString(value)), 64); err == nil {
			result[index] = mm
		}
	}
	return result
}

// parsePlotStyleText parses the nested INI-like body used by CTB and STB
// files. Lists end at a line containing only '}'.
func parsePlotStyleText(text string) map[string]any {
	parser := &plotStyleParser{lines: strings.Split(text, "\n")}
	return parser.parseMapping(false)
}

type plotStyleParser struct {
	lines []string
	pos   int
}

func (p *plotStyleParser) parseMapping(nested bool) map[string]any {
	out := make(map[string]any)
	for p.pos < len(p.lines) {
		line := strings.TrimSpace(p.lines[p.pos])
		if line == "" {
			p.pos++
			continue
		}
		if line == "}" {
			p.pos++
			return out
		}
		p.pos++
		if strings.HasSuffix(line, "{") {
			name := strings.TrimSpace(strings.TrimSuffix(line, "{"))
			out[name] = p.parseMapping(true)
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		out[strings.TrimSpace(key)] = sanitizePlotStyleValue(value)
	}
	_ = nested
	return out
}

func sanitizePlotStyleValue(value string) string {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "\"") {
		return strings.TrimPrefix(value, "\"")
	}
	// AutoCAD may append binary trailer data after the numeric value.
	if strings.HasSuffix(value, ")") {
		if fields := strings.Fields(value); len(fields) > 0 {
			return fields[0]
		}
	}
	return value
}

func stringValue(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func intValue(m map[string]any, key string, fallback int) int {
	raw := strings.TrimSpace(stringValue(m, key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		if parsed, err := strconv.ParseFloat(raw, 64); err == nil {
			return int(parsed)
		}
		return fallback
	}
	return value
}

func floatValue(m map[string]any, key string, fallback float64) float64 {
	raw := strings.TrimSpace(stringValue(m, key))
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fallback
	}
	return value
}
