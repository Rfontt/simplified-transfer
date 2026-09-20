package account

import "fmt"

type AccountAlreadyExistsError struct {
	OwnerID string
}

func (e *AccountAlreadyExistsError) Error() string {
	return fmt.Sprintf("an account already exists for owner %s", e.OwnerID)
}

type OwnerNotFoundError struct {
	OwnerID string
}

func (e *OwnerNotFoundError) Error() string {
	return fmt.Sprintf("owner %s not found", e.OwnerID)
}

type AccountNotFoundError struct {
	AccountID string
}

func (e *AccountNotFoundError) Error() string {
	return fmt.Sprintf("account %s not found", e.AccountID)
}
