package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/codec404/protoguard/internal/model"
)

// WriteJSON writes DiffReport (and optional LLM sections).
func WriteJSON(w io.Writer, report *model.DiffReport, llmParts []string, pretty bool) error {
	type out struct {
		Report      *model.DiffReport `json:"report"`
		LLMSections []string          `json:"llm_sections,omitempty"`
	}
	o := out{Report: report, LLMSections: llmParts}
	enc := json.NewEncoder(w)
	if pretty {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(o)
}

// WriteMarkdown writes human-readable changelog-style output.
func WriteMarkdown(w io.Writer, report *model.DiffReport, llmParts []string) {
	fmt.Fprintf(w, "# ProtoGuard diff (%s)\n\n", report.SpecKind)
	fmt.Fprintf(w, "Changes: **%d**", len(report.Changes))
	if report.HasBreaking() {
		fmt.Fprintf(w, " — includes **BREAKING**")
	}
	fmt.Fprintf(w, "\n\n")
	counts := map[model.Impact]int{}
	for _, c := range report.Changes {
		counts[c.Impact]++
	}
	fmt.Fprintf(w, "| Impact | Count |\n|---|---:|\n")
	for _, imp := range []model.Impact{model.ImpactBreaking, model.ImpactRisky, model.ImpactNonBreaking} {
		if counts[imp] > 0 {
			fmt.Fprintf(w, "| %s | %d |\n", imp, counts[imp])
		}
	}
	fmt.Fprintf(w, "\n## Structured changes\n\n")
	for _, c := range report.Changes {
		fmt.Fprintf(w, "- `%s` **%s** (%s): %s\n", c.Path, c.Kind, c.Impact, c.Summary)
	}
	if len(llmParts) > 0 {
		fmt.Fprintf(w, "\n## AI explanations\n\n")
		for i, p := range llmParts {
			fmt.Fprintf(w, "### Chunk %d\n\n%s\n\n", i+1, strings.TrimSpace(p))
		}
	}
}
