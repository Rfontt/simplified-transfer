package user

import "strings"

type Email string

func NewEmail(value string) (Email, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", &InvalidEmailError{}
	}
	return Email(trimmed), nil
}

func (e Email) String() string {
	return string(e)
}
