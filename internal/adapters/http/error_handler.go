package http

import (
	"errors"
	"net/http"

	"event-driven-architecture/internal/application/account/command"
	usercommand "event-driven-architecture/internal/application/user/command"

	"github.com/gin-gonic/gin"
)

func writeError(ctx *gin.Context, err error) {
	status, message := mapError(err)
	ctx.JSON(status, gin.H{"error": message})
}

func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, command.ErrAccountAlreadyExists),
		errors.Is(err, usercommand.ErrUserAlreadyExists):
		return http.StatusConflict, err.Error()
	case errors.Is(err, command.ErrOwnerNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, command.ErrInvalidOwnerID),
		errors.Is(err, command.ErrInvalidCurrency),
		errors.Is(err, command.ErrInvalidBalance),
		errors.Is(err, usercommand.ErrInvalidFullName),
		errors.Is(err, usercommand.ErrInvalidDocument),
		errors.Is(err, usercommand.ErrInvalidEmail),
		errors.Is(err, usercommand.ErrInvalidPassword),
		errors.Is(err, usercommand.ErrInvalidType):
		return http.StatusBadRequest, err.Error()
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
