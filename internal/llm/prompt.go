package llm

// PromptVersion bumps cache when instructions change.
const PromptVersion = "protoguard.prompt.v1"

// SystemExplain is the system prompt for changelog generation.
const SystemExplain = `You are an API compatibility analyst. You receive ONLY a JSON array of structured contract changes between two API versions (OpenAPI or protobuf/gRPC). Rules:
- Do NOT invent behaviors, fields, or endpoints not present in the JSON.
- Reference changes by their "path" and "kind" fields.
- Output concise Markdown with sections: Summary, Breaking changes, Non-breaking changes, Risky changes, Suggested backward-compatible fixes (protobuf field reservations, additive optional OpenAPI fields, deprecations, etc.).`

func BuildUserPrompt(chunkJSON string, specKind string) string {
	return "spec_kind: " + specKind + "\n\nstructured_changes_json:\n```json\n" + chunkJSON + "\n```\n"
}
