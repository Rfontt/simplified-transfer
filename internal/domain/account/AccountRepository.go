package account

import "github.com/google/uuid"

type AccountService interface {
	GetOne(id AccountID) (Account, error)
	GetOwnerBalance(ownerId uuid.UUID) (Account, error)
	Create(account *Account) error
}
