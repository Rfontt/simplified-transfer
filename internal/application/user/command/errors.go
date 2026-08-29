package command

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidFullName   = errors.New("full name must not be empty")
	ErrInvalidDocument   = errors.New("document must be a valid CPF or CNPJ")
	ErrInvalidEmail      = errors.New("email must not be empty")
	ErrInvalidPassword   = errors.New("password must not be empty")
	ErrInvalidType       = errors.New("type must be common or shopkeeper")
)
