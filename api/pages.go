package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/clariphish/clariphish/db"
	"github.com/clariphish/clariphish/models"
	"github.com/gorilla/mux"
)

func GetPages(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "clariphish-session")
	userID := session.Values["user_id"]

	rows, err := db.DB.Query(`
		SELECT id, name, created_at
		FROM pages WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	pages := []models.Page{}
	for rows.Next() {
		var p models.Page
		rows.Scan(&p.ID, &p.Name, &p.CreatedAt)
		pages = append(pages, p)
	}

	respondJSON(w, pages)
}

func GetPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var p models.Page
	var captureCredentials, capturePasswords int
	err := db.DB.QueryRow(`
		SELECT id, name, html, capture_credentials, capture_passwords, redirect_url, user_id, created_at
		FROM pages WHERE id = ?`, id).Scan(
		&p.ID, &p.Name, &p.HTML, &captureCredentials, &capturePasswords, &p.RedirectURL, &p.UserID, &p.CreatedAt)

	if err == sql.ErrNoRows {
		respondError(w, "Page not found", http.StatusNotFound)
		return
	}
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}

	p.CaptureCredentials = captureCredentials == 1
	p.CapturePasswords = capturePasswords == 1

	respondJSON(w, p)
}

func CreatePage(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "clariphish-session")
	userID := session.Values["user_id"]

	var p models.Page
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	captureCredentials := 0
	if p.CaptureCredentials {
		captureCredentials = 1
	}
	capturePasswords := 0
	if p.CapturePasswords {
		capturePasswords = 1
	}

	result, err := db.DB.Exec(`
		INSERT INTO pages (name, html, capture_credentials, capture_passwords, redirect_url, user_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		p.Name, p.HTML, captureCredentials, capturePasswords, p.RedirectURL, userID, time.Now())

	if err != nil {
		respondError(w, "Failed to create page", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	p.ID = int(id)

	respondJSON(w, p)
}

func UpdatePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var p models.Page
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	captureCredentials := 0
	if p.CaptureCredentials {
		captureCredentials = 1
	}
	capturePasswords := 0
	if p.CapturePasswords {
		capturePasswords = 1
	}

	_, err := db.DB.Exec(`
		UPDATE pages SET name = ?, html = ?, capture_credentials = ?, capture_passwords = ?, redirect_url = ?
		WHERE id = ?`, p.Name, p.HTML, captureCredentials, capturePasswords, p.RedirectURL, id)

	if err != nil {
		respondError(w, "Failed to update page", http.StatusInternalServerError)
		return
	}

	p.ID = id
	respondJSON(w, p)
}

func DeletePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	_, err := db.DB.Exec("DELETE FROM pages WHERE id = ?", id)
	if err != nil {
		respondError(w, "Failed to delete page", http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]bool{"success": true})
}
