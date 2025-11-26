package db

import "log"

// MigrateUserAPIKey adds gemini_api_key column to users table if it doesn't exist
func MigrateUserAPIKey() error {
	// Check if column exists
	var count int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='gemini_api_key'
	`).Scan(&count)

	if err != nil {
		return err
	}

	// Add column if it doesn't exist
	if count == 0 {
		log.Println("Adding gemini_api_key column to users table...")
		_, err = DB.Exec(`
			ALTER TABLE users ADD COLUMN gemini_api_key TEXT NULL
		`)
		if err != nil {
			return err
		}
		log.Println("✓ gemini_api_key column added successfully")
	}

	return nil
}
