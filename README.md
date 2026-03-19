# Colign

A Spec-Driven Development (SDD) workflow platform where developers and non-developers collaboratively discuss and write specs with AI.

## Architecture

```
┌──────────────────┐         ┌──────────────────────┐
│    Next.js 15     │ Connect │     Go + Gin          │
│    (Frontend)     │◄──────►│     (API Server)      │
│                   │ (.proto)│                       │
│  - React 19       │        │  - uptrace/bun (ORM)  │
│  - Tiptap + Y.js  │        │  - connectrpc/connect │
│  - shadcn/ui      │        │  - JWT + OAuth2       │
└──────────────────┘         └───────┬──────────────┘
                                     │
  ┌──────────────────┐              │
  │  Hocuspocus       │  Y.js       │
  │  (Node sidecar)   │◄────────────┘
  └──────────────────┘
                              ┌──────────┬──────────┐
                              │PostgreSQL│  Redis   │
                              └──────────┴──────────┘
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | Next.js 15, React 19, TypeScript, Tailwind CSS, shadcn/ui |
| Editor | Tiptap (ProseMirror) + Y.js (CRDT) |
| API | Connect (buf.build) - gRPC-compatible with JSON support |
| Backend | Go, Gin, uptrace/bun |
| Auth | JWT + OAuth2 (GitHub, Google) |
| Realtime | Hocuspocus (Y.js server), gorilla/websocket |
| AI | Claude API (streaming), MCP Server |
| Database | PostgreSQL, Redis |

## Prerequisites

- Go 1.21+
- Node.js 20+
- Docker & Docker Compose
- [buf](https://buf.build/docs/installation)

## Getting Started

```bash
# Start databases
make up

# Run migrations
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/colign?sslmode=disable" up

# Generate proto code
make proto

# Run API server
make run

# Run frontend (separate terminal)
make web-dev
```

## Project Structure

```
.
├── cmd/
│   ├── api/          # API server entrypoint
│   └── mcp/          # MCP server entrypoint
├── internal/
│   ├── auth/         # Authentication (JWT, OAuth)
│   ├── project/      # Project & Change management
│   ├── workflow/     # Workflow engine (state machine)
│   ├── document/     # Spec editor backend
│   ├── collaboration/# Realtime collaboration
│   ├── chat/         # AI chat
│   ├── specgen/      # AI spec generation
│   ├── task/         # Task tracking
│   ├── mcp/          # MCP server
│   ├── models/       # Database models
│   ├── config/       # Configuration
│   ├── middleware/    # Gin middlewares
│   ├── server/       # Server setup
│   ├── database/     # Database connection
│   ├── cache/        # Redis client
│   └── email/        # Email sending
├── proto/            # Protobuf definitions
├── gen/              # Generated Go code
├── migrations/       # SQL migrations (golang-migrate)
├── web/              # Next.js frontend
├── hocuspocus/       # Y.js collaboration server
├── docker-compose.yml
└── Makefile
```

## Development

```bash
# Generate proto (Go + TypeScript)
make proto

# Run tests
make test

# Build binaries
make build

# Lint proto files
make proto-lint
```

## License

This project is licensed under the [GNU Affero General Public License v3.0](LICENSE) (AGPL-3.0).
