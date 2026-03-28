package metrics

import (
	"regexp"
	"strings"
)

// CountSentences counts the number of sentences in text.
func CountSentences(text string) int {
	if len(strings.TrimSpace(text)) == 0 {
		return 0
	}
	// Split by sentence-ending punctuation
	re := regexp.MustCompile(`[.!?]+\s`)
	parts := re.Split(text, -1)
	count := len(parts)
	if count == 0 {
		return 1
	}
	return count
}

// CountParagraphs counts paragraphs separated by blank lines.
func CountParagraphs(text string) int {
	if len(strings.TrimSpace(text)) == 0 {
		return 0
	}
	paragraphs := 0
	inParagraph := false
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if inParagraph {
				inParagraph = false
			}
		} else {
			if !inParagraph {
				paragraphs++
				inParagraph = true
			}
		}
	}
	return paragraphs
}

// CountQuestions counts question marks in text.
func CountQuestions(text string) int {
	return strings.Count(text, "?")
}

// HasCodeBlock detects markdown code blocks.
func HasCodeBlock(text string) bool {
	return strings.Contains(text, "```") || strings.Contains(text, "~~~")
}

// HasCodeRef detects references to code (file:line, function names, etc).
func HasCodeRef(text string) bool {
	patterns := []string{
		`\w+\.\w+:\d+`,          // file.go:42
		`func\s+\w+`,            // func name
		`\w+\(\)`,               // function()
		`package\s+\w+`,         // package name
		`import\s+`,             // import
		`class\s+\w+`,           // class name
		`def\s+\w+`,             // python def
	}
	for _, p := range patterns {
		if matched, _ := regexp.MatchString(p, text); matched {
			return true
		}
	}
	return false
}

// HasURL detects URLs in text.
func HasURL(text string) bool {
	re := regexp.MustCompile(`https?://\S+`)
	return re.MatchString(text)
}

// HasFilePath detects file paths.
func HasFilePath(text string) bool {
	patterns := []string{
		`[/~]\S+\.\w+`,           // /path/to/file.ext or ~/file.ext
		`\w+/\w+/\w+`,            // dir/subdir/file
		`\w+\.(go|py|js|ts|rs|java|yaml|yml|json|md|txt|sh)`, // file.ext
	}
	for _, p := range patterns {
		if matched, _ := regexp.MatchString(p, text); matched {
			return true
		}
	}
	return false
}

// roleIndicatorPhrases are common phrases that assign a persona or expert role.
var roleIndicatorPhrases = []string{
	"you are",
	"you're",
	"act as",
	"acting as",
	"pretend to be",
	"assume the role",
	"take the role",
	"as an expert",
	"as a senior",
	"as a principal",
	"as an architect",
	"as a security",
	"as a staff",
}

// HasRoleIndicator returns true when the prompt assigns a role or persona to the model.
// These prompts typically carry more context and require higher capability models.
func HasRoleIndicator(text string) bool {
	lower := strings.ToLower(text)
	for _, phrase := range roleIndicatorPhrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

// multiStepPhrases are phrases that signal a multi-step or complex workflow request.
var multiStepPhrases = []string{
	"step by step",
	"step-by-step",
	"first,",
	"then,",
	"finally,",
	"1.",
	"2.",
	"3.",
	"phase 1",
	"phase 2",
	"phase one",
	"phase two",
	"additionally,",
	"furthermore,",
	"after that",
	"following that",
}

// HasMultiStepIndicator returns true when the prompt describes a multi-step process
// or lists multiple sequential tasks.
func HasMultiStepIndicator(text string) bool {
	lower := strings.ToLower(text)
	count := 0
	for _, phrase := range multiStepPhrases {
		if strings.Contains(lower, phrase) {
			count++
			if count >= 2 {
				return true
			}
		}
	}
	return false
}

// constraintPhrases signal that the prompt imposes explicit constraints or requirements.
var constraintPhrases = []string{
	"must",
	"should not",
	"must not",
	"do not",
	"don't",
	"without",
	"only use",
	"only if",
	"ensure that",
	"make sure",
	"requirement",
	"constraint",
	"restriction",
	"forbidden",
	"allowed",
	"no more than",
	"at least",
	"must be",
	"should be",
}

// HasConstraints returns true when the prompt contains multiple constraint phrases,
// indicating a more complex and specific request.
func HasConstraints(text string) bool {
	lower := strings.ToLower(text)
	count := 0
	for _, phrase := range constraintPhrases {
		if strings.Contains(lower, phrase) {
			count++
			if count >= 2 {
				return true
			}
		}
	}
	return false
}
