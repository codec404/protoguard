package run

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/codec404/protoguard/internal/cache"
	"github.com/codec404/protoguard/internal/chunk"
	"github.com/codec404/protoguard/internal/classify"
	"github.com/codec404/protoguard/internal/config"
	odiff "github.com/codec404/protoguard/internal/diff/openapi"
	pdiff "github.com/codec404/protoguard/internal/diff/proto"
	"github.com/codec404/protoguard/internal/llm"
	"github.com/codec404/protoguard/internal/loader"
	"github.com/codec404/protoguard/internal/model"
)

// SpecKind selects loader behavior.
type SpecKind string

const (
	SpecAuto     SpecKind = "auto"
	SpecOpenAPI  SpecKind = "openapi"
	SpecProtobuf SpecKind = "protobuf"
)

// Options configures one diff run.
type Options struct {
	OldPath, NewPath string
	Spec             SpecKind
	SkipLLM          bool
	IncludeFullSpec  bool
	RedactURLs       bool
	CacheDir         string
	// MaxSpecBytes caps bytes read per spec file (0 = default 32 MiB).
	MaxSpecBytes int64
	// AllowOpenAPIExternalRefs enables remote OpenAPI $ref resolution (SSRF risk when enabled).
	AllowOpenAPIExternalRefs bool

	LLM LLMOpts
}

// LLMOpts configures HTTP explanation backend (CLI flags).
type LLMOpts struct {
	TargetStr     string
	BaseURL       string
	Model         string
	APIKey        string
	BackendStr    string
	MaxOutTok     int
	TimeoutSecs   int
	MaxChunkBytes int
}

// Result is returned from Diff after execution.
type Result struct {
	Report   *model.DiffReport
	LLMParts []string
	Breaking bool
}

// Diff performs load → diff → classify → optional LLM → output.
// ctx controls cancellation for outbound LLM HTTP requests; use context.Background() if unused.
func Diff(ctx context.Context, opts Options, stderr io.Writer) (*Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	kind, err := resolveSpec(opts.OldPath, opts.NewPath, opts.Spec)
	if err != nil {
		return nil, err
	}

	var report *model.DiffReport
	switch kind {
	case SpecOpenAPI:
		report, err = diffOpenAPI(opts)
	case SpecProtobuf:
		report, err = diffProto(opts)
	default:
		return nil, fmt.Errorf("unknown spec kind %q", kind)
	}
	if err != nil {
		return nil, err
	}

	classify.ApplyToReport(report)

	var llmParts []string
	if !opts.SkipLLM {
		parts, err := explain(ctx, report, opts, stderr)
		if err != nil {
			return nil, err
		}
		llmParts = parts
	}

	res := &Result{Report: report, LLMParts: llmParts, Breaking: report.HasBreaking()}
	return res, nil
}

func resolveSpec(oldP, newP string, sk SpecKind) (SpecKind, error) {
	switch sk {
	case SpecOpenAPI, SpecProtobuf:
		return sk, nil
	case SpecAuto, "":
		for _, p := range []string{oldP, newP} {
			ext := strings.ToLower(filepath.Ext(p))
			switch ext {
			case ".pb", ".pb.bin", ".fds":
				return SpecProtobuf, nil
			}
		}
		return SpecOpenAPI, nil
	default:
		return "", fmt.Errorf("invalid --spec %q", sk)
	}
}

func diffOpenAPI(opts Options) (*model.DiffReport, error) {
	lo := loader.Options{
		MaxBytes:          opts.MaxSpecBytes,
		AllowExternalRefs: opts.AllowOpenAPIExternalRefs,
	}
	oldD, err := loader.LoadOpenAPI(opts.OldPath, lo)
	if err != nil {
		return nil, err
	}
	newD, err := loader.LoadOpenAPI(opts.NewPath, lo)
	if err != nil {
		return nil, err
	}
	if opts.RedactURLs {
		redactOpenAPIServers(oldD)
		redactOpenAPIServers(newD)
	}
	r := odiff.Diff(oldD, newD)
	r.OldHint = opts.OldPath
	r.NewHint = opts.NewPath
	if opts.IncludeFullSpec {
		r.Changes = append([]model.Change{{
			Path:    "debug.full_spec_warning",
			Kind:    model.ChangeModify,
			Impact:  model.ImpactRisky,
			Summary: "full spec embedding requested — omitted here; use dedicated tooling",
		}}, r.Changes...)
	}
	return r, nil
}

func redactOpenAPIServers(doc *openapi3.T) {
	if doc == nil {
		return
	}
	doc.Servers = openapi3.Servers{}
}

func diffProto(opts Options) (*model.DiffReport, error) {
	lp := loader.Options{MaxBytes: opts.MaxSpecBytes}
	oldF, err := loader.LoadFileDescriptorSet(opts.OldPath, lp)
	if err != nil {
		return nil, err
	}
	newF, err := loader.LoadFileDescriptorSet(opts.NewPath, lp)
	if err != nil {
		return nil, err
	}
	r := pdiff.Diff(oldF, newF)
	r.OldHint = opts.OldPath
	r.NewHint = opts.NewPath
	return r, nil
}

func explain(ctx context.Context, report *model.DiffReport, opts Options, stderr io.Writer) ([]string, error) {
	cfg, err := buildLLMConfig(opts.LLM)
	if err != nil {
		return nil, err
	}
	if cfg.Target == config.TargetCloud && stderr != nil {
		dest := safeOrigin(cfg.BaseURL)
		fmt.Fprintf(stderr, "protoguard: llm-target=cloud — structured diff chunks will be sent to %s (not full specs unless explicitly enabled).\n", dest)
	}

	chunks, err := chunk.Split(report, opts.LLM.MaxChunkBytes)
	if err != nil {
		return nil, err
	}

	cdir, err := cache.Dir(opts.CacheDir)
	if err != nil {
		return nil, err
	}

	client := &llm.Client{
		BaseURL:    cfg.BaseURL,
		APIKey:     cfg.APIKey,
		Model:      cfg.Model,
		Backend:    cfg.Backend,
		HTTPClient: llm.NewHTTPClient(time.Duration(opts.LLM.TimeoutSecs) * time.Second),
	}

	var parts []string
	for _, ck := range chunks {
		if len(ck.Changes) == 0 {
			continue
		}
		raw, err := json.Marshal(ck.Changes)
		if err != nil {
			return nil, err
		}
		user := llm.BuildUserPrompt(string(raw), string(report.SpecKind))

		keyPayload, _ := json.Marshal(map[string]string{
			"prompt_ver": llm.PromptVersion,
			"model":      cfg.Model,
			"target":     string(cfg.Target),
			"backend":    string(cfg.Backend),
			"chunk_id":   ck.ID,
		})
		cacheKey := cache.Key(keyPayload, raw)

		if b, hit, err := cache.Get(cdir, cacheKey); err == nil && hit {
			parts = append(parts, string(b))
			continue
		}

		text, err := client.CompleteContext(ctx, llm.SystemExplain, user, opts.LLM.MaxOutTok)
		if err != nil {
			return nil, err
		}
		_ = cache.Put(cdir, cacheKey, []byte(text))
		parts = append(parts, text)
	}
	return parts, nil
}

type resolvedLLM struct {
	Target  config.LLMTarget
	BaseURL string
	Model   string
	APIKey  string
	Backend config.LLMBackend
}

func buildLLMConfig(o LLMOpts) (*resolvedLLM, error) {
	t, err := config.ResolveLLMTarget(strings.TrimSpace(o.TargetStr), os.Getenv(config.EnvLLMTarget))
	if err != nil {
		return nil, err
	}
	be, err := config.ParseBackend(o.BackendStr)
	if err != nil {
		return nil, err
	}
	llmCfg := &config.LLM{
		Target:      t,
		Backend:     be,
		BaseURL:     strings.TrimSpace(o.BaseURL),
		Model:       strings.TrimSpace(o.Model),
		APIKey:      strings.TrimSpace(o.APIKey),
		MaxOutTok:   o.MaxOutTok,
		TimeoutSecs: o.TimeoutSecs,
	}
	config.ApplyEnv(llmCfg)
	if llmCfg.Model == "" && llmCfg.Target == config.TargetLocal {
		llmCfg.Model = "llama3.2"
	}
	if llmCfg.Model == "" {
		return nil, fmt.Errorf("llm model is required when llm-target=cloud (set --llm-model or %s)", config.EnvModel)
	}
	if err := config.Validate(llmCfg); err != nil {
		return nil, err
	}
	return &resolvedLLM{
		Target:  llmCfg.Target,
		BaseURL: llmCfg.BaseURL,
		Model:   llmCfg.Model,
		APIKey:  llmCfg.APIKey,
		Backend: llmCfg.Backend,
	}, nil
}

func safeOrigin(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Host == "" {
		return "(configured endpoint)"
	}
	scheme := u.Scheme
	if scheme == "" {
		scheme = "https"
	}
	return scheme + "://" + u.Host
}
