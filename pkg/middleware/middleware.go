package middleware

import (
	"github.com/mikeshogin/promptlint/pkg/analyzer"
)

// Router decides which model to use based on prompt metrics.
type Router struct {
	// DefaultModel is used when analysis is inconclusive.
	DefaultModel string
	// Thresholds for routing decisions.
	HighComplexityModel   string
	MediumComplexityModel string
	LowComplexityModel    string
}

// NewRouter creates a router with default settings.
func NewRouter() *Router {
	return &Router{
		DefaultModel:          "haiku",
		HighComplexityModel:   "opus",
		MediumComplexityModel: "sonnet",
		LowComplexityModel:    "haiku",
	}
}

// RouteResult contains the routing decision and analysis.
type RouteResult struct {
	Model    string          `json:"model"`
	Analysis analyzer.Result `json:"analysis"`
}

// Route analyzes a prompt and returns routing decision.
// Uses numeric ComplexityScore for thresholding instead of string labels,
// giving finer control over model selection near tier boundaries.
func (r *Router) Route(prompt string) RouteResult {
	analysis := analyzer.Analyze(prompt)

	model := r.DefaultModel
	switch {
	case analysis.ComplexityScore >= 50:
		model = r.HighComplexityModel
	case analysis.ComplexityScore >= 25:
		model = r.MediumComplexityModel
	default:
		model = r.LowComplexityModel
	}

	return RouteResult{
		Model:    model,
		Analysis: analysis,
	}
}

// ShouldScore returns true if the prompt needs detailed scoring
// (e.g., when basic metrics are ambiguous).
func (r *Router) ShouldScore(prompt string) bool {
	analysis := analyzer.Analyze(prompt)

	// Ambiguous: medium complexity with multiple active domains
	if analysis.Complexity == "medium" {
		activeDomains := 0
		for _, v := range analysis.Domain {
			if v > 0.3 {
				activeDomains++
			}
		}
		return activeDomains > 1
	}

	return false
}
