package transaction

import "event-driven-architecture/internal/domain"

type Type string
type Status string

type ID domain.AggregateID

const (
	DEPOSIT  Type = "deposit"
	WITHDRAW Type = "withdraw"
	TRANSFER Type = "transfer"

	PENDING   Status = "pending"
	COMPLETED Status = "completed"
	FAILED    Status = "failed"
)

type Transaction struct {
	ID      ID
	Type    Type
	Status  Status
	Balance domain.Balance
}

func NewTransaction(id ID, transactionType Type, status Status, balance domain.Balance) Transaction {
	return Transaction{
		ID:      id,
		Type:    transactionType,
		Status:  status,
		Balance: balance,
	}
}
