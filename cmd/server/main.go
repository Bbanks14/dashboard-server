package main

import (
	"github.com/Bbanks14/dashboard-server/internal/api"
	"github.com/Bbanks14/internal/api"
	"github.com/Bbanks14/internal/db"
)

func main() {
	db.Connect()
	// Run Goose migrations: goose -dir ./internal/db/migrations up
	router := api.SetupRouter()
	api.AddSearchRoutes(router)
	router.Run(":8080")
}
