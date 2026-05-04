// Package protoguard is the stable public API for diffing OpenAPI and protobuf
// FileDescriptorSet contracts, classifying compatibility impact, and optionally
// fetching LLM explanations over HTTP.
//
// Security-minded defaults: OpenAPI remote external refs are off unless
// Options.AllowOpenAPIExternalRefs is set (mitigates SSRF via $ref). Spec reads
// respect Options.MaxSpecBytes (zero uses an internal default cap). Never log
// APIKey or full bearer tokens; LLM BaseURL validation rejects unsafe schemes,
// URL userinfo, and common cloud metadata hosts when configuring via LLMConfig.
//
//	import (
//	  "context"
//	  pg "github.com/codec404/protoguard/pkg/protoguard"
//	)
//	res, err := pg.Diff(ctx, pg.Options{OldPath: "old.yaml", NewPath: "new.yaml", SkipLLM: true})
package protoguard
