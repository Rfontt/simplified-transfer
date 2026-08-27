package account

import "fmt"

// AccountAlreadyExistsError indicates that an account already exists for the
// given owner. One user may own at most one account.
type AccountAlreadyExistsError struct {
	OwnerID string
}

func (e *AccountAlreadyExistsError) Error() string {
	return fmt.Sprintf("an account already exists for owner %s", e.OwnerID)
}

// OwnerNotFoundError indicates that the account references an owner that does
// not exist.
type OwnerNotFoundError struct {
	OwnerID string
}

func (e *OwnerNotFoundError) Error() string {
	return fmt.Sprintf("owner %s not found", e.OwnerID)
}
