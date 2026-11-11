// Structured audit report types (parity: ezdxf audited/audit).
package audit

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// MaxIssues bounds the number of issues a single report collects.
// Once reached, further issues are dropped and Truncated is set.
const MaxIssues = 1000

// Severity classifies how serious an audit issue is. It aliases the
// package-wide ErrorSeverity level constants (SeverityInfo ... SeverityCritical).
type Severity = ErrorSeverity

// Issue is a single audit finding.
type Issue struct {
	Severity Severity `json:"severity"`
	Code     string   `json:"code"`
	Handle   string   `json:"handle,omitempty"`
	Message  string   `json:"message"`
	Fixed    bool     `json:"fixed,omitempty"`
}

// String renders the issue as "SEVERITY CODE [handle]: message".
func (i Issue) String() string {
	if i.Handle != "" {
		return fmt.Sprintf("%s %s [%s]: %s", i.Severity, i.Code, i.Handle, i.Message)
	}
	return fmt.Sprintf("%s %s: %s", i.Severity, i.Code, i.Message)
}

// Report is a structured, serializable collection of audit issues.
type Report struct {
	Issues    []Issue `json:"issues"`
	Truncated bool    `json:"truncated,omitempty"`

	dropped int
}

// Add appends an issue, enforcing the MaxIssues bound.
func (r *Report) Add(issue Issue) {
	if len(r.Issues) >= MaxIssues {
		r.Truncated = true
		r.dropped++
		return
	}
	r.Issues = append(r.Issues, issue)
}

// Addf appends an issue with a formatted message.
func (r *Report) Addf(sev Severity, code, handle, format string, args ...any) {
	r.Add(Issue{Severity: sev, Code: code, Handle: handle, Message: fmt.Sprintf(format, args...)})
}

// Count returns the number of issues at severity s.
func (r *Report) Count(s Severity) int {
	n := 0
	for _, it := range r.Issues {
		if it.Severity == s {
			n++
		}
	}
	return n
}

// BySeverity returns issue counts keyed by severity name.
func (r *Report) BySeverity() map[string]int {
	m := map[string]int{"info": 0, "warning": 0, "error": 0, "critical": 0}
	for _, it := range r.Issues {
		m[strings.ToLower(it.Severity.String())]++
	}
	return m
}

// Codes returns issue counts keyed by issue code.
func (r *Report) Codes() map[string]int {
	m := map[string]int{}
	for _, it := range r.Issues {
		m[it.Code]++
	}
	return m
}

// HasErrors reports whether the report contains any issue of
// SeverityError or higher.
func (r *Report) HasErrors() bool {
	for _, it := range r.Issues {
		if it.Severity >= SeverityError {
			return true
		}
	}
	return false
}

// FixedCount returns the number of issues marked as fixed.
func (r *Report) FixedCount() int {
	n := 0
	for _, it := range r.Issues {
		if it.Fixed {
			n++
		}
	}
	return n
}

// JSON serializes the report (counts plus issues).
func (r *Report) JSON() (string, error) {
	payload := struct {
		Counts    map[string]int `json:"counts"`
		Truncated bool           `json:"truncated,omitempty"`
		Issues    []Issue        `json:"issues"`
	}{Counts: r.BySeverity(), Truncated: r.Truncated, Issues: r.Issues}
	b, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// String renders a compact human-readable summary.
func (r *Report) String() string {
	counts := r.BySeverity()
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", k, counts[k]))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "audit report: %d issues (%s)", len(r.Issues), strings.Join(parts, ", "))
	if r.Truncated {
		fmt.Fprintf(&b, " (truncated, %d dropped)", r.dropped)
	}
	return b.String()
}
