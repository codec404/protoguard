package config_test

import (
	"testing"

	"github.com/codec404/protoguard/internal/config"
)

func TestValidateCloudRequiresBaseURL(t *testing.T) {
	llm := &config.LLM{Target: config.TargetCloud, BaseURL: "", Model: "x"}
	config.ApplyEnv(llm)
	llm.BaseURL = ""
	if err := config.Validate(llm); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateLocalBlocksOpenAIHost(t *testing.T) {
	llm := &config.LLM{
		Target:  config.TargetLocal,
		BaseURL: "https://api.openai.com/v1",
		Model:   "gpt-4",
		Backend: config.BackendOpenAIShape,
	}
	if err := config.Validate(llm); err == nil {
		t.Fatal("expected error")
	}
}
