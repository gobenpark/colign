# Colign

## Testing

- 코드 변경 시 반드시 `/tdd` 스킬을 사용하여 테스트 먼저 작성 (RED → GREEN → REFACTOR)
- Go 패키지는 단위 테스트 필수, 80% 이상 커버리지 목표
- 인증/보안 관련 코드(auth, apitoken, oauth, middleware)는 100% 커버리지 목표
- assertion: `testify` (assert/require) 사용
- interface mock: `mockgen`으로 생성 — 인터페이스를 **사용하는 쪽** 파일에 `//go:generate mockgen ...` 주석 추가
- DB/SQL mock: `sqlmock` 사용
- mock 생성: `go generate ./...`

## Build

- Go API: `CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /tmp/colign-api ./cmd/api`
- 크로스 컴파일 전 반드시 `go clean -cache` 실행
- Proto 생성: `cd proto && buf generate`

## Frontend

@web/CLAUDE.md
