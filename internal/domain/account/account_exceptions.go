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

type InvalidCurrencyError struct{}

func (e *InvalidCurrencyError) Error() string {
	return "currency must not be empty"
}

type InvalidBalanceError struct{}

func (e *InvalidBalanceError) Error() string {
	return "balance must not be negative"
}
