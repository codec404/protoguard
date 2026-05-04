# ProtoGuard

CLI tool and **Go library** to diff **OpenAPI 3** specs or **protobuf `FileDescriptorSet`** blobs between versions, classify changes as breaking / non-breaking / risky, optionally explain them via a **local or cloud** HTTP chat API (OpenAI-compatible `/v1/chat/completions` or native **Ollama** `/api/chat`), and cache LLM responses.

## Install

Home: **[github.com/codec404/protoguard](https://github.com/codec404/protoguard)**.

```bash
go install github.com/codec404/protoguard/cmd/protoguard@latest
```

(`@latest` resolves once this module has version tags on the Go module proxy; until then, install from a clone as below.)

From a local clone:

```bash
go install ./cmd/protoguard
```

## Go library (`pkg/protoguard`)

Stable API for embedding in other tools and services. The CLI uses this package.

```go
import (
	"context"
	"os"

	pg "github.com/codec404/protoguard/pkg/protoguard"
)

ctx := context.Background()
res, err := pg.Diff(ctx, pg.Options{
	OldPath: "old.yaml",
	NewPath: "new.yaml",
	Spec:    pg.SpecAuto,
	SkipLLM: true,
})
if err != nil {
	panic(err)
}

pg.WriteReportMarkdown(os.Stdout, res)

if res.Breaking {
	os.Exit(1)
}
```

Use **`DiffWithStderr`** if you want cloud-target banners on `io.Writer`. **`context.Context`** cancels outbound LLM HTTP calls (the CLI wires SIGINT/SIGTERM into this).

Configure explanations with **`LLMConfig`** (`Target`, `BaseURL`, `Model`, `Backend`, token/time limits, chunk size). Types **`DiffReport`**, **`Change`**, **`Impact`** match the structured JSON schema **`protoguard.diff.v1`**.

Packages under **`internal/`** are not import-stable for external modules; depend on **`github.com/codec404/protoguard/pkg/protoguard`** only.

## Setting up Ollama (local LLM)

ProtoGuard defaults to **`http://127.0.0.1:11434/v1`** (OpenAI-compatible) with **`--llm-backend=openai_shape`**. Ollama exposes that API when it is running locally.

### macOS (Homebrew)

```bash
brew install ollama
brew services start ollama    # runs at login; or run `ollama serve` in a terminal
ollama pull llama3.2           # matches ProtoGuard’s local default model name
```

Check that the API is up:

```bash
curl -s http://127.0.0.1:11434/v1/models
```

You should see `llama3.2` (often reported as `llama3.2:latest`). ProtoGuard accepts either `llama3.2` or the full tag.

### macOS / Windows / Linux (installer)

Download the app from [https://ollama.com](https://ollama.com). On macOS the menu-bar app starts the server on port **11434** automatically. Then run **`ollama pull llama3.2`** in a terminal.

### Linux (script install)

See [Ollama Linux docs](https://github.com/ollama/ollama/blob/main/docs/linux.md) for the official install script and GPU notes.

### Native Ollama HTTP (`/api/chat`)

If you prefer the native API instead of `/v1/chat/completions`:

```bash
protoguard diff ... --llm-backend=ollama --llm-base-url=http://127.0.0.1:11434
```

(omit the `/v1` suffix; ProtoGuard appends `/api/chat`.)

### Run ProtoGuard with Ollama

```bash
protoguard diff \
  --old testdata/openapi_old.yaml \
  --new testdata/openapi_new.yaml \
  --format markdown \
  --llm-model llama3.2
```

Omit **`--skip-llm`** so explanations run; first run may take longer while the model loads. Responses are cached under `~/.cache/protoguard` (or `XDG_CACHE_HOME/protoguard`).

## Usage

```bash
protoguard diff --old testdata/openapi_old.yaml --new testdata/openapi_new.yaml --skip-llm --format markdown
```

Outputs structured JSON (`protoguard.diff.v1`) plus Markdown. With `--fail-on-breaking` (default), the process exits `1` when any change is classified `BREAKING` (good for CI).

### Spec selection

- `--spec auto` (default): `.pb` / `.fds` → protobuf descriptor set; otherwise OpenAPI.
- `--spec openapi` / `--spec protobuf` forces the loader.

### LLM explanations

By default the CLI **calls an LLM** using structured diff chunks only (not full specs).

| Flag / env | Meaning |
|------------|---------|
| `--llm-target=local` or omit | Local-first (default). Same as `PROTOGUARD_LLM_TARGET=local`. |
| `--llm-target=cloud` | Sends chunks to your configured hosted API; prints a stderr banner. |
| `--llm-base-url` / `PROTOGUARD_LLM_BASE_URL` | Base URL (local default `http://127.0.0.1:11434/v1` for Ollama-compatible `/v1`). |
| `--llm-model` / `PROTOGUARD_LLM_MODEL` | Model id (local default `llama3.2` if unset). |
| `--llm-api-key` / `PROTOGUARD_LLM_API_KEY` | Optional bearer token. |
| `--llm-backend=openai_shape` | `POST .../chat/completions` (default). |
| `--llm-backend=ollama` | Native `POST .../api/chat` (base URL may omit `/v1`). |
| `--skip-llm` | No HTTP calls; JSON/classification only. |
| `--cache-dir` | Override disk cache (default `XDG_CACHE_HOME/protoguard` or `~/.cache/protoguard`). |

**Guardrail:** with `llm-target=local`, known vendor hosts such as `api.openai.com` are rejected even if set in env—use `--llm-target=cloud` for hosted APIs.

### Security defaults

- **`--max-spec-mb`** bounds how much of each input file is read (default **32**, CLI caps **512**).
- **`--allow-openapi-external-refs`** is **off** by default so OpenAPI cannot fetch remote `$ref` URLs unless you opt in.
- LLM base URLs must be **http/https**, cannot embed **user:password** in the URL, and block common **cloud metadata** endpoints; HTTP client uses **TLS 1.2+** and **does not follow redirects**.
- Disk cache uses restrictive permissions (**`0700`** dirs, **`0600`** files).

See **[SECURITY.md](./SECURITY.md)** for the full threat model and reporting process.

### Protobuf inputs

Build a descriptor set (example with `protoc`):

```bash
protoc --proto_path=. --descriptor_set_out=old.pb --include_imports a.proto
```

Compare `old.pb` vs `new.pb` with `--spec protobuf` or rely on `.pb` auto-detection.

**Bundled examples:** `testdata/proto_old.pb` vs `testdata/proto_new.pb`, built from `testdata/protobuf/v1/catalog.proto` and `testdata/protobuf/v2/catalog.proto` (same file name so diffs show message/RPC deltas, not unrelated file renames). Regenerate after editing:

```bash
make testdata-proto
```

Quick check:

```bash
protoguard diff --old testdata/proto_old.pb --new testdata/proto_new.pb --spec protobuf --skip-llm --format markdown
```

## Development

```bash
go test ./... -count=1
go vet ./...
```

### Ollama integration tests

Requires Ollama running locally with a model available (e.g. `ollama pull llama3.2`):

```bash
PROTOGUARD_TEST_OLLAMA=1 go test ./internal/llm/... ./internal/run/... -v -count=1 -run 'Integration_Ollama|Integration_Run'
```

Optional: `PROTOGUARD_TEST_OLLAMA_BASE_URL`, `PROTOGUARD_TEST_OLLAMA_NATIVE_URL`, `PROTOGUARD_TEST_OLLAMA_MODEL`.

## License

Apache License 2.0. See [LICENSE](./LICENSE).

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md). Security-sensitive issues: [SECURITY.md](./SECURITY.md).
