package main

import (
	"log"
	"net/http"
	"os"

	"github.com/clariphish/clariphish/api"
	"github.com/clariphish/clariphish/db"
	"github.com/gorilla/mux"
)

func main() {
	// Initialize database
	if err := db.Init(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Create default admin user if none exists
	if err := db.CreateDefaultAdmin(); err != nil {
		log.Printf("Admin user setup: %v", err)
	}

	// Setup router
	r := mux.NewRouter()

	// API routes
	apiRouter := r.PathPrefix("/api").Subrouter()
	api.RegisterRoutes(apiRouter)

	// Static files and frontend
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	r.PathPrefix("/").HandlerFunc(api.ServeIndex)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "3333"
	}

	log.Printf("ClaripHish server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
