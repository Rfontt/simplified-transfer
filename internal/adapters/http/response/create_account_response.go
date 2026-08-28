package response

import "event-driven-architecture/internal/application/account/command"

type CreateAccountResponse struct {
	ID       string  `json:"id"`
	OwnerID  string  `json:"owner_id"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
}

func NewCreateAccountResponse(result *command.CreateAccountResult) CreateAccountResponse {
	return CreateAccountResponse{
		ID:       result.ID,
		OwnerID:  result.OwnerID,
		Currency: result.Currency,
		Balance:  result.Balance,
	}
}
