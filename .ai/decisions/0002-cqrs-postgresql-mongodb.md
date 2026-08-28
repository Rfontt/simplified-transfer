# ADR-0002: PostgreSQL for commands, MongoDB for queries

Status: accepted (recorded retrospectively on 2026-08-26)

## Context
CQRS separates writes from reads. Requirement: consistency is not mandatory; availability and partition tolerance are.

## Decision
- Commands (transactional writes) → PostgreSQL.
- Queries (read models, eventually consistent) → MongoDB.

## Consequences
- ACID writes where it matters (transactional transfer with refund on failure).
- Optimized, scalable reads with eventual consistency.
- Two databases to operate (NOT yet integrated into the code).
