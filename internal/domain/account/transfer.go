package account

import (
	"time"

	"event-driven-architecture/internal/domain"
	"event-driven-architecture/internal/domain/user"

	"github.com/google/uuid"
)

type Transfer struct {
	ID     uuid.UUID
	From   AccountID
	To     AccountID
	Amount domain.MonetaryAmount
	Status AccountTransactionStatus

	events []domain.Event
}

func NewTransfer(id uuid.UUID, from, to *Account, payer *user.User, amount domain.MonetaryAmount) (*Transfer, error) {
	transfer := newPendingTransfer(id, from.ID, to.ID, amount)
	if err := transfer.validate(from, to, payer); err != nil {
		return transfer.failed(err.Error()), err
	}
	if err := from.Debit(amount); err != nil {
		return transfer.failed(err.Error()), err
	}
	to.Credit(amount)
	transfer.Status = COMPLETED
	transfer.record(AccountTransferSucceeded{
		ID:        id,
		CreatedAt: time.Now(),
		Amount:    amount.Value,
		From:      from.ID,
		To:        to.ID,
	})
	return transfer, nil
}

func NewFailedTransfer(id uuid.UUID, from, to AccountID, amount domain.MonetaryAmount, reason string) *Transfer {
	return newPendingTransfer(id, from, to, amount).failed(reason)
}

func (t *Transfer) UncommittedEvents() []domain.Event {
	return t.events
}

func (t *Transfer) validate(from, to *Account, payer *user.User) error {
	if !t.Amount.IsPositive() {
		return &domain.ConstraintValidationError{Field: "amount"}
	}
	if t.From == t.To {
		return &SameAccountTransferError{AccountID: t.From}
	}
	if from.Balance.Currency != to.Balance.Currency {
		return &CurrencyMismatchError{FromCurrency: from.Balance.Currency, ToCurrency: to.Balance.Currency}
	}
	if !payer.CanTransfer() {
		return &TransferNotAllowedError{UserType: payer.Type}
	}
	return nil
}

func (t *Transfer) failed(reason string) *Transfer {
	t.Status = FAILED
	t.record(AccountTransferFailed{
		ID:           t.ID,
		CreatedAt:    time.Now(),
		Amount:       t.Amount.Value,
		To:           t.To,
		From:         t.From,
		ErrorMessage: reason,
	})
	return t
}

func (t *Transfer) record(event domain.Event) {
	t.events = append(t.events, event)
}

func newPendingTransfer(id uuid.UUID, from, to AccountID, amount domain.MonetaryAmount) *Transfer {
	transfer := &Transfer{ID: id, From: from, To: to, Amount: amount, Status: PENDING}
	transfer.record(AccountTransferCreated{
		ID:        id,
		CreatedAt: time.Now(),
		Amount:    amount.Value,
		To:        to,
		From:      from,
	})
	return transfer
}
