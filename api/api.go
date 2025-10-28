package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("clariphish-secret-key-change-this"))

func RegisterRoutes(r *mux.Router) {
	// Auth routes
	r.HandleFunc("/login", Login).Methods("POST")
	r.HandleFunc("/logout", Logout).Methods("POST")
	r.HandleFunc("/user", GetCurrentUser).Methods("GET")

	// Campaign routes (Admin only)
	r.HandleFunc("/campaigns", RequireAdmin(GetCampaigns)).Methods("GET")
	r.HandleFunc("/campaigns", RequireAdmin(CreateCampaign)).Methods("POST")
	r.HandleFunc("/campaigns/{id}", RequireAdmin(GetCampaign)).Methods("GET")
	r.HandleFunc("/campaigns/{id}", RequireAdmin(UpdateCampaign)).Methods("PUT")
	r.HandleFunc("/campaigns/{id}", RequireAdmin(DeleteCampaign)).Methods("DELETE")
	r.HandleFunc("/campaigns/{id}/complete", RequireAdmin(CompleteCampaign)).Methods("POST")
	r.HandleFunc("/campaigns/{id}/pdf", RequireAdmin(GenerateCampaignReportPDF)).Methods("GET")
	r.HandleFunc("/campaigns/{id}/credentials", RequireAdmin(GetCampaignCredentials)).Methods("GET")
	r.HandleFunc("/campaigns/{id}/credentials-pdf", RequireAdmin(GenerateCredentialsPDF)).Methods("GET")
	r.HandleFunc("/campaigns/{id}/end", RequireAdmin(EndCampaignEarly)).Methods("POST")
	r.HandleFunc("/campaigns/{id}/launch", RequireAdmin(LaunchCampaignNow)).Methods("POST")
	r.HandleFunc("/campaigns/{id}/debug-events", RequireAdmin(GetCampaignEvents)).Methods("GET")
	r.HandleFunc("/campaigns/{id}/verify-timestamps", RequireAdmin(VerifyCampaignTimestamps)).Methods("GET")
	r.HandleFunc("/campaigns/{id}/cleanup-timestamps", RequireAdmin(CleanupBadTimestamps)).Methods("POST")

	// Template routes (Admin only)
	r.HandleFunc("/templates", RequireAdmin(GetTemplates)).Methods("GET")
	r.HandleFunc("/templates", RequireAdmin(CreateTemplate)).Methods("POST")
	// Specific routes MUST come before parameterized routes
	r.HandleFunc("/templates/generate-random", RequireAdmin(GenerateRandomPhishingTemplate)).Methods("GET")
	r.HandleFunc("/templates/generate", RequireAdmin(GenerateTemplateWithGemini)).Methods("POST")
	// Parameterized routes come last
	r.HandleFunc("/templates/{id}", RequireAdmin(GetTemplate)).Methods("GET")
	r.HandleFunc("/templates/{id}", RequireAdmin(UpdateTemplate)).Methods("PUT")
	r.HandleFunc("/templates/{id}", RequireAdmin(DeleteTemplate)).Methods("DELETE")

	// Page routes (Admin only)
	r.HandleFunc("/pages", RequireAdmin(GetPages)).Methods("GET")
	r.HandleFunc("/pages", RequireAdmin(CreatePage)).Methods("POST")
	// Specific routes MUST come before parameterized routes
	r.HandleFunc("/pages/clone", RequireAdmin(CloneLandingPage)).Methods("POST")
	// Parameterized routes come last
	r.HandleFunc("/pages/{id}", RequireAdmin(GetPage)).Methods("GET")
	r.HandleFunc("/pages/{id}", RequireAdmin(UpdatePage)).Methods("PUT")
	r.HandleFunc("/pages/{id}", RequireAdmin(DeletePage)).Methods("DELETE")

	// Group routes (Admin only)
	r.HandleFunc("/groups", RequireAdmin(GetGroups)).Methods("GET")
	r.HandleFunc("/groups", RequireAdmin(CreateGroup)).Methods("POST")
	r.HandleFunc("/groups/{id}", RequireAdmin(GetGroup)).Methods("GET")
	r.HandleFunc("/groups/{id}", RequireAdmin(UpdateGroup)).Methods("PUT")
	r.HandleFunc("/groups/{id}", RequireAdmin(DeleteGroup)).Methods("DELETE")

	// SMTP routes (Admin only)
	r.HandleFunc("/smtp", RequireAdmin(GetSMTPConfigs)).Methods("GET")
	r.HandleFunc("/smtp", RequireAdmin(CreateSMTPConfig)).Methods("POST")
	r.HandleFunc("/smtp/{id}", RequireAdmin(GetSMTPConfig)).Methods("GET")
	r.HandleFunc("/smtp/{id}", RequireAdmin(UpdateSMTPConfig)).Methods("PUT")
	r.HandleFunc("/smtp/{id}", RequireAdmin(DeleteSMTPConfig)).Methods("DELETE")

	// User Management routes (Admin only)
	r.HandleFunc("/users", RequireAdmin(GetUsers)).Methods("GET")
	r.HandleFunc("/users", RequireAdmin(CreateUser)).Methods("POST")
	r.HandleFunc("/users/{id}", RequireAdmin(GetUser)).Methods("GET")
	r.HandleFunc("/users/{id}", RequireAdmin(UpdateUser)).Methods("PUT")
	r.HandleFunc("/users/{id}", RequireAdmin(DeleteUser)).Methods("DELETE")

	// Assessment routes (Admin)
	r.HandleFunc("/assessments", RequireAdmin(GetAssessments)).Methods("GET")
	r.HandleFunc("/assessments", RequireAdmin(CreateAssessment)).Methods("POST")
	r.HandleFunc("/assessments/{id}", RequireAdmin(GetAssessment)).Methods("GET")
	r.HandleFunc("/assessments/{id}", RequireAdmin(UpdateAssessment)).Methods("PUT")
	r.HandleFunc("/assessments/{id}", RequireAdmin(DeleteAssessment)).Methods("DELETE")
	r.HandleFunc("/assessments/{id}/publish", RequireAdmin(PublishAssessment)).Methods("POST")
	r.HandleFunc("/assessments/{id}/stats", RequireAdmin(GetAssessmentStats)).Methods("GET")
	r.HandleFunc("/assessments/{id}/results", RequireAdmin(GetAssessmentResults)).Methods("GET")
	r.HandleFunc("/assessments/{id}/pdf", RequireAdmin(GenerateAssessmentOverviewPDF)).Methods("GET")

	// User Assessment routes (All authenticated users)
	r.HandleFunc("/user/assessments", RequireAuth(GetUserAssessments)).Methods("GET")
	r.HandleFunc("/user/assessments/{id}", RequireAuth(GetUserAssessmentDetail)).Methods("GET")
	r.HandleFunc("/user/assessments/{id}/start", RequireAuth(StartAssessment)).Methods("POST")
	r.HandleFunc("/user/assessments/attempt/{attemptId}/submit", RequireAuth(SubmitAssessmentResponse)).Methods("POST")
	r.HandleFunc("/user/assessments/attempt/{attemptId}/complete", RequireAuth(CompleteAssessment)).Methods("POST")
	r.HandleFunc("/user/results", RequireAuth(GetUserResults)).Methods("GET")
	r.HandleFunc("/user/results/{attemptId}", RequireAuth(GetUserResultDetail)).Methods("GET")
	r.HandleFunc("/user/results/{attemptId}/pdf", RequireAuth(GenerateResultPDF)).Methods("GET")

	// Knowledge Base Chatbot (All authenticated users)
	r.HandleFunc("/knowledge-base/chat", RequireAuth(KnowledgeBaseChat)).Methods("POST")

	// Tracking routes (public)
	r.HandleFunc("/track", TrackOpen).Methods("GET")
	r.HandleFunc("/click", TrackClick).Methods("GET")
	r.HandleFunc("/report", TrackSubmission).Methods("POST")
	r.HandleFunc("/report-phishing", TrackReportPhishing).Methods("GET")
}

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/index.html")
}

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "clariphish-session")
		userID, ok := session.Values["user_id"]
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add user ID to request context if needed
		log.Printf("User %v accessing %s", userID, r.URL.Path)
		next(w, r)
	}
}

func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "clariphish-session")
		userID, ok := session.Values["user_id"]
		if !ok {
			respondError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		role, ok := session.Values["role"]
		if !ok || role != "admin" {
			respondError(w, "Forbidden: Admin access required", http.StatusForbidden)
			return
		}

		log.Printf("Admin user %v accessing %s", userID, r.URL.Path)
		next(w, r)
	}
}

func getUserFromSession(r *http.Request) (int, string, error) {
	session, _ := store.Get(r, "clariphish-session")
	userID, ok := session.Values["user_id"]
	if !ok {
		return 0, "", http.ErrNoCookie
	}

	role, ok := session.Values["role"]
	if !ok {
		role = "user"
	}

	return userID.(int), role.(string), nil
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
