package main

import (
	"context"
	"log"

	apphttp "event-driven-architecture/internal/adapters/http"
	"event-driven-architecture/internal/adapters/postgres"
	"event-driven-architecture/internal/application/account/command"
	"event-driven-architecture/internal/config"
)

func main() {
	cfg := config.Load()

	db, err := postgres.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer db.Close()

	if err := postgres.Migrate(context.Background(), db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	accountRepo := postgres.NewAccountRepository(db)
	createAccount := command.NewCreateAccountCommandHandler(accountRepo)
	router := apphttp.NewRouter(createAccount)

	log.Printf("simplified-transfer listening on :%s", cfg.HTTPPort)
	if err := router.Run(":" + cfg.HTTPPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
