package metrics

import "testing"

// ClassifyDomain должен различать reasoning vs code на РУССКИХ И английских промптах. Без рус-keywords
// и без reasoning-домена классификатор был слеп на русском (всё -> general), что схлопывало per-prompt
// роутинг claude-router в одну fallback-модель. topDomain — домен с максимальным score.
func topDomain(scores map[string]float64) string {
	best, bestScore := "", -1.0
	for d, s := range scores {
		if s > bestScore || (s == bestScore && d < best) {
			bestScore = s
			best = d
		}
	}
	return best
}

func TestClassifyDomain_ReasoningVsCode(t *testing.T) {
	cases := []struct {
		name, text, wantTop string
	}{
		{"ru-reasoning", "пошагово докажи что 0.999... = 1", "reasoning"},
		{"ru-code", "напиши функцию на Go которая разворачивает связный список", "code"},
		{"en-reasoning", "prove step by step that sqrt(2) is irrational", "reasoning"},
		{"en-code", "write a function to reverse a linked list", "code"},
		{"ru-arch", "спроектируй архитектуру микросервиса с минимальной связностью", "architecture"},
		{"ru-infra", "разверни сервис в kubernetes кластере через helm", "infrastructure"},
	}

	for _, c := range cases {
		got := topDomain(ClassifyDomain(c.text))
		if got != c.wantTop {
			t.Errorf("%s: ClassifyDomain(%q) top=%q, want %q (классификатор слеп?)", c.name, c.text, got, c.wantTop)
		}
	}
}

// reasoning и code НЕ должны давать ОДИНАКОВЫЙ top-домен — иначе scoreModel выберет один winner
// (исходный баг: рус reasoning И рус code -> оба general -> fallback).
func TestClassifyDomain_DistinctTopDomains(t *testing.T) {
	r := topDomain(ClassifyDomain("пошагово докажи что 0.999... = 1"))
	c := topDomain(ClassifyDomain("напиши функцию на Go разворачивающую связный список"))
	if r == c {
		t.Errorf("reasoning и code дали ОДИНАКОВЫЙ top-домен %q — per-prompt различение не работает", r)
	}
	if r == "general" || c == "general" {
		t.Errorf("русский промпт классифицирован как general (классификатор слеп на русском): reasoning=%q code=%q", r, c)
	}
}
