package model

// SchemaVersion is the structured diff JSON version.
const SchemaVersion = "protoguard.diff.v1"

// SpecKind identifies which specification family was compared.
type SpecKind string

const (
	SpecOpenAPI  SpecKind = "openapi"
	SpecProtobuf SpecKind = "protobuf"
	SpecMixed    SpecKind = "mixed" // reserved if comparing unrelated pairs in future
)

// ChangeKind is a machine-readable change category.
type ChangeKind string

const (
	ChangeAdd             ChangeKind = "add"
	ChangeRemove          ChangeKind = "remove"
	ChangeModify          ChangeKind = "modify"
	ChangeTypeChange      ChangeKind = "type_change"
	ChangeRenameSuspected ChangeKind = "rename_suspected"
)

// Impact is rule-based classification before/with LLM narrative.
type Impact string

const (
	ImpactBreaking    Impact = "BREAKING"
	ImpactNonBreaking Impact = "NON_BREAKING"
	ImpactRisky       Impact = "RISKY"
)

// Change is one structured delta with minimal snippets.
type Change struct {
	Path    string     `json:"path"`
	Kind    ChangeKind `json:"kind"`
	Old     any        `json:"old,omitempty"`
	New     any        `json:"new,omitempty"`
	Impact  Impact     `json:"impact"`
	Summary string     `json:"summary,omitempty"`
}

// DiffReport is the machine-readable diff output.
type DiffReport struct {
	SchemaVersion string   `json:"schema_version"`
	SpecKind      SpecKind `json:"spec_kind"`
	OldHint       string   `json:"old_source,omitempty"`
	NewHint       string   `json:"new_source,omitempty"`
	Changes       []Change `json:"changes"`
}

// HasBreaking returns true if any change is classified BREAKING.
func (r *DiffReport) HasBreaking() bool {
	for _, c := range r.Changes {
		if c.Impact == ImpactBreaking {
			return true
		}
	}
	return false
}
