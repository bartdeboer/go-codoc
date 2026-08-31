# codoc

`codoc` exposes small, code-derived documentation records for Go packages.
It retrieves package orientation, executable workflows, explicit behavioral
contracts, and individual API symbols without dumping an entire package.

From inside a Go package, run `codoc` for its overview. Current-package
commands need no package argument:

```sh
codoc
codoc workflows
codoc contract get-thread/not-found
codoc symbol Client.GetThread
```

Pass a package only when inspecting another package, for example
`codoc package contract` from a module root. See `codoc help` for all commands.

## Development

Requires Go 1.24.

```sh
go test ./...
go vet ./...
```

Licensed under Apache-2.0. See `LICENSE` and `NOTICE`.
