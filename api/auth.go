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
	err := db.DB.QueryRow("SELECT id, username, password_hash, role FROM users WHERE username = ?",
		req.Username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role)

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
	session.Values["role"] = user.Role
	session.Save(r, w)

	respondJSON(w, map[string]interface{}{
		"success":  true,
		"message":  "Login successful",
		"user":     user.Username,
		"user_id":  user.ID,
		"role":     user.Role,
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
	err := db.DB.QueryRow("SELECT id, username, role FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Username, &user.Role)

	if err != nil {
		respondError(w, "User not found", http.StatusNotFound)
		return
	}

	respondJSON(w, user)
}
