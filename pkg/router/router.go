package router

import (
	"fmt"
	"strings"

	"github.com/mikeshogin/promptlint/pkg/analyzer"
	"github.com/mikeshogin/promptlint/pkg/config"
)

// RouteResult contains the routing decision and its rationale.
type RouteResult struct {
	Model      string  `json:"model"`
	Tier       string  `json:"tier"`
	Complexity string  `json:"complexity"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
	Score      float64 `json:"score"`
}

// Router performs metric-based model selection.
type Router struct {
	cfg *config.Config
}

// New creates a Router with the provided config. If cfg is nil, the default config is used.
func New(cfg *config.Config) *Router {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	return &Router{cfg: cfg}
}

// NewDefault creates a Router using LoadOrDefault.
func NewDefault() *Router {
	return &Router{cfg: config.LoadOrDefault()}
}

// Route analyzes the prompt and returns a routing decision.
func (r *Router) Route(prompt string) RouteResult {
	a := analyzer.AnalyzeWithConfig(prompt, r.cfg)
	return r.routeFromAnalysis(a)
}

// routeFromAnalysis computes a weighted score and maps it to a model tier.
func (r *Router) routeFromAnalysis(a analyzer.Result) RouteResult {
	score, signals := weightedScore(a)

	// Map score ranges to model tiers.
	// Score is in [0, 1] range where higher = more complex.
	var model, tier, complexity string
	switch {
	case score >= 0.60:
		complexity = "high"
		tier = r.cfg.RouteByComplexity("high")
		model = tier
	case score >= 0.30:
		complexity = "medium"
		tier = r.cfg.RouteByComplexity("medium")
		model = tier
	default:
		complexity = "low"
		tier = r.cfg.RouteByComplexity("low")
		model = tier
	}

	confidence := computeConfidence(score, complexity)
	reasoning := buildReasoning(a, signals, score)

	return RouteResult{
		Model:      model,
		Tier:       tier,
		Complexity: complexity,
		Confidence: confidence,
		Score:      round2(score),
		Reasoning:  reasoning,
	}
}

// signal is a named contribution to the routing score.
type signal struct {
	name   string
	weight float64
}

// weightedScore computes a normalized complexity score from [0, 1] using
// multiple metrics. Returns the score and the triggered signals for reasoning.
func weightedScore(a analyzer.Result) (float64, []signal) {
	var triggered []signal
	total := 0.0

	// --- Length signals (max contribution: 0.20) ---
	switch {
	case a.Words > 300:
		triggered = append(triggered, signal{"very long prompt (>300 words)", 0.20})
		total += 0.20
	case a.Words > 100:
		triggered = append(triggered, signal{"long prompt (>100 words)", 0.14})
		total += 0.14
	case a.Words > 40:
		triggered = append(triggered, signal{"medium prompt (>40 words)", 0.08})
		total += 0.08
	}

	// --- Structural signals (max contribution: 0.20) ---
	if a.HasCodeBlock {
		triggered = append(triggered, signal{"contains code block", 0.12})
		total += 0.12
	}
	if a.HasCodeRef {
		triggered = append(triggered, signal{"references code symbols", 0.08})
		total += 0.08
	}

	// --- Domain signals (max contribution: 0.40) ---
	archScore := a.Domain["architecture"]
	if archScore >= 0.7 {
		triggered = append(triggered, signal{"strong architecture domain", 0.25})
		total += 0.25
	} else if archScore >= 0.3 {
		triggered = append(triggered, signal{"architecture domain detected", 0.15})
		total += 0.15
	}

	activeDomains := 0
	for _, v := range a.Domain {
		if v > 0.3 {
			activeDomains++
		}
	}
	if activeDomains >= 3 {
		triggered = append(triggered, signal{"multi-domain task (3+ domains)", 0.15})
		total += 0.15
	} else if activeDomains == 2 {
		triggered = append(triggered, signal{"cross-domain task (2 domains)", 0.08})
		total += 0.08
	}

	// --- Action signals (max contribution: 0.20) ---
	switch a.Action {
	case "refactor":
		triggered = append(triggered, signal{"action: refactor", 0.20})
		total += 0.20
	case "create":
		// design/architect/plan are mapped to "create" in domain.go
		if archScore > 0 {
			triggered = append(triggered, signal{"action: design/architect", 0.15})
			total += 0.15
		} else {
			triggered = append(triggered, signal{"action: create", 0.10})
			total += 0.10
		}
	case "review":
		triggered = append(triggered, signal{"action: review/analyze", 0.08})
		total += 0.08
	case "explain":
		triggered = append(triggered, signal{"action: explain", 0.05})
		total += 0.05
	case "fix":
		triggered = append(triggered, signal{"action: fix", 0.02})
		total += 0.02
	}

	// --- Question density (max contribution: 0.10) ---
	if a.Questions >= 3 {
		triggered = append(triggered, signal{"many questions (3+)", 0.10})
		total += 0.10
	} else if a.Questions == 2 {
		triggered = append(triggered, signal{"multiple questions (2)", 0.06})
		total += 0.06
	}

	// Cap at 1.0
	if total > 1.0 {
		total = 1.0
	}

	return total, triggered
}

// computeConfidence returns a confidence value based on how decisively the score
// falls into a tier (further from a boundary -> higher confidence).
func computeConfidence(score float64, complexity string) float64 {
	// Boundary points are at 0.30 and 0.60
	var distFromBoundary float64
	switch complexity {
	case "low":
		// [0, 0.30) - distance from 0.30 boundary, normalized over 0.30 range
		distFromBoundary = (0.30 - score) / 0.30
	case "medium":
		// [0.30, 0.60) - distance from nearest boundary, normalized over 0.30 range
		dist1 := score - 0.30
		dist2 := 0.60 - score
		if dist1 < dist2 {
			distFromBoundary = dist1 / 0.30
		} else {
			distFromBoundary = dist2 / 0.30
		}
	case "high":
		// [0.60, 1.0] - distance from 0.60 boundary, normalized over 0.40 range
		distFromBoundary = (score - 0.60) / 0.40
	}

	// Base confidence 0.60, boosted by distance from boundary up to 0.95
	confidence := 0.60 + distFromBoundary*0.35
	if confidence > 0.95 {
		confidence = 0.95
	}
	if confidence < 0.60 {
		confidence = 0.60
	}
	return round2(confidence)
}

// buildReasoning produces a human-readable explanation of the routing decision.
func buildReasoning(a analyzer.Result, signals []signal, score float64) string {
	if len(signals) == 0 {
		return fmt.Sprintf("short, simple task (score=%.2f)", score)
	}

	names := make([]string, 0, len(signals))
	for _, s := range signals {
		names = append(names, s.name)
	}

	// Find top domain for display (skip if already covered by signal names)
	topDomain := ""
	topScore := 0.0
	for d, v := range a.Domain {
		if v > topScore {
			topScore = v
			topDomain = d
		}
	}

	var sb strings.Builder
	sb.WriteString(strings.Join(names, ", "))

	// Only append primary domain if not already mentioned in signal names
	domainMentioned := false
	for _, n := range names {
		if strings.Contains(n, "domain") || strings.Contains(n, topDomain) {
			domainMentioned = true
			break
		}
	}
	if !domainMentioned && topDomain != "" && topDomain != "general" {
		sb.WriteString(fmt.Sprintf("; primary domain: %s", topDomain))
	}

	return sb.String()
}

// round2 rounds a float64 to 2 decimal places.
func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
