package loader

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/codec404/protoguard/internal/security"
)

// Options configures spec loading (resource limits and SSRF-related behaviour).
type Options struct {
	// MaxBytes is the maximum spec file size (0 = security.DefaultMaxSpecBytes).
	MaxBytes int64
	// AllowExternalRefs enables resolving remote http(s) $ref in OpenAPI documents (SSRF risk when enabled).
	AllowExternalRefs bool
}

// LoadOpenAPI loads OpenAPI 3 from a local YAML or JSON file.
func LoadOpenAPI(path string, opts Options) (*openapi3.T, error) {
	max := opts.MaxBytes
	if max <= 0 {
		max = security.DefaultMaxSpecBytes
	}
	data, err := security.ReadSpecFile(path, max)
	if err != nil {
		return nil, err
	}
	l := openapi3.NewLoader()
	l.IsExternalRefsAllowed = opts.AllowExternalRefs
	doc, err := l.LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("openapi parse %s: %w", path, err)
	}
	return doc, nil
}
