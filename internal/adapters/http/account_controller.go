package http

import (
	"net/http"

	"event-driven-architecture/internal/adapters/http/request"
	"event-driven-architecture/internal/adapters/http/response"
	"event-driven-architecture/internal/application/account/command"

	"github.com/gin-gonic/gin"
)

type AccountController struct {
	createAccount command.CreateAccountUseCase
}

func NewAccountController(createAccount command.CreateAccountUseCase) *AccountController {
	return &AccountController{createAccount: createAccount}
}

func (c *AccountController) Create(ctx *gin.Context) {
	var req request.CreateAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := c.createAccount.Handle(ctx.Request.Context(), command.CreateAccountCommand{
		OwnerID:  req.OwnerID,
		Currency: req.Currency,
		Balance:  req.Balance,
	})
	if err != nil {
		writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, response.NewCreateAccountResponse(result))
}
