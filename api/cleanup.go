package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/clariphish/clariphish/db"
	"github.com/gorilla/mux"
)

// CleanupBadTimestamps removes events with invalid timestamps (0001-01-01)
func CleanupBadTimestamps(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignID := vars["id"]

	// Find events with bad timestamps
	rows, err := db.DB.Query(`
		SELECT id, campaign_target_id, message, time
		FROM events
		WHERE campaign_id = ? AND time < ?`, campaignID, "2020-01-01")

	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}

	badEvents := []map[string]interface{}{}
	for rows.Next() {
		var id, ctID int
		var message, timestamp string
		rows.Scan(&id, &ctID, &message, &timestamp)
		badEvents = append(badEvents, map[string]interface{}{
			"id":                 id,
			"campaign_target_id": ctID,
			"message":            message,
			"time":               timestamp,
		})
	}
	rows.Close()

	// Delete bad events
	result, err := db.DB.Exec(`
		DELETE FROM events
		WHERE campaign_id = ? AND time < ?`, campaignID, "2020-01-01")

	if err != nil {
		respondError(w, "Failed to cleanup", http.StatusInternalServerError)
		return
	}

	deleted, _ := result.RowsAffected()

	respondJSON(w, map[string]interface{}{
		"deleted":     deleted,
		"bad_events":  badEvents,
		"message":     fmt.Sprintf("Deleted %d events with invalid timestamps", deleted),
	})
}

// VerifyCampaignTimestamps shows all event timestamps for debugging
func VerifyCampaignTimestamps(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignID := vars["id"]

	rows, err := db.DB.Query(`
		SELECT e.id, e.campaign_target_id, e.message, e.time,
		       t.first_name, t.last_name, t.email
		FROM events e
		LEFT JOIN campaign_targets ct ON ct.id = e.campaign_target_id
		LEFT JOIN targets t ON t.id = ct.target_id
		WHERE e.campaign_id = ?
		ORDER BY e.id`, campaignID)

	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	events := []map[string]interface{}{}
	badCount := 0
	goodCount := 0

	for rows.Next() {
		var id, ctID int
		var message, firstName, lastName, email string
		var timestamp time.Time
		rows.Scan(&id, &ctID, &message, &timestamp, &firstName, &lastName, &email)

		isBad := timestamp.Year() < 2020
		if isBad {
			badCount++
		} else {
			goodCount++
		}

		events = append(events, map[string]interface{}{
			"id":                 id,
			"campaign_target_id": ctID,
			"message":            message,
			"time":               timestamp.Format("2006-01-02 15:04:05"),
			"time_unix":          timestamp.Unix(),
			"target":             fmt.Sprintf("%s %s (%s)", firstName, lastName, email),
			"is_bad":             isBad,
		})
	}

	respondJSON(w, map[string]interface{}{
		"total_events": len(events),
		"bad_events":   badCount,
		"good_events":  goodCount,
		"events":       events,
		"message":      fmt.Sprintf("Found %d events (%d bad, %d good)", len(events), badCount, goodCount),
	})
}
