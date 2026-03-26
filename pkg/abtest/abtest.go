// Package abtest provides A/B testing for prompt routing configurations.
// It compares two different tier configs to measure routing differences and
// potential cost savings.
package abtest

import (
	"crypto/sha256"
	"fmt"

	"github.com/mikeshogin/promptlint/pkg/config"
	"github.com/mikeshogin/promptlint/pkg/router"
)

// ABResult records a single prompt routing result for one variant.
type ABResult struct {
	PromptHash   string  `json:"prompt_hash"`
	Variant      string  `json:"variant"` // "a" or "b"
	RoutedModel  string  `json:"routed_model"`
	Complexity   string  `json:"complexity"`
	Score        float64 `json:"score"`
	CostPer1k    float64 `json:"cost_per_1k"`
}

// ABComparison holds the two results for a single prompt.
type ABComparison struct {
	PromptHash string   `json:"prompt_hash"`
	A          ABResult `json:"a"`
	B          ABResult `json:"b"`
	Differs    bool     `json:"differs"` // true when A and B route to different models
}

// ABTest compares routing decisions between two configurations.
type ABTest struct {
	Name    string
	cfgA    *config.Config
	cfgB    *config.Config
	results []ABComparison
}

// New creates an ABTest with the provided configs. If cfgA is nil, DefaultConfig is used.
func New(name string, cfgA, cfgB *config.Config) *ABTest {
	if cfgA == nil {
		cfgA = config.DefaultConfig()
	}
	if cfgB == nil {
		cfgB = config.DefaultConfig()
	}
	return &ABTest{
		Name: name,
		cfgA: cfgA,
		cfgB: cfgB,
	}
}

// Run routes prompt with both configs and records the comparison.
func (t *ABTest) Run(prompt string) ABComparison {
	hash := promptHash(prompt)

	rA := router.New(t.cfgA).Route(prompt)
	rB := router.New(t.cfgB).Route(prompt)

	a := ABResult{
		PromptHash:  hash,
		Variant:     "a",
		RoutedModel: rA.Model,
		Complexity:  rA.Complexity,
		Score:       rA.Score,
		CostPer1k:   tierCost(t.cfgA, rA.Model),
	}
	b := ABResult{
		PromptHash:  hash,
		Variant:     "b",
		RoutedModel: rB.Model,
		Complexity:  rB.Complexity,
		Score:       rB.Score,
		CostPer1k:   tierCost(t.cfgB, rB.Model),
	}

	cmp := ABComparison{
		PromptHash: hash,
		A:          a,
		B:          b,
		Differs:    a.RoutedModel != b.RoutedModel,
	}
	t.results = append(t.results, cmp)
	return cmp
}

// Summary aggregates all recorded comparisons and returns statistics.
func (t *ABTest) Summary() Summary {
	total := len(t.results)
	if total == 0 {
		return Summary{Name: t.Name}
	}

	var diffCount int
	var totalCostA, totalCostB float64
	modelCountA := map[string]int{}
	modelCountB := map[string]int{}

	for _, c := range t.results {
		if c.Differs {
			diffCount++
		}
		totalCostA += c.A.CostPer1k
		totalCostB += c.B.CostPer1k
		modelCountA[c.A.RoutedModel]++
		modelCountB[c.B.RoutedModel]++
	}

	avgCostA := totalCostA / float64(total)
	avgCostB := totalCostB / float64(total)

	var costDiffPct float64
	if avgCostA > 0 {
		costDiffPct = (avgCostB - avgCostA) / avgCostA * 100
	}

	return Summary{
		Name:          t.Name,
		TotalPrompts:  total,
		DifferentRoutes: diffCount,
		SameRoutes:    total - diffCount,
		AvgCostA:      round2(avgCostA),
		AvgCostB:      round2(avgCostB),
		CostDiffPct:   round2(costDiffPct),
		ModelCountA:   modelCountA,
		ModelCountB:   modelCountB,
	}
}

// Results returns all recorded comparisons.
func (t *ABTest) Results() []ABComparison {
	return t.results
}

// Summary holds aggregated A/B test statistics.
type Summary struct {
	Name            string         `json:"name"`
	TotalPrompts    int            `json:"total_prompts"`
	DifferentRoutes int            `json:"different_routes"`
	SameRoutes      int            `json:"same_routes"`
	AvgCostA        float64        `json:"avg_cost_a"`
	AvgCostB        float64        `json:"avg_cost_b"`
	CostDiffPct     float64        `json:"cost_diff_pct"` // negative means B is cheaper
	ModelCountA     map[string]int `json:"model_count_a"`
	ModelCountB     map[string]int `json:"model_count_b"`
}

// CheaperVariant returns "a", "b", or "equal" based on average cost.
func (s *Summary) CheaperVariant() string {
	switch {
	case s.AvgCostA < s.AvgCostB:
		return "a"
	case s.AvgCostB < s.AvgCostA:
		return "b"
	default:
		return "equal"
	}
}

// promptHash returns a short sha256 hex prefix for the prompt.
func promptHash(prompt string) string {
	h := sha256.Sum256([]byte(prompt))
	return fmt.Sprintf("%x", h[:4])
}

// tierCost looks up the CostPer1k for the given model name in the config.
// Returns 0 if the model is not found.
func tierCost(cfg *config.Config, model string) float64 {
	for _, t := range cfg.Tiers {
		if t.Name == model {
			return t.CostPer1k
		}
	}
	return 0
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
