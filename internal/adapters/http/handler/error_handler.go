package handler

import (
	"errors"
	"net/http"

	"event-driven-architecture/internal/application/account/command"
	usercommand "event-driven-architecture/internal/application/user/command"
	"event-driven-architecture/internal/domain"

	"github.com/gin-gonic/gin"
)

func writeError(ctx *gin.Context, err error) {
	status, message := mapError(err)
	ctx.JSON(status, gin.H{"error": message})
}

func mapError(err error) (int, string) {
	var constraintErr *domain.ConstraintValidationError
	if errors.As(err, &constraintErr) {
		return http.StatusBadRequest, err.Error()
	}
	switch {
	case errors.Is(err, command.ErrAccountAlreadyExists),
		errors.Is(err, usercommand.ErrUserAlreadyExists):
		return http.StatusConflict, err.Error()
	case errors.Is(err, command.ErrOwnerNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, command.ErrInvalidOwnerID):
		return http.StatusBadRequest, err.Error()
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
