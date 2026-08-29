package user

import "strings"

type Password string

func NewPassword(plain string) (Password, error) {
	if strings.TrimSpace(plain) == "" {
		return "", &InvalidPasswordError{}
	}
	return Password(plain), nil
}

func (p Password) String() string {
	return string(p)
}
