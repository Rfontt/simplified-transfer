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
