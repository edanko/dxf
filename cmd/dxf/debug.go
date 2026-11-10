package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/edanko/dxf"
	"github.com/edanko/dxf/drawing"
	"github.com/edanko/dxf/entity"
	"github.com/edanko/dxf/lldxf"
)

type debugSectionInfo struct {
	Name  string `json:"name"`
	Tags  int    `json:"tags"`
	Bytes int    `json:"bytes"`
}

type debugReport struct {
	File              string             `json:"file"`
	Format            string             `json:"format"`
	Encoding          string             `json:"encoding"`
	Version           string             `json:"version"`
	Size              int                `json:"size"`
	Sections          []debugSectionInfo `json:"sections"`
	Fallbacks         []string           `json:"loader_fallbacks"`
	RecoveryIssues    []string           `json:"recovery_issues"`
	RecoveredSections []string           `json:"recovered_sections"`
	SkippedEntities   int                `json:"skipped_entities"`
}

func cmdDebug(args []string) error {
	var format *string
	fs, err := newFlagSet("debug", args, func(fs *flag.FlagSet) {
		format = fs.String("format", "text", "output format: text or json")
	})
	if err != nil || fs == nil {
		return err
	}
	rest := fs.Args()
	if len(rest) < 1 {
		return usageError("usage: dxf debug [options] <file>")
	}

	path := rest[0]
	data, err := os.ReadFile(path)
	if err != nil {
		return dataError("cannot read %s: %v", path, err)
	}

	formatName, _ := dxf.DetectFormat(data)
	tags, tagErr := decodeTags(data)
	sections := splitTagSections(tags)
	version := headerVersion(tags)
	report := debugReport{
		File:              path,
		Format:            strings.ToLower(formatName),
		Encoding:          "ascii",
		Version:           version,
		Size:              len(data),
		Sections:          make([]debugSectionInfo, 0, len(sections)),
		Fallbacks:         []string{},
		RecoveryIssues:    []string{},
		RecoveredSections: []string{},
	}
	if lldxf.IsBinaryDXF(data) {
		report.Encoding = "binary"
	}

	for _, section := range sections {
		info := debugSectionInfo{Name: section.name, Tags: len(section.tags)}
		for _, tag := range section.tags {
			info.Bytes += encodedTagSize(tag)
		}
		report.Sections = append(report.Sections, info)
	}

	if tagErr != nil {
		report.RecoveryIssues = append(report.RecoveryIssues, tagErr.Error())
	}

	doc, loadErr := dxf.FromReaderWithOptions(bytes.NewReader(data), &dxf.LoadOptions{
		Recover:         true,
		PreserveUnknown: true,
	})
	if loadErr != nil {
		report.RecoveryIssues = append(report.RecoveryIssues, loadErr.Error())
	}
	if doc != nil {
		if v := doc.SourceVersion(); v != "" {
			report.Version = v
		}
		if recovery := doc.RecoveryReport(); recovery != nil {
			report.RecoveryIssues = append(report.RecoveryIssues, recovery.Issues...)
			report.RecoveredSections = append(report.RecoveredSections, recovery.RecoveredSections...)
			report.SkippedEntities = recovery.SkippedEntities
		}
		report.Fallbacks = append(report.Fallbacks, loaderFallbacks(doc)...)
	}
	sort.Strings(report.Fallbacks)
	sort.Strings(report.RecoveryIssues)
	sort.Strings(report.RecoveredSections)

	if *format == "json" {
		b, _ := json.MarshalIndent(report, "", "  ")
		if err := emit("", append(b, '\n'), false); err != nil {
			return err
		}
		return nil
	}

	var b bytes.Buffer
	fmt.Fprintf(&b, "file: %s\n", report.File)
	fmt.Fprintf(&b, "version: %s\n", report.Version)
	fmt.Fprintf(&b, "format: %s\n", report.Format)
	fmt.Fprintf(&b, "encoding: %s\n", report.Encoding)
	fmt.Fprintf(&b, "size: %d bytes\n", report.Size)
	fmt.Fprintln(&b, "sections:")
	for _, section := range report.Sections {
		fmt.Fprintf(&b, "  %s: %d tags, %d bytes\n", section.Name, section.Tags, section.Bytes)
	}
	if len(report.Fallbacks) == 0 {
		fmt.Fprintln(&b, "loader fallbacks: none")
	} else {
		fmt.Fprintln(&b, "loader fallbacks:")
		for _, fallback := range report.Fallbacks {
			fmt.Fprintf(&b, "  %s\n", fallback)
		}
	}
	if len(report.RecoveryIssues) == 0 {
		fmt.Fprintln(&b, "recovery issues: none")
	} else {
		fmt.Fprintln(&b, "recovery issues:")
		for _, issue := range report.RecoveryIssues {
			fmt.Fprintf(&b, "  %s\n", issue)
		}
	}
	if len(report.RecoveredSections) > 0 {
		fmt.Fprintf(&b, "recovered sections: %s\n", strings.Join(report.RecoveredSections, ", "))
	}
	if report.SkippedEntities > 0 {
		fmt.Fprintf(&b, "skipped entities: %d\n", report.SkippedEntities)
	}
	return emit("", b.Bytes(), false)
}

func decodeTags(data []byte) ([]lldxf.Tag, error) {
	tagger := taggerForData(data)
	var tags []lldxf.Tag
	for tagger.Next() {
		if tag := tagger.Tag(); tag != nil {
			tags = append(tags, tag)
		}
	}
	return tags, tagger.Err()
}

func headerVersion(tags []lldxf.Tag) string {
	for i, tag := range tags {
		if int(tag.Code()) != 9 || strings.TrimSpace(fmt.Sprint(tag.Value())) != "$ACADVER" {
			continue
		}
		for _, next := range tags[i+1:] {
			if int(next.Code()) == 1 {
				return strings.TrimSpace(fmt.Sprint(next.Value()))
			}
		}
	}
	return ""
}

func encodedTagSize(tag lldxf.Tag) int {
	return len(strconv.Itoa(int(tag.Code()))) + len(fmt.Sprint(tag.Value())) + 4
}

func loaderFallbacks(d *drawing.Drawing) []string {
	var fallbacks []string
	for _, e := range d.Entities() {
		if _, ok := e.(*entity.RetainedRecordType); ok {
			fallbacks = append(fallbacks, fmt.Sprintf("entity %s retained as raw record", e.DXFType()))
		}
	}
	return fallbacks
}
