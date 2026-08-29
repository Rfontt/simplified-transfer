package command

import "errors"

var (
	ErrAccountAlreadyExists = errors.New("account already exists for this owner")
	ErrOwnerNotFound        = errors.New("owner not found")
	ErrInvalidOwnerID       = errors.New("owner id must be a valid uuid")
)
