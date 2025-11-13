package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run setup_api_key.go <your-api-key>")
		os.Exit(1)
	}

	apiKey := os.Args[1]

	// Open database
	db, err := sql.Open("sqlite3", "./clariphish.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Check if database exists and has tables
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='settings'").Scan(&count)
	if err != nil {
		log.Fatal("Database not initialized. Please start the application first to create the database.")
	}

	if count == 0 {
		log.Fatal("Settings table not found. Please start the application first to create the database.")
	}

	// Check if API key already exists
	var existingKey string
	err = db.QueryRow("SELECT value FROM settings WHERE key = 'gemini_api_key'").Scan(&existingKey)

	if err == sql.ErrNoRows {
		// Insert new API key
		_, err = db.Exec(
			"INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)",
			"gemini_api_key", apiKey, time.Now())
		if err != nil {
			log.Fatal("Failed to insert API key:", err)
		}
		fmt.Println("✓ Gemini API key configured successfully!")
	} else if err != nil {
		log.Fatal("Failed to check existing API key:", err)
	} else {
		// Update existing API key
		_, err = db.Exec(
			"UPDATE settings SET value = ?, updated_at = ? WHERE key = ?",
			apiKey, time.Now(), "gemini_api_key")
		if err != nil {
			log.Fatal("Failed to update API key:", err)
		}
		fmt.Println("✓ Gemini API key updated successfully!")
	}

	fmt.Println("\nYou can now use the AI features in the application.")
	fmt.Println("The application will be available at: http://localhost:3333")
}
