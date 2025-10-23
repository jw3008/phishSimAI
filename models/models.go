package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	APIKey       string    `json:"api_key,omitempty"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Campaign struct {
	ID            int        `json:"id"`
	Name          string     `json:"name"`
	Status        string     `json:"status"`
	CreatedDate   time.Time  `json:"created_date"`
	LaunchDate    *time.Time `json:"launch_date,omitempty"`
	CompletedDate *time.Time `json:"completed_date,omitempty"`
	TemplateID    int        `json:"template_id"`
	PageID        int        `json:"page_id"`
	SMTPID        int        `json:"smtp_id"`
	URL           string     `json:"url"`
	UserID        int        `json:"user_id"`
	Template      *Template  `json:"template,omitempty"`
	Page          *Page      `json:"page,omitempty"`
	SMTP          *SMTP      `json:"smtp,omitempty"`
	Groups        []int      `json:"groups,omitempty"`
	Results       []Result   `json:"results,omitempty"`
	Stats         *Stats     `json:"stats,omitempty"`
}

type Template struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Subject   string    `json:"subject"`
	Text      string    `json:"text"`
	HTML      string    `json:"html"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Page struct {
	ID                 int       `json:"id"`
	Name               string    `json:"name"`
	HTML               string    `json:"html"`
	CaptureCredentials bool      `json:"capture_credentials"`
	CapturePasswords   bool      `json:"capture_passwords"`
	RedirectURL        string    `json:"redirect_url"`
	UserID             int       `json:"user_id"`
	CreatedAt          time.Time `json:"created_at"`
}

type Group struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	UserID    int       `json:"user_id"`
	Targets   []Target  `json:"targets,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Target struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Position  string `json:"position"`
	GroupID   int    `json:"group_id"`
}

type CampaignTarget struct {
	ID         int       `json:"id"`
	CampaignID int       `json:"campaign_id"`
	TargetID   int       `json:"target_id"`
	RID        string    `json:"rid"`
	Status     string    `json:"status"`
	SendDate   time.Time `json:"send_date"`
}

type Event struct {
	ID               int       `json:"id"`
	CampaignID       int       `json:"campaign_id"`
	CampaignTargetID int       `json:"campaign_target_id"`
	Email            string    `json:"email"`
	Time             time.Time `json:"time"`
	Message          string    `json:"message"`
	Details          string    `json:"details"`
}

type Result struct {
	ID         int       `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
	Position   string    `json:"position"`
	Status     string    `json:"status"`
	SendDate   time.Time `json:"send_date,omitempty"`
	OpenDate   time.Time `json:"open_date,omitempty"`
	ClickDate  time.Time `json:"click_date,omitempty"`
	SubmitDate time.Time `json:"submit_date,omitempty"`
}

type Stats struct {
	Total       int `json:"total"`
	Sent        int `json:"sent"`
	Opened      int `json:"opened"`
	Clicked     int `json:"clicked"`
	Submitted   int `json:"submitted"`
	Error       int `json:"error"`
	OpenRate    int `json:"open_rate"`
	ClickRate   int `json:"click_rate"`
	SubmitRate  int `json:"submit_rate"`
}

type SMTP struct {
	ID               int       `json:"id"`
	Name             string    `json:"name"`
	Host             string    `json:"host"`
	Username         string    `json:"username"`
	Password         string    `json:"password,omitempty"`
	FromAddress      string    `json:"from_address"`
	UserID           int       `json:"user_id"`
	IgnoreCertErrors bool      `json:"ignore_cert_errors"`
	CreatedAt        time.Time `json:"created_at"`
}

type Assessment struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	IsPublished bool       `json:"is_published"`
	CreatedBy   int        `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Questions   []Question `json:"questions,omitempty"`
}

type Question struct {
	ID            int            `json:"id"`
	AssessmentID  int            `json:"assessment_id"`
	QuestionText  string         `json:"question_text"`
	QuestionOrder int            `json:"question_order"`
	Points        int            `json:"points"`
	CreatedAt     time.Time      `json:"created_at"`
	AnswerOptions []AnswerOption `json:"answer_options,omitempty"`
}

type AnswerOption struct {
	ID          int    `json:"id"`
	QuestionID  int    `json:"question_id"`
	OptionText  string `json:"option_text"`
	IsCorrect   bool   `json:"is_correct,omitempty"`
	OptionOrder int    `json:"option_order"`
}

type UserAssessmentAttempt struct {
	ID           int        `json:"id"`
	UserID       int        `json:"user_id"`
	AssessmentID int        `json:"assessment_id"`
	StartedAt    time.Time  `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	Score        int        `json:"score"`
	TotalPoints  int        `json:"total_points"`
	Responses    []UserResponse `json:"responses,omitempty"`
}

type UserResponse struct {
	ID               int       `json:"id"`
	AttemptID        int       `json:"attempt_id"`
	QuestionID       int       `json:"question_id"`
	SelectedOptionID int       `json:"selected_option_id"`
	IsCorrect        bool      `json:"is_correct"`
	PointsEarned     int       `json:"points_earned"`
	AnsweredAt       time.Time `json:"answered_at"`
}

type AssessmentStats struct {
	TotalUsers     int     `json:"total_users"`
	CompletedUsers int     `json:"completed_users"`
	PendingUsers   int     `json:"pending_users"`
	AverageScore   float64 `json:"average_score"`
	PassRate       float64 `json:"pass_rate"`
}

type UserAssessmentResult struct {
	UserID       int        `json:"user_id"`
	Username     string     `json:"username"`
	AttemptID    int        `json:"attempt_id"`
	Score        int        `json:"score"`
	TotalPoints  int        `json:"total_points"`
	Percentage   float64    `json:"percentage"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	Status       string     `json:"status"`
}
