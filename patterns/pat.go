package patterns

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

// PATError reports one malformed line in a hatch pattern file. Parsing
// continues so that callers receive the patterns that were read before and
// after the bad line.
type PATError struct {
	File    string
	Line    int
	Message string
}

func (e PATError) Error() string {
	if e.File == "" {
		return fmt.Sprintf("line %d: %s", e.Line, e.Message)
	}
	return fmt.Sprintf("%s:%d: %s", e.File, e.Line, e.Message)
}

// PATFile contains the patterns and recoverable errors from one .pat file.
type PATFile struct {
	Patterns []PatternDefinition
	Errors   []PATError
}

// ParsePAT parses AutoCAD .pat content. Comment and blank lines are ignored,
// each *NAME line starts a pattern, and malformed pattern lines are reported
// in Errors while the remaining patterns still parse.
func ParsePAT(data []byte, filename string) *PATFile {
	return ParsePATReader(bytes.NewReader(data), filename)
}

// ParsePATReader parses .pat content from r.
func ParsePATReader(r io.Reader, filename string) *PATFile {
	out := &PATFile{}
	scanner := bufio.NewScanner(r)
	lineNo := 0
	var current *PatternDefinition
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\ufeff"))
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "*") {
			name, description := splitPATHeader(strings.TrimSpace(line[1:]))
			if name == "" {
				out.Errors = append(out.Errors, PATError{File: filename, Line: lineNo, Message: "empty pattern name"})
				current = nil
				continue
			}
			out.Patterns = append(out.Patterns, PatternDefinition{
				Name:        name,
				Description: description,
				Category:    PatternCategoryCustom,
			})
			current = &out.Patterns[len(out.Patterns)-1]
			continue
		}
		if current == nil {
			out.Errors = append(out.Errors, PATError{File: filename, Line: lineNo, Message: "pattern line before any *NAME header"})
			continue
		}
		pl, err := parsePATLine(line)
		if err != nil {
			// The first non-pattern line after *NAME,description is the
			// human-readable pattern description.
			if current.Description == "" && len(current.Lines) == 0 {
				current.Description = strings.TrimSpace(line)
				continue
			}
			out.Errors = append(out.Errors, PATError{File: filename, Line: lineNo, Message: err.Error()})
			continue
		}
		current.Lines = append(current.Lines, pl)
	}
	if err := scanner.Err(); err != nil {
		out.Errors = append(out.Errors, PATError{File: filename, Line: lineNo, Message: err.Error()})
	}
	return out
}

// Register adds every successfully parsed pattern to the custom lookup table.
func (p *PATFile) Register() error {
	if p == nil {
		return nil
	}
	for _, def := range p.Patterns {
		if err := RegisterPattern(def); err != nil {
			return err
		}
	}
	return nil
}

func splitPATHeader(s string) (string, string) {
	name, description, ok := strings.Cut(s, ",")
	if !ok {
		return strings.TrimSpace(s), ""
	}
	return strings.TrimSpace(name), strings.TrimSpace(description)
}

func parsePATLine(line string) (PatternLine, error) {
	if i := strings.IndexByte(line, ';'); i >= 0 {
		line = strings.TrimSpace(line[:i])
	}
	parts := strings.Split(line, ",")
	if len(parts) < 5 {
		return PatternLine{}, fmt.Errorf("pattern line needs angle, base point, and offset: %q", line)
	}
	values := make([]float64, len(parts))
	for i, part := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil {
			return PatternLine{}, fmt.Errorf("invalid number %q", strings.TrimSpace(part))
		}
		values[i] = v
	}
	return PatternLine{
		Angle:           values[0],
		BasePoint:       []float64{values[1], values[2]},
		Offset:          []float64{values[3], values[4]},
		DashLengthItems: append([]float64(nil), values[5:]...),
	}, nil
}

var (
	customMu       sync.RWMutex
	customPatterns = map[string]PatternDefinition{}
)

// RegisterPattern adds or replaces a custom pattern. Lookup remains
// case-insensitive and returns the name exactly as registered.
func RegisterPattern(def PatternDefinition) error {
	if strings.TrimSpace(def.Name) == "" {
		return fmt.Errorf("patterns: empty pattern name")
	}
	customMu.Lock()
	defer customMu.Unlock()
	customPatterns[strings.ToUpper(def.Name)] = clonePattern(def)
	return nil
}

// UnregisterPattern removes a custom pattern.
func UnregisterPattern(name string) {
	customMu.Lock()
	defer customMu.Unlock()
	delete(customPatterns, strings.ToUpper(name))
}

// LookupPattern finds a custom or predefined pattern without regard to case.
// The returned definition keeps its original (source) name.
func LookupPattern(name string) (PatternDefinition, bool) {
	customMu.RLock()
	def, ok := customPatterns[strings.ToUpper(name)]
	customMu.RUnlock()
	if ok {
		return clonePattern(def), true
	}
	for key, def := range PredefinedPatterns {
		if strings.EqualFold(key, name) {
			return clonePattern(def), true
		}
	}
	return PatternDefinition{}, false
}

// Scaled returns a copy of the definition with base points, offsets, and dash
// lengths multiplied by factor. The receiver is unchanged.
func (p PatternDefinition) Scaled(factor float64) PatternDefinition {
	out := clonePattern(p)
	if factor == 1 {
		return out
	}
	for i := range out.Lines {
		scaleSlice(out.Lines[i].BasePoint, factor)
		scaleSlice(out.Lines[i].Offset, factor)
		scaleSlice(out.Lines[i].DashLengthItems, factor)
	}
	return out
}

// ConvertMeasurement returns a copy scaled from one measurement system to
// another. The receiver is unchanged.
func (p PatternDefinition) ConvertMeasurement(from, to MeasurementSystem) PatternDefinition {
	factor := 1.0
	if to.ScaleFactor != 0 {
		factor = from.ScaleFactor / to.ScaleFactor
	}
	return p.Scaled(factor)
}

func clonePattern(p PatternDefinition) PatternDefinition {
	out := p
	out.Lines = make([]PatternLine, len(p.Lines))
	for i, line := range p.Lines {
		out.Lines[i] = PatternLine{
			Angle:           line.Angle,
			BasePoint:       append([]float64(nil), line.BasePoint...),
			Offset:          append([]float64(nil), line.Offset...),
			DashLengthItems: append([]float64(nil), line.DashLengthItems...),
		}
	}
	return out
}

func scaleSlice(v []float64, factor float64) {
	for i := range v {
		v[i] *= factor
	}
}
