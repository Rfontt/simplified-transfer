# HTTP API (CONTEXT)

Recorded: 2026-08-29

## Source of truth
**`docs/openapi.yaml` is the single source of truth for the HTTP API contract** — all endpoints, request/response schemas and error responses live there. When adding or changing endpoints, update `docs/openapi.yaml` first, then implement.

## Base
- Dev server: `http://localhost:8080` (Gin, port from `HTTP_PORT`).
- Error contract: single JSON object `{"error": "<message>"}`.

## Current endpoints
| Method | Path | Summary | Status codes |
|---|---|---|---|
| POST | `/accounts` | Create an account for an existing user (`{owner_id, currency, balance}`) | 201, 400, 404, 409, 500 |
| POST | `/users` | Create a user (`{full_name, document, email, password, type}`) | 201, 400, 409, 500 |

See `docs/openapi.yaml` for the full schemas (`CreateAccountRequest`, `Account`, `CreateUserRequest`, `User`, `Error`).

## Error mapping (HTTP adapter, `error_handler.go`)
- 400 — malformed body (bind failure) or validation: invalid owner_id UUID, empty currency, negative balance; for users: empty fields, invalid document, invalid type.
- 404 — owner referenced by `owner_id` does not exist.
- 409 — account already exists for this owner; user with this document/email already exists.
- 500 — unknown error; generic message, internals never leaked.

## Planned (per README / business rules; NOT in openapi.yaml yet)
- `POST /transfer` with `{ "value": 100.0, "payer": 4, "payee": 15 }` (transfer authorization + notification flows).
