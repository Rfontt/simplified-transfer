package http

import (
	"event-driven-architecture/internal/application/account/command"
	usercommand "event-driven-architecture/internal/application/user/command"

	"github.com/gin-gonic/gin"
)

func NewRouter(createAccount command.CreateAccountUseCase, createUser usercommand.CreateUserUseCase) *gin.Engine {
	router := gin.Default()

	accountController := NewAccountController(createAccount)
	router.POST("/accounts", accountController.Create)

	userController := NewUserController(createUser)
	router.POST("/users", userController.Create)

	return router
}
