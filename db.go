package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// DB represents the global persistent database connection pool
var DB *sql.DB

// InitDB configures the SQLite file storage connection safely
func InitDB() error {
	var err error
	
	// Establishes a local database file instance
	DB, err = sql.Open("sqlite", "./records.db")
	if err != nil {
		return fmt.Errorf("database connection establishment failed: %w", err)
	}

	// Verify datastore network availability
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("database availability verification failed: %w", err)
	}

	// Schema configuration matching Nigerian banking dispute parameters
	createTableSQL := `CREATE TABLE IF NOT EXISTS transaction_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sender_bank TEXT NOT NULL,
		receiver_bank TEXT NOT NULL,
		amount REAL NOT NULL,
		session_id TEXT NOT NULL UNIQUE,
		transfer_date DATETIME NOT NULL
	);`

	_, err = DB.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to provision transaction storage schema: %w", err)
	}

	fmt.Println("✔ Database infrastructure provisioned successfully.")
	return nil
}
