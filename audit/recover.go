// Recover: tolerant loading merged with the audit report (parity: ezdxf
// recover).
package audit

import (
	"io"
	"os"
	"strings"

	"github.com/edanko/dxf"
	"github.com/edanko/dxf/drawing"
)

// Issue codes contributed by the loader recovery pass.
const (
	CodeLoadIssue        = "LOAD_ISSUE"
	CodeTruncatedSection = "TRUNCATED_SECTION"
	CodeSkippedEntity    = "SKIPPED_ENTITY"
)

// RecoverReader loads a DXF file tolerously (loader Recover mode), merges
// the loader's recovery notes into a structured report and runs the audit
// checks over the recovered document. It never returns a non-nil document
// together with a fatal load error; on fatal errors the report explains why.
func RecoverReader(r io.Reader, opts *Options) (*drawing.Drawing, *Report, error) {
	rep := &Report{}
	doc, loadErr := dxf.FromReaderWithOptions(r, &dxf.LoadOptions{Recover: true})
	if doc == nil {
		msg := "file could not be loaded"
		if loadErr != nil {
			msg = loadErr.Error()
		}
		rep.Addf(SeverityCritical, CodeLoadIssue, "", "%s", msg)
		return nil, rep, loadErr
	}
	if rr := doc.RecoveryReport(); rr != nil {
		for _, issue := range rr.Issues {
			code := CodeLoadIssue
			upper := strings.ToUpper(issue)
			if strings.Contains(upper, "ENDSEC") || strings.Contains(upper, "EOF") ||
				strings.Contains(upper, "SECTION") {
				code = CodeTruncatedSection
			}
			rep.Addf(SeverityWarning, code, "", "%s", issue)
		}
		for _, section := range rr.RecoveredSections {
			rep.Addf(SeverityWarning, CodeTruncatedSection, "",
				"truncated section %s closed and document retained", section)
		}
		if rr.SkippedEntities > 0 {
			rep.Addf(SeverityWarning, CodeSkippedEntity, "",
				"%d malformed entities skipped during recovery", rr.SkippedEntities)
		}
	}
	auditRep := Run(doc, opts)
	rep.Issues = append(rep.Issues, auditRep.Issues...)
	rep.Truncated = rep.Truncated || auditRep.Truncated
	if loadErr != nil {
		rep.Addf(SeverityWarning, CodeLoadIssue, "", "%s", loadErr.Error())
	}
	return doc, rep, loadErr
}

// RecoverFile recovers a file from disk (see RecoverReader).
func RecoverFile(filename string, opts *Options) (*drawing.Drawing, *Report, error) {
	f, err := os.Open(filename)
	if err != nil {
		rep := &Report{}
		rep.Addf(SeverityCritical, CodeLoadIssue, "", "%v", err)
		return nil, rep, err
	}
	defer f.Close()
	return RecoverReader(f, opts)
}
