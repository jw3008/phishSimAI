package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/clariphish/clariphish/db"
	"github.com/clariphish/clariphish/models"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// User Management (Admin Only)

// GetUsers returns all users (admin only)
func GetUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT id, username, role, created_at
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Username, &u.Role, &u.CreatedAt)
		if err != nil {
			continue
		}
		users = append(users, u)
	}

	respondJSON(w, users)
}

// CreateUser creates a new user (admin only)
func CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Username == "" {
		respondError(w, "Username is required", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		respondError(w, "Password is required", http.StatusBadRequest)
		return
	}

	if len(req.Password) < 6 {
		respondError(w, "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	// Validate role - only allow 'admin' or 'user' (principle of least privilege)
	if req.Role != "admin" && req.Role != "user" {
		respondError(w, "Invalid role. Must be 'admin' or 'user'", http.StatusBadRequest)
		return
	}

	// Check if username already exists
	var existingID int
	err := db.DB.QueryRow("SELECT id FROM users WHERE username = ?", req.Username).Scan(&existingID)
	if err == nil {
		respondError(w, "Username already exists", http.StatusConflict)
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Create user with specified role (least privilege - explicit role assignment)
	result, err := db.DB.Exec(`
		INSERT INTO users (username, password_hash, role, created_at)
		VALUES (?, ?, ?, ?)
	`, req.Username, string(hash), req.Role, time.Now())

	if err != nil {
		respondError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	userID, _ := result.LastInsertId()

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "User created successfully",
		"user_id": userID,
		"username": req.Username,
		"role":     req.Role,
	})
}

// GetUser returns a specific user (admin only)
func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var user models.User
	err := db.DB.QueryRow(`
		SELECT id, username, role, created_at
		FROM users WHERE id = ?
	`, id).Scan(&user.ID, &user.Username, &user.Role, &user.CreatedAt)

	if err == sql.ErrNoRows {
		respondError(w, "User not found", http.StatusNotFound)
		return
	}
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}

	respondJSON(w, user)
}

// UpdateUser updates a user's role or resets password (admin only)
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req struct {
		Role        string `json:"role"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate role if provided (principle of least privilege)
	if req.Role != "" && req.Role != "admin" && req.Role != "user" {
		respondError(w, "Invalid role. Must be 'admin' or 'user'", http.StatusBadRequest)
		return
	}

	// Prevent deleting the last admin
	if req.Role == "user" {
		var adminCount int
		db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&adminCount)

		var currentRole string
		db.DB.QueryRow("SELECT role FROM users WHERE id = ?", id).Scan(&currentRole)

		if currentRole == "admin" && adminCount <= 1 {
			respondError(w, "Cannot change role: at least one admin user must exist", http.StatusBadRequest)
			return
		}
	}

	// Update role if provided
	if req.Role != "" {
		_, err := db.DB.Exec("UPDATE users SET role = ? WHERE id = ?", req.Role, id)
		if err != nil {
			respondError(w, "Failed to update role", http.StatusInternalServerError)
			return
		}
	}

	// Update password if provided
	if req.NewPassword != "" {
		if len(req.NewPassword) < 6 {
			respondError(w, "Password must be at least 6 characters", http.StatusBadRequest)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			respondError(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		_, err = db.DB.Exec("UPDATE users SET password_hash = ? WHERE id = ?", string(hash), id)
		if err != nil {
			respondError(w, "Failed to update password", http.StatusInternalServerError)
			return
		}
	}

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "User updated successfully",
	})
}

// DeleteUser deletes a user (admin only)
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get current user from session
	currentUserID, _, err := getUserFromSession(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Prevent self-deletion
	if currentUserID == parseInt(id) {
		respondError(w, "Cannot delete your own account", http.StatusBadRequest)
		return
	}

	// Check if user is admin and if they're the last admin
	var role string
	err = db.DB.QueryRow("SELECT role FROM users WHERE id = ?", id).Scan(&role)
	if err == sql.ErrNoRows {
		respondError(w, "User not found", http.StatusNotFound)
		return
	}

	if role == "admin" {
		var adminCount int
		db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&adminCount)

		if adminCount <= 1 {
			respondError(w, "Cannot delete the last admin user", http.StatusBadRequest)
			return
		}
	}

	// Delete user
	_, err = db.DB.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		respondError(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "User deleted successfully",
	})
}

// Helper function to parse string to int
func parseInt(s string) int {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		}
	}
	return result
}
