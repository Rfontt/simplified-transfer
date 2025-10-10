package domain

type UserKind string

const (
	COMMON     UserKind = "common"
	SHOPKEEPER UserKind = "shopkeeper"
)

type User struct {
	ID       string
	Name     string
	Document string
	Email    string
	Password string
	Kind     UserKind
}
