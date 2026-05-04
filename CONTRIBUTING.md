# Contributing

Thanks for helping improve ProtoGuard.

## Development

```bash
go test ./... -count=1
go vet ./...
```

Optional (matches CI):

```bash
go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...
```

Optional Ollama integration tests require a running server and model (see `README.md`).

## Pull requests

- Keep changes focused; match existing style and formatting (`gofmt`).
- Avoid logging or persisting secrets (`APIKey`, bearer tokens).
- For security-sensitive reports, see **[SECURITY.md](./SECURITY.md)** instead of filing a public issue.

## Module path (forks and mirrors)

The canonical Go module is **`github.com/codec404/protoguard`**. If you publish under a different repository URL, update the `module` line in **`go.mod`** and replace the import prefix **`github.com/codec404/protoguard/`** across `.go` files (for example with `perl`/`sed`, then `gofmt`), so consumers get a consistent import path.
