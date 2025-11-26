package api

import (
	"fmt"
	"net/http"

	"github.com/clariphish/clariphish/db"
	"github.com/gorilla/mux"
)

// GetCampaignEvents returns all events for a campaign (DEBUG ONLY)
func GetCampaignEvents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignID := vars["id"]

	rows, err := db.DB.Query(`
		SELECT e.id, e.campaign_target_id, e.email, e.time, e.message, e.details,
		       t.first_name, t.last_name, ct.rid
		FROM events e
		LEFT JOIN campaign_targets ct ON ct.id = e.campaign_target_id
		LEFT JOIN targets t ON t.id = ct.target_id
		WHERE e.campaign_id = ?
		ORDER BY e.time DESC`, campaignID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	html := `
<!DOCTYPE html>
<html>
<head>
	<title>Campaign Events Debug</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 20px; }
		table { border-collapse: collapse; width: 100%; }
		th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
		th { background-color: #4CAF50; color: white; }
		tr:nth-child(even) { background-color: #f2f2f2; }
		.event-opened { background-color: #e3f2fd; }
		.event-clicked { background-color: #fff3e0; }
		.event-submitted { background-color: #ffebee; }
		.event-reported { background-color: #e8f5e9; }
	</style>
</head>
<body>
	<h1>Campaign Events Debug</h1>
	<p>This page shows all raw events from the database for this campaign.</p>
	<table>
		<tr>
			<th>Event ID</th>
			<th>Time</th>
			<th>Target</th>
			<th>Email</th>
			<th>RID</th>
			<th>Message</th>
			<th>Details</th>
			<th>CT_ID</th>
		</tr>
`

	count := 0
	for rows.Next() {
		var id, ctID int
		var email, timestamp, message, details, firstName, lastName, rid string
		rows.Scan(&id, &ctID, &email, &timestamp, &message, &details, &firstName, &lastName, &rid)

		rowClass := ""
		switch message {
		case "Email Opened":
			rowClass = "event-opened"
		case "Clicked Link":
			rowClass = "event-clicked"
		case "Submitted Data":
			rowClass = "event-submitted"
		case "Reported Phishing":
			rowClass = "event-reported"
		}

		html += fmt.Sprintf(`
		<tr class="%s">
			<td>%d</td>
			<td>%s</td>
			<td>%s %s</td>
			<td>%s</td>
			<td>%s</td>
			<td><strong>%s</strong></td>
			<td>%s</td>
			<td>%d</td>
		</tr>
`, rowClass, id, timestamp, firstName, lastName, email, rid, message, details, ctID)
		count++
	}

	html += fmt.Sprintf(`
	</table>
	<p><strong>Total events: %d</strong></p>
	<p><a href="/dashboard">Back to Dashboard</a></p>
</body>
</html>
`, count)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
