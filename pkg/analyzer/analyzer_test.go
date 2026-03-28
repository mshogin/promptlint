package analyzer

import (
	"testing"
)

func TestAnalyzeSimplePrompt(t *testing.T) {
	r := Analyze("Fix the bug in server.go:57")

	if r.Action != "fix" {
		t.Errorf("expected action 'fix', got '%s'", r.Action)
	}
	if !r.HasCodeRef {
		t.Error("expected HasCodeRef to be true")
	}
	if !r.HasFilePath {
		t.Error("expected HasFilePath to be true")
	}
	if r.Complexity != "low" {
		t.Errorf("expected complexity 'low', got '%s'", r.Complexity)
	}
	if r.SuggestedModel != "haiku" {
		t.Errorf("expected model 'haiku', got '%s'", r.SuggestedModel)
	}
}

func TestAnalyzeComplexPrompt(t *testing.T) {
	prompt := `Review the architecture of our microservice system.
We have coupling issues between the payment service and order service.
The dependency graph shows circular dependencies.
Can you analyze the SOLID violations?
What refactoring pattern would you suggest?
Also check the Docker deployment pipeline.

` + "```go\nfunc main() {\n  // code\n}\n```"

	r := Analyze(prompt)

	// Complex prompt should be at least medium
	if r.Complexity == "low" {
		t.Errorf("expected complexity >= 'medium', got '%s'", r.Complexity)
	}
	if r.SuggestedModel == "haiku" {
		t.Errorf("expected model != 'haiku', got '%s'", r.SuggestedModel)
	}
	if r.Questions < 2 {
		t.Errorf("expected at least 2 questions, got %d", r.Questions)
	}
	if !r.HasCodeBlock {
		t.Error("expected HasCodeBlock to be true")
	}
	if r.Domain["architecture"] == 0 {
		t.Error("expected architecture domain score > 0")
	}
}

func TestAnalyzeEmptyPrompt(t *testing.T) {
	r := Analyze("")

	if r.Length != 0 {
		t.Errorf("expected length 0, got %d", r.Length)
	}
	if r.SuggestedModel != "haiku" {
		t.Errorf("expected model 'haiku', got '%s'", r.SuggestedModel)
	}
}

func TestComplexityScoreNumeric(t *testing.T) {
	// Simple bug fix should score low
	simple := Analyze("Fix bug in server.go")
	if simple.ComplexityScore >= 25 {
		t.Errorf("expected complexity_score < 25 for simple prompt, got %d", simple.ComplexityScore)
	}
	if simple.Complexity != "low" {
		t.Errorf("expected complexity 'low' for simple prompt, got '%s'", simple.Complexity)
	}

	// Architecture prompt should score high
	arch := Analyze("Design microservices architecture with CQRS pattern, event sourcing, and saga orchestration. Include dependency graph, API gateway, and circuit breaker patterns.")
	if arch.ComplexityScore < 50 {
		t.Errorf("expected complexity_score >= 50 for architecture prompt, got %d", arch.ComplexityScore)
	}
	if arch.Complexity != "high" {
		t.Errorf("expected complexity 'high' for architecture prompt, got '%s'", arch.Complexity)
	}
	if arch.SuggestedModel != "opus" {
		t.Errorf("expected model 'opus' for architecture prompt, got '%s'", arch.SuggestedModel)
	}
}

func TestComplexityScoreRoleIndicator(t *testing.T) {
	// Role/persona prompt should score higher than equivalent without role
	withRole := Analyze("You are an expert security auditor. Analyze this codebase for vulnerabilities.")
	withoutRole := Analyze("Analyze this codebase for vulnerabilities.")

	if withRole.ComplexityScore <= withoutRole.ComplexityScore {
		t.Errorf("expected role prompt (score=%d) to score higher than no-role prompt (score=%d)",
			withRole.ComplexityScore, withoutRole.ComplexityScore)
	}
}

func TestComplexityScoreMicroservicesArch(t *testing.T) {
	// The canonical "before/after" case from issue #15:
	// "Design microservices arch with CQRS..." should route to opus, not haiku
	prompt := "Design microservices architecture with CQRS and event sourcing. Include service boundaries, coupling analysis, and deployment pipeline."
	r := Analyze(prompt)

	if r.Complexity == "low" {
		t.Errorf("microservices architecture prompt should NOT be 'low' complexity, got '%s' (score=%d)",
			r.Complexity, r.ComplexityScore)
	}
	if r.SuggestedModel == "haiku" {
		t.Errorf("microservices architecture prompt should NOT route to 'haiku', got '%s' (score=%d)",
			r.SuggestedModel, r.ComplexityScore)
	}
}
