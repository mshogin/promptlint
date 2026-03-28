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
	mux.HandleFunc("/analyze/batch", handleAnalyzeBatch)
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

// BatchRequest is the input for POST /analyze/batch.
type BatchRequest struct {
	Prompts []string `json:"prompts"`
}

// BatchItemResult is the per-prompt result in a batch response.
type BatchItemResult struct {
	Index  int                  `json:"index"`
	Text   string               `json:"text"`
	Route  router.RouteResult   `json:"route"`
	Result analyzer.Result      `json:"result"`
}

// BatchSummary contains aggregated statistics for a batch.
type BatchSummary struct {
	Total             int                `json:"total"`
	ModelDistribution map[string]int     `json:"model_distribution"`
	ComplexityBreakdown map[string]int   `json:"complexity_breakdown"`
	AvgScore          float64            `json:"avg_score"`
}

// BatchResponse is the response for POST /analyze/batch.
type BatchResponse struct {
	Items   []BatchItemResult `json:"items"`
	Summary BatchSummary      `json:"summary"`
}

// handleAnalyzeBatch accepts POST /analyze/batch with a JSON body {"prompts": ["...", "..."]}.
// It returns per-prompt routing decisions and aggregated summary statistics.
func handleAnalyzeBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	var req BatchRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid JSON: expected {\"prompts\": [...]}", http.StatusBadRequest)
		return
	}
	if len(req.Prompts) == 0 {
		http.Error(w, "prompts array is empty", http.StatusBadRequest)
		return
	}

	rt := router.NewDefault()

	items := make([]BatchItemResult, 0, len(req.Prompts))
	modelDist := make(map[string]int)
	complexityDist := make(map[string]int)
	totalScore := 0.0

	for i, text := range req.Prompts {
		text = strings.TrimSpace(text)
		result := analyzer.Analyze(text)
		route := rt.Route(text)

		items = append(items, BatchItemResult{
			Index:  i,
			Text:   text,
			Route:  route,
			Result: result,
		})

		modelDist[route.Model]++
		complexityDist[route.Complexity]++
		totalScore += route.Score
	}

	avgScore := 0.0
	if len(req.Prompts) > 0 {
		avgScore = totalScore / float64(len(req.Prompts))
	}

	resp := BatchResponse{
		Items: items,
		Summary: BatchSummary{
			Total:               len(req.Prompts),
			ModelDistribution:   modelDist,
			ComplexityBreakdown: complexityDist,
			AvgScore:            round2(avgScore),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// round2 rounds a float64 to 2 decimal places.
func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

// handleHealth returns a simple health check response.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}
