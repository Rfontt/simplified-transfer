package transaction

type TransferService interface {
	GetOne(id ID) (Transfer, error)
	Create(transfer Transfer) (Transfer, error)
}
