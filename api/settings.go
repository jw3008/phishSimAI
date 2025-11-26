package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/clariphish/clariphish/db"
)

type Setting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetSettings retrieves all settings (admin only)
func GetSettings(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT key, value, updated_at FROM settings")
	if err != nil {
		respondError(w, "Failed to fetch settings", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	settings := make(map[string]string)
	for rows.Next() {
		var key, value string
		var updatedAt time.Time
		if err := rows.Scan(&key, &value, &updatedAt); err != nil {
			continue
		}
		settings[key] = value
	}

	respondJSON(w, settings)
}

// UpdateSetting updates or creates a setting (admin only)
func UpdateSetting(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		respondError(w, "Key is required", http.StatusBadRequest)
		return
	}

	// Check if setting exists
	var existingValue string
	err := db.DB.QueryRow("SELECT value FROM settings WHERE key = ?", req.Key).Scan(&existingValue)

	if err == sql.ErrNoRows {
		// Insert new setting
		_, err = db.DB.Exec(
			"INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)",
			req.Key, req.Value, time.Now())
	} else {
		// Update existing setting
		_, err = db.DB.Exec(
			"UPDATE settings SET value = ?, updated_at = ? WHERE key = ?",
			req.Value, time.Now(), req.Key)
	}

	if err != nil {
		respondError(w, "Failed to update setting", http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]bool{"success": true})
}

// GetGeminiAPIKey retrieves the Gemini API key from settings
func GetGeminiAPIKey() string {
	var apiKey string
	err := db.DB.QueryRow("SELECT value FROM settings WHERE key = 'gemini_api_key'").Scan(&apiKey)
	if err != nil {
		return ""
	}
	return apiKey
}
