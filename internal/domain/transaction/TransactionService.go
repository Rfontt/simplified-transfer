package transaction

type TransactionService interface {
	GetOne(id ID) (Transaction, error)
}
