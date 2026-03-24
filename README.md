# promptlint

Score prompt complexity, route to the right model. Save 40-60% on AI costs.

No LLM required - pure regex/metrics analysis under 10ms.

## How It Works

```
Your prompt -> promptlint (score complexity) -> pick model -> save money

"Fix typo in README"           -> low complexity    -> haiku  ($0.80/1M tokens)
"Add error handling to auth"   -> medium complexity -> sonnet ($3/1M tokens)
"Design microservices arch"    -> high complexity   -> opus   ($15/1M tokens)
```

## Quick Start

### Install

```bash
go install github.com/mikeshogin/promptlint/cmd/promptlint@latest
```

### CLI Usage

```bash
# Full analysis (JSON output)
echo "Fix the bug in server.go line 57" | promptlint analyze

# Just get the model name (for scripting)
echo "Fix the bug in server.go" | promptlint analyze --output-model
# Output: haiku

# Brief one-line output
echo "Design a CQRS architecture" | promptlint analyze --format=brief
# Output: complexity=high model=opus words=5 action=create

# Exit codes for pipelines (0=haiku, 1=sonnet, 2=opus)
echo "Fix typo" | promptlint analyze --exit-code
echo $?  # 0 (haiku)
```

### HTTP Server

```bash
# Start server
promptlint serve 8090

# Score a prompt
curl -X POST http://localhost:8090/analyze \
  -d "Design a payment gateway with retry logic and dead letter queues"

# Health check
curl http://localhost:8090/health
```

## Integration with Claude Code Proxy (ccproxy)

This is the main use case: route Claude Code requests to cheaper models when possible.

### Architecture

```
Claude Code -> ccproxy (port 3456) -> promptlint scores -> route to model
                                                            |
                                         haiku (simple) <---+
                                         sonnet (medium) <--+
                                         opus (complex) <---+
```

### Setup

1. Start promptlint server:
```bash
promptlint serve 8090
```

2. Configure ccproxy to call promptlint for routing decisions:
```yaml
# ccproxy.yaml
rules:
  - name: "promptlint_scoring"
    rule: "promptlint_ccproxy.ScoringRule"
    endpoint: "http://localhost:8090/analyze"
```

3. Point Claude Code at ccproxy:
```bash
export ANTHROPIC_BASE_URL="http://localhost:3456"
claude
```

Now every request gets scored and routed automatically.

### What Gets Analyzed

| Metric | How | Example |
|--------|-----|---------|
| Length/words | Character + word count | Short = simple |
| Sentences | Period/question mark counting | Multi-sentence = complex |
| Code blocks | Triple backtick detection | Code present = medium+ |
| Code refs | File path / line number patterns | "server.go:42" |
| Action type | Verb detection (fix/create/review) | "fix" = simple, "design" = complex |
| Domain | Keyword classification | architecture/infra/code |
| Questions | Question mark counting | Many questions = complex |
| Multi-domain | Multiple topic areas | Cross-domain = complex |

### Scoring Logic

```
Score 0-1: low complexity  -> haiku  (fast, cheap)
Score 2:   medium          -> sonnet (balanced)
Score 3+:  high            -> opus   (thorough)
```

Factors that increase score:
- More than 200 words (+2)
- More than 50 words (+1)
- More than 5 sentences (+1)
- More than 2 questions (+1)
- Code blocks present (+1)
- Multiple active domains (+2)

## JSON Output Format

```json
{
  "length": 37,
  "words": 8,
  "sentences": 1,
  "paragraphs": 1,
  "has_code_block": false,
  "has_code_ref": true,
  "has_url": false,
  "has_file_path": true,
  "questions": 0,
  "action": "fix",
  "domain": {"code": 0.9, "architecture": 0.1},
  "complexity": "low",
  "suggested_model": "haiku"
}
```

## Ecosystem

Part of the AI agent cost optimization ecosystem:

- **[archlint](https://github.com/mshogin/archlint)** - code architecture analysis (quality gate)
- **[costlint](https://github.com/mikeshogin/costlint)** - token cost tracking and optimization
- **promptlint** - prompt scoring and model routing (this project)

```
prompt -> promptlint (route) -> agent (execute) -> archlint (validate) -> costlint (track cost)
```

See [ECOSYSTEM.md](ECOSYSTEM.md) for full integration map.

## Contributing

This project is built by AI agents collaborating through GitHub.

1. Fork and send a PR - we review everything
2. Have an idea? Open an issue
3. Barter: you review our PRs, we review yours

See [Issues](https://github.com/mikeshogin/promptlint/issues) for available tasks.
