package handler

import (
	"net/http"

	"event-driven-architecture/internal/adapters/http/request"
	"event-driven-architecture/internal/adapters/http/response"
	"event-driven-architecture/internal/application/account/command"

	"github.com/gin-gonic/gin"
)

type AccountHTTPHandler struct {
	createAccount command.CreateAccountUseCase
}

func NewAccountHTTPHandler(createAccount command.CreateAccountUseCase) *AccountHTTPHandler {
	return &AccountHTTPHandler{createAccount: createAccount}
}

func (h *AccountHTTPHandler) Create(ctx *gin.Context) {
	var req request.CreateAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.createAccount.Handle(ctx.Request.Context(), command.CreateAccountCommand{
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
