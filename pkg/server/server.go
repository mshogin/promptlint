package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/mikeshogin/promptlint/pkg/analyzer"
	"github.com/mikeshogin/promptlint/pkg/router"
	"github.com/mikeshogin/promptlint/pkg/validator"
)

// New returns an http.Handler with all routes registered.
func New() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/analyze", handleAnalyze)
	mux.HandleFunc("/validate", handleValidate)
	mux.HandleFunc("/route", handleRoute)
	mux.HandleFunc("/health", handleHealth)
	return mux
}

// handleAnalyze accepts POST /analyze with a JSON body {"text": "..."} or plain text.
// It returns the analysis result as JSON.
func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	// Try to parse as JSON {"text": "..."}
	var req struct {
		Text string `json:"text"`
	}
	text := string(body)
	if json.Unmarshal(body, &req) == nil && req.Text != "" {
		text = req.Text
	}

	result := analyzer.Analyze(text)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleValidate accepts POST /validate with a JSON body {"text": "..."} or plain text.
// It returns a JSON array of ValidationResult objects (empty array when no violations found).
func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	// Try to parse as JSON {"text": "..."}
	var req struct {
		Text string `json:"text"`
	}
	text := string(body)
	if json.Unmarshal(body, &req) == nil && req.Text != "" {
		text = req.Text
	}

	v := validator.New()
	results := v.Validate(text)
	if results == nil {
		results = []validator.ValidationResult{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// handleRoute accepts POST /route with a JSON body {"text": "..."} or plain text.
// It returns a routing decision as JSON.
func handleRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	// Try to parse as JSON {"text": "..."}
	var req struct {
		Text string `json:"text"`
	}
	text := strings.TrimSpace(string(body))
	if json.Unmarshal(body, &req) == nil && req.Text != "" {
		text = req.Text
	}

	rt := router.NewDefault()
	result := rt.Route(text)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleHealth returns a simple health check response.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
