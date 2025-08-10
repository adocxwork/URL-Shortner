package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// URL represents a shortened URL entry
type URL struct {
	ID           string    `json:"id"`
	Original     string    `json:"original_url"`
	ShortCode    string    `json:"short_url"`
	CreatedAt    time.Time `json:"creation_date"`
}

// In-memory storage for shortened URLs
var urlStore = make(map[string]URL)

// generateShortCode creates an MD5-based 8-character code from the original URL
func generateShortCode(original string) string {
	hasher := md5.New()
	hasher.Write([]byte(original))
	hash := hex.EncodeToString(hasher.Sum(nil))
	return hash[:8]
}

// saveURL creates and stores a new shortened URL entry
func saveURL(original string) string {
	code := generateShortCode(original)
	urlStore[code] = URL{
		ID:        code,
		Original:  original,
		ShortCode: code,
		CreatedAt: time.Now(),
	}
	return code
}

// fetchURL retrieves a URL entry by its short code
func fetchURL(code string) (URL, error) {
	url, found := urlStore[code]
	if !found {
		return URL{}, errors.New("URL not found")
	}
	return url, nil
}

// handleHome handles requests to the root endpoint
func handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to the URL Shortener!")
}

// handleShorten processes a POST request to shorten a URL
func handleShorten(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	code := saveURL(req.URL)

	resp := struct {
		ShortURL string `json:"short_url"`
	}{
		ShortURL: code,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleRedirect handles redirection from short URL to original URL
func handleRedirect(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[len("/redirect/"):]

	url, err := fetchURL(code)
	if err != nil {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, url.Original, http.StatusFound)
}

func main() {
	http.HandleFunc("/", handleHome)
	http.HandleFunc("/shorten", handleShorten)
	http.HandleFunc("/redirect/", handleRedirect)

	fmt.Println("Server running on http://localhost:3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		fmt.Println("Server error:", err)
	}
}
