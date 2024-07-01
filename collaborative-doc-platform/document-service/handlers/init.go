package handlers

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB() error {
	err := godotenv.Load(".env")
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}
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

	

    createDocumentsTable := `
    CREATE TABLE IF NOT EXISTS documents (
        id SERIAL PRIMARY KEY,
        title TEXT,
        content JSONB,
        created_at TIMESTAMP,
        updated_at TIMESTAMP
    );`
    _, err = db.Exec(createDocumentsTable)
    if err != nil {
        return err
    }

    createChangesTable := `
    CREATE TABLE IF NOT EXISTS changes (
        id SERIAL PRIMARY KEY,
    document_id INTEGER NOT NULL REFERENCES documents(id),
    delta JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL
    );`
    _, err = db.Exec(createChangesTable)
    return err

}
