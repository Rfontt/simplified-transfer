package transaction

import (
	"event-driven-architecture/internal/domain/user"
	"fmt"
)

type Transfer struct {
	TransactionID ID
	From          user.ID
	To            user.ID
}

type NewTransfer struct {
	userService     user.UserService
	transferService TransferService
}

func (transfer NewTransfer) Create(
	transactionID ID,
	from user.ID,
	to user.ID,
) error {
	userFrom, err := transfer.userService.GetOne(from)

	if err != nil {
		return err
	}

	if userFrom.Type == user.SHOPKEEPER {
		return fmt.Errorf("%s user type is not allowed", user.SHOPKEEPER)
	}

	err = transfer.transferService.Create(
		Transfer{
			TransactionID: transactionID,
			From:          from,
			To:            to,
		},
	)

	if err != nil {
		return err
	}

	// TODO(rfontt): emit transfer events here

	return nil
}
