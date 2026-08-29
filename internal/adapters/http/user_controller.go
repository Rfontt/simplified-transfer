package http

import (
	"net/http"

	"event-driven-architecture/internal/adapters/http/request"
	"event-driven-architecture/internal/adapters/http/response"
	"event-driven-architecture/internal/application/user/command"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	createUser command.CreateUserUseCase
}

func NewUserController(createUser command.CreateUserUseCase) *UserController {
	return &UserController{createUser: createUser}
}

func (c *UserController) Create(ctx *gin.Context) {
	var req request.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := c.createUser.Handle(ctx.Request.Context(), command.CreateUserCommand{
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
