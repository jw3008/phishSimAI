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

	// Campaign routes
	r.HandleFunc("/campaigns", RequireAuth(GetCampaigns)).Methods("GET")
	r.HandleFunc("/campaigns", RequireAuth(CreateCampaign)).Methods("POST")
	r.HandleFunc("/campaigns/{id}", RequireAuth(GetCampaign)).Methods("GET")
	r.HandleFunc("/campaigns/{id}", RequireAuth(UpdateCampaign)).Methods("PUT")
	r.HandleFunc("/campaigns/{id}", RequireAuth(DeleteCampaign)).Methods("DELETE")
	r.HandleFunc("/campaigns/{id}/complete", RequireAuth(CompleteCampaign)).Methods("POST")

	// Template routes
	r.HandleFunc("/templates", RequireAuth(GetTemplates)).Methods("GET")
	r.HandleFunc("/templates", RequireAuth(CreateTemplate)).Methods("POST")
	r.HandleFunc("/templates/{id}", RequireAuth(GetTemplate)).Methods("GET")
	r.HandleFunc("/templates/{id}", RequireAuth(UpdateTemplate)).Methods("PUT")
	r.HandleFunc("/templates/{id}", RequireAuth(DeleteTemplate)).Methods("DELETE")

	// Page routes
	r.HandleFunc("/pages", RequireAuth(GetPages)).Methods("GET")
	r.HandleFunc("/pages", RequireAuth(CreatePage)).Methods("POST")
	r.HandleFunc("/pages/{id}", RequireAuth(GetPage)).Methods("GET")
	r.HandleFunc("/pages/{id}", RequireAuth(UpdatePage)).Methods("PUT")
	r.HandleFunc("/pages/{id}", RequireAuth(DeletePage)).Methods("DELETE")

	// Group routes
	r.HandleFunc("/groups", RequireAuth(GetGroups)).Methods("GET")
	r.HandleFunc("/groups", RequireAuth(CreateGroup)).Methods("POST")
	r.HandleFunc("/groups/{id}", RequireAuth(GetGroup)).Methods("GET")
	r.HandleFunc("/groups/{id}", RequireAuth(UpdateGroup)).Methods("PUT")
	r.HandleFunc("/groups/{id}", RequireAuth(DeleteGroup)).Methods("DELETE")

	// SMTP routes
	r.HandleFunc("/smtp", RequireAuth(GetSMTPConfigs)).Methods("GET")
	r.HandleFunc("/smtp", RequireAuth(CreateSMTPConfig)).Methods("POST")
	r.HandleFunc("/smtp/{id}", RequireAuth(GetSMTPConfig)).Methods("GET")
	r.HandleFunc("/smtp/{id}", RequireAuth(UpdateSMTPConfig)).Methods("PUT")
	r.HandleFunc("/smtp/{id}", RequireAuth(DeleteSMTPConfig)).Methods("DELETE")

	// Tracking routes (public)
	r.HandleFunc("/track", TrackOpen).Methods("GET")
	r.HandleFunc("/click", TrackClick).Methods("GET")
	r.HandleFunc("/report", TrackSubmission).Methods("POST")
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

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
