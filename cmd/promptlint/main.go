package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/mikeshogin/promptlint/pkg/analyzer"
	"github.com/mikeshogin/promptlint/pkg/server"
)

// Exit codes for pipeline integration.
const (
	ExitHaiku  = 0
	ExitSonnet = 1
	ExitOpus   = 2
	ExitError  = 3
)

func modelExitCode(model string) int {
	switch model {
	case "haiku":
		return ExitHaiku
	case "sonnet":
		return ExitSonnet
	case "opus":
		return ExitOpus
	default:
		return ExitError
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: promptlint {analyze|serve}\n")
		fmt.Fprintf(os.Stderr, "\nanalyze flags:\n")
		fmt.Fprintf(os.Stderr, "  --output-model   print only model name\n")
		fmt.Fprintf(os.Stderr, "  --format=json    output format: json (default), brief\n")
		fmt.Fprintf(os.Stderr, "  --exit-code      use model-based exit codes (0=haiku,1=sonnet,2=opus)\n")
		os.Exit(ExitError)
	}

	cmd := os.Args[1]

	switch cmd {
	case "analyze":
		// Parse flags
		outputModel := false
		exitCode := false
		format := "json"
		for _, arg := range os.Args[2:] {
			switch {
			case arg == "--output-model":
				outputModel = true
			case arg == "--exit-code":
				exitCode = true
			case strings.HasPrefix(arg, "--format="):
				format = strings.TrimPrefix(arg, "--format=")
			}
		}

		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(ExitError)
		}

		result := analyzer.Analyze(string(input))

		if outputModel {
			fmt.Println(result.SuggestedModel)
		} else {
			switch format {
			case "brief":
				fmt.Printf("complexity=%s model=%s words=%d action=%s\n",
					result.Complexity, result.SuggestedModel, result.Words, result.Action)
			default:
				out, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(out))
			}
		}

		if exitCode {
			os.Exit(modelExitCode(result.SuggestedModel))
		}

	case "serve":
		port := "8080"
		for _, arg := range os.Args[2:] {
			if strings.HasPrefix(arg, "--port=") {
				port = strings.TrimPrefix(arg, "--port=")
			} else if arg == "--port" {
				// handled by next iteration below
			}
		}
		// Support --port 8080 (space-separated)
		args := os.Args[2:]
		for i, arg := range args {
			if arg == "--port" && i+1 < len(args) {
				port = args[i+1]
			}
		}

		fmt.Fprintf(os.Stderr, "promptlint server on :%s\n", port)

		if err := http.ListenAndServe(":"+port, server.New()); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\nUsage: promptlint {analyze|serve [port]}\n", cmd)
		os.Exit(1)
	}
}

