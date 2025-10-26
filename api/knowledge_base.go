package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// KnowledgeBaseChat handles chatbot questions about phishing and security
func KnowledgeBaseChat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Question string `json:"question"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Question == "" {
		respondError(w, "Question is required", http.StatusBadRequest)
		return
	}

	// Get Gemini API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		// Check if provided in request header
		apiKey = r.Header.Get("X-Gemini-API-Key")
		if apiKey == "" {
			respondError(w, "Gemini API key not configured. Please set GEMINI_API_KEY environment variable or provide X-Gemini-API-Key header", http.StatusInternalServerError)
			return
		}
	}

	// Create system prompt for security knowledge base
	systemPrompt := `You are a cybersecurity expert assistant specializing in phishing detection and security awareness training.
Your role is to help users understand:
- What phishing is and how it works
- Common phishing techniques and red flags
- How to identify suspicious emails, websites, and messages
- Best practices for online security
- Password security and management
- Social engineering tactics
- How to report suspected phishing attempts
- Multi-factor authentication (MFA) importance
- Safe browsing habits
- Mobile security considerations

Provide clear, concise, and educational responses. Use examples when helpful.
Keep answers professional but friendly and easy to understand for non-technical users.
If asked about creating phishing attacks or malicious content, politely decline and redirect to defensive security education.`

	fullPrompt := fmt.Sprintf("%s\n\nUser Question: %s", systemPrompt, req.Question)

	// Prepare Gemini API request
	geminiReq := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": fullPrompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.7,
			"topK":            40,
			"topP":            0.95,
			"maxOutputTokens": 1024,
		},
		"safetySettings": []map[string]interface{}{
			{
				"category":  "HARM_CATEGORY_HARASSMENT",
				"threshold": "BLOCK_MEDIUM_AND_ABOVE",
			},
			{
				"category":  "HARM_CATEGORY_HATE_SPEECH",
				"threshold": "BLOCK_MEDIUM_AND_ABOVE",
			},
			{
				"category":  "HARM_CATEGORY_SEXUALLY_EXPLICIT",
				"threshold": "BLOCK_MEDIUM_AND_ABOVE",
			},
			{
				"category":  "HARM_CATEGORY_DANGEROUS_CONTENT",
				"threshold": "BLOCK_MEDIUM_AND_ABOVE",
			},
		},
	}

	jsonData, err := json.Marshal(geminiReq)
	if err != nil {
		respondError(w, "Failed to prepare AI request", http.StatusInternalServerError)
		return
	}

	// Call Gemini API
	geminiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=%s", apiKey)
	resp, err := http.Post(geminiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		respondError(w, fmt.Sprintf("Failed to call Gemini API: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		respondError(w, "Failed to read AI response", http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		respondError(w, fmt.Sprintf("Gemini API error: %s", string(body)), http.StatusInternalServerError)
		return
	}

	// Parse Gemini response
	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &geminiResp); err != nil {
		respondError(w, "Failed to parse AI response", http.StatusInternalServerError)
		return
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		respondError(w, "No response generated from AI", http.StatusInternalServerError)
		return
	}

	answer := geminiResp.Candidates[0].Content.Parts[0].Text

	respondJSON(w, map[string]interface{}{
		"question": req.Question,
		"answer":   answer,
	})
}
