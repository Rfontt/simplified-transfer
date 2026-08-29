package user

import "strings"

type FullName string

func NewFullName(value string) (FullName, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", &InvalidFullNameError{}
	}
	return FullName(trimmed), nil
}

func (n FullName) String() string {
	return string(n)
}
