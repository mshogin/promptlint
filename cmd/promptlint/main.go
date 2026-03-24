package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/mikeshogin/promptlint/pkg/analyzer"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: promptlint analyze < prompt.txt\n")
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

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nUsage: promptlint analyze\n", cmd)
		os.Exit(1)
	}
}
