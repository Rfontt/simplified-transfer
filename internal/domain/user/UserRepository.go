package user

import "context"

type UserRepository interface {
	ByID(ctx context.Context, id ID) (*User, error)
	Add(ctx context.Context, user *User) error
}
