package domain

type MonetaryAmount struct {
	Currency string
	Value    float64
}

func (m MonetaryAmount) IsPositive() bool {
	return m.Value > 0
}

func (m MonetaryAmount) IsGreaterThanOrEqual(other MonetaryAmount) bool {
	return m.Value >= other.Value
}
