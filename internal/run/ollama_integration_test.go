package run_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codec404/protoguard/internal/config"
	"github.com/codec404/protoguard/internal/run"
)

// End-to-end LLM path through protoguard diff (structured chunks → Ollama).
//
//	PROTOGUARD_TEST_OLLAMA=1 go test ./internal/run/... -v -count=1 -run Integration_RunDiffWithOllama

func TestIntegration_RunDiffWithOllamaExplain(t *testing.T) {
	if os.Getenv("PROTOGUARD_TEST_OLLAMA") != "1" {
		t.Skip("set PROTOGUARD_TEST_OLLAMA=1")
	}
	root := filepath.Join("..", "..", "testdata")
	oldP := filepath.Join(root, "openapi_old.yaml")
	newP := filepath.Join(root, "openapi_new.yaml")

	base := strings.TrimSpace(os.Getenv("PROTOGUARD_TEST_OLLAMA_BASE_URL"))
	if base == "" {
		base = config.DefaultLocalBaseURL
	}
	model := strings.TrimSpace(os.Getenv("PROTOGUARD_TEST_OLLAMA_MODEL"))
	if model == "" {
		model = "llama3.2"
	}

	res, err := run.Diff(context.Background(), run.Options{
		OldPath: oldP,
		NewPath: newP,
		Spec:    run.SpecOpenAPI,
		SkipLLM: false,
		LLM: run.LLMOpts{
			TargetStr:     "local",
			BaseURL:       base,
			Model:         model,
			BackendStr:    "openai_shape",
			MaxOutTok:     512,
			TimeoutSecs:   180,
			MaxChunkBytes: 8000,
		},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.LLMParts) == 0 {
		t.Fatal("expected at least one LLM chunk response")
	}
	combined := strings.Join(res.LLMParts, "\n")
	if len(strings.TrimSpace(combined)) < 20 {
		t.Fatalf("LLM output suspiciously short: %q", combined)
	}
	t.Logf("received %d LLM chunk(s), total chars=%d", len(res.LLMParts), len(combined))
}
