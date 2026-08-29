package domain

import "fmt"

type ConstraintValidationError struct {
	Field string
}

func (e *ConstraintValidationError) Error() string {
	return fmt.Sprintf("invalid value for field %q", e.Field)
}
