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
