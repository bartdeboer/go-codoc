# codoc

`codoc` exposes small, code-derived documentation records for Go packages. It
uses Go's own documentation model to retrieve package orientation, executable
workflows, explicit behavioral contracts, and focused public API symbols.

From inside a Go package, bare `codoc` renders the package narrative and runs
every valid Go test marked `//codoc:doc` once. The directive is an explicit
promise that the test is isolated, deterministic, safe, documented, and suitable
for automatic execution. If no documented tests exist, codoc says so without
treating their absence as a gap.

```go
// TestTrustedDeviceElevation explains the live architectural path.
//
//codoc:doc
func TestTrustedDeviceElevation(t *testing.T) { /* ... */ }
```

Use `//codoc:code omit` in the same comment to run a documented test without
rendering its body.

```sh
codoc
codoc package
codoc workflows
codoc workflow get-thread
codoc contracts
codoc contract get-thread/not-found
codoc symbol Client.GetThread
codoc search 'missing thread'
```

The current package is the default. Package context and output format are
global, orthogonal options that may appear anywhere:

```sh
codoc -C ./internal/thread
codoc --package ./internal/thread workflows
codoc symbol Client.GetThread --json
codoc --format json search GetThread
```

`codoc verify` explicitly runs package tests. Narrow verification is available
for a workflow or contract:

```sh
codoc verify workflow get-thread
codoc verify contract get-thread/not-found
```

Explicit record retrieval (`codoc package`, `workflow`, `contract`, `symbol`, and `search`) never runs tests and reports executable records as `not run`.

## Documentation sources

- Package comments provide orientation.
- Go `Example...` functions with `// Output:` provide executable workflows and expected output. Output-less examples are illustrative Go documentation, not codoc workflows.
- Public declarations and Go doc comments provide the API.
- `//codoc:contract <stable-id>` marks only a small set of deliberately
  documented behavioral promises. Ordinary implementation tests should remain
  unmarked and are not documentation records.

Codoc follows the files selected by `go list`, including the active platform
and build tags. It does not generate README files or maintain a parallel source
of truth.

## Installation

Beginning with v0.1.0, add codoc as a Go tool dependency and invoke it through
the Go toolchain:

```sh
go get -tool github.com/bartdeboer/go-codoc/cmd/codoc@v0.1.0
go tool codoc
```

## Development

Requires Go 1.24.

```sh
gofmt -w .
go test ./...
go vet ./...
git diff --check
```

Licensed under Apache-2.0. See `LICENSE` and `NOTICE`.
