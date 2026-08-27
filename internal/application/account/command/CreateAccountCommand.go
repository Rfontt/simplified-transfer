package command

type CreateAccountCommand struct {
	OwnerID  string
	Currency string
	Balance  float64
}
