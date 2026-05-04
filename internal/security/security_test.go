package security

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateLLMBaseURL_RejectsUserinfo(t *testing.T) {
	err := ValidateLLMBaseURL("https://user:pass@api.example.com/v1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateLLMBaseURL_RejectsMetadataIP(t *testing.T) {
	err := ValidateLLMBaseURL("http://169.254.169.254/latest/meta-data/")
	if !errors.Is(err, ErrMetadataEndpointBlocked) {
		t.Fatalf("got %v", err)
	}
}

func TestValidateLLMBaseURL_RejectsBlockedHostname(t *testing.T) {
	err := ValidateLLMBaseURL("http://metadata.google.internal/")
	if !errors.Is(err, ErrMetadataEndpointBlocked) {
		t.Fatalf("got %v", err)
	}
}

func TestValidateLLMBaseURL_AllowsLocalhost(t *testing.T) {
	if err := ValidateLLMBaseURL("http://127.0.0.1:11434/v1"); err != nil {
		t.Fatal(err)
	}
}

func TestReadSpecFile_RespectsLimit(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "chunk.bin")
	if err := os.WriteFile(path, []byte("abcd"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := ReadSpecFile(path, 2)
	if err == nil {
		t.Fatal("expected over-limit error")
	}
}
