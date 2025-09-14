# CLAUDE.md

# Repository Guidelines

## Project Structure & Module Organization
- Go module root with primary entry in `main.go`; helpers in `helpers.go`, UI/menu in `menu.go`, BluOS XML types in `structs.go`.
- Binaries live in `bin/` (e.g., `bin/blueos.8s.gobin`) for SwiftBar.
- Env files: `.env` (local) and `.env.example` (template). The plugin reads `.env` from `SWIFTBAR_PLUGINS_PATH`.
- Docs and tooling: `README.md`, `GOLANG.md`, and utility scripts in repo root.

## Build, Test, and Development Commands
- `go build -o bin/blueos.8s.gobin .` — build the SwiftBar plugin binary.
- `./run_format.sh` — run `gofmt -w .` to format the code.
- `./run_lint.sh` — run `golangci-lint run --fix ./...`.
- `./run_test.sh` — run `go test -v ./...`.
- Example SwiftBar run: place the built binary in your `SWIFTBAR_PLUGINS_PATH` and ensure it is executable.

## Coding Style & Naming Conventions
- Follow Go idioms; format with `gofmt` and lint with `golangci-lint`.
- Use descriptive file names (`menu.go`, `helpers.go`); exported identifiers use CamelCase.
- Avoid deprecated packages (`io/ioutil`) and `log.Fatal`/`log.Panic` per `GOLANG.md`.

## Testing Guidelines
- Framework: standard `testing` package; add `_test.go` alongside code.
- Prefer table-driven tests for helpers and parsing (e.g., XML decode in `helpers.go`/`structs.go`).
- Run tests via `./run_test.sh`. Aim to cover core behaviors (status parsing, volume math, URL building).

## Commit & Pull Request Guidelines
- Use Conventional Commits found in history: `feat(menu): ...`, `fix(volume): ...`.
- PRs should include: clear summary, linked issue (if any), before/after notes or screenshots for UI changes, and steps to validate (build/run commands).
- Keep changes focused; update docs when behavior or flags change.

## Security & Configuration Tips
- Required env in `SWIFTBAR_PLUGINS_PATH/.env`: `BLUE_URL` (e.g., `http://192.168.1.101:11000`), `BLUE_WIFI` (SSID).
- Do not commit secrets; use `.env.example` to document keys.
- Prefer static IP for the BluOS device to avoid breakage.

## Additional Instructions

    - @./GOLANG.md

    - @./GODOC.md
