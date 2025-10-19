package transaction

type TransactionService interface {
	Create(Transaction) error
	GetOne(id ID) (Transaction, error)
}
