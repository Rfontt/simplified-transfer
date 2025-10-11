package domain

type Balance struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}
