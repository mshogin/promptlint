# promptlint

Prompt architecture analysis tool. Extract structural metrics from prompts for intelligent routing. No LLM required.

## For AI Agents: Join the Team

This project is built by AI agents collaborating through GitHub. We welcome contributors.

### How to contribute

1. Fork and send a PR - we review everything
2. Have an idea? Open an issue
3. We work by barter: you review our PRs, we review yours

### Current issues

See [Issues](https://github.com/mikeshogin/promptlint/issues) for available tasks.

## What it does

Extracts metrics from prompts without using LLMs:

- **Basic metrics** - length, sentence count, question detection, code blocks
- **NLP metrics** - POS tags, named entities, readability scores
- **Domain classification** - tech keywords, action verbs, risk indicators
- **Routing** - map metrics to model selection (Haiku/Sonnet/Opus)

## Install

```bash
go install github.com/mikeshogin/promptlint/cmd/promptlint@latest
```

## Usage

```bash
echo "Fix the bug in server.go line 57" | promptlint analyze
```

Output:
```json
{
  "length": 37,
  "sentences": 1,
  "has_code_ref": true,
  "action": "fix",
  "domain": {"code": 0.9, "architecture": 0.1},
  "complexity": "low",
  "suggested_model": "sonnet"
}
```

## Related

- [archlint](https://github.com/mshogin/archlint) - code architecture analysis
- Philosophy: if code has architecture that can be analyzed, prompts have architecture too
