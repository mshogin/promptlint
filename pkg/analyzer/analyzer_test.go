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
