# Colign

[![CI](https://github.com/gobenpark/colign/actions/workflows/ci.yml/badge.svg)](https://github.com/gobenpark/colign/actions/workflows/ci.yml)
[![License: AGPL-3.0](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Next.js](https://img.shields.io/badge/Next.js-16-000000?logo=next.js&logoColor=white)](https://nextjs.org)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

## Why Colign?

AI vibe coding has made individual developers incredibly productive. Claude Code, Cursor, Copilot — there are already powerful tools for writing code on your own.

But real software is never built alone.

Working in a team means aligning on **"what to build"** before writing any code. You gather requirements, discuss specs, define acceptance criteria, get reviews, and only then start implementing. Most AI development tools don't address this **upstream collaboration** — the discussion, alignment, and spec writing that happens before code.

That's where Colign comes in.

- Discuss and write specs together with AI
- Co-edit specs in real-time with your team
- Structure the alignment process with a Draft → Design → Review → Ready workflow
- Once specs are finalized, AI helps implement based on them

**AI that writes code already exists. Colign makes sure your team is looking at the same thing before the code gets written.**

## Spec-Driven Development

Colign follows a **Spec-Driven Development (SDD)** approach. When AI can generate code in minutes, the bottleneck shifts from "can we build it?" to "have we defined it correctly?"

### Two-layer spec architecture

Traditional PRDs tried to put everything in one 30-page document. Colign splits this into two layers:

**Project Memory** — Strategic context that rarely changes. Domain rules, business constraints, target users, and technical decisions. Written once, referenced by every Change.

**Structured Proposal** — Tactical spec for each Change. Lightweight, structured, and designed for both humans and AI agents to read:

| Section | Required | Purpose |
|---------|:--------:|---------|
| **Problem** | Yes | Why is this change needed? |
| **Scope** | Yes | What specifically will change? |
| **Out of Scope** | No | What is explicitly NOT part of this? |
| **Approach** | No | Technical direction and rationale |
| **Acceptance Criteria** | Yes | Given/When/Then scenarios |

### Two paths for AI integration

1. **Platform AI** — Use your own API key to generate structured specs directly in Colign
2. **External AI** — Connect Claude Code, Cursor, or other AI tools via MCP Server to read/write specs in Colign

> For design decisions and competitive analysis behind this structure, see [docs/structured-proposal.md](docs/structured-proposal.md).

## Getting Started

```bash
# Start all services (API + DB + Redis)
docker-compose up --build

# Run frontend (separate terminal)
cd web && npm install && npm run dev
```

Open http://localhost:3000 to access the app.

## Development

```bash
# Generate proto (Go + TypeScript)
cd proto && buf generate

# Run API server locally
go run ./cmd/api

# Run tests
go test ./...
```

## Prerequisites

- Go 1.26+
- Node.js 20+
- Docker & Docker Compose
- [buf](https://buf.build/docs/installation)

## License

This project is licensed under the [GNU Affero General Public License v3.0](LICENSE) (AGPL-3.0).
