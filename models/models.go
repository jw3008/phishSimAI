package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	APIKey       string    `json:"api_key,omitempty"`
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
