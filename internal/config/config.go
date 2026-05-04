package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/codec404/protoguard/internal/security"
)

// LLMBackend selects HTTP wire shape for chat requests.
type LLMBackend string

const (
	BackendOpenAIShape LLMBackend = "openai_shape"
	BackendOllama      LLMBackend = "ollama"
)

// LLMTarget is local vs cloud routing.
type LLMTarget string

const (
	TargetLocal LLMTarget = "local"
	TargetCloud LLMTarget = "cloud"
)

// EnvLLMTarget selects default llm-target when flag omitted.
const EnvLLMTarget = "PROTOGUARD_LLM_TARGET"
const EnvBaseURL = "PROTOGUARD_LLM_BASE_URL"
const EnvModel = "PROTOGUARD_LLM_MODEL"
const EnvAPIKey = "PROTOGUARD_LLM_API_KEY"

// LLM holds resolved LLM client settings.
type LLM struct {
	Target      LLMTarget
	Backend     LLMBackend
	BaseURL     string
	Model       string
	APIKey      string
	MaxInTok    int
	MaxOutTok   int
	TimeoutSecs int
}

// DefaultLocalBaseURL is Ollama OpenAI-compatible endpoint prefix.
const DefaultLocalBaseURL = "http://127.0.0.1:11434/v1"

var blockedLocalHosts = []string{
	"api.openai.com",
	"openai.azure.com",
}

// ResolveLLMTarget parses --llm-target and PROTOGUARD_LLM_TARGET (flag wins).
func ResolveLLMTarget(flagVal string, envVal string) (LLMTarget, error) {
	if strings.TrimSpace(flagVal) != "" {
		return parseTarget(flagVal)
	}
	ev := strings.TrimSpace(envVal)
	if ev == "" {
		return TargetLocal, nil
	}
	return parseTarget(ev)
}

func parseTarget(s string) (LLMTarget, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "local":
		return TargetLocal, nil
	case "cloud":
		return TargetCloud, nil
	default:
		return "", fmt.Errorf("invalid llm target %q: want local or cloud", s)
	}
}

// ApplyEnv fills defaults from environment when flags leave values empty.
func ApplyEnv(llm *LLM) {
	if llm.BaseURL == "" {
		llm.BaseURL = strings.TrimSpace(os.Getenv(EnvBaseURL))
	}
	if llm.Model == "" {
		llm.Model = strings.TrimSpace(os.Getenv(EnvModel))
	}
	if llm.APIKey == "" {
		llm.APIKey = strings.TrimSpace(os.Getenv(EnvAPIKey))
	}
	if llm.Target == TargetLocal && llm.BaseURL == "" {
		llm.BaseURL = DefaultLocalBaseURL
	}
}

// Validate checks guardrails (local must not hit vendor defaults accidentally).
func Validate(llm *LLM) error {
	if strings.TrimSpace(llm.BaseURL) != "" {
		if err := security.ValidateLLMBaseURL(llm.BaseURL); err != nil {
			return fmt.Errorf("llm base URL: %w", err)
		}
	}
	u := strings.ToLower(llm.BaseURL)
	for _, h := range blockedLocalHosts {
		if llm.Target == TargetLocal && strings.Contains(u, h) {
			return fmt.Errorf("refusing %s as base URL while llm-target is local (set PROTOGUARD_LLM_TARGET=cloud or --llm-target=cloud to use hosted APIs)", h)
		}
	}
	if llm.Target == TargetCloud && llm.BaseURL == "" {
		return fmt.Errorf("cloud llm-target requires base URL (--llm-base-url or %s)", EnvBaseURL)
	}
	return nil
}

// ParseBackend normalizes llm-backend flag value.
func ParseBackend(s string) (LLMBackend, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "openai_shape", "":
		return BackendOpenAIShape, nil
	case "ollama":
		return BackendOllama, nil
	default:
		return "", fmt.Errorf("invalid llm-backend %q", s)
	}
}
