package user

import "fmt"

type AlreadyExistsError struct {
	Email    string
	Document string
}

func (e *AlreadyExistsError) Error() string {
	switch {
	case e.Email != "":
		return fmt.Sprintf("user with email %s already exists", e.Email)
	case e.Document != "":
		return fmt.Sprintf("user with document %s already exists", e.Document)
	default:
		return "user already exists"
	}
}

type InvalidTypeError struct {
	Type string
}

func (e *InvalidTypeError) Error() string {
	return fmt.Sprintf("type %s is not valid (expected common or shopkeeper)", e.Type)
}
