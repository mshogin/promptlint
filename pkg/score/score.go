package score

import (
	"strings"

	"github.com/mikeshogin/promptlint/pkg/metrics"
)

// Breakdown holds the 4 category sub-scores (0-25 each).
type Breakdown struct {
	Readability int `json:"readability"` // based on NLP readability_score
	Clarity     int `json:"clarity"`     // instruction verb, length, non-empty
	Technical   int `json:"technical"`   // technical_density (some good, too much bad)
	Vocabulary  int `json:"vocabulary"`  // vocabulary_richness
}

// PromptScore is the overall quality score for a prompt.
type PromptScore struct {
	Total     int       `json:"total"`     // 0-100
	Breakdown Breakdown `json:"breakdown"` // 4 categories, 0-25 each
}

// imperativeVerbs mirrors the starters in metrics/nlp.go (subset used for clarity check).
var imperativeVerbs = map[string]bool{
	"create": true, "fix": true, "add": true, "remove": true, "update": true,
	"delete": true, "implement": true, "write": true, "build": true, "refactor": true,
	"review": true, "check": true, "analyze": true, "design": true, "deploy": true,
	"migrate": true, "optimize": true, "explain": true, "describe": true, "generate": true,
	"run": true, "test": true, "install": true, "configure": true, "setup": true,
	"make": true, "list": true, "show": true, "find": true, "get": true, "set": true,
	"use": true, "edit": true, "change": true, "move": true, "rename": true,
	"replace": true, "ensure": true, "handle": true, "return": true, "define": true,
	"print": true, "parse": true, "load": true, "save": true, "read": true,
	"send": true, "fetch": true, "connect": true, "open": true, "close": true,
	"start": true, "stop": true, "restart": true, "enable": true, "disable": true,
	"convert": true, "validate": true, "verify": true, "initialize": true,
	"extend": true, "override": true, "append": true, "insert": true, "extract": true,
	"split": true, "merge": true, "sort": true, "filter": true, "map": true,
	"reduce": true, "transform": true,
}

// ComputePromptScore computes a 0-100 quality score from NLP metrics and raw prompt.
//
// The score has four equal-weight categories (0-25 each):
//   - Readability: maps the Flesch-Kincaid readability_score (0-100) -> 0-25
//   - Clarity: instruction verb present, non-empty, reasonable length (10-5000 chars)
//   - Technical: technical_density in a "sweet spot" (some technical = good, too much = bad)
//   - Vocabulary: vocabulary_richness (0-1) mapped to 0-25
func ComputePromptScore(prompt string, nlp metrics.NLPMetrics) PromptScore {
	b := Breakdown{
		Readability: readabilityScore(nlp.ReadabilityScore),
		Clarity:     clarityScore(prompt),
		Technical:   technicalScore(nlp.TechnicalDensity),
		Vocabulary:  vocabularyScore(nlp.VocabularyRichness),
	}
	return PromptScore{
		Total:     b.Readability + b.Clarity + b.Technical + b.Vocabulary,
		Breakdown: b,
	}
}

// readabilityScore maps Flesch-Kincaid score (0-100) to 0-25 linearly.
func readabilityScore(fk float64) int {
	v := int(fk * 25.0 / 100.0)
	return clamp(v, 0, 25)
}

// clarityScore awards up to 25 points based on:
//   - non-empty prompt:               5 pts
//   - length between 10-5000 chars:   5 pts
//   - starts with an imperative verb: 10 pts
//   - no more than one question mark (action-oriented, not open-ended): 5 pts
func clarityScore(prompt string) int {
	trimmed := strings.TrimSpace(prompt)
	if trimmed == "" {
		return 0
	}

	score := 0

	// Non-empty
	score += 5

	// Reasonable length
	l := len(trimmed)
	if l >= 10 && l <= 5000 {
		score += 5
	}

	// Has imperative verb (action word at start of any sentence)
	words := strings.Fields(trimmed)
	if len(words) > 0 {
		first := strings.ToLower(strings.Trim(words[0], ".,!?;:\"'()[]{}"))
		if imperativeVerbs[first] {
			score += 10
		}
	}

	// Clear action (not purely question-based): reward prompts without trailing "?" only
	if !strings.HasSuffix(strings.TrimRight(trimmed, " \t"), "?") {
		score += 5
	}

	return clamp(score, 0, 25)
}

// technicalScore maps technical_density (0-1) to 0-25 using a bell-curve shape:
//   - 0.0-0.05  (no technical): low score   (0-5)
//   - 0.05-0.30 (ideal range):  peaks at 25
//   - 0.30-1.0  (too dense):    drops back down
func technicalScore(density float64) int {
	var v float64
	switch {
	case density <= 0.0:
		v = 0
	case density <= 0.05:
		// 0 -> 10 linearly
		v = density / 0.05 * 10
	case density <= 0.30:
		// 10 -> 25 linearly
		v = 10 + (density-0.05)/(0.30-0.05)*15
	case density <= 1.0:
		// 25 -> 5 linearly as density goes high
		v = 25 - (density-0.30)/(1.0-0.30)*20
		if v < 5 {
			v = 5
		}
	default:
		v = 5
	}
	return clamp(int(v), 0, 25)
}

// vocabularyScore maps vocabulary_richness (0-1) to 0-25 linearly.
func vocabularyScore(richness float64) int {
	v := int(richness * 25.0)
	return clamp(v, 0, 25)
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
