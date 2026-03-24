# promptlint + ccproxy Integration

Route Claude Code requests to the cheapest sufficient model based on prompt complexity scoring.

## How it works

```
Claude Code -> ccproxy:3456 -> promptlint:8090 (score) -> route to model
                                                           |
                                        haiku  (simple) <--+
                                        sonnet (medium) <--+
                                        opus   (complex) <-+
```

1. Claude Code sends request to ccproxy (via ANTHROPIC_BASE_URL)
2. ccproxy calls promptlint HTTP API with the prompt text
3. promptlint scores complexity (no LLM, <10ms)
4. ccproxy routes request to the appropriate model
5. Response flows back to Claude Code transparently

## Setup

```bash
./setup.sh
```

Or manually:

```bash
# Terminal 1: start promptlint server
promptlint serve 8090

# Terminal 2: start ccproxy
ccproxy start

# Terminal 3: use Claude Code
export ANTHROPIC_BASE_URL="http://localhost:3456"
claude
```

## Routing Rules

| Complexity | Score | Model | Input cost per 1M tokens |
|------------|-------|-------|--------------------------|
| low | 0-1 | haiku | $0.80 |
| medium | 2 | sonnet | $3.00 |
| high | 3+ | opus | $15.00 |

## Customization

Edit `promptlint_rule.py` to change model mapping:

```python
MODEL_MAP = {
    "low": "claude-haiku-4-5-20251001",
    "medium": "claude-sonnet-4-6-20260320",
    "high": "claude-opus-4-6-20260320",
}
```

## Telemetry

Each scored request adds `_promptlint` metadata:

```json
{
  "complexity": "low",
  "suggested_model": "haiku",
  "words": 8,
  "action": "fix"
}
```

Use with [costlint](https://github.com/mikeshogin/costlint) for cost tracking.
