package validator

import (
	"strings"
	"unicode"
)

// Severity indicates how critical a validation violation is.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// ValidationResult represents a single rule violation.
type ValidationResult struct {
	Rule        string   `json:"rule"`
	Description string   `json:"description"`
	Severity    Severity `json:"severity"`
	Message     string   `json:"message"`
}

// Rule defines a single validation rule.
type Rule struct {
	Name        string
	Description string
	Severity    Severity
	Check       func(text string) (ok bool, message string)
}

// Validator holds a set of rules and runs them against prompts.
type Validator struct {
	rules []Rule
}

// New returns a Validator with the built-in rule set.
func New() *Validator {
	return &Validator{rules: builtinRules()}
}

// Validate runs all rules against text and returns any violations.
// An empty slice means the prompt passed all checks.
func (v *Validator) Validate(text string) []ValidationResult {
	var results []ValidationResult
	for _, rule := range v.rules {
		ok, msg := rule.Check(text)
		if !ok {
			results = append(results, ValidationResult{
				Rule:        rule.Name,
				Description: rule.Description,
				Severity:    rule.Severity,
				Message:     msg,
			})
		}
	}
	return results
}

// builtinRules returns the standard set of validation rules.
func builtinRules() []Rule {
	return []Rule{
		{
			Name:        "not_empty",
			Description: "Prompt must not be empty",
			Severity:    SeverityError,
			Check: func(text string) (bool, string) {
				if strings.TrimSpace(text) == "" {
					return false, "prompt is empty"
				}
				return true, ""
			},
		},
		{
			Name:        "min_length",
			Description: "Prompt should be at least 10 characters",
			Severity:    SeverityWarning,
			Check: func(text string) (bool, string) {
				trimmed := strings.TrimSpace(text)
				if len(trimmed) > 0 && len(trimmed) < 10 {
					return false, "prompt is too short (less than 10 characters)"
				}
				return true, ""
			},
		},
		{
			Name:        "max_length",
			Description: "Prompt should be under 100000 characters",
			Severity:    SeverityWarning,
			Check: func(text string) (bool, string) {
				if len(text) > 100000 {
					return false, "prompt exceeds 100000 characters"
				}
				return true, ""
			},
		},
		{
			Name:        "has_instruction",
			Description: "Prompt should contain an action verb or instruction",
			Severity:    SeverityWarning,
			Check:       checkHasInstruction,
		},
		{
			Name:        "no_injection",
			Description: "Prompt should not contain common prompt injection patterns",
			Severity:    SeverityError,
			Check:       checkNoInjection,
		},
	}
}

// actionVerbs is a list of common English and instruction verbs.
var actionVerbs = []string{
	"write", "create", "make", "build", "generate", "produce",
	"explain", "describe", "summarize", "analyze", "analyse",
	"fix", "debug", "correct", "update", "change", "modify",
	"list", "show", "find", "search", "get", "fetch",
	"convert", "translate", "transform", "format",
	"review", "check", "test", "validate",
	"help", "tell", "give", "provide", "suggest",
	"compare", "evaluate", "assess", "calculate", "compute",
	"implement", "add", "remove", "delete", "refactor",
}

// checkHasInstruction returns false when the prompt contains no recognisable
// action verb. Short prompts under 10 chars are skipped (covered by min_length).
func checkHasInstruction(text string) (bool, string) {
	trimmed := strings.TrimSpace(text)
	if len(trimmed) < 10 {
		return true, "" // let min_length handle it
	}
	lower := strings.ToLower(trimmed)
	// Tokenise into words, stripping leading/trailing punctuation.
	words := tokenise(lower)
	wordSet := make(map[string]struct{}, len(words))
	for _, w := range words {
		wordSet[w] = struct{}{}
	}
	for _, verb := range actionVerbs {
		if _, ok := wordSet[verb]; ok {
			return true, ""
		}
	}
	return false, "prompt does not appear to contain a clear instruction or action verb"
}

// injectionPatterns are substrings that signal common prompt injection attacks.
var injectionPatterns = []string{
	"ignore previous",
	"ignore all previous",
	"disregard previous",
	"forget previous",
	"system prompt",
	"you are now",
	"new persona",
	"pretend you are",
	"act as if",
	"override instructions",
	"disregard instructions",
	"ignore instructions",
}

func checkNoInjection(text string) (bool, string) {
	lower := strings.ToLower(text)
	for _, pattern := range injectionPatterns {
		if strings.Contains(lower, pattern) {
			return false, "prompt contains a possible injection pattern: \"" + pattern + "\""
		}
	}
	return true, ""
}

// tokenise splits text into lowercase words, removing non-letter/digit runes
// from word boundaries so "write," becomes "write".
func tokenise(text string) []string {
	var words []string
	var current strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
		}
	}
	if current.Len() > 0 {
		words = append(words, current.String())
	}
	return words
}
