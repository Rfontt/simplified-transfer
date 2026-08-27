package command

// CreateAccountCommand is an immutable request to create a new account for an
// existing user. It carries only the data required to perform the operation.
type CreateAccountCommand struct {
	OwnerID  string
	Currency string
	Balance  float64
}
