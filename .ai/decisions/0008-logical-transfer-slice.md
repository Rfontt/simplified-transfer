# ADR-0008: Logical transfer slice — domain factory owns the rules, events are collected in the domain

Status: accepted (2026-09-20)

## Context
`POST /transfer` is a documented but unbuilt requirement. The domain had a transfer flow
(`NewTransfer` struct holding repositories plus a `TransferService`) that never loaded the
payee, never validated the amount, returned untyped errors and never raised the
`AccountTransfer*` events it declared. Persistence, authorization and notification
infrastructure do not exist in the repository.

## Decision
- The slice is **logical only**: domain rules + domain events + one application use case. No
  HTTP route, no Kafka, no event store, no balance persistence, no notification.
- **Domain factory** `account.NewTransfer(id, from, to, payer, amount)` enforces every rule
  (amount positive, payer different from payee, same currency, payer is COMMON, account ACTIVE,
  sufficient balance) and moves the money through `Account.Debit`/`Account.Credit`.
  `HasBalance` compares with `>=`, so transferring the whole balance is legal.
- A refused attempt returns a **non-nil FAILED transfer together with the typed error**: the
  refused attempt is the audit record (`AccountTransferCreated` + `AccountTransferFailed`).
- **Uncommitted events**: the domain collects events on the transfer (`UncommittedEvents()`) and
  the application handler publishes them through the `TransferEventPublisher` port. The domain
  never holds a publisher.
- `domain.Event` (`EventName() string`) is introduced in the root package and implemented by the
  three transfer events only; the remaining events are a follow-up.
- **Authorization** is an outbound port (`account.TransferAuthorizer`) called by the handler
  before any balance change; `NewFailedTransfer` records the refusal. The real gateway adapter
  is deferred.
- **Single write seam**: `TransferRepository.Add` was replaced by
  `Save(ctx, transfer, from, to)` — one call per attempt, and the place where the future
  Postgres adapter must be atomic (DDD-07). No balance is persisted in this slice.
- `TransferService` and the `NewTransfer` struct were **deleted** (dead after the rework).
- Business-rule refusals are mapped to application sentinels (`ErrInsufficientBalance`,
  `ErrSameAccountTransfer`, `ErrCurrencyMismatch`, `ErrAccountNotTransactable`,
  `ErrTransferNotAllowed`); field violations keep flowing as `ConstraintValidationError`.
- `AccountRepository.ByID` translates `sql.ErrNoRows` into `AccountNotFoundError`, so the
  not-found contract holds for the real adapter and not only for the fakes.

## Consequences
- The transfer rules live in one place and are proven by pure domain tests; the handler has no
  rule branch.
- The service still does not expose transfers: nothing is wired in `main.go` and only the read
  path has a real implementation.
- Known gaps: balance persistence and its atomicity ADR, concurrency safety, idempotency,
  authorization adapter, notification + DLQ, broker publishing/event store, `EventName()` for
  the remaining events, migrating `NewDeposit`/`DepositService` to this pattern.
- **Open invariant gap found in review (2026-09-20)**: the domain compares the two accounts'
  currencies (`transfer.go` validate) but never compares the transfer's own amount currency
  against the payer account, and `Account.Debit` subtracts `amount.Value` regardless of
  `amount.Currency`. Today the rule "the transfer currency is the payer account's currency"
  lives only in `TransferMoneyCommandHandler` (mapping the command value into a
  `MonetaryAmount`). Any other caller (the upcoming HTTP adapter, a retry job) could build an
  amount whose currency differs from the account and the domain would accept it. Fix belongs in
  `NewTransfer` validation (plus a `TestNewTransfer_AmountCurrencyMismatch`) before a second
  entry point exists. See `.ai/lessons/validate-amount-currency-in-the-domain.md`.
