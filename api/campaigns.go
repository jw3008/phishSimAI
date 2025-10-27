package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/clariphish/clariphish/db"
	"github.com/clariphish/clariphish/mailer"
	"github.com/clariphish/clariphish/models"
	"github.com/gorilla/mux"
)

func GetCampaigns(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "clariphish-session")
	userID := session.Values["user_id"]

	rows, err := db.DB.Query(`
		SELECT id, name, status, created_date, launch_date, completed_date
		FROM campaigns WHERE user_id = ? ORDER BY created_date DESC`, userID)
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	campaigns := []models.Campaign{}
	for rows.Next() {
		var c models.Campaign
		var launchDate, completedDate sql.NullTime
		rows.Scan(&c.ID, &c.Name, &c.Status, &c.CreatedDate, &launchDate, &completedDate)
		if launchDate.Valid {
			c.LaunchDate = &launchDate.Time
		}
		if completedDate.Valid {
			c.CompletedDate = &completedDate.Time
		}

		// Get stats
		c.Stats = getCampaignStats(c.ID)
		campaigns = append(campaigns, c)
	}

	respondJSON(w, campaigns)
}

func GetCampaign(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var c models.Campaign
	var launchDate, completedDate sql.NullTime
	err := db.DB.QueryRow(`
		SELECT id, name, status, created_date, launch_date, completed_date,
		       template_id, page_id, smtp_id, url, user_id
		FROM campaigns WHERE id = ?`, id).Scan(
		&c.ID, &c.Name, &c.Status, &c.CreatedDate, &launchDate, &completedDate,
		&c.TemplateID, &c.PageID, &c.SMTPID, &c.URL, &c.UserID)

	if err == sql.ErrNoRows {
		respondError(w, "Campaign not found", http.StatusNotFound)
		return
	}
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}

	if launchDate.Valid {
		c.LaunchDate = &launchDate.Time
	}
	if completedDate.Valid {
		c.CompletedDate = &completedDate.Time
	}

	// Get campaign results
	c.Results = getCampaignResults(id)
	c.Stats = getCampaignStats(id)

	respondJSON(w, c)
}

func CreateCampaign(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "clariphish-session")
	userID := session.Values["user_id"]

	var c models.Campaign
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	result, err := db.DB.Exec(`
		INSERT INTO campaigns (name, status, template_id, page_id, smtp_id, url, user_id, created_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		c.Name, "draft", c.TemplateID, c.PageID, c.SMTPID, c.URL, userID, time.Now())

	if err != nil {
		respondError(w, "Failed to create campaign", http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	c.ID = int(id)

	// Add targets from groups
	if len(c.Groups) > 0 {
		for _, groupID := range c.Groups {
			_, err := db.DB.Exec(`
				INSERT INTO campaign_targets (campaign_id, target_id, rid, status)
				SELECT ?, id, lower(hex(randomblob(16))), 'scheduled'
				FROM targets WHERE group_id = ?`, id, groupID)
			if err != nil {
				respondError(w, "Failed to add targets", http.StatusInternalServerError)
				return
			}
		}
	}

	// Launch campaign if requested
	if c.Status == "launched" {
		go mailer.LaunchCampaign(int(id))
		db.DB.Exec("UPDATE campaigns SET status = 'launched', launch_date = ? WHERE id = ?", time.Now(), id)
	}

	respondJSON(w, c)
}

func UpdateCampaign(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	var c models.Campaign
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	_, err := db.DB.Exec(`
		UPDATE campaigns
		SET name = ?, template_id = ?, page_id = ?, smtp_id = ?, url = ?
		WHERE id = ?`, c.Name, c.TemplateID, c.PageID, c.SMTPID, c.URL, id)

	if err != nil {
		respondError(w, "Failed to update campaign", http.StatusInternalServerError)
		return
	}

	c.ID = id
	respondJSON(w, c)
}

func DeleteCampaign(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	_, err := db.DB.Exec("DELETE FROM campaigns WHERE id = ?", id)
	if err != nil {
		respondError(w, "Failed to delete campaign", http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]bool{"success": true})
}

func CompleteCampaign(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	_, err := db.DB.Exec("UPDATE campaigns SET status = 'completed', completed_date = ? WHERE id = ?",
		time.Now(), id)
	if err != nil {
		respondError(w, "Failed to complete campaign", http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]bool{"success": true})
}

func getCampaignResults(campaignID int) []models.Result {
	rows, err := db.DB.Query(`
		SELECT t.id, t.first_name, t.last_name, t.email, t.position, ct.status, ct.send_date
		FROM campaign_targets ct
		JOIN targets t ON t.id = ct.target_id
		WHERE ct.campaign_id = ?`, campaignID)

	if err != nil {
		return []models.Result{}
	}
	defer rows.Close()

	results := []models.Result{}
	for rows.Next() {
		var r models.Result
		var sendDate sql.NullTime
		rows.Scan(&r.ID, &r.FirstName, &r.LastName, &r.Email, &r.Position, &r.Status, &sendDate)
		if sendDate.Valid {
			r.SendDate = sendDate.Time
		}

		// Get event times
		db.DB.QueryRow(`
			SELECT MIN(CASE WHEN message = 'Email Opened' THEN time END),
			       MIN(CASE WHEN message = 'Clicked Link' THEN time END),
			       MIN(CASE WHEN message = 'Submitted Data' THEN time END),
			       MIN(CASE WHEN message = 'Reported Phishing' THEN time END)
			FROM events WHERE campaign_target_id = ?`, r.ID).
			Scan(&r.OpenDate, &r.ClickDate, &r.SubmitDate, &r.ReportDate)

		results = append(results, r)
	}

	return results
}

func getCampaignStats(campaignID int) *models.Stats {
	stats := &models.Stats{}

	db.DB.QueryRow("SELECT COUNT(*) FROM campaign_targets WHERE campaign_id = ?", campaignID).
		Scan(&stats.Total)

	db.DB.QueryRow("SELECT COUNT(*) FROM campaign_targets WHERE campaign_id = ? AND status != 'scheduled'", campaignID).
		Scan(&stats.Sent)

	db.DB.QueryRow(`
		SELECT COUNT(DISTINCT campaign_target_id)
		FROM events WHERE campaign_id = ? AND message = 'Email Opened'`, campaignID).
		Scan(&stats.Opened)

	db.DB.QueryRow(`
		SELECT COUNT(DISTINCT campaign_target_id)
		FROM events WHERE campaign_id = ? AND message = 'Clicked Link'`, campaignID).
		Scan(&stats.Clicked)

	db.DB.QueryRow(`
		SELECT COUNT(DISTINCT campaign_target_id)
		FROM events WHERE campaign_id = ? AND message = 'Submitted Data'`, campaignID).
		Scan(&stats.Submitted)

	db.DB.QueryRow(`
		SELECT COUNT(DISTINCT campaign_target_id)
		FROM events WHERE campaign_id = ? AND message = 'Reported Phishing'`, campaignID).
		Scan(&stats.Reported)

	if stats.Sent > 0 {
		stats.OpenRate = (stats.Opened * 100) / stats.Sent
		stats.ClickRate = (stats.Clicked * 100) / stats.Sent
		stats.SubmitRate = (stats.Submitted * 100) / stats.Sent
		stats.ReportRate = (stats.Reported * 100) / stats.Sent
	}

	return stats
}

// GetCampaignCredentials returns all submitted credentials for a campaign
func GetCampaignCredentials(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignID := vars["id"]

	rows, err := db.DB.Query(`
		SELECT e.time, t.first_name, t.last_name, t.email, e.details
		FROM events e
		JOIN campaign_targets ct ON ct.id = e.campaign_target_id
		JOIN targets t ON t.id = ct.target_id
		WHERE e.campaign_id = ? AND e.message = 'Submitted Data'
		ORDER BY e.time DESC`, campaignID)

	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type SubmittedData struct {
		Time        time.Time              `json:"time"`
		FirstName   string                 `json:"first_name"`
		LastName    string                 `json:"last_name"`
		Email       string                 `json:"email"`
		Credentials map[string]interface{} `json:"credentials"`
	}

	submissions := []SubmittedData{}
	for rows.Next() {
		var s SubmittedData
		var detailsJSON string
		rows.Scan(&s.Time, &s.FirstName, &s.LastName, &s.Email, &detailsJSON)

		// Parse JSON credentials
		json.Unmarshal([]byte(detailsJSON), &s.Credentials)
		submissions = append(submissions, s)
	}

	respondJSON(w, submissions)
}
