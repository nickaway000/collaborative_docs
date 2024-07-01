package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func InitDB() error {
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	jwtKey = []byte(os.Getenv("JWT_SECRET"))

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("error connecting to the database: %w", err)
	}

	errPing := db.Ping()
	if errPing != nil {
		return fmt.Errorf("error pinging the database: %w", errPing)
	}

	log.Println("Successfully connected to the database")
	return nil
}
