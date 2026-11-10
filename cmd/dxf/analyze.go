package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	dxf "github.com/edanko/dxf"
	"github.com/edanko/dxf/lldxf"
)

type analyzeTagStats struct {
	Total    int            `json:"total"`
	Sections map[string]int `json:"sections"`
	Codes    map[string]int `json:"codes"`
}

type analyzeReport struct {
	Format   string          `json:"format"`
	Version  string          `json:"version"`
	Tags     analyzeTagStats `json:"tags"`
	Entities map[string]int  `json:"entities"`
	Tables   map[string]int  `json:"tables"`
	Blocks   map[string]int  `json:"blocks"`
	Objects  map[string]int  `json:"objects"`
}

func cmdAnalyze(args []string) error {
	var format *string
	fs, err := newFlagSet("analyze", args, func(fs *flag.FlagSet) {
		format = fs.String("format", "text", "output format: text or json")
	})
	if err != nil || fs == nil {
		return err
	}
	in, _, _, err := inputOutput(fs, false)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(in)
	if err != nil {
		return dataError("cannot read %s: %v", in, err)
	}
	d, err := loadFile(in)
	if err != nil {
		return err
	}
	tagStats, err := collectTagStats(data)
	if err != nil {
		return dataError("cannot analyze tags in %s: %v", in, err)
	}
	fileFormat, err := dxf.DetectFormat(data)
	if err != nil {
		return dataError("cannot detect format of %s: %v", in, err)
	}

	report := analyzeReport{
		Format:   fileFormat,
		Version:  d.Header().Version,
		Tags:     tagStats,
		Entities: map[string]int{},
		Tables:   map[string]int{},
		Blocks:   map[string]int{},
		Objects:  map[string]int{},
	}
	for _, e := range d.Entities() {
		report.Entities[e.DXFType()]++
	}
	for _, t := range d.Tables() {
		if t == nil {
			continue
		}
		report.Tables[t.Name()] = len(t.Entries())
	}
	for _, b := range d.Blocks() {
		if b == nil {
			continue
		}
		report.Blocks[b.Name] = len(b.Entities())
	}
	for _, o := range d.Objects() {
		report.Objects[objectTypeName(o)]++
	}

	if *format == "json" {
		b, _ := json.MarshalIndent(report, "", "  ")
		return emit("", append(b, '\n'), false)
	}

	var b bytes.Buffer
	fmt.Fprintf(&b, "format: %s\n", report.Format)
	fmt.Fprintf(&b, "version: %s\n", report.Version)
	fmt.Fprintf(&b, "tags: %d\n", report.Tags.Total)
	fmt.Fprintln(&b, "  sections:")
	for _, section := range sortedKeys(report.Tags.Sections) {
		fmt.Fprintf(&b, "    %s: %d\n", section, report.Tags.Sections[section])
	}
	fmt.Fprintln(&b, "  codes:")
	for _, code := range sortedKeys(report.Tags.Codes) {
		fmt.Fprintf(&b, "    %s: %d\n", code, report.Tags.Codes[code])
	}
	writeAnalyzeCounts(&b, "entities", report.Entities)
	writeAnalyzeCounts(&b, "tables", report.Tables)
	writeAnalyzeCounts(&b, "blocks", report.Blocks)
	writeAnalyzeCounts(&b, "objects", report.Objects)
	_, err = os.Stdout.Write(b.Bytes())
	return err
}

func writeAnalyzeCounts(b *bytes.Buffer, label string, counts map[string]int) {
	total := 0
	for _, count := range counts {
		total += count
	}
	fmt.Fprintf(b, "%s: %d\n", label, total)
	for _, name := range sortedKeys(counts) {
		fmt.Fprintf(b, "  %s: %d\n", name, counts[name])
	}
}

func collectTagStats(data []byte) (analyzeTagStats, error) {
	tagger := taggerForData(data)

	stats := analyzeTagStats{
		Sections: map[string]int{},
		Codes:    map[string]int{},
	}
	section := "PREAMBLE"
	awaitingSectionName := false
	for tagger.Next() {
		tag := tagger.Tag()
		if tag == nil {
			continue
		}
		stats.Total++
		code := int(tag.Code())
		stats.Codes[strconv.Itoa(code)]++

		value := strings.TrimSpace(fmt.Sprint(tag.Value()))
		if code == 0 {
			switch strings.ToUpper(value) {
			case "SECTION":
				awaitingSectionName = true
			case "ENDSEC":
				section = "PREAMBLE"
				awaitingSectionName = false
			}
		} else if code == 2 && awaitingSectionName {
			section = strings.ToUpper(value)
			awaitingSectionName = false
		}
		stats.Sections[section]++
	}
	if err := tagger.Err(); err != nil {
		return stats, err
	}
	return stats, nil
}

func objectTypeName(o interface{}) string {
	if named, ok := o.(interface{ WireType() string }); ok {
		if name := named.WireType(); name != "" {
			return name
		}
	}
	return strings.TrimPrefix(fmt.Sprintf("%T", o), "*object.")
}

func taggerForData(data []byte) lldxf.Tagger {
	if lldxf.IsBinaryDXF(data) {
		return lldxf.NewBinaryTaggerFromBytes(data)
	}
	return lldxf.NewASCIITagger(bytes.NewReader(data))
}
