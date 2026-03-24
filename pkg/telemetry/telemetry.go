package telemetry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/mikeshogin/promptlint/pkg/analyzer"
)

// Record represents one telemetry entry.
type Record struct {
	Timestamp  string          `json:"timestamp"`
	PromptHash string          `json:"prompt_hash"`
	Analysis   analyzer.Result `json:"analysis"`
	Routed     string          `json:"routed_to"`
	Source     string          `json:"source,omitempty"`
}

// Collector collects and stores telemetry records.
type Collector struct {
	filePath string
}

// NewCollector creates a collector that writes to the given JSONL file.
func NewCollector(filePath string) *Collector {
	return &Collector{filePath: filePath}
}

// Record stores an analysis result with routing decision.
func (c *Collector) Record(prompt string, analysis analyzer.Result, routedTo string, source string) error {
	record := Record{
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		PromptHash: hashPrompt(prompt),
		Analysis:   analysis,
		Routed:     routedTo,
		Source:     source,
	}

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal telemetry: %w", err)
	}

	f, err := os.OpenFile(c.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o640)
	if err != nil {
		return fmt.Errorf("open telemetry file: %w", err)
	}
	defer f.Close()

	_, err = f.Write(append(data, '\n'))
	return err
}

// Stats returns basic statistics from collected telemetry.
func (c *Collector) Stats() (map[string]int, error) {
	data, err := os.ReadFile(c.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]int{}, nil
		}
		return nil, err
	}

	stats := map[string]int{
		"total":  0,
		"haiku":  0,
		"sonnet": 0,
		"opus":   0,
	}

	for _, line := range splitLines(data) {
		if len(line) == 0 {
			continue
		}
		var r Record
		if err := json.Unmarshal(line, &r); err != nil {
			continue
		}
		stats["total"]++
		stats[r.Routed]++
	}

	return stats, nil
}

func hashPrompt(prompt string) string {
	h := sha256.Sum256([]byte(prompt))
	return hex.EncodeToString(h[:8])
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
