# Lesson: validate the value object's own fields, not only the ones you compare

Problem: `NewTransfer` validated that the payer and payee accounts hold the **same currency**,
but never that the transfer's `MonetaryAmount.Currency` matches the account it debits. The rule
"the transfer currency is the payer account's currency" was only true because the single caller
(`TransferMoneyCommandHandler`) happened to build the amount from `from.Balance.Currency`.

Root cause: the validation compared two *aggregates* (a cross-account rule) and never the
**value object against the aggregate it mutates**. Reviewing the rule list ("different accounts,
same currency, positive amount, COMMON payer") felt complete, so the missing half — the amount's
own currency — was invisible.

Consequence if left: a second entry point (HTTP adapter, retry job) could record a USD transfer
against BRL accounts; `Account.Debit` subtracts `Amount.Value` without looking at the currency,
so the domain would accept it.

How to avoid it again: for every value object that a domain operation consumes, ask "which of its
fields does the operation actually use, and is each one validated against the target?" A rule list
that only mentions cross-entity comparisons is incomplete until the consumed value object's fields
are checked against the entity too. When a rule can only be satisfied by one specific caller, that
is the smell: it belongs in the aggregate.

Found in: review of the logical transfer slice (ADR-0008), `internal/domain/account/transfer.go`.
