package main

import (
	"invoice-generator-backend/config"
	"invoice-generator-backend/routes"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Local: use .env, Production: use .env.local
	appEnv := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	goEnv := strings.ToLower(strings.TrimSpace(os.Getenv("GO_ENV")))
	envName := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
	isProduction := appEnv == "production" || goEnv == "production" || envName == "production"

	if isProduction {
		_ = godotenv.Overload(".env.local")
		gin.SetMode(gin.ReleaseMode)
	} else {
		_ = godotenv.Overload(".env")
	}

	config.ConnectDB()

	r := routes.SetupRoutes()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}