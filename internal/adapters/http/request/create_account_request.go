package request

type CreateAccountRequest struct {
	OwnerID  string  `json:"owner_id" binding:"required"`
	Currency string  `json:"currency" binding:"required"`
	Balance  float64 `json:"balance"`
}
