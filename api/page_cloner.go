package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/clariphish/clariphish/db"
	"golang.org/x/net/html"
)

// CloneLandingPage clones a landing page from a URL
func CloneLandingPage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL  string `json:"url"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate URL
	parsedURL, err := url.Parse(req.URL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		respondError(w, "Invalid URL. Must be http or https", http.StatusBadRequest)
		return
	}

	// Fetch the page
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Get(req.URL)
	if err != nil {
		respondError(w, fmt.Sprintf("Failed to fetch URL: %v", err), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respondError(w, fmt.Sprintf("Failed to fetch URL: HTTP %d", resp.StatusCode), http.StatusBadRequest)
		return
	}

	// Read the HTML content
	htmlContent, err := io.ReadAll(resp.Body)
	if err != nil {
		respondError(w, fmt.Sprintf("Failed to read page content: %v", err), http.StatusInternalServerError)
		return
	}

	// Parse and process HTML to make URLs absolute
	processedHTML := processHTML(string(htmlContent), req.URL)

	// Add tracking pixel to the HTML
	trackingPixel := `<img src="{{.TrackingURL}}" style="display:none" />`
	processedHTML = strings.Replace(processedHTML, "</body>", trackingPixel+"\n</body>", 1)

	// Generate a name if not provided
	pageName := req.Name
	if pageName == "" {
		pageName = fmt.Sprintf("Cloned from %s", parsedURL.Host)
	}

	// Get user from session
	session, _ := store.Get(r, "clariphish-session")
	userID := session.Values["user_id"]

	// Save to database (with default values for other fields)
	result, err := db.DB.Exec(`
		INSERT INTO pages (name, html, capture_credentials, capture_passwords, redirect_url, user_id, created_at)
		VALUES (?, ?, 1, 1, '', ?, ?)
	`, pageName, processedHTML, userID, time.Now())

	if err != nil {
		respondError(w, fmt.Sprintf("Failed to save page: %v", err), http.StatusInternalServerError)
		return
	}

	pageID, _ := result.LastInsertId()

	respondJSON(w, map[string]interface{}{
		"id":      pageID,
		"name":    pageName,
		"message": "Landing page cloned successfully",
		"user_id": userID,
	})
}

// processHTML converts relative URLs to absolute URLs
func processHTML(htmlContent, baseURL string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}

	parsedBase, err := url.Parse(baseURL)
	if err != nil {
		return htmlContent
	}

	var process func(*html.Node)
	process = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Process different attributes based on element type
			switch n.Data {
			case "a", "link":
				updateAttribute(n, "href", parsedBase)
			case "img", "script":
				updateAttribute(n, "src", parsedBase)
			case "form":
				updateAttribute(n, "action", parsedBase)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			process(c)
		}
	}

	process(doc)

	// Convert back to HTML string
	var buf strings.Builder
	html.Render(&buf, doc)
	return buf.String()
}

// updateAttribute updates an attribute to use absolute URL
func updateAttribute(n *html.Node, attrName string, baseURL *url.URL) {
	for i, attr := range n.Attr {
		if attr.Key == attrName && attr.Val != "" {
			// Skip data URLs, javascript, and absolute URLs
			if strings.HasPrefix(attr.Val, "data:") ||
				strings.HasPrefix(attr.Val, "javascript:") ||
				strings.HasPrefix(attr.Val, "http://") ||
				strings.HasPrefix(attr.Val, "https://") {
				continue
			}

			// Convert relative URL to absolute
			relURL, err := url.Parse(attr.Val)
			if err == nil {
				absURL := baseURL.ResolveReference(relURL)
				n.Attr[i].Val = absURL.String()
			}
		}
	}
}

// ExtractPagePreview extracts a preview/description from HTML
func ExtractPagePreview(htmlContent string) string {
	// Remove HTML tags for preview
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(htmlContent, " ")

	// Remove extra whitespace
	re = regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	// Limit to 200 characters
	if len(text) > 200 {
		text = text[:200] + "..."
	}

	return text
}
