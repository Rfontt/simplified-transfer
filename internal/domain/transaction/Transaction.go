package transaction

import "event-driven-architecture/internal/domain"

type Type string
type Status string

const (
	DEPOSIT  Type = "deposit"
	WITHDRAW Type = "withdraw"
	TRANSFER Type = "transfer"

	PENDING   Status = "pending"
	COMPLETED Status = "completed"
	FAILED    Status = "failed"
)

type Transaction struct {
	ID      string
	Type    Type
	Status  Status
	Balance domain.Balance
}

type Deposit struct {
	From string
	To   string
}
