package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/edanko/dxf/standards"
)

func cmdStandards(args []string) error {
	var format *string
	var layerPattern *string
	var severityName *string
	var disable *string
	fs, err := newFlagSet("standards", args, func(fs *flag.FlagSet) {
		format = fs.String("format", "text", "output format: text or json")
		layerPattern = fs.String("layer-pattern", "^[^<>/\\\\\":;?*|,=]{1,255}$", "regular expression for valid layer names")
		severityName = fs.String("severity", "error", "severity for layer-name violations: info, warning, error or critical")
		disable = fs.String("disable", "", "comma-separated rule codes to disable")
	})
	if err != nil || fs == nil {
		return err
	}
	if fs.NArg() < 1 {
		return usageError("usage: dxf standards [options] <file>")
	}

	severity, err := parseStandardSeverity(*severityName)
	if err != nil {
		return err
	}
	rules := standards.NewRuleSet(standards.NewLayerNameRule(*layerPattern, severity))
	if *disable != "" {
		for code := range strings.SplitSeq(*disable, ",") {
			code = strings.TrimSpace(code)
			if code == "" {
				continue
			}
			if err := rules.Disable(code); err != nil {
				return usageError("%v", err)
			}
		}
	}

	doc, err := loadFile(fs.Arg(0))
	if err != nil {
		return err
	}
	report := rules.Run(doc)

	if *format == "json" {
		data, _ := json.MarshalIndent(report, "", "  ")
		if err := emit("", append(data, '\n'), false); err != nil {
			return err
		}
	} else {
		var b bytes.Buffer
		if len(report.Violations) == 0 {
			fmt.Fprintln(&b, "no standards violations")
		} else {
			for _, violation := range report.Violations {
				location := violation.Path
				if violation.Handle != "" {
					location += " (handle " + violation.Handle + ")"
				}
				fmt.Fprintf(&b, "%s %s %s: %s\n", violation.Severity, violation.Code, location, violation.Message)
			}
			fmt.Fprintf(&b, "%d standards violation(s)\n", len(report.Violations))
		}
		if err := emit("", b.Bytes(), false); err != nil {
			return err
		}
	}

	if report.HasViolations() {
		return &exitError{code: 1, msg: "standards: violations found"}
	}
	return nil
}

func parseStandardSeverity(value string) (standards.Severity, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(standards.SeverityInfo):
		return standards.SeverityInfo, nil
	case string(standards.SeverityWarning):
		return standards.SeverityWarning, nil
	case string(standards.SeverityError):
		return standards.SeverityError, nil
	case string(standards.SeverityCritical):
		return standards.SeverityCritical, nil
	default:
		return "", usageError("invalid severity %q", value)
	}
}
