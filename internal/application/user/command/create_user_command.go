package command

type CreateUserCommand struct {
	FullName string
	Document string
	Email    string
	Password string
	Type     string
}
