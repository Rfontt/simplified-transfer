package user

import "fmt"

type AlreadyExistsError struct {
	Email string
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("user with email %s already exists", e.Email)
}
