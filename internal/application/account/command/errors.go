package command

import "errors"

// Application-level error sentinels exposed to driving adapters (e.g. HTTP).
// The handler translates domain/DB errors into these so the transport layer
// does not depend on domain types.
var (
	ErrAccountAlreadyExists = errors.New("account already exists for this owner")
	ErrOwnerNotFound        = errors.New("owner not found")
	ErrInvalidOwnerID       = errors.New("owner id must be a valid uuid")
	ErrInvalidCurrency      = errors.New("currency must not be empty")
	ErrInvalidBalance       = errors.New("balance must not be negative")
)
