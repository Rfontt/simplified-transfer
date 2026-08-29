package account

type AccountTransactionStatus string

const (
	PENDING   AccountTransactionStatus = "PENDING"
	COMPLETED AccountTransactionStatus = "COMPLETED"
	FAILED    AccountTransactionStatus = "FAILED"
)
