package http

import (
	"event-driven-architecture/internal/application/account/command"

	"github.com/gin-gonic/gin"
)

func NewRouter(createAccount command.CreateAccountUseCase) *gin.Engine {
	router := gin.Default()
	controller := NewAccountController(createAccount)
	router.POST("/accounts", controller.Create)
	return router
}
