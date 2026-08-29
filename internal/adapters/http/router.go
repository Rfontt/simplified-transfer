package http

import (
	"event-driven-architecture/internal/adapters/http/handler"
	"event-driven-architecture/internal/application/account/command"
	usercommand "event-driven-architecture/internal/application/user/command"

	"github.com/gin-gonic/gin"
)

func NewRouter(createAccount command.CreateAccountUseCase, createUser usercommand.CreateUserUseCase) *gin.Engine {
	router := gin.Default()

	accountHandler := handler.NewAccountHTTPHandler(createAccount)
	router.POST("/accounts", accountHandler.Create)

	userHandler := handler.NewUserHTTPHandler(createUser)
	router.POST("/users", userHandler.Create)

	return router
}
