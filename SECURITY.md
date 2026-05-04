# Security

This document describes how ProtoGuard treats untrusted inputs, network boundaries, and secrets. It is aimed at operators and integrators embedding **`github.com/codec404/protoguard/pkg/protoguard`** or running the CLI in CI.

## Threat model

ProtoGuard reads **your** old/new contract artifacts (OpenAPI YAML/JSON or protobuf `FileDescriptorSet` binaries), optionally sends **structured diff chunks** (not full specs by default) to an HTTP LLM endpoint, and writes a **local disk cache**. Typical risks:

| Surface | Risk | Mitigation (defaults) |
|---------|------|----------------------|
| OpenAPI documents | Remote `$ref` resolution can trigger SSRF | Remote refs are **disabled** unless you opt in (`AllowOpenAPIExternalRefs` / `--allow-openapi-external-refs`). |
| Large inputs | Memory/DoS | Reads are **bounded** (`MaxSpecBytes` / `--max-spec-mb`, capped). |
| LLM `BaseURL` | SSRF to cloud metadata or internal networks | Only **http/https**; **no URL userinfo**; known **metadata endpoints** blocked in validation. |
| LLM HTTP client | TLS downgrade, credential leakage via redirects | TLS **1.2+**; **redirects disabled**. |
| API keys | Exposure via logs or env dumps | Keys come from **flags/env only**; library docs warn **not to log** `APIKey`. |
| Cache directory | Sensitive explanations readable by other users | Cache dirs **`0700`**, files **`0600`**. |

This is **not** a formal penetration-test report. Adjust flags and options if your threat model differs (e.g. trusted specs only, air-gapped LLM).

## Secrets

- Prefer **`PROTOGUARD_LLM_API_KEY`** (or equivalent env) over committing keys; rotate if leaked.
- Treat cached LLM responses as potentially sensitive; restrict filesystem permissions and backup policies accordingly.

## Reporting issues

Report vulnerabilities privately to the maintainers (for example via the repository’s **Security → Advisories** flow on GitHub, or your org’s coordinated disclosure process). Avoid public issues with exploit details until a fix is coordinated.

## Supply chain

CI runs **`go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...`** on the **stable** Go toolchain; findings depend on that toolchain’s standard library. Run the same locally after upgrading Go. Keep dependencies updated per your policy.
