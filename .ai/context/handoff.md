---
status: open
owner: pi
created: 2026-09-20
---

> Single-use: the next session that reads this must act on it and then mark it
> `status: claimed` (append `claimed_by` + `claimed_at`). Do not reuse a claimed handoff.

# Handoff — transfer logical slice (specs/001-transfer-endpoint)

Workflow state: `.pi/workflow.json` → `slug=001-transfer-endpoint`, `phase=review`.
In `review` the source is read-only; only `.ai/` may be written. `workflow action=advance`
closes `review → done` (no human approval needed for that step).

## Done

- Full spec → plan → tasks chain written and human-approved:
  `specs/001-transfer-endpoint/{spec.md,plan.md,tasks.md}` (26 tasks).
- Implemented all 26 tasks — the **logical** transfer slice (ADR-0008): no HTTP, no Kafka, no
  persistence, no notification; all externals are ports with hand-rolled fakes.
  - `internal/domain/account/transfer.go` rewritten: `NewTransfer(id, from, to, payer, amount)`
    enforces every rule and moves the money via `Account.Debit`/`Credit`;
    `NewFailedTransfer(...)` records an attempt refused before execution; refused attempts
    return a **non-nil FAILED transfer together with the typed error**; `UncommittedEvents()`
    exposes the collected events.
  - `Account.HasBalance` fixed to `>=` (transferring the whole balance is legal), `Debit`
    refuses a non-ACTIVE account and an insufficient balance, `Credit` adds.
  - New: `domain.Event` + `EventName()` on the three transfer events;
    `AccountTransferSucceeded` gained `To`; `transfer_exceptions.go` (5 typed errors);
    `transfer_authorizer.go` (`TransferAuthorizer` port); `AccountNotFoundError`.
  - `transfer_service.go` **deleted** (dead interface) and `TransferRepository.Add` replaced by
    `Save(ctx, transfer, from, to)` — the single write seam / future transaction boundary.
  - Application: `TransferMoneyCommand`, `TransferMoneyResult`, `TransferMoneyUseCase`,
    `TransferMoneyCommandHandler` (ports: account repo, user repo, transfer repo, authorizer,
    publisher), `TransferEventPublisher`, 11 transfer sentinels + `mapTransferError`.
  - Postgres: `AccountRepository.ByID` translates `sql.ErrNoRows` → `AccountNotFoundError`.
- Tests: 12 new domain tests, 9 new handler tests, 1 new sqlmock test.
- Verification RUN (not claimed): `go build ./cmd/simplified-transfer/`, `go vet ./...`,
  `go test ./...` all PASS; rules `domain-purity` and `gofmt` PASS. Plus a mutation check:
  reverting `>=` to `>` made `TestAccount_HasBalanceExactAmount` and
  `TestNewTransfer_ExactBalance` fail (then reverted).
- `.pi/gates.json` created (verify commands + static rules), so `workflow action=verify` runs the
  real project checks instead of passing vacuously.
- Memory: `.ai/decisions/0008-logical-transfer-slice.md`,
  `.ai/lessons/validate-amount-currency-in-the-domain.md`, and updated
  `.ai/architecture/{domain,application,adapters}.md`,
  `.ai/context/{development,testing}.md`.

## Failed / pending

- **Review finding #1 is unresolved (important, DDD-01/DDD-04).** The domain compares the two
  accounts' currencies (`transfer.go` `validate`) but **never the amount's own currency against
  the payer account**, and `Account.Debit` subtracts `amount.Value` regardless of
  `amount.Currency`. The rule "the transfer currency is the payer's currency" exists only in
  `transfer_money_command_handler.go` (`amount := domain.MonetaryAmount{Currency: from.Balance.Currency, ...}`).
  Fix = add the amount-currency check to `NewTransfer.validate` + `TestNewTransfer_AmountCurrencyMismatch`.
- **Harness patch not applied.** The planned `/approve` auto-continue (calling
  `pi.sendUserMessage(...)` after the state flip in `~/.pi/agent/extensions/workflow-gate.ts`)
  was **blocked**: `safety-gate: edit outside project root is blocked`. Still true that after
  `/approve` nothing runs until the human sends another message.
- **Nothing is committed.** All work is in the worktree: 8 new Go files, 1 deleted, 8 modified,
  plus untracked `specs/`, `.pi/`, `.ai/decisions/0008-...`, `.ai/lessons/validate-...`.
- **Smaller review suggestions, not acted on:** `loadAccount` can panic if a port returns
  `(nil, nil)`; `Account.Credit` has no status guard (BLOCKED/CLOSED accounts can receive);
  `fakeTransferRepository.payerBalances`/`payeeBalances` are recorded but never asserted.

## Next

1. `workflow action=advance` to close `review → done` (writes become allowed), then apply the
   finding #1 fix: in `internal/domain/account/transfer.go` add to `validate`
   `if from.Balance.Currency != t.Amount.Currency { return &CurrencyMismatchError{FromCurrency: from.Balance.Currency, ToCurrency: t.Amount.Currency} }`,
   add `TestNewTransfer_AmountCurrencyMismatch` to `internal/domain/account/transfer_test.go`,
   run `workflow action=verify`, and update ADR-0008 (move the gap from "open" to fixed).
2. Commit on a branch (`feat/transfer-logical-slice`): source + `specs/` + `.ai/`; propose
   gitignoring `.pi/workflow.json` (transient state) while keeping `.pi/gates.json`. Do not push
   without being asked.
3. Next slice (deferred, in spec US4): `POST /transfer` HTTP route + `docs/openapi.yaml` entry
   (body `{value, payer, payee}` with **account UUIDs**), balance persistence with the atomic
   `TransferRepository.Save` implementation, authorization adapter, notification + DLQ.

## Open questions

- Fix review finding #1 now (as step 1 above) or make it the first task of the HTTP slice? Rita
  was asked; the recommendation is to fix now.
- Commit/branch as proposed in step 2, or leave the worktree uncommitted?
- Apply the `guard.json` change (`boundary: "project"` blocks harness edits — add
  `"allowPaths": ["/Users/ritadecassiafontenele/.pi/agent/extensions"]`) so the `/approve`
  auto-continue patch can be applied?
- Business question: may a non-ACTIVE account **receive** a transfer? (`Account.Credit` currently
  allows it; the spec only restricts the payer.)
- Money stays `float64` (ADR-0004 TODO: migrate to a decimal type) — decide before persisting
  balances.

## Decisions / lessons

- `.ai/decisions/0008-logical-transfer-slice.md` — the whole slice: logical-only scope, domain
  factory owns the rules, non-nil FAILED transfer + typed error, uncommitted events in the domain
  with the publisher as an application port, single `Save` write seam, `TransferService`/`NewTransfer`
  struct deleted, sentinel mapping so HTTP never imports the account context, and the open
  currency invariant gap.
- `.ai/lessons/validate-amount-currency-in-the-domain.md` — validate the consumed value object's
  own fields, not only the cross-entity comparisons.
- `.ai/lessons/fail-fast-ordering-when-moving-validation.md` — still relevant: the transfer
  handler authorizes **before** the domain validates the amount (deliberate trade-off, recorded in
  `plan.md` Alternatives); revisit if the gateway becomes rate-limited/expensive.
- `.ai/context/code-review.md` — the ruleset used for the review (findings reference `[DDD-01]`,
  `[DDD-04]`, `[PROCESS-03]`).
- Episodic memory `harness:workflow-gate` (expires 2026-11-30) — `/approve` only advances when
  `pendingApproval` is set, which only `workflow action=advance` writes, not `action=check`.
