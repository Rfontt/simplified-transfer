package domain

type AggregateEvent[T any] interface {
	Handle(event T) error
}
