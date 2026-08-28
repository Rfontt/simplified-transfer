# AGENTS.md — event-driven-architecture project

Project-specific instructions for the harness. Full domain context is in CLAUDE.md (read it).

## First step
Read `.ai/index.md` and the relevant memory files before acting. Engineering memory lives in `.ai/` (decisions, architecture, incidents, lessons, context).

## How to verify (before claiming success)
- `go test ./...`
- `go build ./cmd/simplified-transfer/`
- `go vet ./...` and `go fmt ./...`

## Project state (important)
- Early stage: skeletal domain/application. Much is stubbed (persistence, event store, HTTP, adapters, projections).
- Do NOT assume something is implemented — check the code before claiming.

## Conventions
- Go, DDD + CQRS + Event Sourcing. Events are the source of truth.
- Business rules and development order: see CLAUDE.md.
