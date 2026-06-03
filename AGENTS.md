# rixi - Instruction for Agents

Rixi is a **CLI tool that generates MVC web applications** for Go. It is not an MVC project itself — it is a framework generator.

The generated projects follow an MVC-lite pattern (model, view, controller) with an embedded web server, HTML templates, and a development server with hot reload.

Rixi is an AI-native web framework. Make your decisions accordingly.

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

### Low external dependencies

Rixi minimizes dependencies but does not require zero.
- The CLI tool itself should use only the Go standard library.
- Generated projects may use minimal, well-audited dependencies (e.g., `modernc.org/sqlite` for embedded databases, `go-redis` for caching).
- Prefer pure Go implementations to avoid CGo and external system libraries.
- Justify every new dependency. If a standard library solution exists, use it.

### Single Binary

The generated project's artifact is a single binary.
HTML templates, static files, etc. are compiled into the binary via `go:embed`.

### Provider pattern for extensibility

Core interfaces (`Store`, `Cache`) are defined in the rixi package.
Implementations live in separate driver packages (`driver/sqlite`, `cache/redis`, etc.).
Users import only what they need — no bloat in the final binary.

## Commits

Follow `type(scope): subject` format.

- Subject is lowercase, no period, imperative mood ("add" not "added")
- Keep subject under 72 characters
- One logical change per commit

## Misc

- Read README.md to understand the philosophy of the project.
