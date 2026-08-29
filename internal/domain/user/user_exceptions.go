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

type InvalidFullNameError struct{}

func (e *InvalidFullNameError) Error() string {
	return "full name must not be empty"
}

type InvalidDocumentError struct {
	Document string
}

func (e *InvalidDocumentError) Error() string {
	return fmt.Sprintf("document %s is not a valid CPF or CNPJ", e.Document)
}

type InvalidEmailError struct{}

func (e *InvalidEmailError) Error() string {
	return "email must not be empty"
}

type InvalidPasswordError struct{}

func (e *InvalidPasswordError) Error() string {
	return "password must not be empty"
}

type InvalidTypeError struct {
	Type string
}

func (e *InvalidTypeError) Error() string {
	return fmt.Sprintf("type %s is not valid (expected common or shopkeeper)", e.Type)
}
