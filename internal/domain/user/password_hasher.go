package user

type PasswordHasher interface {
	Hash(plain string) (string, error)
}
