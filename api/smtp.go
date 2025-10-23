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

func GetSMTPConfigs(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "clariphish-session")
	userID := session.Values["user_id"]

	rows, err := db.DB.Query(`
		SELECT id, name, host, username, from_address, ignore_cert_errors, created_at
		FROM smtp WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	configs := []models.SMTP{}
	for rows.Next() {
		var s models.SMTP
		var ignoreCert int
		rows.Scan(&s.ID, &s.Name, &s.Host, &s.Username, &s.FromAddress, &ignoreCert, &s.CreatedAt)
		s.IgnoreCertErrors = ignoreCert == 1
		configs = append(configs, s)
	}

	respondJSON(w, configs)
}

func GetSMTPConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var s models.SMTP
	var ignoreCert int
	err := db.DB.QueryRow(`
		SELECT id, name, host, username, password, from_address, user_id, ignore_cert_errors, created_at
		FROM smtp WHERE id = ?`, id).Scan(
		&s.ID, &s.Name, &s.Host, &s.Username, &s.Password, &s.FromAddress, &s.UserID, &ignoreCert, &s.CreatedAt)

	if err == sql.ErrNoRows {
		respondError(w, "SMTP config not found", http.StatusNotFound)
		return
	}
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}

	s.IgnoreCertErrors = ignoreCert == 1
	respondJSON(w, s)
}

func CreateSMTPConfig(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "clariphish-session")
	userID := session.Values["user_id"]

	var s models.SMTP
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ignoreCert := 0
	if s.IgnoreCertErrors {
		ignoreCert = 1
	}

	result, err := db.DB.Exec(`
		INSERT INTO smtp (name, host, username, password, from_address, user_id, ignore_cert_errors, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		s.Name, s.Host, s.Username, s.Password, s.FromAddress, userID, ignoreCert, time.Now())

	if err != nil {
		respondError(w, "Failed to create SMTP config", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	s.ID = int(id)

	respondJSON(w, s)
}

func UpdateSMTPConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var s models.SMTP
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ignoreCert := 0
	if s.IgnoreCertErrors {
		ignoreCert = 1
	}

	_, err := db.DB.Exec(`
		UPDATE smtp SET name = ?, host = ?, username = ?, password = ?, from_address = ?, ignore_cert_errors = ?
		WHERE id = ?`, s.Name, s.Host, s.Username, s.Password, s.FromAddress, ignoreCert, id)

	if err != nil {
		respondError(w, "Failed to update SMTP config", http.StatusInternalServerError)
		return
	}

	s.ID = id
	respondJSON(w, s)
}

func DeleteSMTPConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	_, err := db.DB.Exec("DELETE FROM smtp WHERE id = ?", id)
	if err != nil {
		respondError(w, "Failed to delete SMTP config", http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]bool{"success": true})
}
