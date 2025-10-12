package user

type UserRepostiory interface {
	GetOne(id ID) (User, error)
	Save(User) error
}
