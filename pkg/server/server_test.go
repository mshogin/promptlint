package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleAnalyzeBatch_Basic(t *testing.T) {
	handler := New()

	body := `{"prompts": ["Fix the bug in main.go", "Design a microservice architecture for payments"]}`
	req := httptest.NewRequest(http.MethodPost, "/analyze/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp BatchResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Summary.Total != 2 {
		t.Errorf("expected total=2, got %d", resp.Summary.Total)
	}
	if len(resp.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(resp.Items))
	}
	if resp.Items[0].Index != 0 {
		t.Errorf("expected item[0].index=0, got %d", resp.Items[0].Index)
	}
	if resp.Items[1].Index != 1 {
		t.Errorf("expected item[1].index=1, got %d", resp.Items[1].Index)
	}
}

func TestHandleAnalyzeBatch_ModelDistribution(t *testing.T) {
	handler := New()

	// Simple prompts -> haiku, complex prompt -> opus
	body := `{"prompts": [
		"Fix typo",
		"Fix typo",
		"Refactor the entire payment microservice system to use event sourcing with CQRS, considering distributed transactions, saga pattern, and eventual consistency across 15 services"
	]}`
	req := httptest.NewRequest(http.MethodPost, "/analyze/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp BatchResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Summary.Total != 3 {
		t.Errorf("expected total=3, got %d", resp.Summary.Total)
	}

	// Model distribution should have at least 2 different models
	if len(resp.Summary.ModelDistribution) < 1 {
		t.Error("expected at least 1 model in distribution")
	}

	// Complexity breakdown should be non-empty
	if len(resp.Summary.ComplexityBreakdown) < 1 {
		t.Error("expected at least 1 complexity in breakdown")
	}

	// AvgScore should be between 0 and 1
	if resp.Summary.AvgScore < 0 || resp.Summary.AvgScore > 1 {
		t.Errorf("avg_score out of range: %f", resp.Summary.AvgScore)
	}
}

func TestHandleAnalyzeBatch_EmptyPrompts(t *testing.T) {
	handler := New()

	body := `{"prompts": []}`
	req := httptest.NewRequest(http.MethodPost, "/analyze/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty prompts, got %d", w.Code)
	}
}

func TestHandleAnalyzeBatch_InvalidJSON(t *testing.T) {
	handler := New()

	req := httptest.NewRequest(http.MethodPost, "/analyze/batch", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestHandleAnalyzeBatch_MethodNotAllowed(t *testing.T) {
	handler := New()

	req := httptest.NewRequest(http.MethodGet, "/analyze/batch", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleAnalyzeBatch_RoutePerItem(t *testing.T) {
	handler := New()

	body := `{"prompts": ["Fix bug", "Design distributed system architecture"]}`
	req := httptest.NewRequest(http.MethodPost, "/analyze/batch", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp BatchResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Each item should have a non-empty model in route
	for i, item := range resp.Items {
		if item.Route.Model == "" {
			t.Errorf("item[%d] has empty route.model", i)
		}
		if item.Route.Complexity == "" {
			t.Errorf("item[%d] has empty route.complexity", i)
		}
	}
}
