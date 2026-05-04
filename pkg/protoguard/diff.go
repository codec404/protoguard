package protoguard

import (
	"context"
	"io"

	"github.com/codec404/protoguard/internal/run"
)

// Result is the outcome of Diff.
type Result struct {
	Report   *DiffReport
	LLMParts []string
	Breaking bool
}

// Diff compares old vs new artifacts and optionally calls the configured LLM.
// Pass context.Background() if you do not need cancellation.
func Diff(ctx context.Context, opts Options) (*Result, error) {
	return DiffWithStderr(ctx, opts, nil)
}

// DiffWithStderr is like Diff but writes optional diagnostics (e.g. cloud banner) to stderr.
func DiffWithStderr(ctx context.Context, opts Options, stderr io.Writer) (*Result, error) {
	rr, err := run.Diff(ctx, opts.toRun(), stderr)
	if err != nil {
		return nil, err
	}
	return &Result{
		Report:   rr.Report,
		LLMParts: rr.LLMParts,
		Breaking: rr.Breaking,
	}, nil
}
