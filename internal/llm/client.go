package llm

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/codec404/protoguard/internal/config"
)

// Client calls local or hosted chat HTTP APIs.
type Client struct {
	BaseURL    string
	APIKey     string
	Model      string
	Backend    config.LLMBackend
	HTTPClient *http.Client
}

// Complete returns assistant markdown/text for one chunk (no cancellation).
func (c *Client) Complete(system, user string, maxOut int) (string, error) {
	return c.CompleteContext(context.Background(), system, user, maxOut)
}

// CompleteContext is like Complete but respects ctx for the HTTP request.
func (c *Client) CompleteContext(ctx context.Context, system, user string, maxOut int) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	base := strings.TrimRight(c.BaseURL, "/")
	switch c.Backend {
	case config.BackendOllama:
		return c.doOllamaChat(ctx, base, system, user, maxOut)
	default:
		return c.doOpenAICompat(ctx, base, system, user, maxOut)
	}
}

func (c *Client) hdr(req *http.Request) {
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")
}

func (c *Client) doOpenAICompat(ctx context.Context, base, system, user string, maxOut int) (string, error) {
	u := base + "/chat/completions"
	body := map[string]any{
		"model": c.Model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"temperature": 0.2,
	}
	if maxOut > 0 {
		body["max_tokens"] = maxOut
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	c.hdr(req)
	resp, err := c.http().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("chat/completions %s: %s", resp.Status, truncate(string(b), 500))
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return "", fmt.Errorf("decode response: %w body=%s", err, truncate(string(b), 400))
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("empty choices from LLM")
	}
	return strings.TrimSpace(out.Choices[0].Message.Content), nil
}

func (c *Client) doOllamaChat(ctx context.Context, base, system, user string, maxOut int) (string, error) {
	uBase := strings.TrimSuffix(base, "/v1")
	u := uBase + "/api/chat"
	body := map[string]any{
		"model": c.Model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"stream": false,
	}
	if maxOut > 0 {
		body["options"] = map[string]any{"num_predict": maxOut}
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	c.hdr(req)
	resp, err := c.http().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("ollama /api/chat %s: %s", resp.Status, truncate(string(b), 500))
	}
	var out struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return "", fmt.Errorf("decode ollama response: %w", err)
	}
	return strings.TrimSpace(out.Message.Content), nil
}

func (c *Client) http() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// NewHTTPClient returns an HTTP client suitable for LLM backends: TLS ≥1.2,
// bounded lifetime, and redirects disabled (SSRF hardening).
func NewHTTPClient(timeout time.Duration) *http.Client {
	tr, ok := http.DefaultTransport.(*http.Transport)
	var rt http.RoundTripper = http.DefaultTransport
	if ok {
		cloned := tr.Clone()
		if cloned.TLSClientConfig == nil {
			cloned.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		} else {
			tc := cloned.TLSClientConfig.Clone()
			if tc.MinVersion < tls.VersionTLS12 {
				tc.MinVersion = tls.VersionTLS12
			}
			cloned.TLSClientConfig = tc
		}
		rt = cloned
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: rt,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return fmt.Errorf("protoguard: HTTP redirects disabled for LLM requests")
		},
	}
}
