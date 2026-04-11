package main

import (
	"invoice-generator-backend/config"
	"invoice-generator-backend/routes"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// .env.local overrides .env — use it for local development
	_ = godotenv.Load(".env.local")
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env")
	}

	config.ConnectDB()

	r := routes.SetupRoutes()
	r.Run(":8080")
}