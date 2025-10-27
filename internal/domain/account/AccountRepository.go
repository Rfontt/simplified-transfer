package account

type AccountService interface {
	GetOne(id AccountID) (Account, error)
	Create(account *Account) error
}
