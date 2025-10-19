package transaction

type DepositService interface {
	GetOne(id ID) (Deposit, error)
	Create(deposit Deposit) error
}
