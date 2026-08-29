package handler

import (
	"net/http"

	"event-driven-architecture/internal/adapters/http/request"
	"event-driven-architecture/internal/adapters/http/response"
	"event-driven-architecture/internal/application/user/command"

	"github.com/gin-gonic/gin"
)

type UserHTTPHandler struct {
	createUser command.CreateUserUseCase
}

func NewUserHTTPHandler(createUser command.CreateUserUseCase) *UserHTTPHandler {
	return &UserHTTPHandler{createUser: createUser}
}

func (h *UserHTTPHandler) Create(ctx *gin.Context) {
	var req request.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.createUser.Handle(ctx.Request.Context(), command.CreateUserCommand{
		FullName: req.FullName,
		Document: req.Document,
		Email:    req.Email,
		Password: req.Password,
		Type:     req.Type,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, response.NewCreateUserResponse(result))
}
