package analyzer

import (
	"strings"
	"unicode"

	"github.com/mikeshogin/promptlint/pkg/config"
	"github.com/mikeshogin/promptlint/pkg/metrics"
	"github.com/mikeshogin/promptlint/pkg/score"
)

// Result contains all extracted metrics from a prompt.
type Result struct {
	// Basic metrics
	Length     int `json:"length"`
	Words      int `json:"words"`
	Sentences  int `json:"sentences"`
	Paragraphs int `json:"paragraphs"`

	// Content detection
	HasCodeBlock bool `json:"has_code_block"`
	HasCodeRef   bool `json:"has_code_ref"`
	HasURL       bool `json:"has_url"`
	HasFilePath  bool `json:"has_file_path"`
	Questions    int  `json:"questions"`

	// Classification
	Action          string             `json:"action"`
	Domain          map[string]float64 `json:"domain"`
	Complexity      string             `json:"complexity"`
	ComplexityScore int                `json:"complexity_score"` // numeric 0-100

	// NLP metrics
	NLPMetrics metrics.NLPMetrics `json:"nlp_metrics"`

	// Routing suggestion
	SuggestedModel string `json:"suggested_model"`

	// Quality score
	PromptScore score.PromptScore `json:"prompt_score"`
}

// Analyze extracts metrics from a prompt string using the default config.
func Analyze(prompt string) Result {
	return AnalyzeWithConfig(prompt, config.LoadOrDefault())
}

// AnalyzeWithConfig extracts metrics from a prompt string using the provided config.
func AnalyzeWithConfig(prompt string, cfg *config.Config) Result {
	r := Result{
		Domain: make(map[string]float64),
	}

	// Basic metrics
	r.Length = len(prompt)
	r.Words = countWords(prompt)
	r.Sentences = metrics.CountSentences(prompt)
	r.Paragraphs = metrics.CountParagraphs(prompt)

	// Content detection
	r.HasCodeBlock = metrics.HasCodeBlock(prompt)
	r.HasCodeRef = metrics.HasCodeRef(prompt)
	r.HasURL = metrics.HasURL(prompt)
	r.HasFilePath = metrics.HasFilePath(prompt)
	r.Questions = metrics.CountQuestions(prompt)

	// Classification
	r.Action = metrics.DetectAction(prompt)
	r.Domain = metrics.ClassifyDomain(prompt)
	r.ComplexityScore = complexityScore(r, prompt)
	r.Complexity = complexityLabel(r.ComplexityScore)

	// NLP metrics
	r.NLPMetrics = metrics.AnalyzeNLP(prompt)

	// Quality score
	r.PromptScore = score.ComputePromptScore(prompt, r.NLPMetrics)

	// Routing
	r.SuggestedModel = suggestModel(r, cfg)

	return r
}

func countWords(s string) int {
	return len(strings.Fields(s))
}

// complexityScore computes a numeric complexity score in [0, 100].
// Higher values mean more complex prompts requiring more capable models.
func complexityScore(r Result, prompt string) int {
	total := 0

	// --- Length signals (max 20 pts) ---
	switch {
	case r.Words > 300:
		total += 20
	case r.Words > 100:
		total += 14
	case r.Words > 40:
		total += 8
	case r.Words > 10:
		total += 3
	}

	// --- Structural signals (max 14 pts) ---
	if r.HasCodeBlock {
		total += 8
	}
	if r.HasCodeRef {
		total += 6
	}

	// --- Domain signals (max 25 pts) ---
	archScore := r.Domain["architecture"]
	switch {
	case archScore >= 0.7:
		total += 25
	case archScore >= 0.3:
		total += 15
	case archScore > 0:
		total += 7
	}

	activeDomains := 0
	for _, v := range r.Domain {
		if v > 0.3 {
			activeDomains++
		}
	}
	if activeDomains >= 3 {
		total += 10
	} else if activeDomains == 2 {
		total += 5
	}

	// --- Action signals (max 15 pts) ---
	switch r.Action {
	case "refactor":
		total += 15
	case "create":
		if archScore >= 0.7 {
			// "design/architect" + strong architecture domain = highly complex
			total += 15
		} else if archScore > 0 {
			total += 10
		} else {
			total += 7
		}
	case "review":
		total += 6
	case "explain":
		total += 4
	case "fix":
		total += 2
	}

	// --- Question density (max 8 pts) ---
	if r.Questions >= 3 {
		total += 8
	} else if r.Questions == 2 {
		total += 5
	} else if r.Questions == 1 {
		total += 2
	}

	// --- Technical term density boost (max 8 pts) ---
	// High technical density in architecture domain signals specialized expertise needed.
	if archScore >= 0.7 && activeDomains >= 2 {
		total += 5
	} else if archScore >= 0.5 {
		total += 2
	}

	// --- Role / persona indicators (max 10 pts) ---
	// Prompts that assign a role ("you are an expert ...", "act as ...") tend to be
	// multi-step or context-heavy.
	if metrics.HasRoleIndicator(prompt) {
		total += 10
	}

	// --- Multi-step / constraint indicators (max 8 pts) ---
	if metrics.HasMultiStepIndicator(prompt) {
		total += 5
	}
	if metrics.HasConstraints(prompt) {
		total += 3
	}

	if total > 100 {
		total = 100
	}
	return total
}

// complexityLabel maps a numeric score to a string label.
func complexityLabel(score int) string {
	switch {
	case score >= 50:
		return "high"
	case score >= 25:
		return "medium"
	default:
		return "low"
	}
}

// classifyComplexity is kept for backwards compatibility but delegates to the
// new numeric scoring approach. Callers that already have a Result should use
// complexityLabel(complexityScore(r, prompt)) directly.
func classifyComplexity(r Result) string {
	return r.Complexity
}

func suggestModel(r Result, cfg *config.Config) string {
	if cfg != nil {
		return cfg.RouteByComplexity(r.Complexity)
	}
	switch r.Complexity {
	case "high":
		return "opus"
	case "medium":
		return "sonnet"
	default:
		return "haiku"
	}
}

// isLetter checks if a rune is a letter (unused but kept for future use).
var _ = unicode.IsLetter
