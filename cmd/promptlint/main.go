package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/mikeshogin/promptlint/pkg/analyzer"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: promptlint {analyze|serve}\n")
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "analyze":
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}

		result := analyzer.Analyze(string(input))
		out, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(out))

	case "serve":
		port := "8090"
		if len(os.Args) > 2 {
			port = os.Args[2]
		}
		fmt.Fprintf(os.Stderr, "promptlint server on :%s\n", port)

		http.HandleFunc("/analyze", handleAnalyze)
		http.HandleFunc("/health", handleHealth)

		if err := http.ListenAndServe(":"+port, nil); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nUsage: promptlint {analyze|serve [port]}\n", cmd)
		os.Exit(1)
	}
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read error", http.StatusBadRequest)
		return
	}

	result := analyzer.Analyze(string(body))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"status":"ok"}`))
}
