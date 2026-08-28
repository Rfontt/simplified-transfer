# ADR-0003: AP (availability + partition tolerance), eventual consistency

Status: accepted (recorded retrospectively on 2026-08-26)

## Context
Non-functional requirement: "consistency is not mandatory; the system must be available and partition-tolerant".

## Decision
Prioritize A + P of the CAP theorem. Eventual consistency between writes and read models is acceptable.

## Consequences
- The transfer itself is transactional (ACID) with refund on failure.
- Notification (SMS/email) may fail → use DLQ/retry.
- Reads may be slightly behind the written state.
