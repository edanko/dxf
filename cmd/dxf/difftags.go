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

	"github.com/edanko/dxf/lldxf"
)

type tagDifference struct {
	Path string `json:"path"`
	Old  string `json:"old,omitempty"`
	New  string `json:"new,omitempty"`
}

type diffTagsReport struct {
	Added     []tagDifference `json:"added"`
	Removed   []tagDifference `json:"removed"`
	Changed   []tagDifference `json:"changed"`
	Different bool            `json:"different"`
}

func cmdDiffTags(args []string) error {
	var format *string
	fs, err := newFlagSet("difftags", args, func(fs *flag.FlagSet) {
		format = fs.String("format", "text", "output format: text or json")
	})
	if err != nil || fs == nil {
		return err
	}
	rest := fs.Args()
	if len(rest) < 2 {
		return usageError("usage: dxf difftags [options] <file-a> <file-b>")
	}

	first, err := readTagsFromFile(rest[0])
	if err != nil {
		return err
	}
	second, err := readTagsFromFile(rest[1])
	if err != nil {
		return err
	}
	left := tagPathValues(first)
	right := tagPathValues(second)

	report := diffTagsReport{
		Added:   []tagDifference{},
		Removed: []tagDifference{},
		Changed: []tagDifference{},
	}
	for path, value := range right {
		if old, ok := left[path]; !ok {
			report.Added = append(report.Added, tagDifference{Path: path, New: value})
		} else if old != value {
			report.Changed = append(report.Changed, tagDifference{Path: path, Old: old, New: value})
		}
	}
	for path, value := range left {
		if _, ok := right[path]; !ok {
			report.Removed = append(report.Removed, tagDifference{Path: path, Old: value})
		}
	}
	sortDifferences(report.Added)
	sortDifferences(report.Removed)
	sortDifferences(report.Changed)
	report.Different = len(report.Added)+len(report.Removed)+len(report.Changed) > 0

	if *format == "json" {
		b, _ := json.MarshalIndent(report, "", "  ")
		if err := emit("", append(b, '\n'), false); err != nil {
			return err
		}
	} else {
		var b bytes.Buffer
		for _, diff := range report.Added {
			fmt.Fprintf(&b, "added %s: %q\n", diff.Path, diff.New)
		}
		for _, diff := range report.Removed {
			fmt.Fprintf(&b, "removed %s: %q\n", diff.Path, diff.Old)
		}
		for _, diff := range report.Changed {
			fmt.Fprintf(&b, "changed %s: %q -> %q\n", diff.Path, diff.Old, diff.New)
		}
		if !report.Different {
			fmt.Fprintln(&b, "no differences")
		}
		if err := emit("", b.Bytes(), false); err != nil {
			return err
		}
	}

	if report.Different {
		return &exitError{code: 1, msg: "difftags: differences found"}
	}
	return nil
}

func readTagsFromFile(path string) ([]lldxf.Tag, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, dataError("cannot read %s: %v", path, err)
	}
	tagger := taggerForData(data)
	var tags []lldxf.Tag
	for tagger.Next() {
		if tag := tagger.Tag(); tag != nil {
			tags = append(tags, tag)
		}
	}
	if err := tagger.Err(); err != nil {
		return nil, dataError("cannot parse %s: %v", path, err)
	}
	return tags, nil
}

type tagSection struct {
	name string
	tags []lldxf.Tag
}

func splitTagSections(tags []lldxf.Tag) []tagSection {
	var sections []tagSection
	current := tagSection{name: "PREAMBLE"}
	awaitingName := false
	for _, tag := range tags {
		code := int(tag.Code())
		value := strings.TrimSpace(fmt.Sprint(tag.Value()))
		if code == 0 {
			switch strings.ToUpper(value) {
			case "SECTION":
				if len(current.tags) > 0 {
					sections = append(sections, current)
				}
				current = tagSection{name: "PREAMBLE"}
				awaitingName = true
			case "ENDSEC":
				sections = append(sections, current)
				current = tagSection{name: "PREAMBLE"}
				awaitingName = false
			}
		} else if code == 2 && awaitingName {
			current.name = strings.ToUpper(value)
			awaitingName = false
		}
		current.tags = append(current.tags, tag)
	}
	if len(current.tags) > 0 {
		sections = append(sections, current)
	}
	return sections
}

func tagPathValues(tags []lldxf.Tag) map[string]string {
	result := map[string]string{}
	for _, section := range splitTagSections(tags) {
		if section.name == "HEADER" {
			indexTagSection(section.name, section.tags, result, true)
			continue
		}
		indexTagSection(section.name, section.tags, result, false)
	}
	return result
}

func indexTagSection(section string, tags []lldxf.Tag, result map[string]string, header bool) {
	prefix := section
	occurrences := map[string]int{}
	currentHeader := ""
	for _, record := range splitRecords(tags) {
		if len(record) == 0 {
			continue
		}
		typeName := strings.ToUpper(strings.TrimSpace(fmt.Sprint(record[0].Value())))
		if header {
			currentHeader = ""
			headerCounts := map[string]int{}
			for _, tag := range record {
				code := int(tag.Code())
				value := fmt.Sprint(tag.Value())
				if code == 9 {
					currentHeader = strings.ToUpper(strings.TrimSpace(value))
					result["HEADER/"+currentHeader+"/9"] = value
					continue
				}
				if currentHeader == "" {
					currentHeader = "PREAMBLE"
				}
				key := "HEADER/" + currentHeader + "/" + strconv.Itoa(code)
				key += occurrenceSuffix(headerCounts, strconv.Itoa(code))
				result[key] = value
			}
			continue
		}

		switch typeName {
		case "BLOCK":
			if name := firstTagValue(record, 2); name != "" {
				prefix = "BLOCKS/" + strings.ToUpper(name)
			}
		case "TABLE":
			if name := firstTagValue(record, 2); name != "" {
				prefix = "TABLES/" + strings.ToUpper(name)
			}
		case "ENDBLK":
			prefix = "BLOCKS"
		case "ENDTAB":
			prefix = "TABLES"
		}

		recordKey := typeName
		if handle := firstTagValue(record, 5); handle != "" {
			recordKey += "[" + strings.ToUpper(handle) + "]"
		} else if handle := firstTagValue(record, 105); handle != "" {
			recordKey += "[" + strings.ToUpper(handle) + "]"
		} else {
			recordKey += "[" + strconv.Itoa(occurrences[typeName]) + "]"
		}
		occurrences[typeName]++
		recordPath := prefix + "/" + recordKey

		codeCounts := map[string]int{}
		for _, tag := range record {
			code := int(tag.Code())
			if code == 0 {
				continue
			}
			codeString := strconv.Itoa(code)
			key := recordPath + "/" + codeString + occurrenceSuffix(codeCounts, codeString)
			result[key] = fmt.Sprint(tag.Value())
		}
	}
}

func splitRecords(tags []lldxf.Tag) [][]lldxf.Tag {
	var records [][]lldxf.Tag
	var current []lldxf.Tag
	for _, tag := range tags {
		if int(tag.Code()) == 0 && len(current) > 0 {
			records = append(records, current)
			current = nil
		}
		current = append(current, tag)
	}
	if len(current) > 0 {
		records = append(records, current)
	}
	return records
}

func firstTagValue(tags []lldxf.Tag, code int) string {
	for _, tag := range tags {
		if int(tag.Code()) == code {
			return strings.TrimSpace(fmt.Sprint(tag.Value()))
		}
	}
	return ""
}

func occurrenceSuffix(counts map[string]int, key string) string {
	count := counts[key]
	counts[key]++
	if count == 0 {
		return ""
	}
	return "#" + strconv.Itoa(count)
}

func sortDifferences(diffs []tagDifference) {
	sort.Slice(diffs, func(i, j int) bool {
		return diffs[i].Path < diffs[j].Path
	})
}
