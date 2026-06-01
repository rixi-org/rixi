# rixi - Instruction for Agents

Rixi is going to be the future of golang framework for building web applications.
It will be an ai native web framework so make your decisions accordingly.

## Commands

- **Build**: `make build` (runs `go build`)
- **Install**: `make install` (runs `go install`, places binary in `$GOBIN`/`$GOPATH/bin`; run after each changes to create fresh binary)
- **Dev**: `rixi dev` (runs the project with auto-reload on `.go` file changes — uses `.dev-timestamp` polling, no external deps)
- **Format**: `go fmt` (run it after each code change)

## Go API inspection

- **`go doc`** — Use for already-imported or standard library packages. Fast, offline, shows exact API for what's in the module.
- **`gopls`** — LSP-driven. Powers editor hovers, completions, go-to-definition. Not directly invoked — your editor talks to it.
- **`pkgsite-cli`** — Use for **discovering** or **evaluating** packages you don't have yet. Searches pkg.go.dev: browse symbols, check versions, find importers, detect vulns.

## Philosophy

### Zero external dependencies

This project uses only the Go standard library.
Do not add external packages — zero supply chain attack surface.
In case of any dependency is needed like an ORM, search existing mature ORMs and replicate only the features we need.

### Single Binary

The eventual artifact is going to be a single binary.
Any files like html templates, static files, etc. are going to be compiled into the binary.

## Misc

- Read README.md to understand the philosophy of the project.
