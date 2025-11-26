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

func GetTemplates(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "clariphish-session")
	userID := session.Values["user_id"]

	rows, err := db.DB.Query(`
		SELECT id, name, subject, created_at
		FROM templates WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	templates := []models.Template{}
	for rows.Next() {
		var t models.Template
		rows.Scan(&t.ID, &t.Name, &t.Subject, &t.CreatedAt)
		templates = append(templates, t)
	}

	respondJSON(w, templates)
}

func GetTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var t models.Template
	err := db.DB.QueryRow(`
		SELECT id, name, subject, text, html, user_id, created_at
		FROM templates WHERE id = ?`, id).Scan(
		&t.ID, &t.Name, &t.Subject, &t.Text, &t.HTML, &t.UserID, &t.CreatedAt)

	if err == sql.ErrNoRows {
		respondError(w, "Template not found", http.StatusNotFound)
		return
	}
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}

	respondJSON(w, t)
}

func CreateTemplate(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "clariphish-session")
	userID := session.Values["user_id"]

	var t models.Template
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result, err := db.DB.Exec(`
		INSERT INTO templates (name, subject, text, html, user_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		t.Name, t.Subject, t.Text, t.HTML, userID, time.Now())

	if err != nil {
		respondError(w, "Failed to create template", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	t.ID = int(id)

	respondJSON(w, t)
}

func UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var t models.Template
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	_, err := db.DB.Exec(`
		UPDATE templates SET name = ?, subject = ?, text = ?, html = ?
		WHERE id = ?`, t.Name, t.Subject, t.Text, t.HTML, id)

	if err != nil {
		respondError(w, "Failed to update template", http.StatusInternalServerError)
		return
	}

	t.ID = id
	respondJSON(w, t)
}

func DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	_, err := db.DB.Exec("DELETE FROM templates WHERE id = ?", id)
	if err != nil {
		respondError(w, "Failed to delete template", http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]bool{"success": true})
}
