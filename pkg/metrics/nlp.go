package metrics

import (
	"regexp"
	"strings"
	"unicode"
)

// NLPMetrics holds NLP-style heuristic metrics for a prompt.
type NLPMetrics struct {
	ReadabilityScore      float64 `json:"readability_score"`
	VocabularyRichness    float64 `json:"vocabulary_richness"`
	AvgSentenceLength     float64 `json:"avg_sentence_length"`
	QuestionDensity       float64 `json:"question_density"`
	ImperativeRatio       float64 `json:"imperative_ratio"`
	TechnicalDensity      float64 `json:"technical_density"`
}

// imperativeStarters is the list of common imperative verb starters.
var imperativeStarters = map[string]bool{
	"create":     true,
	"fix":        true,
	"add":        true,
	"remove":     true,
	"update":     true,
	"delete":     true,
	"implement":  true,
	"write":      true,
	"build":      true,
	"refactor":   true,
	"review":     true,
	"check":      true,
	"analyze":    true,
	"design":     true,
	"deploy":     true,
	"migrate":    true,
	"optimize":   true,
	"explain":    true,
	"describe":   true,
	"generate":   true,
	"run":        true,
	"test":       true,
	"install":    true,
	"configure":  true,
	"setup":      true,
	"make":       true,
	"list":       true,
	"show":       true,
	"find":       true,
	"get":        true,
	"set":        true,
	"use":        true,
	"edit":       true,
	"change":     true,
	"move":       true,
	"rename":     true,
	"replace":    true,
	"ensure":     true,
	"handle":     true,
	"return":     true,
	"define":     true,
	"print":      true,
	"parse":      true,
	"load":       true,
	"save":       true,
	"read":       true,
	"send":       true,
	"fetch":      true,
	"connect":    true,
	"open":       true,
	"close":      true,
	"start":      true,
	"stop":       true,
	"restart":    true,
	"enable":     true,
	"disable":    true,
	"convert":    true,
	"validate":   true,
	"verify":     true,
	"initialize": true,
	"extend":     true,
	"override":   true,
	"append":     true,
	"insert":     true,
	"extract":    true,
	"split":      true,
	"merge":      true,
	"sort":       true,
	"filter":     true,
	"map":        true,
	"reduce":     true,
	"transform":  true,
}

// reToken matches camelCase, snake_case, or dotted.paths technical tokens.
var reCamelCase = regexp.MustCompile(`[a-z][A-Z]`)
var reSnakeCase = regexp.MustCompile(`[a-z_]{2,}[_][a-z]{2,}`)
var reDottedPath = regexp.MustCompile(`[a-zA-Z]+\.[a-zA-Z]+`)

// AnalyzeNLP computes NLP-style heuristic metrics for the given text.
func AnalyzeNLP(text string) NLPMetrics {
	if strings.TrimSpace(text) == "" {
		return NLPMetrics{}
	}

	words := strings.Fields(text)
	totalWords := len(words)
	sentences := splitSentences(text)
	totalSentences := len(sentences)

	nm := NLPMetrics{}

	// VocabularyRichness: type-token ratio (unique / total words)
	nm.VocabularyRichness = vocabularyRichness(words)

	// AvgSentenceLength: words per sentence
	if totalSentences > 0 {
		nm.AvgSentenceLength = float64(totalWords) / float64(totalSentences)
	}

	// QuestionDensity: question sentences / total sentences
	if totalSentences > 0 {
		questions := countQuestionSentences(sentences)
		nm.QuestionDensity = float64(questions) / float64(totalSentences)
	}

	// ImperativeRatio: sentences starting with imperative verbs / total sentences
	if totalSentences > 0 {
		imperatives := countImperativeSentences(sentences)
		nm.ImperativeRatio = float64(imperatives) / float64(totalSentences)
	}

	// TechnicalDensity: technical tokens / total words
	if totalWords > 0 {
		techCount := countTechnicalTokens(words)
		nm.TechnicalDensity = float64(techCount) / float64(totalWords)
		if nm.TechnicalDensity > 1.0 {
			nm.TechnicalDensity = 1.0
		}
	}

	// ReadabilityScore: Flesch-Kincaid approximation
	// FK Reading Ease = 206.835 - 1.015*(words/sentences) - 84.6*(syllables/words)
	// Score clamped to [0, 100].
	if totalSentences > 0 && totalWords > 0 {
		syllables := countSyllables(words)
		avgWordsPerSentence := float64(totalWords) / float64(totalSentences)
		avgSyllablesPerWord := float64(syllables) / float64(totalWords)
		score := 206.835 - 1.015*avgWordsPerSentence - 84.6*avgSyllablesPerWord
		if score < 0 {
			score = 0
		}
		if score > 100 {
			score = 100
		}
		nm.ReadabilityScore = score
	}

	return nm
}

// splitSentences splits text into sentences on .!? followed by whitespace or end-of-string.
func splitSentences(text string) []string {
	re := regexp.MustCompile(`[^.!?]*[.!?]+`)
	parts := re.FindAllString(text, -1)
	if len(parts) == 0 {
		trimmed := strings.TrimSpace(text)
		if trimmed != "" {
			return []string{trimmed}
		}
		return nil
	}
	// Include any trailing text that lacks a terminator.
	matched := strings.Join(parts, "")
	remaining := strings.TrimSpace(text[len(matched):])
	if remaining != "" {
		parts = append(parts, remaining)
	}
	return parts
}

// vocabularyRichness returns unique words / total words (type-token ratio).
func vocabularyRichness(words []string) float64 {
	if len(words) == 0 {
		return 0
	}
	seen := make(map[string]struct{}, len(words))
	for _, w := range words {
		norm := strings.ToLower(strings.Trim(w, ".,!?;:\"'()[]{}"))
		if norm != "" {
			seen[norm] = struct{}{}
		}
	}
	return float64(len(seen)) / float64(len(words))
}

// countQuestionSentences counts sentences that end with '?'.
func countQuestionSentences(sentences []string) int {
	count := 0
	for _, s := range sentences {
		trimmed := strings.TrimSpace(s)
		if strings.HasSuffix(trimmed, "?") {
			count++
		}
	}
	return count
}

// countImperativeSentences counts sentences whose first word is an imperative verb.
func countImperativeSentences(sentences []string) int {
	count := 0
	for _, s := range sentences {
		words := strings.Fields(s)
		if len(words) == 0 {
			continue
		}
		first := strings.ToLower(strings.Trim(words[0], ".,!?;:\"'()[]{}"))
		if imperativeStarters[first] {
			count++
		}
	}
	return count
}

// countTechnicalTokens counts words that look like camelCase, snake_case, dotted.paths,
// or contain digits mixed with letters (e.g. JWT, OAuth2).
func countTechnicalTokens(words []string) int {
	count := 0
	for _, w := range words {
		clean := strings.Trim(w, ".,!?;:\"'()[]{}*`")
		if clean == "" {
			continue
		}
		if isTechnicalToken(clean) {
			count++
		}
	}
	return count
}

func isTechnicalToken(s string) bool {
	// camelCase
	if reCamelCase.MatchString(s) {
		return true
	}
	// snake_case
	if reSnakeCase.MatchString(s) {
		return true
	}
	// dotted.path (e.g. pkg.Name, some.func)
	if reDottedPath.MatchString(s) {
		return true
	}
	// ALL CAPS abbreviation of 2+ chars (e.g. JWT, REST, API, SQL)
	if len(s) >= 2 && isAllUpper(s) {
		return true
	}
	// Mixed letters and digits (e.g. OAuth2, HTTP2, k8s, v1)
	hasLetter := false
	hasDigit := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if hasLetter && hasDigit {
		return true
	}
	return false
}

func isAllUpper(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) && unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// countSyllables approximates syllable count for all words using a vowel-group heuristic.
func countSyllables(words []string) int {
	total := 0
	for _, w := range words {
		total += syllablesInWord(strings.ToLower(strings.Trim(w, ".,!?;:\"'()[]{}*`")))
	}
	return total
}

// syllablesInWord counts syllables in a single word using vowel-group heuristics.
func syllablesInWord(word string) int {
	if len(word) == 0 {
		return 0
	}
	vowels := "aeiouy"
	count := 0
	prevWasVowel := false
	for _, ch := range word {
		isVowel := strings.ContainsRune(vowels, ch)
		if isVowel && !prevWasVowel {
			count++
		}
		prevWasVowel = isVowel
	}
	// Silent trailing 'e'
	if strings.HasSuffix(word, "e") && count > 1 {
		count--
	}
	if count == 0 {
		return 1
	}
	return count
}
