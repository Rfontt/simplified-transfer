package user

type UserService interface {
	GetOne(id ID) (User, error)
}
