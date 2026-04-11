package main

import (
	"invoice-generator-backend/config"
	"invoice-generator-backend/routes"

	"github.com/joho/godotenv"
)

func main() {
	// Load shared defaults first; .env.local (if present) can override for local dev.
	_ = godotenv.Load()
	_ = godotenv.Overload(".env.local")

	config.ConnectDB()

	r := routes.SetupRoutes()
	r.Run(":8080")
}