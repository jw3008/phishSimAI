package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/clariphish/clariphish/db"
)

// GetUserSettings gets the current user's settings
func GetUserSettings(w http.ResponseWriter, r *http.Request) {
	userID, _, err := getUserFromSession(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var apiKey sql.NullString
	err = db.DB.QueryRow(`
		SELECT gemini_api_key FROM users WHERE id = ?
	`, userID).Scan(&apiKey)

	if err != nil {
		respondError(w, "Failed to get settings", http.StatusInternalServerError)
		return
	}

	// Don't send the full API key, just indicate if one is set
	hasAPIKey := apiKey.Valid && apiKey.String != ""

	respondJSON(w, map[string]interface{}{
		"has_api_key": hasAPIKey,
	})
}

// SaveUserAPIKey saves the user's personal Gemini API key
func SaveUserAPIKey(w http.ResponseWriter, r *http.Request) {
	userID, _, err := getUserFromSession(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		APIKey string `json:"api_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Update user's API key
	_, err = db.DB.Exec(`
		UPDATE users SET gemini_api_key = ? WHERE id = ?
	`, req.APIKey, userID)

	if err != nil {
		respondError(w, "Failed to save API key", http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "API key saved successfully",
	})
}

// GetUserAPIKey retrieves the user's Gemini API key
// This is used internally by other API functions
func GetUserAPIKey(userID int) string {
	var apiKey sql.NullString
	err := db.DB.QueryRow(`
		SELECT gemini_api_key FROM users WHERE id = ?
	`, userID).Scan(&apiKey)

	if err != nil || !apiKey.Valid {
		// Fallback to global API key from settings
		return GetGeminiAPIKey()
	}

	return apiKey.String
}
