package command

type TransferMoneyCommand struct {
	PayerAccountID string
	PayeeAccountID string
	Value          float64
}
