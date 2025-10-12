package transaction

import "event-driven-architecture/internal/domain/user"

type Deposit struct {
	TransactionID ID
	UserID        user.ID
}
