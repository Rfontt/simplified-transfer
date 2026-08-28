# ADR-0001: DDD + CQRS + Event Sourcing architecture

Status: accepted (recorded retrospectively on 2026-08-26)

## Context
A study/implementation project for simplified transfers. The architecture decision was made via Event Storming (see README and Medium articles).

## Decision
Adopt Domain-Driven Design with two bounded contexts (User, Account), CQRS (commands vs queries separated) and Event Sourcing (immutable events as the source of truth).

## Consequences
- Events are the audit trail; aggregates are reconstructed from history.
- Separating writes/reads allows scaling and choosing distinct databases per side.
- Cost: higher complexity (projections, replay, eventual consistency).
