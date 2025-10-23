package mailer

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"

	"github.com/clariphish/clariphish/db"
	"github.com/clariphish/clariphish/models"
)

func LaunchCampaign(campaignID int) {
	log.Printf("Launching campaign %d", campaignID)

	// Get campaign details
	var c models.Campaign
	err := db.DB.QueryRow(`
		SELECT id, name, url, template_id, smtp_id
		FROM campaigns WHERE id = ?`, campaignID).
		Scan(&c.ID, &c.Name, &c.URL, &c.TemplateID, &c.SMTPID)

	if err != nil {
		log.Printf("Failed to get campaign: %v", err)
		return
	}

	// Get template
	var template models.Template
	err = db.DB.QueryRow(`
		SELECT id, subject, html, text FROM templates WHERE id = ?`, c.TemplateID).
		Scan(&template.ID, &template.Subject, &template.HTML, &template.Text)

	if err != nil {
		log.Printf("Failed to get template: %v", err)
		return
	}

	// Get SMTP config
	var smtpConfig models.SMTP
	err = db.DB.QueryRow(`
		SELECT id, host, username, password, from_address
		FROM smtp WHERE id = ?`, c.SMTPID).
		Scan(&smtpConfig.ID, &smtpConfig.Host, &smtpConfig.Username,
		&smtpConfig.Password, &smtpConfig.FromAddress)

	if err != nil {
		log.Printf("Failed to get SMTP config: %v", err)
		return
	}

	// Get campaign targets
	rows, err := db.DB.Query(`
		SELECT ct.id, ct.rid, t.email, t.first_name, t.last_name
		FROM campaign_targets ct
		JOIN targets t ON t.id = ct.target_id
		WHERE ct.campaign_id = ? AND ct.status = 'scheduled'`, campaignID)

	if err != nil {
		log.Printf("Failed to get targets: %v", err)
		return
	}
	defer rows.Close()

	targets := []struct {
		ID        int
		RID       string
		Email     string
		FirstName string
		LastName  string
	}{}

	for rows.Next() {
		var t struct {
			ID        int
			RID       string
			Email     string
			FirstName string
			LastName  string
		}
		rows.Scan(&t.ID, &t.RID, &t.Email, &t.FirstName, &t.LastName)
		targets = append(targets, t)
	}

	// Send emails to each target
	for _, target := range targets {
		err := sendEmail(smtpConfig, target.Email, target.FirstName, target.LastName,
			target.RID, template, c.URL)

		if err != nil {
			log.Printf("Failed to send to %s: %v", target.Email, err)
			db.DB.Exec(`
				INSERT INTO events (campaign_id, campaign_target_id, email, time, message, details)
				VALUES (?, ?, ?, ?, 'Error Sending Email', ?)`,
				campaignID, target.ID, target.Email, time.Now(), err.Error())

			db.DB.Exec("UPDATE campaign_targets SET status = 'error' WHERE id = ?", target.ID)
		} else {
			log.Printf("Sent email to %s", target.Email)
			db.DB.Exec(`
				INSERT INTO events (campaign_id, campaign_target_id, email, time, message, details)
				VALUES (?, ?, ?, ?, 'Email Sent', '')`,
				campaignID, target.ID, target.Email, time.Now())

			db.DB.Exec("UPDATE campaign_targets SET status = 'sent', send_date = ? WHERE id = ?",
				time.Now(), target.ID)
		}

		// Small delay between emails
		time.Sleep(100 * time.Millisecond)
	}

	log.Printf("Campaign %d launch completed", campaignID)
}

func sendEmail(smtpConfig models.SMTP, to, firstName, lastName, rid string,
	template models.Template, baseURL string) error {

	// Replace template variables
	subject := replaceVariables(template.Subject, firstName, lastName, rid, baseURL)
	htmlBody := replaceVariables(template.HTML, firstName, lastName, rid, baseURL)
	textBody := replaceVariables(template.Text, firstName, lastName, rid, baseURL)

	// Add tracking pixel
	trackingPixel := fmt.Sprintf(`<img src="%s/api/track?rid=%s" width="1" height="1" />`,
		baseURL, rid)
	htmlBody = strings.Replace(htmlBody, "</body>", trackingPixel+"</body>", 1)
	if !strings.Contains(htmlBody, "</body>") {
		htmlBody += trackingPixel
	}

	// Build email message
	message := ""
	message += fmt.Sprintf("From: %s\r\n", smtpConfig.FromAddress)
	message += fmt.Sprintf("To: %s\r\n", to)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "MIME-Version: 1.0\r\n"
	message += "Content-Type: multipart/alternative; boundary=\"boundary123\"\r\n"
	message += "\r\n"
	message += "--boundary123\r\n"
	message += "Content-Type: text/plain; charset=\"UTF-8\"\r\n"
	message += "\r\n"
	message += textBody + "\r\n"
	message += "\r\n"
	message += "--boundary123\r\n"
	message += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	message += "\r\n"
	message += htmlBody + "\r\n"
	message += "\r\n"
	message += "--boundary123--\r\n"

	// Send via SMTP
	auth := smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, strings.Split(smtpConfig.Host, ":")[0])
	err := smtp.SendMail(smtpConfig.Host, auth, smtpConfig.FromAddress, []string{to}, []byte(message))

	return err
}

func replaceVariables(text, firstName, lastName, rid, baseURL string) string {
	text = strings.ReplaceAll(text, "{{.FirstName}}", firstName)
	text = strings.ReplaceAll(text, "{{.LastName}}", lastName)
	text = strings.ReplaceAll(text, "{{.Email}}", "")
	text = strings.ReplaceAll(text, "{{.Position}}", "")
	text = strings.ReplaceAll(text, "{{.RId}}", rid)
	text = strings.ReplaceAll(text, "{{.URL}}", fmt.Sprintf("%s/api/click?rid=%s", baseURL, rid))

	return text
}
