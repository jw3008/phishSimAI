package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/clariphish/clariphish/db"
)

// 1x1 transparent pixel
var trackingPixel, _ = base64.StdEncoding.DecodeString(
	"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==")

func TrackOpen(w http.ResponseWriter, r *http.Request) {
	rid := r.URL.Query().Get("rid")
	if rid == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Get campaign target info
	var campaignID, targetID int
	var email string
	err := db.DB.QueryRow(`
		SELECT ct.campaign_id, ct.target_id, t.email
		FROM campaign_targets ct
		JOIN targets t ON t.id = ct.target_id
		WHERE ct.rid = ?`, rid).Scan(&campaignID, &targetID, &email)

	if err == nil {
		// Check if already opened
		var count int
		db.DB.QueryRow(`
			SELECT COUNT(*) FROM events
			WHERE campaign_target_id = (SELECT id FROM campaign_targets WHERE rid = ?)
			AND message = 'Email Opened'`, rid).Scan(&count)

		if count == 0 {
			// Record the open event
			var ctID int
			db.DB.QueryRow("SELECT id FROM campaign_targets WHERE rid = ?", rid).Scan(&ctID)

			db.DB.Exec(`
				INSERT INTO events (campaign_id, campaign_target_id, email, time, message, details)
				VALUES (?, ?, ?, ?, 'Email Opened', '')`,
				campaignID, ctID, email, time.Now())

			// Update campaign target status
			db.DB.Exec("UPDATE campaign_targets SET status = 'opened' WHERE rid = ?", rid)
		}
	}

	// Return tracking pixel
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Write(trackingPixel)
}

func TrackClick(w http.ResponseWriter, r *http.Request) {
	rid := r.URL.Query().Get("rid")
	if rid == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Get campaign target info
	var campaignID, targetID, pageID int
	var email, redirectURL string
	err := db.DB.QueryRow(`
		SELECT ct.campaign_id, ct.target_id, t.email, c.page_id
		FROM campaign_targets ct
		JOIN targets t ON t.id = ct.target_id
		JOIN campaigns c ON c.id = ct.campaign_id
		WHERE ct.rid = ?`, rid).Scan(&campaignID, &targetID, &email, &pageID)

	if err == nil {
		// Record the click event
		var ctID int
		db.DB.QueryRow("SELECT id FROM campaign_targets WHERE rid = ?", rid).Scan(&ctID)

		db.DB.Exec(`
			INSERT INTO events (campaign_id, campaign_target_id, email, time, message, details)
			VALUES (?, ?, ?, ?, 'Clicked Link', '')`,
			campaignID, ctID, email, time.Now())

		// Update campaign target status
		db.DB.Exec("UPDATE campaign_targets SET status = 'clicked' WHERE rid = ?", rid)

		// Get page HTML
		var pageHTML string
		db.DB.QueryRow("SELECT html FROM pages WHERE id = ?", pageID).Scan(&pageHTML)

		// Inject tracking form if capture is enabled
		db.DB.QueryRow("SELECT redirect_url FROM pages WHERE id = ?", pageID).Scan(&redirectURL)

		// Return the landing page with RID embedded
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(injectRID(pageHTML, rid)))
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func TrackSubmission(w http.ResponseWriter, r *http.Request) {
	rid := r.FormValue("rid")
	if rid == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Parse form data
	r.ParseForm()
	formData := make(map[string]interface{})
	for key, values := range r.Form {
		if key != "rid" && len(values) > 0 {
			formData[key] = values[0]
		}
	}

	// Get campaign target info
	var campaignID int
	var email string
	err := db.DB.QueryRow(`
		SELECT ct.campaign_id, t.email
		FROM campaign_targets ct
		JOIN targets t ON t.id = ct.target_id
		WHERE ct.rid = ?`, rid).Scan(&campaignID, &email)

	if err == nil {
		// Record the submission event
		var ctID int
		db.DB.QueryRow("SELECT id FROM campaign_targets WHERE rid = ?", rid).Scan(&ctID)

		detailsJSON, _ := json.Marshal(formData)
		db.DB.Exec(`
			INSERT INTO events (campaign_id, campaign_target_id, email, time, message, details)
			VALUES (?, ?, ?, ?, 'Submitted Data', ?)`,
			campaignID, ctID, email, time.Now(), string(detailsJSON))

		// Update campaign target status
		db.DB.Exec("UPDATE campaign_targets SET status = 'submitted' WHERE rid = ?", rid)

		// Get redirect URL
		var redirectURL string
		db.DB.QueryRow(`
			SELECT p.redirect_url FROM pages p
			JOIN campaigns c ON c.page_id = p.id
			WHERE c.id = ?`, campaignID).Scan(&redirectURL)

		if redirectURL != "" {
			respondJSON(w, map[string]string{"redirect": redirectURL})
			return
		}
	}

	respondJSON(w, map[string]bool{"success": true})
}

func injectRID(html, rid string) string {
	// Inject RID into forms
	injectionScript := `<script>
document.addEventListener('DOMContentLoaded', function() {
	var forms = document.getElementsByTagName('form');
	for (var i = 0; i < forms.length; i++) {
		var ridInput = document.createElement('input');
		ridInput.type = 'hidden';
		ridInput.name = 'rid';
		ridInput.value = '` + rid + `';
		forms[i].appendChild(ridInput);

		forms[i].addEventListener('submit', function(e) {
			e.preventDefault();
			var formData = new FormData(this);
			fetch('/api/report', {
				method: 'POST',
				body: formData
			}).then(function(response) {
				return response.json();
			}).then(function(data) {
				if (data.redirect) {
					window.location.href = data.redirect;
				} else {
					alert('Thank you for your submission!');
				}
			});
		});
	}
});
</script></body>`

	return replaceOnce(html, "</body>", injectionScript)
}

func replaceOnce(s, old, new string) string {
	i := len(s) - len(old)
	for i >= 0 {
		if s[i:i+len(old)] == old {
			return s[:i] + new + s[i+len(old):]
		}
		i--
	}
	return s + new
}

// TrackReportPhishing tracks when a user reports the email as phishing
func TrackReportPhishing(w http.ResponseWriter, r *http.Request) {
	rid := r.URL.Query().Get("rid")
	if rid == "" {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Get campaign target info
	var campaignID int
	var email string
	err := db.DB.QueryRow(`
		SELECT ct.campaign_id, t.email
		FROM campaign_targets ct
		JOIN targets t ON t.id = ct.target_id
		WHERE ct.rid = ?`, rid).Scan(&campaignID, &email)

	if err == nil {
		// Check if already reported
		var count int
		db.DB.QueryRow(`
			SELECT COUNT(*) FROM events
			WHERE campaign_target_id = (SELECT id FROM campaign_targets WHERE rid = ?)
			AND message = 'Reported Phishing'`, rid).Scan(&count)

		if count == 0 {
			// Record the report event
			var ctID int
			db.DB.QueryRow("SELECT id FROM campaign_targets WHERE rid = ?", rid).Scan(&ctID)

			db.DB.Exec(`
				INSERT INTO events (campaign_id, campaign_target_id, email, time, message, details)
				VALUES (?, ?, ?, ?, 'Reported Phishing', '')`,
				campaignID, ctID, email, time.Now())

			// Update campaign target status
			db.DB.Exec("UPDATE campaign_targets SET status = 'reported' WHERE rid = ?", rid)
		}
	}

	respondJSON(w, map[string]bool{"success": true})
}
