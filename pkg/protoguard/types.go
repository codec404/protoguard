package protoguard

import "github.com/codec404/protoguard/internal/model"

// Re-export structured diff types from the internal model.
type (
	DiffReport = model.DiffReport
	Change     = model.Change
	ChangeKind = model.ChangeKind
	Impact     = model.Impact
)

// SchemaVersion is the JSON schema tag for DiffReport.SchemaVersion.
const SchemaVersion = model.SchemaVersion

// Change kinds.
const (
	ChangeAdd             = model.ChangeAdd
	ChangeRemove          = model.ChangeRemove
	ChangeModify          = model.ChangeModify
	ChangeTypeChange      = model.ChangeTypeChange
	ChangeRenameSuspected = model.ChangeRenameSuspected
)

// Impact labels.
const (
	ImpactBreaking    = model.ImpactBreaking
	ImpactNonBreaking = model.ImpactNonBreaking
	ImpactRisky       = model.ImpactRisky
)
