package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() {
	log.Println("DB INFO: Starting database connection")

	connStr := fmt.Sprintf(
  "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
  os.Getenv("DB_HOST"),
  os.Getenv("DB_PORT"),
  os.Getenv("DB_USERNAME"),
  os.Getenv("DB_PASSWORD"),
  os.Getenv("DB_NAME"),
  os.Getenv("DB_SSLMODE"),
)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Println("DB ERROR:", err)
		log.Fatal("DB connection error:", err)
	}

	err = db.Ping()
	if err != nil {
		log.Println("DB ERROR:", err)
		log.Fatal("DB not reachable:", err)
	}

	DB = db
	log.Println("DB SUCCESS: Connected to PostgreSQL")

	if err := runMigrations(); err != nil {
		log.Println("DB ERROR:", err)
		log.Fatal("DB migration error:", err)
	}
}