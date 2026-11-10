// Package standards provides a reusable rule engine for drawing standards
// validation.
package standards

import (
	"fmt"
	"sort"

	"github.com/edanko/dxf/drawing"
)

// Severity classifies a standards violation.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// Violation is one standards finding.
type Violation struct {
	Code     string   `json:"code"`
	Severity Severity `json:"severity"`
	Path     string   `json:"path"`
	Handle   string   `json:"handle,omitempty"`
	Message  string   `json:"message"`
}

// Rule is a single standards predicate.
type Rule interface {
	// Code is the stable identifier used in reports and configuration.
	Code() string
	// Description explains what the rule checks.
	Description() string
	// Severity is the default severity for violations produced by the rule.
	Severity() Severity
	// Check evaluates the rule against a drawing.
	Check(*drawing.Drawing) []Violation
}

// RuleSet runs an ordered collection of rules with optional severity and
// enabled/disabled overrides.
type RuleSet struct {
	rules      []Rule
	severities map[string]Severity
	disabled   map[string]bool
}

// NewRuleSet creates a rule set from the given rules.
func NewRuleSet(rules ...Rule) *RuleSet {
	set := &RuleSet{
		severities: map[string]Severity{},
		disabled:   map[string]bool{},
	}
	for _, rule := range rules {
		set.Add(rule)
	}
	return set
}

// DefaultRuleSet returns the built-in rules used by the CLI.
func DefaultRuleSet() *RuleSet {
	return NewRuleSet(NewLayerNameRule(defaultLayerNamePattern, SeverityError))
}

// Add appends a rule. A nil rule is ignored.
func (s *RuleSet) Add(rule Rule) {
	if rule != nil {
		s.rules = append(s.rules, rule)
	}
}

// Rules returns the configured rules in order.
func (s *RuleSet) Rules() []Rule {
	return append([]Rule(nil), s.rules...)
}

// SetSeverity overrides the severity emitted for the named rule.
func (s *RuleSet) SetSeverity(code string, severity Severity) error {
	if !validSeverity(severity) {
		return fmt.Errorf("standards: invalid severity %q", severity)
	}
	if !s.hasRule(code) {
		return fmt.Errorf("standards: unknown rule %q", code)
	}
	s.severities[code] = severity
	return nil
}

// Disable disables the named rule for future runs.
func (s *RuleSet) Disable(code string) error {
	if !s.hasRule(code) {
		return fmt.Errorf("standards: unknown rule %q", code)
	}
	s.disabled[code] = true
	return nil
}

// Enable re-enables a previously disabled rule.
func (s *RuleSet) Enable(code string) error {
	if !s.hasRule(code) {
		return fmt.Errorf("standards: unknown rule %q", code)
	}
	delete(s.disabled, code)
	return nil
}

func (s *RuleSet) hasRule(code string) bool {
	for _, rule := range s.rules {
		if rule.Code() == code {
			return true
		}
	}
	return false
}

// Report is a structured standards validation result.
type Report struct {
	Violations []Violation `json:"violations"`
}

// HasViolations reports whether the report contains any finding.
func (r Report) HasViolations() bool {
	return len(r.Violations) > 0
}

// Count returns the number of violations at severity.
func (r Report) Count(severity Severity) int {
	count := 0
	for _, violation := range r.Violations {
		if violation.Severity == severity {
			count++
		}
	}
	return count
}

// Run evaluates every enabled rule and returns a deterministic report.
func (s *RuleSet) Run(d *drawing.Drawing) Report {
	report := Report{Violations: []Violation{}}
	if d == nil {
		return report
	}
	for _, rule := range s.rules {
		if s.disabled[rule.Code()] {
			continue
		}
		severity := rule.Severity()
		if override, ok := s.severities[rule.Code()]; ok {
			severity = override
		}
		for _, violation := range rule.Check(d) {
			if violation.Code == "" {
				violation.Code = rule.Code()
			}
			if override, ok := s.severities[rule.Code()]; ok {
				violation.Severity = override
			} else if violation.Severity == "" {
				violation.Severity = severity
			}
			if !validSeverity(violation.Severity) {
				violation.Severity = severity
			}
			report.Violations = append(report.Violations, violation)
		}
	}
	sort.SliceStable(report.Violations, func(i, j int) bool {
		left, right := report.Violations[i], report.Violations[j]
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.Code != right.Code {
			return left.Code < right.Code
		}
		return left.Handle < right.Handle
	})
	return report
}

func validSeverity(severity Severity) bool {
	switch severity {
	case SeverityInfo, SeverityWarning, SeverityError, SeverityCritical:
		return true
	default:
		return false
	}
}
