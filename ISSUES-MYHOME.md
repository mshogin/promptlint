# promptlint: myhome Integration Issues

Issues for integrating promptlint with myhome orchestrator.
These will become GitHub issues when mshogin/promptlint repo is created.

---

## Issue 1: CLI exit code and machine-readable output

Status: todo

### Problem
Current CLI outputs JSON but doesn't use exit codes for routing decisions.

### Tasks
- [ ] Exit code based on suggested model: 0=haiku, 1=sonnet, 2=opus
- [ ] `--format` flag: json (default), text, brief
- [ ] `--output-model` flag: print only model name (for shell scripting)
- [ ] Stable JSON schema for pipeline consumption

### Usage
```bash
# Shell integration
MODEL=$(echo "Fix bug" | promptlint analyze --output-model)
# Returns: "haiku"
```

---

## Issue 2: HTTP server - batch analysis endpoint

Status: todo

### Problem
Current `/analyze` endpoint handles one prompt at a time. myhome may need batch scoring.

### Tasks
- [ ] `POST /analyze/batch` - array of prompts
- [ ] Response includes routing decision per prompt
- [ ] Performance: <10ms per prompt (no LLM, pure regex/metrics)

---

## Issue 3: myhome config integration

Status: todo

### Problem
promptlint needs to know about available model tiers and their cost/capability.

### Tasks
- [ ] Config file: model tiers with names, cost weights, capability scores
- [ ] `--config` flag to load custom routing rules
- [ ] Default config embedded (haiku/sonnet/opus)

### Config format
```yaml
models:
  - name: haiku
    tier: low
    cost_weight: 1
    max_complexity: 30
  - name: sonnet
    tier: standard
    cost_weight: 10
    max_complexity: 70
  - name: opus
    tier: high
    cost_weight: 30
    max_complexity: 100
```

---

## Issue 4: Telemetry export for myhome dashboard

Status: todo

### Problem
Telemetry is JSONL file-based. myhome may want structured metrics.

### Tasks
- [ ] `GET /stats` endpoint with routing accuracy
- [ ] Export: total requests, per-model distribution, cost savings estimate
- [ ] Prometheus-compatible metrics endpoint (optional)
