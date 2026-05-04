package security

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

const (
	// DefaultMaxSpecBytes caps OpenAPI / FileDescriptorSet inputs (DoS safety).
	DefaultMaxSpecBytes int64 = 32 << 20 // 32 MiB
)

// ErrMetadataEndpointBlocked is returned when the LLM base URL points at a known metadata / SSRF sink.
var ErrMetadataEndpointBlocked = fmt.Errorf("llm base URL host is blocked (cloud metadata / SSRF sink)")

var blockedHosts = []string{
	"metadata.google.internal",
	"metadata.goog",
}

var blockedHostIPs = []net.IP{
	net.ParseIP("169.254.169.254"),
}

// ValidateLLMBaseURL enforces safe outbound HTTP targets for chat completions.
// - Only http/https schemes.
// - No username/password embedded in the URL (secrets belong in headers / env).
// - Blocks common cloud metadata endpoints (SSRF) for every target mode.
func ValidateLLMBaseURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("empty LLM base URL")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("parse LLM base URL: %w", err)
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
	default:
		return fmt.Errorf("LLM base URL scheme must be http or https, got %q", u.Scheme)
	}
	if u.User != nil {
		return fmt.Errorf("LLM base URL must not embed credentials (use PROTOGUARD_LLM_API_KEY / Authorization header)")
	}
	host := strings.ToLower(strings.TrimSpace(u.Hostname()))
	if host == "" {
		return fmt.Errorf("LLM base URL missing host")
	}
	if ip := net.ParseIP(host); ip != nil {
		for _, b := range blockedHostIPs {
			if b != nil && ip.Equal(b) {
				return ErrMetadataEndpointBlocked
			}
		}
	}
	for _, b := range blockedHosts {
		if host == b || strings.HasSuffix(host, "."+b) {
			return ErrMetadataEndpointBlocked
		}
	}
	return nil
}
