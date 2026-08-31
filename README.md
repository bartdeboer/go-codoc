# codoc

`codoc` exposes small, code-derived documentation records for Go packages.
It retrieves package orientation, executable workflows, explicit behavioral
contracts, and individual API symbols without dumping an entire package.

The project is at foundation stage. See `codoc help` for its command surface.

## Development

Requires Go 1.24.

```sh
go test ./...
go vet ./...
```

Licensed under Apache-2.0. See `LICENSE` and `NOTICE`.
