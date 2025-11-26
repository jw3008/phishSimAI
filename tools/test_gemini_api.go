package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GeminiAPIRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
}

type GeminiAPIResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func main() {
	apiKey := "AIzaSyBxLTGTPAwKG0fPjdojuEsDJ77y_fUqxCI"

	// Create test request
	req := GeminiAPIRequest{
		Contents: []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			{
				Parts: []struct {
					Text string `json:"text"`
				}{
					{Text: "Say 'Hello! The API is working correctly.' if you can read this."},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(req)
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1/models/gemini-2.5-flash:generateContent?key=%s", apiKey)

	fmt.Println("Testing Gemini API connection...")
	fmt.Println("URL:", "https://generativelanguage.googleapis.com/v1/models/gemini-2.5-flash:generateContent")
	fmt.Println()

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("✗ Failed to connect to Gemini API:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Println("✗ API request failed with status:", resp.StatusCode)
		fmt.Println("Response:", string(body))
		return
	}

	var geminiResp GeminiAPIResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		fmt.Println("✗ Failed to parse response:", err)
		return
	}

	if geminiResp.Error != nil {
		fmt.Println("✗ API Error:", geminiResp.Error.Message)
		return
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		fmt.Println("✓ API Connection Successful!")
		fmt.Println()
		fmt.Println("Response from Gemini:")
		fmt.Println(geminiResp.Candidates[0].Content.Parts[0].Text)
		fmt.Println()
		fmt.Println("✓ Your Gemini API key is working correctly!")
		fmt.Println("✓ The phishing simulation AI features are ready to use!")
	} else {
		fmt.Println("✗ Unexpected response format")
	}
}
