package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/clariphish/clariphish/db"
	"github.com/clariphish/clariphish/models"
	"github.com/gorilla/mux"
)

// User Assessment Taking

func GetUserAssessments(w http.ResponseWriter, r *http.Request) {
	userID, _, err := getUserFromSession(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := db.DB.Query(`
		SELECT
			a.id, a.title, a.description, a.deadline, a.is_published,
			COALESCE(att.id, 0) as attempt_id,
			att.completed_at,
			COALESCE(att.score, 0) as score,
			COALESCE(att.total_points, 0) as total_points
		FROM assessments a
		LEFT JOIN user_assessment_attempts att ON a.id = att.assessment_id AND att.user_id = ?
		WHERE a.is_published = 1
		ORDER BY a.deadline ASC
	`, userID)
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type AssessmentWithStatus struct {
		models.Assessment
		AttemptID   int        `json:"attempt_id"`
		Status      string     `json:"status"`
		Score       int        `json:"score"`
		TotalPoints int        `json:"total_points"`
		Percentage  float64    `json:"percentage"`
		CompletedAt *time.Time `json:"completed_at,omitempty"`
	}

	assessments := []AssessmentWithStatus{}
	for rows.Next() {
		var a AssessmentWithStatus
		var deadline sql.NullTime
		var completedAt sql.NullTime
		err := rows.Scan(&a.ID, &a.Title, &a.Description, &deadline, &a.IsPublished,
			&a.AttemptID, &completedAt, &a.Score, &a.TotalPoints)
		if err != nil {
			continue
		}

		if deadline.Valid {
			a.Deadline = &deadline.Time
		}

		if completedAt.Valid {
			a.CompletedAt = &completedAt.Time
			a.Status = "Completed"
		} else if a.AttemptID > 0 {
			a.Status = "In Progress"
		} else {
			a.Status = "Not Started"
		}

		if a.TotalPoints > 0 {
			a.Percentage = float64(a.Score) / float64(a.TotalPoints) * 100
		}

		assessments = append(assessments, a)
	}

	respondJSON(w, assessments)
}

func GetUserAssessmentDetail(w http.ResponseWriter, r *http.Request) {
	userID, _, err := getUserFromSession(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]

	// Check if assessment is published
	var isPublished int
	err = db.DB.QueryRow("SELECT is_published FROM assessments WHERE id = ?", id).Scan(&isPublished)
	if err == sql.ErrNoRows || isPublished == 0 {
		respondError(w, "Assessment not found or not published", http.StatusNotFound)
		return
	}

	// Get assessment details without correct answers
	var assessment models.Assessment
	var deadline sql.NullTime
	err = db.DB.QueryRow(`
		SELECT id, title, description, deadline, is_published, created_at
		FROM assessments WHERE id = ?
	`, id).Scan(&assessment.ID, &assessment.Title, &assessment.Description, &deadline, &assessment.IsPublished, &assessment.CreatedAt)

	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}

	if deadline.Valid {
		assessment.Deadline = &deadline.Time
	}

	// Get questions (without showing correct answers to users)
	qRows, err := db.DB.Query(`
		SELECT id, assessment_id, question_text, question_order, points
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
		qRows.Scan(&q.ID, &q.AssessmentID, &q.QuestionText, &q.QuestionOrder, &q.Points)

		// Get answer options (without is_correct field for security)
		optRows, err := db.DB.Query(`
			SELECT id, question_id, option_text, option_order
			FROM answer_options WHERE question_id = ?
			ORDER BY option_order
		`, q.ID)
		if err == nil {
			options := []models.AnswerOption{}
			for optRows.Next() {
				var opt models.AnswerOption
				optRows.Scan(&opt.ID, &opt.QuestionID, &opt.OptionText, &opt.OptionOrder)
				opt.IsCorrect = false // Don't expose correct answers
				options = append(options, opt)
			}
			optRows.Close()
			q.AnswerOptions = options
		}

		questions = append(questions, q)
	}
	assessment.Questions = questions

	// Check if user has an existing attempt
	var attemptID int
	var completedAt sql.NullTime
	err = db.DB.QueryRow(`
		SELECT id, completed_at FROM user_assessment_attempts
		WHERE user_id = ? AND assessment_id = ?
		ORDER BY started_at DESC LIMIT 1
	`, userID, id).Scan(&attemptID, &completedAt)

	response := map[string]interface{}{
		"assessment": assessment,
		"attempt_id": attemptID,
	}

	if completedAt.Valid {
		response["status"] = "completed"
	} else if attemptID > 0 {
		response["status"] = "in_progress"
	} else {
		response["status"] = "not_started"
	}

	respondJSON(w, response)
}

func StartAssessment(w http.ResponseWriter, r *http.Request) {
	userID, _, err := getUserFromSession(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	assessmentID := vars["id"]

	// Check if assessment is published
	var isPublished int
	err = db.DB.QueryRow("SELECT is_published FROM assessments WHERE id = ?", assessmentID).Scan(&isPublished)
	if err == sql.ErrNoRows || isPublished == 0 {
		respondError(w, "Assessment not found or not published", http.StatusNotFound)
		return
	}

	// Check if user already has an attempt
	var existingAttemptID int
	var completedAt sql.NullTime
	err = db.DB.QueryRow(`
		SELECT id, completed_at FROM user_assessment_attempts
		WHERE user_id = ? AND assessment_id = ?
		ORDER BY started_at DESC LIMIT 1
	`, userID, assessmentID).Scan(&existingAttemptID, &completedAt)

	if err == nil && completedAt.Valid {
		respondError(w, "You have already completed this assessment", http.StatusBadRequest)
		return
	}

	if err == nil && existingAttemptID > 0 {
		// Return existing in-progress attempt
		respondJSON(w, map[string]interface{}{
			"success":    true,
			"attempt_id": existingAttemptID,
			"message":    "Continuing existing attempt",
		})
		return
	}

	// Create new attempt
	result, err := db.DB.Exec(`
		INSERT INTO user_assessment_attempts (user_id, assessment_id, started_at)
		VALUES (?, ?, ?)
	`, userID, assessmentID, time.Now())
	if err != nil {
		respondError(w, "Failed to start assessment", http.StatusInternalServerError)
		return
	}

	attemptID, _ := result.LastInsertId()

	respondJSON(w, map[string]interface{}{
		"success":    true,
		"attempt_id": attemptID,
		"message":    "Assessment started successfully",
	})
}

func SubmitAssessmentResponse(w http.ResponseWriter, r *http.Request) {
	userID, _, err := getUserFromSession(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	attemptID := vars["attemptId"]

	var req struct {
		QuestionID       int `json:"question_id"`
		SelectedOptionID int `json:"selected_option_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Verify the attempt belongs to this user and is not completed
	var attemptUserID int
	var completedAt sql.NullTime
	err = db.DB.QueryRow(`
		SELECT user_id, completed_at FROM user_assessment_attempts WHERE id = ?
	`, attemptID).Scan(&attemptUserID, &completedAt)

	if err == sql.ErrNoRows {
		respondError(w, "Attempt not found", http.StatusNotFound)
		return
	}

	if attemptUserID != userID {
		respondError(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if completedAt.Valid {
		respondError(w, "Assessment already completed", http.StatusBadRequest)
		return
	}

	// Check if the selected option is correct
	var isCorrect int
	var questionPoints int
	err = db.DB.QueryRow(`
		SELECT ao.is_correct, q.points
		FROM answer_options ao
		JOIN questions q ON ao.question_id = q.id
		WHERE ao.id = ? AND ao.question_id = ?
	`, req.SelectedOptionID, req.QuestionID).Scan(&isCorrect, &questionPoints)

	if err != nil {
		respondError(w, "Invalid option selected", http.StatusBadRequest)
		return
	}

	pointsEarned := 0
	if isCorrect == 1 {
		pointsEarned = questionPoints
	}

	// Delete existing response for this question if any
	db.DB.Exec("DELETE FROM user_responses WHERE attempt_id = ? AND question_id = ?", attemptID, req.QuestionID)

	// Insert response
	_, err = db.DB.Exec(`
		INSERT INTO user_responses (attempt_id, question_id, selected_option_id, is_correct, points_earned, answered_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, attemptID, req.QuestionID, req.SelectedOptionID, isCorrect, pointsEarned, time.Now())

	if err != nil {
		respondError(w, "Failed to save response", http.StatusInternalServerError)
		return
	}

	respondJSON(w, map[string]interface{}{
		"success": true,
		"message": "Response saved successfully",
	})
}

func CompleteAssessment(w http.ResponseWriter, r *http.Request) {
	userID, _, err := getUserFromSession(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	attemptID := vars["attemptId"]

	// Verify the attempt belongs to this user
	var attemptUserID int
	var completedAt sql.NullTime
	var assessmentID int
	err = db.DB.QueryRow(`
		SELECT user_id, completed_at, assessment_id FROM user_assessment_attempts WHERE id = ?
	`, attemptID).Scan(&attemptUserID, &completedAt, &assessmentID)

	if err == sql.ErrNoRows {
		respondError(w, "Attempt not found", http.StatusNotFound)
		return
	}

	if attemptUserID != userID {
		respondError(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if completedAt.Valid {
		respondError(w, "Assessment already completed", http.StatusBadRequest)
		return
	}

	// Calculate total score
	var totalScore int
	db.DB.QueryRow(`
		SELECT COALESCE(SUM(points_earned), 0)
		FROM user_responses WHERE attempt_id = ?
	`, attemptID).Scan(&totalScore)

	// Calculate total possible points
	var totalPoints int
	db.DB.QueryRow(`
		SELECT COALESCE(SUM(points), 0)
		FROM questions WHERE assessment_id = ?
	`, assessmentID).Scan(&totalPoints)

	// Update attempt
	_, err = db.DB.Exec(`
		UPDATE user_assessment_attempts
		SET completed_at = ?, score = ?, total_points = ?
		WHERE id = ?
	`, time.Now(), totalScore, totalPoints, attemptID)

	if err != nil {
		respondError(w, "Failed to complete assessment", http.StatusInternalServerError)
		return
	}

	percentage := 0.0
	if totalPoints > 0 {
		percentage = float64(totalScore) / float64(totalPoints) * 100
	}

	respondJSON(w, map[string]interface{}{
		"success":     true,
		"message":     "Assessment completed successfully",
		"score":       totalScore,
		"total":       totalPoints,
		"percentage":  percentage,
		"attempt_id":  attemptID,
	})
}

func GetUserResults(w http.ResponseWriter, r *http.Request) {
	userID, _, err := getUserFromSession(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := db.DB.Query(`
		SELECT
			att.id, att.assessment_id, a.title, att.score, att.total_points, att.completed_at
		FROM user_assessment_attempts att
		JOIN assessments a ON att.assessment_id = a.id
		WHERE att.user_id = ? AND att.completed_at IS NOT NULL
		ORDER BY att.completed_at DESC
	`, userID)
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ResultSummary struct {
		AttemptID    int        `json:"attempt_id"`
		AssessmentID int        `json:"assessment_id"`
		Title        string     `json:"title"`
		Score        int        `json:"score"`
		TotalPoints  int        `json:"total_points"`
		Percentage   float64    `json:"percentage"`
		CompletedAt  *time.Time `json:"completed_at"`
	}

	results := []ResultSummary{}
	for rows.Next() {
		var r ResultSummary
		var completedAt sql.NullTime
		rows.Scan(&r.AttemptID, &r.AssessmentID, &r.Title, &r.Score, &r.TotalPoints, &completedAt)

		if completedAt.Valid {
			r.CompletedAt = &completedAt.Time
		}

		if r.TotalPoints > 0 {
			r.Percentage = float64(r.Score) / float64(r.TotalPoints) * 100
		}

		results = append(results, r)
	}

	respondJSON(w, results)
}

func GetUserResultDetail(w http.ResponseWriter, r *http.Request) {
	userID, _, err := getUserFromSession(r)
	if err != nil {
		respondError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	attemptID := vars["attemptId"]

	// Verify the attempt belongs to this user
	var attempt models.UserAssessmentAttempt
	var completedAt sql.NullTime
	err = db.DB.QueryRow(`
		SELECT id, user_id, assessment_id, started_at, completed_at, score, total_points
		FROM user_assessment_attempts WHERE id = ? AND user_id = ?
	`, attemptID, userID).Scan(&attempt.ID, &attempt.UserID, &attempt.AssessmentID, &attempt.StartedAt, &completedAt, &attempt.Score, &attempt.TotalPoints)

	if err == sql.ErrNoRows {
		respondError(w, "Result not found", http.StatusNotFound)
		return
	}
	if err != nil {
		respondError(w, "Database error", http.StatusInternalServerError)
		return
	}

	if completedAt.Valid {
		attempt.CompletedAt = &completedAt.Time
	}

	// Get assessment info
	var assessmentTitle string
	db.DB.QueryRow("SELECT title FROM assessments WHERE id = ?", attempt.AssessmentID).Scan(&assessmentTitle)

	// Get responses with question details
	respRows, err := db.DB.Query(`
		SELECT
			r.question_id, q.question_text, q.points,
			r.selected_option_id, ao.option_text,
			r.is_correct, r.points_earned,
			(SELECT option_text FROM answer_options WHERE question_id = r.question_id AND is_correct = 1 LIMIT 1) as correct_answer
		FROM user_responses r
		JOIN questions q ON r.question_id = q.id
		JOIN answer_options ao ON r.selected_option_id = ao.id
		WHERE r.attempt_id = ?
		ORDER BY q.question_order
	`, attemptID)

	type ResponseDetail struct {
		QuestionID       int    `json:"question_id"`
		QuestionText     string `json:"question_text"`
		QuestionPoints   int    `json:"question_points"`
		SelectedOption   string `json:"selected_option"`
		CorrectAnswer    string `json:"correct_answer"`
		IsCorrect        bool   `json:"is_correct"`
		PointsEarned     int    `json:"points_earned"`
	}

	responses := []ResponseDetail{}
	if err == nil {
		defer respRows.Close()
		for respRows.Next() {
			var rd ResponseDetail
			var selectedOptionID int
			var isCorrect int
			respRows.Scan(&rd.QuestionID, &rd.QuestionText, &rd.QuestionPoints,
				&selectedOptionID, &rd.SelectedOption,
				&isCorrect, &rd.PointsEarned, &rd.CorrectAnswer)
			rd.IsCorrect = isCorrect == 1
			responses = append(responses, rd)
		}
	}

	percentage := 0.0
	if attempt.TotalPoints > 0 {
		percentage = float64(attempt.Score) / float64(attempt.TotalPoints) * 100
	}

	respondJSON(w, map[string]interface{}{
		"attempt_id":       attempt.ID,
		"assessment_id":    attempt.AssessmentID,
		"assessment_title": assessmentTitle,
		"started_at":       attempt.StartedAt,
		"completed_at":     attempt.CompletedAt,
		"score":            attempt.Score,
		"total_points":     attempt.TotalPoints,
		"percentage":       percentage,
		"responses":        responses,
	})
}

// GenerateResultPDF is implemented in pdf_reports.go
