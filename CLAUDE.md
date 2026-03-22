# Colign

## Testing

- Always use `/tdd` skill when changing code — write tests first (RED → GREEN → REFACTOR)
- Unit tests required for all Go packages, target 80%+ coverage
- Auth/security code (auth, apitoken, oauth, middleware) targets 100% coverage
- Assertions: use `testify` (assert/require)
- Interface mocks: generate with `mockgen` — place `//go:generate mockgen ...` in the file that **uses** the interface
- DB/SQL mocks: use `sqlmock`
- Generate mocks: `go generate ./...`

## Code Quality

- Always handle error returns. Never use `_, _ =` to discard errors — check or log them
- Logging: use `log/slog` (structured logging). Do not use `fmt.Println`, `log.Printf`, or third-party loggers
- Must pass golangci-lint (errcheck) — if an error must be ignored, add an explicit comment explaining why
- goimports: group imports as stdlib → external packages → internal packages (separated by blank lines)
- commitlint: conventional commits format, subject must start lowercase

## Build

- Go API: `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /tmp/colign-api ./cmd/api`
- Always run `go clean -cache` before cross-compilation
- Proto generation: `cd proto && buf generate`

## Frontend

@web/CLAUDE.md
