package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/clariphish/clariphish/db"
	"github.com/clariphish/clariphish/models"
	"github.com/gorilla/mux"
)

// Admin Assessment Management

func GetAssessments(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT id, title, description, deadline, is_published,
		       COALESCE(is_dynamic, 0) as is_dynamic,
		       COALESCE(behavior_type, '') as behavior_type,
		       COALESCE(target_email, '') as target_email,
		       COALESCE(campaign_id, 0) as campaign_id,
		       created_by, created_at, updated_at
		FROM assessments
		ORDER BY created_at DESC
	`)
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	assessments := []models.Assessment{}
	for rows.Next() {
		var a models.Assessment
		var deadline sql.NullTime
		var isDynamic int
		err := rows.Scan(&a.ID, &a.Title, &a.Description, &deadline, &a.IsPublished, &isDynamic, &a.BehaviorType, &a.TargetEmail, &a.CampaignID, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			continue
		}
		if deadline.Valid {
			a.Deadline = &deadline.Time
		}
		a.IsDynamic = isDynamic == 1
		assessments = append(assessments, a)
	}

	respondJSON(w, assessments)
}

func CreateAssessment(w http.ResponseWriter, r *http.Request) {
	userID, _, err := getUserFromSession(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Title       string     `json:"title"`
		Description string     `json:"description"`
		Deadline    *time.Time `json:"deadline"`
		Questions   []struct {
			QuestionText  string `json:"question_text"`
			QuestionOrder int    `json:"question_order"`
			Points        int    `json:"points"`
			AnswerOptions []struct {
				OptionText  string `json:"option_text"`
				IsCorrect   bool   `json:"is_correct"`
				OptionOrder int    `json:"option_order"`
			} `json:"answer_options"`
		} `json:"questions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := db.DB.Begin()
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Create assessment
	result, err := tx.Exec(`
		INSERT INTO assessments (title, description, deadline, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, req.Title, req.Description, req.Deadline, userID, time.Now(), time.Now())
	if err != nil {
		respondError(w, "Failed to create assessment", http.StatusInternalServerError)
		return
	}

	assessmentID, _ := result.LastInsertId()

	// Create questions and answer options
	for _, q := range req.Questions {
		qResult, err := tx.Exec(`
			INSERT INTO questions (assessment_id, question_text, question_order, points, created_at)
			VALUES (?, ?, ?, ?, ?)
		`, assessmentID, q.QuestionText, q.QuestionOrder, q.Points, time.Now())
		if err != nil {
			respondError(w, "Failed to create question", http.StatusInternalServerError)
			return
		}

		questionID, _ := qResult.LastInsertId()

		// Create answer options
		for _, opt := range q.AnswerOptions {
			isCorrect := 0
			if opt.IsCorrect {
				isCorrect = 1
			}
			_, err := tx.Exec(`
				INSERT INTO answer_options (question_id, option_text, is_correct, option_order)
				VALUES (?, ?, ?, ?)
			`, questionID, opt.OptionText, isCorrect, opt.OptionOrder)
			if err != nil {
				respondError(w, "Failed to create answer option", http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		respondError(w, "Failed to save assessment", http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]interface{}{
		"success": true,
		"id":      assessmentID,
		"message": "Assessment created successfully",
	})
}

func GetAssessment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var assessment models.Assessment
	var deadline sql.NullTime
	err := db.DB.QueryRow(`
		SELECT id, title, description, deadline, is_published, created_by, created_at, updated_at
		FROM assessments WHERE id = ?
	`, id).Scan(&assessment.ID, &assessment.Title, &assessment.Description, &deadline, &assessment.IsPublished, &assessment.CreatedBy, &assessment.CreatedAt, &assessment.UpdatedAt)

	if err == sql.ErrNoRows {
		respondError(w, "Assessment not found", http.StatusNotFound)
		return
	}
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}

	if deadline.Valid {
		assessment.Deadline = &deadline.Time
	}

	// Get questions
	qRows, err := db.DB.Query(`
		SELECT id, assessment_id, question_text, question_order, points, created_at
		FROM questions WHERE assessment_id = ?
		ORDER BY question_order
	`, id)
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer qRows.Close()

	questions := []models.Question{}
	for qRows.Next() {
		var q models.Question
		qRows.Scan(&q.ID, &q.AssessmentID, &q.QuestionText, &q.QuestionOrder, &q.Points, &q.CreatedAt)

		// Get answer options for this question
		optRows, err := db.DB.Query(`
			SELECT id, question_id, option_text, is_correct, option_order
			FROM answer_options WHERE question_id = ?
			ORDER BY option_order
		`, q.ID)
		if err == nil {
			options := []models.AnswerOption{}
			for optRows.Next() {
				var opt models.AnswerOption
				var isCorrect int
				optRows.Scan(&opt.ID, &opt.QuestionID, &opt.OptionText, &isCorrect, &opt.OptionOrder)
				opt.IsCorrect = isCorrect == 1
				options = append(options, opt)
			}
			optRows.Close()
			q.AnswerOptions = options
		}

		questions = append(questions, q)
	}
	assessment.Questions = questions

	respondJSON(w, assessment)
}

func UpdateAssessment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req struct {
		Title       string     `json:"title"`
		Description string     `json:"description"`
		Deadline    *time.Time `json:"deadline"`
		Questions   []struct {
			ID            int    `json:"id,omitempty"`
			QuestionText  string `json:"question_text"`
			QuestionOrder int    `json:"question_order"`
			Points        int    `json:"points"`
			AnswerOptions []struct {
				ID          int    `json:"id,omitempty"`
				OptionText  string `json:"option_text"`
				IsCorrect   bool   `json:"is_correct"`
				OptionOrder int    `json:"option_order"`
			} `json:"answer_options"`
		} `json:"questions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := db.DB.Begin()
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Update assessment
	_, err = tx.Exec(`
		UPDATE assessments
		SET title = ?, description = ?, deadline = ?, updated_at = ?
		WHERE id = ?
	`, req.Title, req.Description, req.Deadline, time.Now(), id)
	if err != nil {
		respondError(w, "Failed to update assessment", http.StatusInternalServerError)
		return
	}

	// Delete existing questions and options (cascade will handle options)
	_, err = tx.Exec("DELETE FROM questions WHERE assessment_id = ?", id)
	if err != nil {
		respondError(w, "Failed to update questions", http.StatusInternalServerError)
		return
	}

	// Recreate questions and answer options
	for _, q := range req.Questions {
		qResult, err := tx.Exec(`
			INSERT INTO questions (assessment_id, question_text, question_order, points, created_at)
			VALUES (?, ?, ?, ?, ?)
		`, id, q.QuestionText, q.QuestionOrder, q.Points, time.Now())
		if err != nil {
			respondError(w, "Failed to create question", http.StatusInternalServerError)
			return
		}

		questionID, _ := qResult.LastInsertId()

		for _, opt := range q.AnswerOptions {
			isCorrect := 0
			if opt.IsCorrect {
				isCorrect = 1
			}
			_, err := tx.Exec(`
				INSERT INTO answer_options (question_id, option_text, is_correct, option_order)
				VALUES (?, ?, ?, ?)
			`, questionID, opt.OptionText, isCorrect, opt.OptionOrder)
			if err != nil {
				respondError(w, "Failed to create answer option", http.StatusInternalServerError)
				return
			}
		}
	}

	if err := tx.Commit(); err != nil {
		respondError(w, "Failed to save assessment", http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Assessment updated successfully",
	})
}

func DeleteAssessment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, err := db.DB.Exec("DELETE FROM assessments WHERE id = ?", id)
	if err != nil {
		respondError(w, "Failed to delete assessment", http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Assessment deleted successfully",
	})
}

func PublishAssessment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, err := db.DB.Exec("UPDATE assessments SET is_published = 1, updated_at = ? WHERE id = ?", time.Now(), id)
	if err != nil {
		respondError(w, "Failed to publish assessment", http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Assessment published successfully",
	})
}

func GetAssessmentStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get total users (non-admin)
	var totalUsers int
	db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'user'").Scan(&totalUsers)

	// Get completed attempts
	var completedUsers int
	db.DB.QueryRow(`
		SELECT COUNT(DISTINCT user_id)
		FROM user_assessment_attempts
		WHERE assessment_id = ? AND completed_at IS NOT NULL
	`, id).Scan(&completedUsers)

	// Get pending users
	pendingUsers := totalUsers - completedUsers

	// Get average score
	var avgScore sql.NullFloat64
	db.DB.QueryRow(`
		SELECT AVG(CAST(score AS FLOAT) / CAST(total_points AS FLOAT) * 100)
		FROM user_assessment_attempts
		WHERE assessment_id = ? AND completed_at IS NOT NULL AND total_points > 0
	`, id).Scan(&avgScore)

	averageScore := 0.0
	if avgScore.Valid {
		averageScore = avgScore.Float64
	}

	// Get pass rate (assuming 70% is passing)
	var passCount int
	db.DB.QueryRow(`
		SELECT COUNT(*)
		FROM user_assessment_attempts
		WHERE assessment_id = ? AND completed_at IS NOT NULL
		AND (CAST(score AS FLOAT) / CAST(total_points AS FLOAT) * 100) >= 70
	`, id).Scan(&passCount)

	passRate := 0.0
	if completedUsers > 0 {
		passRate = float64(passCount) / float64(completedUsers) * 100
	}

	stats := models.AssessmentStats{
		TotalUsers:     totalUsers,
		CompletedUsers: completedUsers,
		PendingUsers:   pendingUsers,
		AverageScore:   averageScore,
		PassRate:       passRate,
	}

	respondJSON(w, stats)
}

func GetAssessmentResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	rows, err := db.DB.Query(`
		SELECT
			u.id, u.username,
			COALESCE(a.id, 0) as attempt_id,
			COALESCE(a.score, 0) as score,
			COALESCE(a.total_points, 0) as total_points,
			a.completed_at
		FROM users u
		LEFT JOIN user_assessment_attempts a ON u.id = a.user_id AND a.assessment_id = ?
		WHERE u.role = 'user'
		ORDER BY u.username
	`, id)
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	results := []models.UserAssessmentResult{}
	for rows.Next() {
		var r models.UserAssessmentResult
		var completedAt sql.NullTime
		err := rows.Scan(&r.UserID, &r.Username, &r.AttemptID, &r.Score, &r.TotalPoints, &completedAt)
		if err != nil {
			continue
		}

		if completedAt.Valid {
			r.CompletedAt = &completedAt.Time
			r.Status = "Completed"
		} else if r.AttemptID > 0 {
			r.Status = "In Progress"
		} else {
			r.Status = "Not Started"
		}

		if r.TotalPoints > 0 {
			r.Percentage = float64(r.Score) / float64(r.TotalPoints) * 100
		}

		results = append(results, r)
	}

	respondJSON(w, results)
}

// GenerateDynamicAssessment generates personalized assessments based on user behavior in phishing simulation
func GenerateDynamicAssessment(w http.ResponseWriter, r *http.Request) {
	userID, _, err := getUserFromSession(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		CampaignID int `json:"campaign_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.CampaignID == 0 {
		respondError(w, "Campaign ID is required", http.StatusBadRequest)
		return
	}

	// Get all targets and their behavior from this campaign
	rows, err := db.DB.Query(`
		SELECT DISTINCT
			t.email,
			MAX(CASE WHEN e.message = 'Clicked Link' THEN 1 ELSE 0 END) as clicked,
			MAX(CASE WHEN e.message = 'Submitted Data' THEN 1 ELSE 0 END) as submitted,
			MAX(CASE WHEN e.message = 'Reported Phishing' THEN 1 ELSE 0 END) as reported
		FROM targets t
		JOIN campaign_targets ct ON ct.target_id = t.id
		LEFT JOIN events e ON e.campaign_target_id = ct.id
		WHERE ct.campaign_id = ?
		GROUP BY t.email
	`, req.CampaignID)

	if err != nil {
		respondError(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Analyze behavior and generate assessments for each target
	var generatedCount int
	tx, err := db.DB.Begin()
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	for rows.Next() {
		var email string
		var clicked, submitted, reported int

		if err := rows.Scan(&email, &clicked, &submitted, &reported); err != nil {
			continue
		}

		// Determine behavior type
		var behaviorType, title, description string
		if submitted == 1 {
			behaviorType = "clicked_and_submitted"
			title = "Security Awareness Assessment - Credential Protection"
			description = "Personalized assessment focusing on credential security and phishing awareness"
		} else if clicked == 1 && reported == 1 {
			behaviorType = "clicked_and_reported"
			title = "Security Awareness Assessment - Threat Recognition"
			description = "Personalized assessment focusing on identifying and reporting threats"
		} else if reported == 1 {
			behaviorType = "only_reported"
			title = "Security Awareness Assessment - Proactive Security"
			description = "Personalized assessment recognizing proactive security behavior"
		} else {
			// Skip targets who didn't interact
			continue
		}

		// Generate assessment with Gemini API
		questions, err := generateQuestionsForBehavior(behaviorType, email)
		if err != nil {
			continue
		}

		// Create the assessment
		result, err := tx.Exec(`
			INSERT INTO assessments (title, description, is_published, is_dynamic, behavior_type, target_email, campaign_id, created_by, created_at, updated_at)
			VALUES (?, ?, 1, 1, ?, ?, ?, ?, ?, ?)
		`, title, description, behaviorType, email, req.CampaignID, userID, time.Now(), time.Now())

		if err != nil {
			continue
		}

		assessmentID, _ := result.LastInsertId()

		// Insert generated questions
		for i, q := range questions {
			qResult, err := tx.Exec(`
				INSERT INTO questions (assessment_id, question_text, question_order, points, created_at)
				VALUES (?, ?, ?, ?, ?)
			`, assessmentID, q.QuestionText, i+1, q.Points, time.Now())

			if err != nil {
				continue
			}

			questionID, _ := qResult.LastInsertId()

			// Insert answer options
			for j, opt := range q.AnswerOptions {
				isCorrect := 0
				if opt.IsCorrect {
					isCorrect = 1
				}
				_, err := tx.Exec(`
					INSERT INTO answer_options (question_id, option_text, is_correct, option_order)
					VALUES (?, ?, ?, ?)
				`, questionID, opt.OptionText, isCorrect, j+1)

				if err != nil {
					break
				}
			}
		}

		generatedCount++
	}

	if err := tx.Commit(); err != nil {
		respondError(w, "Failed to save assessments", http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Dynamic assessments generated successfully",
		"count":   generatedCount,
	})
}

type GeneratedQuestion struct {
	QuestionText  string                   `json:"question_text"`
	Points        int                      `json:"points"`
	AnswerOptions []GeneratedAnswerOption  `json:"answer_options"`
}

type GeneratedAnswerOption struct {
	OptionText string `json:"option_text"`
	IsCorrect  bool   `json:"is_correct"`
}

// generateQuestionsForBehavior uses Gemini API to generate personalized questions based on behavior type
func generateQuestionsForBehavior(behaviorType, targetEmail string) ([]GeneratedQuestion, error) {
	// Get Gemini API key from settings
	apiKey := GetGeminiAPIKey()
	if apiKey == "" {
		return nil, http.ErrNotSupported
	}

	// Build behavior-specific prompt
	var prompt string
	switch behaviorType {
	case "clicked_and_submitted":
		prompt = `You are a cybersecurity expert creating a personalized security awareness assessment for an employee who clicked on a phishing link AND submitted their credentials.

This person needs focused training on:
1. Recognizing phishing emails
2. Understanding the consequences of credential theft
3. Learning to verify legitimacy before entering credentials
4. Best practices for password security

Generate 5 multiple-choice questions that:
- Test understanding of phishing tactics
- Reinforce why submitting credentials is dangerous
- Teach how to verify legitimate requests
- Cover password security best practices

Each question should have 4 answer options with only 1 correct answer.

Respond in JSON format:
{
  "questions": [
    {
      "question_text": "...",
      "points": 1,
      "answer_options": [
        {"option_text": "...", "is_correct": true},
        {"option_text": "...", "is_correct": false},
        {"option_text": "...", "is_correct": false},
        {"option_text": "...", "is_correct": false}
      ]
    }
  ]
}`

	case "clicked_and_reported":
		prompt = `You are a cybersecurity expert creating a personalized security awareness assessment for an employee who clicked on a phishing link BUT then reported it as suspicious.

This person showed good judgment by reporting but needs training on:
1. Identifying phishing emails BEFORE clicking
2. Understanding red flags in emails and links
3. Safe browsing practices
4. When and how to report suspicious emails

Generate 5 multiple-choice questions that:
- Teach early phishing detection
- Reinforce the importance of reporting
- Cover safe link-clicking practices
- Test understanding of email security indicators

Each question should have 4 answer options with only 1 correct answer.

Respond in JSON format:
{
  "questions": [
    {
      "question_text": "...",
      "points": 1,
      "answer_options": [
        {"option_text": "...", "is_correct": true},
        {"option_text": "...", "is_correct": false},
        {"option_text": "...", "is_correct": false},
        {"option_text": "...", "is_correct": false}
      ]
    }
  ]
}`

	case "only_reported":
		prompt = `You are a cybersecurity expert creating a personalized security awareness assessment for an employee who immediately reported a phishing email without clicking.

This person showed excellent security awareness! The assessment should:
1. Reinforce their good behavior
2. Test advanced phishing detection skills
3. Cover emerging phishing tactics
4. Explore social engineering techniques

Generate 5 multiple-choice questions that:
- Test advanced phishing recognition
- Cover sophisticated social engineering tactics
- Explore spear phishing and CEO fraud
- Include scenario-based questions

Each question should have 4 answer options with only 1 correct answer.

Respond in JSON format:
{
  "questions": [
    {
      "question_text": "...",
      "points": 1,
      "answer_options": [
        {"option_text": "...", "is_correct": true},
        {"option_text": "...", "is_correct": false},
        {"option_text": "...", "is_correct": false},
        {"option_text": "...", "is_correct": false}
      ]
    }
  ]
}`

	default:
		return nil, http.ErrNotSupported
	}

	// Call Gemini API using the existing structure from gemini.go
	geminiReq := struct {
		Contents []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"contents"`
	}{
		Contents: []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			{
				Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(geminiReq)
	geminiURL := "https://generativelanguage.googleapis.com/v1/models/gemini-2.5-flash:generateContent?key=" + apiKey

	resp, err := http.Post(geminiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, http.ErrNotSupported
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, http.ErrNotSupported
	}

	generatedText := geminiResp.Candidates[0].Content.Parts[0].Text

	// Clean up markdown code blocks
	cleanedText := generatedText
	if bytes.Contains([]byte(generatedText), []byte("```json")) {
		start := bytes.Index([]byte(generatedText), []byte("```json"))
		end := bytes.LastIndex([]byte(generatedText), []byte("```"))
		if start != -1 && end != -1 && end > start {
			cleanedText = generatedText[start+7 : end]
		}
	} else if bytes.Contains([]byte(generatedText), []byte("```")) {
		start := bytes.Index([]byte(generatedText), []byte("```"))
		end := bytes.LastIndex([]byte(generatedText), []byte("```"))
		if start != -1 && end != -1 && end > start {
			cleanedText = generatedText[start+3 : end]
		}
	}

	// Parse the response
	var result struct {
		Questions []GeneratedQuestion `json:"questions"`
	}

	if err := json.Unmarshal([]byte(cleanedText), &result); err != nil {
		return nil, err
	}

	return result.Questions, nil
}
