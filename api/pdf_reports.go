package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/clariphish/clariphish/db"
	"github.com/gorilla/mux"
	"github.com/jung-kurt/gofpdf"
)

// GenerateResultPDF generates a PDF report for a user's assessment result
func GenerateResultPDF(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	attemptID := vars["attemptId"]

	userID, _, err := getUserFromSession(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get attempt details
	var attempt struct {
		ID            int
		UserID        int
		AssessmentID  int
		Score         sql.NullFloat64
		TotalPoints   float64
		StartedAt     time.Time
		CompletedAt   sql.NullTime
		Username      string
		AssessmentTitle string
	}

	err = db.DB.QueryRow(`
		SELECT ua.id, ua.user_id, ua.assessment_id, ua.score, ua.total_points,
		       ua.started_at, ua.completed_at, u.username, a.title
		FROM user_assessment_attempts ua
		JOIN users u ON ua.user_id = u.id
		JOIN assessments a ON ua.assessment_id = a.id
		WHERE ua.id = ?
	`, attemptID).Scan(
		&attempt.ID, &attempt.UserID, &attempt.AssessmentID, &attempt.Score,
		&attempt.TotalPoints, &attempt.StartedAt, &attempt.CompletedAt,
		&attempt.Username, &attempt.AssessmentTitle,
	)

	if err != nil {
		respondError(w, "Assessment attempt not found", http.StatusNotFound)
		return
	}

	// Verify user owns this attempt (or is admin)
	if attempt.UserID != userID {
		session, _ := store.Get(r, "clariphish-session")
		role, _ := session.Values["role"]
		if role != "admin" {
			respondError(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	if !attempt.CompletedAt.Valid {
		respondError(w, "Assessment not completed yet", http.StatusBadRequest)
		return
	}

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(190, 10, "Assessment Result Report")
	pdf.Ln(15)

	// Assessment Info
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Assessment: "+attempt.AssessmentTitle)
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(190, 7, "Student: "+attempt.Username)
	pdf.Ln(7)
	pdf.Cell(190, 7, "Completed: "+attempt.CompletedAt.Time.Format("2006-01-02 15:04:05"))
	pdf.Ln(7)

	// Score
	percentage := 0.0
	if attempt.TotalPoints > 0 && attempt.Score.Valid {
		percentage = (attempt.Score.Float64 / attempt.TotalPoints) * 100
	}

	pdf.SetFont("Arial", "B", 14)
	scoreText := fmt.Sprintf("Score: %.1f / %.1f (%.1f%%)", attempt.Score.Float64, attempt.TotalPoints, percentage)
	pdf.Cell(190, 10, scoreText)
	pdf.Ln(15)

	// Performance indicator
	pdf.SetFont("Arial", "", 12)
	var performance string
	if percentage >= 90 {
		performance = "Excellent! You have a strong understanding of security concepts."
	} else if percentage >= 70 {
		performance = "Good job! Consider reviewing the topics you missed."
	} else if percentage >= 50 {
		performance = "Fair. Please review the security training materials."
	} else {
		performance = "Needs improvement. We recommend retaking the training modules."
	}
	pdf.MultiCell(190, 7, "Performance: "+performance, "", "L", false)
	pdf.Ln(10)

	// Get detailed results
	rows, err := db.DB.Query(`
		SELECT
			q.question_text,
			q.points,
			ur.is_correct,
			ur.points_earned,
			selected.option_text as selected_option,
			correct.option_text as correct_option
		FROM user_responses ur
		JOIN questions q ON ur.question_id = q.id
		LEFT JOIN answer_options selected ON ur.selected_option_id = selected.id
		LEFT JOIN answer_options correct ON q.id = correct.question_id AND correct.is_correct = 1
		WHERE ur.attempt_id = ?
		ORDER BY q.question_order
	`, attemptID)

	if err == nil {
		defer rows.Close()

		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(190, 8, "Detailed Results:")
		pdf.Ln(10)

		questionNum := 1
		for rows.Next() {
			var questionText, selectedOption, correctOption sql.NullString
			var points, pointsEarned sql.NullFloat64
			var isCorrect sql.NullInt64

			rows.Scan(&questionText, &points, &isCorrect, &pointsEarned, &selectedOption, &correctOption)

			pdf.SetFont("Arial", "B", 11)
			qText := "N/A"
			if questionText.Valid {
				qText = questionText.String
			}
			pdf.MultiCell(190, 6, fmt.Sprintf("Q%d. %s", questionNum, qText), "", "L", false)

			pdf.SetFont("Arial", "", 10)
			status := "Incorrect"
			isCorrectBool := false
			if isCorrect.Valid && isCorrect.Int64 == 1 {
				status = "Correct"
				isCorrectBool = true
			}

			selectedText := "N/A"
			if selectedOption.Valid {
				selectedText = selectedOption.String
			}

			pdf.MultiCell(190, 6, fmt.Sprintf("   Your answer: %s (%s)", selectedText, status), "", "L", false)

			// Show correct answer if user was wrong
			if !isCorrectBool && correctOption.Valid {
				pdf.SetFont("Arial", "I", 10)
				pdf.MultiCell(190, 6, fmt.Sprintf("   Correct answer: %s", correctOption.String), "", "L", false)
			}

			// Show points
			earnedPts := 0.0
			totalPts := 0.0
			if pointsEarned.Valid {
				earnedPts = pointsEarned.Float64
			}
			if points.Valid {
				totalPts = points.Float64
			}
			pdf.SetFont("Arial", "", 9)
			pdf.Cell(190, 5, fmt.Sprintf("   Points: %.0f / %.0f", earnedPts, totalPts))
			pdf.Ln(6)
			pdf.Ln(4)

			questionNum++
		}
	}

	// Footer
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 10, fmt.Sprintf("Generated by clariphish on %s", time.Now().Format("2006-01-02 15:04:05")))

	// Output PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="assessment_result_%s.pdf"`, attemptID))

	err = pdf.Output(w)
	if err != nil {
		respondError(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}
}

// GenerateCampaignReportPDF generates a PDF report for a campaign
func GenerateCampaignReportPDF(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignID := vars["id"]

	// Get campaign details
	var campaign struct {
		ID        int
		Name      string
		CreatedAt time.Time
		Status    string
	}

	err := db.DB.QueryRow(`
		SELECT id, name, created_at, status
		FROM campaigns
		WHERE id = ?
	`, campaignID).Scan(&campaign.ID, &campaign.Name, &campaign.CreatedAt, &campaign.Status)

	if err != nil {
		respondError(w, "Campaign not found", http.StatusNotFound)
		return
	}

	// Get campaign statistics
	var stats struct {
		TotalTargets int
		EmailsSent   int
		EmailsOpened int
		LinksClicked int
		DataSubmitted int
		Reported     int
	}

	db.DB.QueryRow(`
		SELECT COUNT(*) FROM campaign_targets WHERE campaign_id = ?
	`, campaignID).Scan(&stats.TotalTargets)

	db.DB.QueryRow(`
		SELECT COUNT(*) FROM campaign_targets WHERE campaign_id = ? AND sent_at IS NOT NULL
	`, campaignID).Scan(&stats.EmailsSent)

	db.DB.QueryRow(`
		SELECT COUNT(*) FROM campaign_targets WHERE campaign_id = ? AND opened_at IS NOT NULL
	`, campaignID).Scan(&stats.EmailsOpened)

	db.DB.QueryRow(`
		SELECT COUNT(*) FROM campaign_targets WHERE campaign_id = ? AND clicked_at IS NOT NULL
	`, campaignID).Scan(&stats.LinksClicked)

	db.DB.QueryRow(`
		SELECT COUNT(*) FROM campaign_targets WHERE campaign_id = ? AND submitted_at IS NOT NULL
	`, campaignID).Scan(&stats.DataSubmitted)

	db.DB.QueryRow(`
		SELECT COUNT(*) FROM campaign_targets WHERE campaign_id = ? AND reported_at IS NOT NULL
	`, campaignID).Scan(&stats.Reported)

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(190, 10, "Campaign Report")
	pdf.Ln(15)

	// Campaign Info
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Campaign: "+campaign.Name)
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(190, 7, "Created: "+campaign.CreatedAt.Format("2006-01-02 15:04:05"))
	pdf.Ln(7)
	pdf.Cell(190, 7, "Status: "+campaign.Status)
	pdf.Ln(15)

	// Statistics
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Campaign Statistics")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(90, 7, "Total Targets:")
	pdf.Cell(100, 7, strconv.Itoa(stats.TotalTargets))
	pdf.Ln(7)

	pdf.Cell(90, 7, "Emails Sent:")
	pdf.Cell(100, 7, fmt.Sprintf("%d (%.1f%%)", stats.EmailsSent, percentage(stats.EmailsSent, stats.TotalTargets)))
	pdf.Ln(7)

	pdf.Cell(90, 7, "Emails Opened:")
	pdf.Cell(100, 7, fmt.Sprintf("%d (%.1f%%)", stats.EmailsOpened, percentage(stats.EmailsOpened, stats.EmailsSent)))
	pdf.Ln(7)

	pdf.Cell(90, 7, "Links Clicked:")
	pdf.Cell(100, 7, fmt.Sprintf("%d (%.1f%%)", stats.LinksClicked, percentage(stats.LinksClicked, stats.EmailsSent)))
	pdf.Ln(7)

	pdf.Cell(90, 7, "Data Submitted:")
	pdf.Cell(100, 7, fmt.Sprintf("%d (%.1f%%)", stats.DataSubmitted, percentage(stats.DataSubmitted, stats.EmailsSent)))
	pdf.Ln(7)

	pdf.Cell(90, 7, "Reported as Phishing:")
	pdf.Cell(100, 7, fmt.Sprintf("%d (%.1f%%)", stats.Reported, percentage(stats.Reported, stats.EmailsSent)))
	pdf.Ln(15)

	// Analysis
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Analysis")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	clickRate := percentage(stats.LinksClicked, stats.EmailsSent)
	submitRate := percentage(stats.DataSubmitted, stats.EmailsSent)
	reportRate := percentage(stats.Reported, stats.EmailsSent)

	analysis := ""
	if reportRate > 50 {
		analysis = "Excellent! More than half of users reported the phishing attempt. This shows strong security awareness."
	} else if clickRate > 30 {
		analysis = "Warning: High click rate indicates users need more phishing awareness training."
	} else if submitRate > 10 {
		analysis = "Concern: A significant number of users submitted data. Additional training recommended."
	} else {
		analysis = "Overall performance is acceptable, but continuous training is recommended."
	}

	pdf.MultiCell(190, 7, analysis, "", "L", false)

	// Footer
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 10, fmt.Sprintf("Generated by clariphish on %s", time.Now().Format("2006-01-02 15:04:05")))

	// Output PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="campaign_report_%s.pdf"`, campaignID))

	err = pdf.Output(w)
	if err != nil {
		respondError(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}
}

// GenerateAssessmentOverviewPDF generates a PDF report for assessment overview (admin)
func GenerateAssessmentOverviewPDF(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	assessmentID := vars["id"]

	// Get assessment details
	var assessment struct {
		ID          int
		Title       string
		Description string
		CreatedAt   time.Time
	}

	err := db.DB.QueryRow(`
		SELECT id, title, description, created_at
		FROM assessments
		WHERE id = ?
	`, assessmentID).Scan(&assessment.ID, &assessment.Title, &assessment.Description, &assessment.CreatedAt)

	if err != nil {
		respondError(w, "Assessment not found", http.StatusNotFound)
		return
	}

	// Get statistics
	var stats struct {
		TotalUsers    int
		Completed     int
		InProgress    int
		NotStarted    int
		AverageScore  sql.NullFloat64
		HighestScore  sql.NullFloat64
		LowestScore   sql.NullFloat64
	}

	// Count total users
	db.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'user'`).Scan(&stats.TotalUsers)

	// Count completed attempts
	db.DB.QueryRow(`
		SELECT COUNT(*) FROM user_assessment_attempts
		WHERE assessment_id = ? AND completed_at IS NOT NULL
	`, assessmentID).Scan(&stats.Completed)

	// Count in progress
	db.DB.QueryRow(`
		SELECT COUNT(*) FROM user_assessment_attempts
		WHERE assessment_id = ? AND completed_at IS NULL
	`, assessmentID).Scan(&stats.InProgress)

	stats.NotStarted = stats.TotalUsers - stats.Completed - stats.InProgress

	// Get score statistics
	db.DB.QueryRow(`
		SELECT AVG(score), MAX(score), MIN(score)
		FROM user_assessment_attempts
		WHERE assessment_id = ? AND completed_at IS NOT NULL
	`, assessmentID).Scan(&stats.AverageScore, &stats.HighestScore, &stats.LowestScore)

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(190, 10, "Assessment Overview Report")
	pdf.Ln(15)

	// Assessment Info
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Assessment: "+assessment.Title)
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(190, 6, assessment.Description, "", "L", false)
	pdf.Ln(5)

	pdf.Cell(190, 7, "Created: "+assessment.CreatedAt.Format("2006-01-02"))
	pdf.Ln(15)

	// Statistics
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Completion Statistics")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(90, 7, "Total Users:")
	pdf.Cell(100, 7, strconv.Itoa(stats.TotalUsers))
	pdf.Ln(7)

	pdf.Cell(90, 7, "Completed:")
	pdf.Cell(100, 7, fmt.Sprintf("%d (%.1f%%)", stats.Completed, percentage(stats.Completed, stats.TotalUsers)))
	pdf.Ln(7)

	pdf.Cell(90, 7, "In Progress:")
	pdf.Cell(100, 7, fmt.Sprintf("%d (%.1f%%)", stats.InProgress, percentage(stats.InProgress, stats.TotalUsers)))
	pdf.Ln(7)

	pdf.Cell(90, 7, "Not Started:")
	pdf.Cell(100, 7, fmt.Sprintf("%d (%.1f%%)", stats.NotStarted, percentage(stats.NotStarted, stats.TotalUsers)))
	pdf.Ln(15)

	// Score Statistics
	if stats.Completed > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(190, 8, "Score Statistics")
		pdf.Ln(12)

		pdf.SetFont("Arial", "", 12)
		if stats.AverageScore.Valid {
			pdf.Cell(90, 7, "Average Score:")
			pdf.Cell(100, 7, fmt.Sprintf("%.2f", stats.AverageScore.Float64))
			pdf.Ln(7)
		}

		if stats.HighestScore.Valid {
			pdf.Cell(90, 7, "Highest Score:")
			pdf.Cell(100, 7, fmt.Sprintf("%.2f", stats.HighestScore.Float64))
			pdf.Ln(7)
		}

		if stats.LowestScore.Valid {
			pdf.Cell(90, 7, "Lowest Score:")
			pdf.Cell(100, 7, fmt.Sprintf("%.2f", stats.LowestScore.Float64))
			pdf.Ln(7)
		}
	}

	// Footer
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 10, fmt.Sprintf("Generated by clariphish on %s", time.Now().Format("2006-01-02 15:04:05")))

	// Output PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="assessment_overview_%s.pdf"`, assessmentID))

	err = pdf.Output(w)
	if err != nil {
		respondError(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}
}

// GenerateCredentialsPDF generates a PDF report of campaign statistics and user actions
func GenerateCredentialsPDF(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	campaignID := vars["id"]

	// Get campaign details
	var campaign struct {
		Name        string
		CreatedDate time.Time
		LaunchDate  sql.NullTime
		Status      string
	}
	err := db.DB.QueryRow(`
		SELECT name, created_date, launch_date, status
		FROM campaigns WHERE id = ?`, campaignID).Scan(
		&campaign.Name, &campaign.CreatedDate, &campaign.LaunchDate, &campaign.Status)
	if err != nil {
		respondError(w, "Campaign not found", http.StatusNotFound)
		return
	}

	// Get campaign statistics
	var stats struct {
		Sent       int
		Opened     int
		Clicked    int
		Submitted  int
		Reported   int
		OpenRate   int
		ClickRate  int
		SubmitRate int
		ReportRate int
	}

	db.DB.QueryRow(`SELECT COUNT(*) FROM campaign_targets WHERE campaign_id = ?`, campaignID).Scan(&stats.Sent)
	db.DB.QueryRow(`SELECT COUNT(DISTINCT campaign_target_id) FROM events WHERE campaign_id = ? AND message = 'Email Opened'`, campaignID).Scan(&stats.Opened)
	db.DB.QueryRow(`SELECT COUNT(DISTINCT campaign_target_id) FROM events WHERE campaign_id = ? AND message = 'Clicked Link'`, campaignID).Scan(&stats.Clicked)
	db.DB.QueryRow(`SELECT COUNT(DISTINCT campaign_target_id) FROM events WHERE campaign_id = ? AND message = 'Submitted Data'`, campaignID).Scan(&stats.Submitted)
	db.DB.QueryRow(`SELECT COUNT(DISTINCT campaign_target_id) FROM events WHERE campaign_id = ? AND message = 'Reported Phishing'`, campaignID).Scan(&stats.Reported)

	if stats.Sent > 0 {
		stats.OpenRate = (stats.Opened * 100) / stats.Sent
		stats.ClickRate = (stats.Clicked * 100) / stats.Sent
		stats.SubmitRate = (stats.Submitted * 100) / stats.Sent
		stats.ReportRate = (stats.Reported * 100) / stats.Sent
	}

	// Get user details
	rows, err := db.DB.Query(`
		SELECT t.first_name, t.last_name, t.email, ct.status,
		       EXISTS(SELECT 1 FROM events WHERE campaign_target_id = ct.id AND message = 'Clicked Link') as clicked,
		       EXISTS(SELECT 1 FROM events WHERE campaign_target_id = ct.id AND message = 'Submitted Data') as submitted,
		       EXISTS(SELECT 1 FROM events WHERE campaign_target_id = ct.id AND message = 'Reported Phishing') as reported
		FROM campaign_targets ct
		JOIN targets t ON t.id = ct.target_id
		WHERE ct.campaign_id = ?
		ORDER BY t.last_name, t.first_name`, campaignID)

	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Header
	pdf.SetFont("Arial", "B", 20)
	pdf.Cell(190, 10, "Campaign Report")
	pdf.Ln(15)

	// Campaign Info
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Campaign: "+campaign.Name)
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(90, 7, "Created:")
	pdf.Cell(100, 7, campaign.CreatedDate.Format("2006-01-02 15:04:05"))
	pdf.Ln(7)

	if campaign.LaunchDate.Valid {
		pdf.Cell(90, 7, "Launched:")
		pdf.Cell(100, 7, campaign.LaunchDate.Time.Format("2006-01-02 15:04:05"))
		pdf.Ln(7)
	}

	pdf.Cell(90, 7, "Status:")
	pdf.Cell(100, 7, campaign.Status)
	pdf.Ln(12)

	// Statistics Section
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Campaign Statistics")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(90, 7, "Total Sent:")
	pdf.Cell(100, 7, fmt.Sprintf("%d", stats.Sent))
	pdf.Ln(7)

	pdf.Cell(90, 7, "Clicked:")
	pdf.Cell(100, 7, fmt.Sprintf("%d (%d%%)", stats.Clicked, stats.ClickRate))
	pdf.Ln(7)

	pdf.Cell(90, 7, "Submitted Credentials:")
	pdf.Cell(100, 7, fmt.Sprintf("%d (%d%%)", stats.Submitted, stats.SubmitRate))
	pdf.Ln(7)

	pdf.Cell(90, 7, "Reported as Phishing:")
	pdf.Cell(100, 7, fmt.Sprintf("%d (%d%%)", stats.Reported, stats.ReportRate))
	pdf.Ln(15)

	// Users Section
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Target Users")
	pdf.Ln(10)

	// Table header
	pdf.SetFont("Arial", "B", 9)
	pdf.Cell(60, 7, "Name")
	pdf.Cell(35, 7, "Clicked")
	pdf.Cell(35, 7, "Submitted")
	pdf.Cell(35, 7, "Reported")
	pdf.Ln(7)

	pdf.SetFont("Arial", "", 9)
	for rows.Next() {
		var firstName, lastName, email, status string
		var clicked, submitted, reported bool
		rows.Scan(&firstName, &lastName, &email, &status, &clicked, &submitted, &reported)

		pdf.Cell(60, 6, firstName+" "+lastName)
		pdf.Cell(35, 6, boolToCheckmark(clicked))
		pdf.Cell(35, 6, boolToCheckmark(submitted))
		pdf.Cell(35, 6, boolToCheckmark(reported))
		pdf.Ln(6)
	}

	// Harvested Credentials Section
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Harvested Credentials")
	pdf.Ln(10)

	// Query for harvested credentials
	credRows, err := db.DB.Query(`
		SELECT e.time, t.first_name, t.last_name, t.email, e.details
		FROM events e
		JOIN campaign_targets ct ON ct.id = e.campaign_target_id
		JOIN targets t ON t.id = ct.target_id
		WHERE e.campaign_id = ? AND e.message = 'Submitted Data'
		ORDER BY e.time DESC`, campaignID)

	if err == nil {
		defer credRows.Close()

		credCount := 0
		for credRows.Next() {
			var timestamp time.Time
			var firstName, lastName, email, detailsJSON string
			credRows.Scan(&timestamp, &firstName, &lastName, &email, &detailsJSON)

			credCount++

			// Parse credentials to extract email and password
			var credentials map[string]interface{}
			var submittedEmail, submittedPassword string
			if json.Unmarshal([]byte(detailsJSON), &credentials) == nil {
				// Try common field names for email
				if val, ok := credentials["email"]; ok {
					submittedEmail = fmt.Sprintf("%v", val)
				} else if val, ok := credentials["username"]; ok {
					submittedEmail = fmt.Sprintf("%v", val)
				} else if val, ok := credentials["user"]; ok {
					submittedEmail = fmt.Sprintf("%v", val)
				}

				// Try common field names for password
				if val, ok := credentials["password"]; ok {
					submittedPassword = fmt.Sprintf("%v", val)
				} else if val, ok := credentials["pass"]; ok {
					submittedPassword = fmt.Sprintf("%v", val)
				} else if val, ok := credentials["pwd"]; ok {
					submittedPassword = fmt.Sprintf("%v", val)
				}
			}

			// Entry
			pdf.SetFont("Arial", "B", 10)
			pdf.Cell(190, 6, fmt.Sprintf("%d. %s %s (%s)", credCount, firstName, lastName, email))
			pdf.Ln(6)

			pdf.SetFont("Arial", "", 9)
			pdf.Cell(190, 5, "  Time: "+timestamp.Format("2006-01-02 15:04:05"))
			pdf.Ln(5)
			if submittedEmail != "" {
				pdf.Cell(190, 5, "  Email/Username: "+submittedEmail)
				pdf.Ln(5)
			}
			if submittedPassword != "" {
				pdf.Cell(190, 5, "  Password: "+submittedPassword)
				pdf.Ln(5)
			}
			pdf.Ln(3)
		}

		if credCount == 0 {
			pdf.SetFont("Arial", "I", 10)
			pdf.Cell(190, 7, "No credentials submitted.")
			pdf.Ln(7)
		}
	}

	// Footer
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(190, 10, fmt.Sprintf("Generated by clariphish on %s - Page %d", time.Now().Format("2006-01-02 15:04:05"), pdf.PageNo()))

	// Output PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="campaign_report_%s.pdf"`, campaignID))

	err = pdf.Output(w)
	if err != nil {
		respondError(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}
}

// GenerateEmailAnalysisPDF generates a PDF report for email phishing analysis
func GenerateEmailAnalysisPDF(w http.ResponseWriter, r *http.Request) {
	userID, _, err := getUserFromSession(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		EmailText       string   `json:"email_text"`
		IsPhishing      bool     `json:"is_phishing"`
		ConfidenceScore int      `json:"confidence_score"`
		RiskLevel       string   `json:"risk_level"`
		Indicators      []string `json:"indicators"`
		Explanation     string   `json:"explanation"`
		Recommendations []string `json:"recommendations"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Get user info
	var username string
	err = db.DB.QueryRow(`SELECT username FROM users WHERE id = ?`, userID).Scan(&username)
	if err != nil {
		respondError(w, "User not found", http.StatusNotFound)
		return
	}

	// Create PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Title
	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(190, 10, "Email Phishing Analysis Report")
	pdf.Ln(12)

	// Analysis Info
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(50, 6, "Analyzed by:")
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(140, 6, username)
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(50, 6, "Analysis Date:")
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(140, 6, time.Now().Format("2006-01-02 15:04:05"))
	pdf.Ln(12)

	// Verdict Section
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Verdict")
	pdf.Ln(8)

	verdictText := "LEGITIMATE EMAIL"
	if req.IsPhishing {
		if req.RiskLevel == "critical" || req.ConfidenceScore > 90 {
			verdictText = "PHISHING - HIGH RISK"
		} else if req.RiskLevel == "high" || req.ConfidenceScore > 70 {
			verdictText = "LIKELY PHISHING"
		} else {
			verdictText = "SUSPICIOUS EMAIL"
		}
	}

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, verdictText)
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(50, 6, "Confidence Score:")
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(140, 6, fmt.Sprintf("%d%%", req.ConfidenceScore))
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(50, 6, "Risk Level:")
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(140, 6, req.RiskLevel)
	pdf.Ln(12)

	// Indicators Section
	if len(req.Indicators) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(190, 8, "Phishing Indicators")
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 10)
		for _, indicator := range req.Indicators {
			pdf.MultiCell(190, 5, "- "+indicator, "", "L", false)
			pdf.Ln(2)
		}
		pdf.Ln(6)
	}

	// Explanation Section
	if req.Explanation != "" {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(190, 8, "Analysis Explanation")
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 10)
		pdf.MultiCell(190, 5, req.Explanation, "", "L", false)
		pdf.Ln(6)
	}

	// Recommendations Section
	if len(req.Recommendations) > 0 {
		pdf.SetFont("Arial", "B", 14)
		pdf.Cell(190, 8, "Recommendations")
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 10)
		for _, rec := range req.Recommendations {
			pdf.MultiCell(190, 5, "- "+rec, "", "L", false)
			pdf.Ln(2)
		}
		pdf.Ln(6)
	}

	// Email Content Section
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(190, 8, "Analyzed Email Content")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 9)
	pdf.SetFillColor(245, 245, 245)
	pdf.MultiCell(190, 4, req.EmailText, "1", "L", true)
	pdf.Ln(8)

	// Footer
	pdf.SetY(-20)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(128, 128, 128)
	pdf.Cell(190, 10, fmt.Sprintf("Generated by clariphish on %s", time.Now().Format("2006-01-02 15:04:05")))

	// Output PDF
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="email_analysis_%s.pdf"`, time.Now().Format("20060102_150405")))

	err = pdf.Output(w)
	if err != nil {
		respondError(w, "Failed to generate PDF", http.StatusInternalServerError)
		return
	}
}

// boolToCheckmark converts boolean to checkmark symbol
func boolToCheckmark(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// percentage calculates percentage safely
func percentage(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}
