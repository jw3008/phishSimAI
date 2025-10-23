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

func GetGroups(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "clariphish-session")
	userID := session.Values["user_id"]

	rows, err := db.DB.Query(`
		SELECT id, name, created_at
		FROM groups WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	groups := []models.Group{}
	for rows.Next() {
		var g models.Group
		rows.Scan(&g.ID, &g.Name, &g.CreatedAt)

		// Get target count
		var count int
		db.DB.QueryRow("SELECT COUNT(*) FROM targets WHERE group_id = ?", g.ID).Scan(&count)
		g.Targets = make([]models.Target, 0)

		groups = append(groups, g)
	}

	respondJSON(w, groups)
}

func GetGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var g models.Group
	err := db.DB.QueryRow(`
		SELECT id, name, user_id, created_at
		FROM groups WHERE id = ?`, id).Scan(&g.ID, &g.Name, &g.UserID, &g.CreatedAt)

	if err == sql.ErrNoRows {
		respondError(w, "Group not found", http.StatusNotFound)
		return
	}
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get targets
	rows, err := db.DB.Query(`
		SELECT id, first_name, last_name, email, position
		FROM targets WHERE group_id = ?`, id)
	if err == nil {
		defer rows.Close()
		g.Targets = []models.Target{}
		for rows.Next() {
			var t models.Target
			rows.Scan(&t.ID, &t.FirstName, &t.LastName, &t.Email, &t.Position)
			t.GroupID = id
			g.Targets = append(g.Targets, t)
		}
	}

	respondJSON(w, g)
}

func CreateGroup(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "clariphish-session")
	userID := session.Values["user_id"]

	var g models.Group
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result, err := db.DB.Exec(`
		INSERT INTO groups (name, user_id, created_at)
		VALUES (?, ?, ?)`, g.Name, userID, time.Now())

	if err != nil {
		respondError(w, "Failed to create group", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	g.ID = int(id)

	// Add targets if provided
	if len(g.Targets) > 0 {
		for _, target := range g.Targets {
			db.DB.Exec(`
				INSERT INTO targets (first_name, last_name, email, position, group_id)
				VALUES (?, ?, ?, ?, ?)`,
				target.FirstName, target.LastName, target.Email, target.Position, id)
		}
	}

	respondJSON(w, g)
}

func UpdateGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var g models.Group
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	_, err := db.DB.Exec("UPDATE groups SET name = ? WHERE id = ?", g.Name, id)
	if err != nil {
		respondError(w, "Failed to update group", http.StatusInternalServerError)
		return
	}

	// Update targets
	if len(g.Targets) > 0 {
		// Delete existing targets
		db.DB.Exec("DELETE FROM targets WHERE group_id = ?", id)

		// Insert new targets
		for _, target := range g.Targets {
			db.DB.Exec(`
				INSERT INTO targets (first_name, last_name, email, position, group_id)
				VALUES (?, ?, ?, ?, ?)`,
				target.FirstName, target.LastName, target.Email, target.Position, id)
		}
	}

	g.ID = id
	respondJSON(w, g)
}

func DeleteGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	// Delete targets first
	db.DB.Exec("DELETE FROM targets WHERE group_id = ?", id)

	_, err := db.DB.Exec("DELETE FROM groups WHERE id = ?", id)
	if err != nil {
		respondError(w, "Failed to delete group", http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]bool{"success": true})
}
