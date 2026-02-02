# Repository Guidelines

Terma is a Go module for terminal UI widgets plus many runnable demos. Use this guide as the default contributor reference.

## Project Structure & Module Organization
- Core library code lives at repo root in `*.go` (package `terma`).
- Runnable demos and examples live in `cmd/` (for example `cmd/todo-app`, `cmd/list-demo`).
- Additional demos/examples are top-level folders like `*-demo`, `*-example`, and numbered tutorials (`01-hello-world`, etc.).
- Tests are `*_test.go` at the root and in subpackages. Snapshot goldens live in `testdata/` and output artifacts in `snapshot-output/`.
- Documentation sources live in `docs/` with `mkdocs.yml`.

## Build, Test, and Development Commands
- `go test ./...` runs the full test suite.
- `UPDATE_SNAPSHOTS=1 go test ./...` updates snapshot golden files.
- `go test -run TestName ./...` runs a specific test.
- `go run ./cmd/todo-app` runs the main todo demo.
- `go run ./cmd/list-demo` runs the list demo.  

## Coding Style & Naming Conventions
- Use standard Go formatting (`gofmt`); tabs are expected.
- Exported identifiers use `CamelCase`; unexported use `camelCase`.
- File names follow Go conventions (short, descriptive, `*_test.go` for tests).
- Keep widget IDs stable and explicit when focus or state persistence matters (see `cmd/` demos for patterns).
- Do not mutate reactive state (Signals or AnySignals) inside `Build()` methods. Setting signal values during `Build()` triggers re-renders and can cause infinite refresh loops. Do updates in handlers, effects, or setup code instead.

## Testing Guidelines
- Tests use Go’s `testing` package and `stretchr/testify`.
- Snapshot tests render UI to SVG and compare against `testdata/<TestName>.svg`.
- When UI changes are intentional, update snapshots via `UPDATE_SNAPSHOTS=1`.
- Keep snapshot test names stable; they map directly to golden filenames.

## Commit & Pull Request Guidelines
- Recent history uses short, imperative summaries (for example: “Use autocomplete in edit mode”).
- Keep commits focused and prefer one logical change per commit.
- PRs should include a concise summary, test results, and note any snapshot updates.
- If UI output changes, include a brief description and reference the updated snapshots.

## Debugging & Configuration Tips
- `TERMA_DEBUG_OVERLAY=1` shows the live render overlay and last render cause.
