package protoguard

import "github.com/codec404/protoguard/internal/run"

// SpecKind selects which loader to use.
type SpecKind = run.SpecKind

const (
	SpecAuto     = run.SpecAuto
	SpecOpenAPI  = run.SpecOpenAPI
	SpecProtobuf = run.SpecProtobuf
)
