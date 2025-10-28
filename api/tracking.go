package api

import (
	"encoding/base64"
	"encoding/json"
	"log"
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
	log.Printf("TrackClick called with RID: %s", rid)
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
		log.Printf("TrackClick: Found campaign_id=%d, email=%s", campaignID, email)

		// Record the click event
		var ctID int
		db.DB.QueryRow("SELECT id FROM campaign_targets WHERE rid = ?", rid).Scan(&ctID)
		log.Printf("TrackClick: campaign_target_id=%d", ctID)

		result, err := db.DB.Exec(`
			INSERT INTO events (campaign_id, campaign_target_id, email, time, message, details)
			VALUES (?, ?, ?, ?, 'Clicked Link', '')`,
			campaignID, ctID, email, time.Now())

		if err != nil {
			log.Printf("TrackClick: Error inserting event: %v", err)
		} else {
			eventID, _ := result.LastInsertId()
			log.Printf("TrackClick: Event created with ID=%d", eventID)
		}

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
	} else {
		log.Printf("TrackClick: Error finding campaign target: %v", err)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func TrackSubmission(w http.ResponseWriter, r *http.Request) {
	rid := r.FormValue("rid")
	log.Printf("TrackSubmission called with RID: %s", rid)
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
	log.Printf("TrackSubmission: Form data: %+v", formData)

	// Get campaign target info
	var campaignID int
	var email string
	err := db.DB.QueryRow(`
		SELECT ct.campaign_id, t.email
		FROM campaign_targets ct
		JOIN targets t ON t.id = ct.target_id
		WHERE ct.rid = ?`, rid).Scan(&campaignID, &email)

	if err == nil {
		log.Printf("TrackSubmission: Found campaign_id=%d, email=%s", campaignID, email)

		// Record the submission event
		var ctID int
		db.DB.QueryRow("SELECT id FROM campaign_targets WHERE rid = ?", rid).Scan(&ctID)
		log.Printf("TrackSubmission: campaign_target_id=%d", ctID)

		detailsJSON, _ := json.Marshal(formData)
		result, err := db.DB.Exec(`
			INSERT INTO events (campaign_id, campaign_target_id, email, time, message, details)
			VALUES (?, ?, ?, ?, 'Submitted Data', ?)`,
			campaignID, ctID, email, time.Now(), string(detailsJSON))

		if err != nil {
			log.Printf("TrackSubmission: Error inserting event: %v", err)
		} else {
			eventID, _ := result.LastInsertId()
			log.Printf("TrackSubmission: Event created with ID=%d, details=%s", eventID, string(detailsJSON))
		}

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
	} else {
		log.Printf("TrackSubmission: Error finding campaign target: %v", err)
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
	log.Printf("TrackReportPhishing called with RID: %s", rid)
	if rid == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
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
		log.Printf("TrackReportPhishing: Found campaign_id=%d, email=%s", campaignID, email)

		// Check if already reported
		var count int
		db.DB.QueryRow(`
			SELECT COUNT(*) FROM events
			WHERE campaign_target_id = (SELECT id FROM campaign_targets WHERE rid = ?)
			AND message = 'Reported Phishing'`, rid).Scan(&count)

		log.Printf("TrackReportPhishing: Existing report count=%d", count)

		if count == 0 {
			// Record the report event
			var ctID int
			db.DB.QueryRow("SELECT id FROM campaign_targets WHERE rid = ?", rid).Scan(&ctID)
			log.Printf("TrackReportPhishing: campaign_target_id=%d", ctID)

			result, err := db.DB.Exec(`
				INSERT INTO events (campaign_id, campaign_target_id, email, time, message, details)
				VALUES (?, ?, ?, ?, 'Reported Phishing', '')`,
				campaignID, ctID, email, time.Now())

			if err != nil {
				log.Printf("TrackReportPhishing: Error inserting event: %v", err)
			} else {
				eventID, _ := result.LastInsertId()
				log.Printf("TrackReportPhishing: Event created with ID=%d", eventID)
			}

			// Update campaign target status
			db.DB.Exec("UPDATE campaign_targets SET status = 'reported' WHERE rid = ?", rid)
			log.Printf("TrackReportPhishing: Updated status to 'reported'")
		} else {
			log.Printf("TrackReportPhishing: Already reported, skipping")
		}
	} else {
		log.Printf("TrackReportPhishing: Error finding campaign target: %v", err)
	}

	// Return a user-friendly HTML page
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
	<title>Thank You for Reporting</title>
	<style>
		body { font-family: Arial, sans-serif; text-align: center; padding: 50px; background: #f5f5f5; }
		.container { background: white; padding: 40px; border-radius: 8px; max-width: 500px; margin: 0 auto; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		h1 { color: #28a745; }
		p { color: #666; line-height: 1.6; }
	</style>
</head>
<body>
	<div class="container">
		<h1>✓ Thank You!</h1>
		<p>Your report has been recorded successfully.</p>
		<p>Thank you for being vigilant about suspicious emails and helping us improve security awareness.</p>
	</div>
</body>
</html>
	`))
}
