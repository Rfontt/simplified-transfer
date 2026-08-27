package http

import (
	"errors"
	"net/http"

	"event-driven-architecture/internal/application/account/command"

	"github.com/gin-gonic/gin"
)

// AccountController is a driving adapter exposing account use cases over HTTP.
// It depends only on the application port, not on domain types.
type AccountController struct {
	createAccount command.CreateAccountUseCase
}

func NewAccountController(createAccount command.CreateAccountUseCase) *AccountController {
	return &AccountController{createAccount: createAccount}
}

type createAccountRequest struct {
	OwnerID  string  `json:"owner_id" binding:"required"`
	Currency string  `json:"currency" binding:"required"`
	Balance  float64 `json:"balance"`
}

type createAccountResponse struct {
	ID       string  `json:"id"`
	OwnerID  string  `json:"owner_id"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
}

func (c *AccountController) Create(ctx *gin.Context) {
	var req createAccountRequest
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
		c.writeError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, createAccountResponse{
		ID:       result.ID,
		OwnerID:  result.OwnerID,
		Currency: result.Currency,
		Balance:  result.Balance,
	})
}

func (c *AccountController) writeError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, command.ErrAccountAlreadyExists):
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, command.ErrOwnerNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, command.ErrInvalidOwnerID),
		errors.Is(err, command.ErrInvalidCurrency),
		errors.Is(err, command.ErrInvalidBalance):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
