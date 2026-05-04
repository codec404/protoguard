package llm_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/codec404/protoguard/internal/config"
	"github.com/codec404/protoguard/internal/llm"
)

// Integration tests against a running Ollama instance.
//
// Enable with:
//
//	PROTOGUARD_TEST_OLLAMA=1 go test ./internal/llm/... -v -count=1 -run Integration_Ollama
//
// Optional env:
//   - PROTOGUARD_TEST_OLLAMA_BASE_URL (default http://127.0.0.1:11434/v1 for openai_shape)
//   - PROTOGUARD_TEST_OLLAMA_NATIVE_URL (default http://127.0.0.1:11434 for native backend)
//   - PROTOGUARD_TEST_OLLAMA_MODEL (default llama3.2)

func TestIntegration_Ollama_ListModels(t *testing.T) {
	if os.Getenv("PROTOGUARD_TEST_OLLAMA") != "1" {
		t.Skip("set PROTOGUARD_TEST_OLLAMA=1 to run Ollama integration tests")
	}
	base := strings.TrimSuffix(strings.TrimSpace(os.Getenv("PROTOGUARD_TEST_OLLAMA_BASE_URL")), "/")
	if base == "" {
		base = strings.TrimSuffix(config.DefaultLocalBaseURL, "/")
	}
	cli := llm.NewHTTPClient(15 * time.Second)
	resp, err := cli.Get(base + "/models")
	if err != nil {
		t.Fatalf("GET %s/models: %v", base, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("GET %s/models: %s", base, resp.Status)
	}
}

func TestIntegration_Ollama_OpenAIShapeChat(t *testing.T) {
	if os.Getenv("PROTOGUARD_TEST_OLLAMA") != "1" {
		t.Skip("set PROTOGUARD_TEST_OLLAMA=1 to run Ollama integration tests")
	}
	base := strings.TrimSpace(os.Getenv("PROTOGUARD_TEST_OLLAMA_BASE_URL"))
	if base == "" {
		base = config.DefaultLocalBaseURL
	}
	model := strings.TrimSpace(os.Getenv("PROTOGUARD_TEST_OLLAMA_MODEL"))
	if model == "" {
		model = "llama3.2"
	}
	c := &llm.Client{
		BaseURL:    base,
		Model:      model,
		Backend:    config.BackendOpenAIShape,
		HTTPClient: llm.NewHTTPClient(120 * time.Second),
	}
	out, err := c.Complete(
		"You are a test harness. Follow instructions exactly.",
		"Reply with exactly these two characters and nothing else: OK",
		64,
	)
	if err != nil {
		t.Fatal(err)
	}
	out = strings.TrimSpace(out)
	if len(out) < 2 {
		t.Fatalf("expected non-trivial completion, got %q", out)
	}
	// Small models may paraphrase; connectivity + non-empty output is enough.
	t.Logf("openai_shape completion: %q", out)
}

func TestIntegration_Ollama_NativeChat(t *testing.T) {
	if os.Getenv("PROTOGUARD_TEST_OLLAMA") != "1" {
		t.Skip("set PROTOGUARD_TEST_OLLAMA=1 to run Ollama integration tests")
	}
	base := strings.TrimSpace(os.Getenv("PROTOGUARD_TEST_OLLAMA_NATIVE_URL"))
	if base == "" {
		base = "http://127.0.0.1:11434"
	}
	model := strings.TrimSpace(os.Getenv("PROTOGUARD_TEST_OLLAMA_MODEL"))
	if model == "" {
		model = "llama3.2"
	}
	c := &llm.Client{
		BaseURL:    base,
		Model:      model,
		Backend:    config.BackendOllama,
		HTTPClient: llm.NewHTTPClient(120 * time.Second),
	}
	out, err := c.Complete(
		"You are a test harness.",
		"Reply with exactly the word YES and nothing else.",
		32,
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.ToUpper(strings.TrimSpace(out)), "YES") {
		t.Fatalf("unexpected completion %q (want substring YES)", out)
	}
}
