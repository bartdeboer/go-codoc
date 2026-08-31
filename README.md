# codoc

`codoc` exposes small, code-derived documentation records for Go packages. It
uses Go's own documentation model to retrieve package orientation, executable
workflows, explicit behavioral contracts, and focused public API symbols.

From inside a Go package:

```sh
codoc
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

Ordinary reading never runs tests and reports records as `not run`.

## Documentation sources

- Package comments provide orientation.
- Go `Example...` functions provide executable workflows and expected output.
- Public declarations and Go doc comments provide the API.
- `//codoc:contract <stable-id>` marks only a small set of deliberately
  documented behavioral promises. Ordinary implementation tests should remain
  unmarked and are not documentation records.

Codoc follows the files selected by `go list`, including the active platform
and build tags. It does not generate README files or maintain a parallel source
of truth.

## Development

Requires Go 1.24.

```sh
gofmt -w .
go test ./...
go vet ./...
git diff --check
```

Licensed under Apache-2.0. See `LICENSE` and `NOTICE`.
