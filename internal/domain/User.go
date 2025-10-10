package domain

type UserKind string

const (
	COMMON     UserKind = "common"
	SHOPKEEPER UserKind = "shopkeeper"
)

type User struct {
	ID       string
	Name     string
	LastName string
	Document string
	Email    string
	Password string
	Kind     UserKind
}
