package main

import (
	"invoice-generator-backend/config"
	"invoice-generator-backend/routes"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load shared defaults first; .env.local (if present) can override for local dev.
	_ = godotenv.Load()
	_ = godotenv.Overload(".env.local")

	config.ConnectDB()

	r := routes.SetupRoutes()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}