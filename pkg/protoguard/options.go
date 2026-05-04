package protoguard

import (
	"github.com/codec404/protoguard/internal/run"
)

// LLMConfig configures HTTP chat backends for explanations (Ollama, OpenAI-compatible, etc.).
// Never log fields verbatim in production — APIKey is sensitive.
type LLMConfig struct {
	// Target is "local" or "cloud". Empty uses env PROTOGUARD_LLM_TARGET or defaults to local.
	Target  string
	BaseURL string
	Model   string
	// APIKey is sent as Authorization: Bearer only; do not log or persist outside process env.
	APIKey string
	// Backend is "openai_shape" (default) or "ollama".
	Backend string
	// MaxOutputTokens caps completion length per chunk (e.g. 2048).
	MaxOutputTokens int
	// TimeoutSecs is per-chunk HTTP timeout.
	TimeoutSecs int
	// MaxChunkBytes bounds serialized JSON per LLM request (~8k default).
	MaxChunkBytes int
}

// Options configures a single Diff run (library use — no CLI output format flags).
type Options struct {
	OldPath, NewPath string
	Spec             SpecKind

	SkipLLM         bool
	IncludeFullSpec bool
	RedactURLs      bool
	CacheDir        string

	// MaxSpecBytes caps each spec file read (0 = default 32 MiB).
	MaxSpecBytes int64
	// AllowOpenAPIExternalRefs enables remote OpenAPI $ref fetching (SSRF risk; default false).
	AllowOpenAPIExternalRefs bool

	LLM LLMConfig
}

func (o Options) toRun() run.Options {
	return run.Options{
		OldPath:                  o.OldPath,
		NewPath:                  o.NewPath,
		Spec:                     run.SpecKind(o.Spec),
		SkipLLM:                  o.SkipLLM,
		IncludeFullSpec:          o.IncludeFullSpec,
		RedactURLs:               o.RedactURLs,
		CacheDir:                 o.CacheDir,
		MaxSpecBytes:             o.MaxSpecBytes,
		AllowOpenAPIExternalRefs: o.AllowOpenAPIExternalRefs,
		LLM: run.LLMOpts{
			TargetStr:     o.LLM.Target,
			BaseURL:       o.LLM.BaseURL,
			Model:         o.LLM.Model,
			APIKey:        o.LLM.APIKey,
			BackendStr:    o.LLM.Backend,
			MaxOutTok:     o.LLM.MaxOutputTokens,
			TimeoutSecs:   o.LLM.TimeoutSecs,
			MaxChunkBytes: o.LLM.MaxChunkBytes,
		},
	}
}
