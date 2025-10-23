package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/clariphish/clariphish/db"
	"github.com/clariphish/clariphish/models"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var user models.User
	err := db.DB.QueryRow("SELECT id, username, password_hash FROM users WHERE username = ?",
		req.Username).Scan(&user.ID, &user.Username, &user.PasswordHash)

	if err == sql.ErrNoRows {
		respondError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		respondError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	session, _ := store.Get(r, "clariphish-session")
	session.Values["user_id"] = user.ID
	session.Values["username"] = user.Username
	session.Save(r, w)

	respondJSON(w, map[string]interface{}{
		"success":  true,
		"message":  "Login successful",
		"user":     user.Username,
		"user_id":  user.ID,
	})
}

func Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "clariphish-session")
	session.Values["user_id"] = nil
	session.Options.MaxAge = -1
	session.Save(r, w)

	respondJSON(w, map[string]bool{"success": true})
}

func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "clariphish-session")
	userID, ok := session.Values["user_id"]
	if !ok {
		respondError(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	var user models.User
	err := db.DB.QueryRow("SELECT id, username FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Username)

	if err != nil {
		respondError(w, "User not found", http.StatusNotFound)
		return
	}

	respondJSON(w, user)
}
