package metrics

import (
	"strings"
)

var actionVerbs = map[string]string{
	"fix":        "fix",
	"repair":     "fix",
	"debug":      "fix",
	"create":     "create",
	"add":        "create",
	"implement":  "create",
	"build":      "create",
	"write":      "create",
	"review":     "review",
	"check":      "review",
	"analyze":    "review",
	"refactor":   "refactor",
	"rewrite":    "refactor",
	"restructure": "refactor",
	"delete":     "delete",
	"remove":     "delete",
	"design":     "create",
	"architect":  "create",
	"plan":       "create",
	"propose":    "create",
	"deploy":     "deploy",
	"migrate":    "refactor",
	"optimize":   "refactor",
	"improve":    "refactor",
	"explain":    "explain",
	"describe":   "explain",
	// RU action-verbs (корпоративные промпты русские).
	"исправь":    "fix",
	"почини":     "fix",
	"создай":     "create",
	"напиши":     "create",
	"добавь":     "create",
	"реализуй":   "create",
	"проверь":    "review",
	"проанализируй": "review",
	"отрефактори":   "refactor",
	"перепиши":   "refactor",
	"удали":      "delete",
	"спроектируй":   "create",
	"объясни":    "explain",
	"опиши":      "explain",
	"докажи":     "explain",
	"выведи":     "explain",
	"реши":       "explain",
}

// domainKeywords — keyword-подстроки по доменам. Русские заданы КОРНЯМИ (strings.Count ловит
// подстроку -> «функци» матчит функция/функцию/функции). WB-промпты русские: без рус-корней
// классификатор слеп (рус-текст -> 0 совпадений -> general), что схлопывало per-prompt роутинг.
var domainKeywords = map[string][]string{
	"reasoning": {
		// EN
		"prove", "proof", "derive", "theorem", "lemma", "reasoning", "rationale",
		"step by step", "step-by-step", "deduce", "infer", "logic", "calculate",
		"compute", "solve", "equation", "math", "formula", "induction",
		// RU (корни — ловят словоформы)
		"докаж", "доказательств", "вывед", "теорем", "лемм", "рассужд",
		"пошагов", "вычисл", "посчитай", "реши", "уравнен", "формул",
		"логическ", "обоснуй", "обоснован", "выведи", "почему",
	},
	"code": {
		// EN
		"function", "method", "variable", "class", "struct", "interface",
		"loop", "array", "string", "int", "bool", "error", "return",
		"import", "package", "module", "test", "unittest", "assert",
		"algorithm", "list", "slice", "map", "pointer", "goroutine",
		// RU (корни)
		"функци", "метод", "класс", "структур", "интерфейс", "цикл",
		"массив", "строк", "переменн", "пакет", "тест", "ошибк",
		"возврат", "список", "алгоритм", "указател", "горутин", "срез",
	},
	"architecture": {
		"architecture", "design", "pattern", "solid", "dip", "srp",
		"coupling", "cohesion", "dependency", "layer", "boundary",
		"component", "service", "microservice", "monolith", "graph",
		"cycle", "fan-out", "fan-in", "metric", "cqrs", "event sourcing",
		"saga", "domain driven", "hexagonal", "clean architecture",
		"event-driven", "distributed", "scalab", "resilien",
		"load balanc", "api gateway", "circuit breaker", "retry",
		"dead letter", "idempoten",
		// RU (корни)
		"архитектур", "паттерн", "проектир", "зависимост", "связност",
		"слой", "компонент", "сервис", "микросервис", "монолит", "граф",
		"масштабир", "отказоустойч", "распределённ", "распределенн",
	},
	"infrastructure": {
		"docker", "kubernetes", "k8s", "nginx", "deploy", "ci", "cd",
		"pipeline", "server", "vps", "ssh", "container", "pod",
		"helm", "terraform", "ansible",
		// RU (корни)
		"деплой", "разверн", "развёрт", "разверт", "контейнер", "кластер",
		"сервер", "пайплайн", "инфраструктур",
	},
	"content": {
		"article", "post", "blog", "linkedin", "twitter", "write",
		"publish", "draft", "headline", "summary", "translate",
		// RU (корни)
		"статья", "стать", "пост", "блог", "опубликуй", "опубликов",
		"черновик", "заголовок", "перевод", "переведи", "текст", "напиши пост",
	},
}

// DetectAction identifies the primary action requested in the prompt.
func DetectAction(text string) string {
	lower := strings.ToLower(text)
	words := strings.Fields(lower)

	for _, word := range words {
		clean := strings.Trim(word, ".,!?;:")
		if action, ok := actionVerbs[clean]; ok {
			return action
		}
	}

	return "unknown"
}

// ClassifyDomain returns a vector of domain scores.
func ClassifyDomain(text string) map[string]float64 {
	lower := strings.ToLower(text)
	result := make(map[string]float64)

	for domain, keywords := range domainKeywords {
		count := 0
		for _, kw := range keywords {
			count += strings.Count(lower, kw)
		}
		if count > 0 {
			// Normalize: more keywords = higher score, cap at 1.0
			score := float64(count) / 5.0
			if score > 1.0 {
				score = 1.0
			}
			result[domain] = score
		}
	}

	// Default if nothing detected
	if len(result) == 0 {
		result["general"] = 1.0
	}

	return result
}
