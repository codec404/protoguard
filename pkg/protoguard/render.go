package protoguard

import (
	"io"

	"github.com/codec404/protoguard/internal/output"
)

// WriteReportJSON writes Result as JSON (report + llm_sections).
func WriteReportJSON(w io.Writer, res *Result, pretty bool) error {
	return output.WriteJSON(w, res.Report, res.LLMParts, pretty)
}

// WriteReportMarkdown writes a Markdown summary and optional AI sections.
func WriteReportMarkdown(w io.Writer, res *Result) {
	output.WriteMarkdown(w, res.Report, res.LLMParts)
}
