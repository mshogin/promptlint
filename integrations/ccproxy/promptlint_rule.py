"""
promptlint scoring rule for ccproxy.

Intercepts Claude Code requests, scores prompt complexity via promptlint HTTP API,
and routes to the appropriate model tier.

Setup:
    1. Start promptlint server: promptlint serve 8090
    2. Copy this file to your ccproxy plugins directory
    3. Add to ccproxy.yaml rules section
"""

import json
import urllib.request
import urllib.error

PROMPTLINT_URL = "http://localhost:8090/analyze"
TIMEOUT_SECONDS = 2

# Model mapping: complexity -> model name
MODEL_MAP = {
    "low": "claude-haiku-4-5-20251001",
    "medium": "claude-sonnet-4-6-20260320",
    "high": "claude-opus-4-6-20260320",
}


def extract_prompt_text(request_body: dict) -> str:
    """Extract user message text from Anthropic API request."""
    messages = request_body.get("messages", [])
    if not messages:
        return ""

    # Get the last user message
    for msg in reversed(messages):
        if msg.get("role") == "user":
            content = msg.get("content", "")
            if isinstance(content, str):
                return content
            if isinstance(content, list):
                # Handle content blocks
                texts = []
                for block in content:
                    if isinstance(block, dict) and block.get("type") == "text":
                        texts.append(block.get("text", ""))
                return " ".join(texts)
    return ""


def score_prompt(text: str) -> dict:
    """Call promptlint HTTP API to score prompt complexity."""
    try:
        req = urllib.request.Request(
            PROMPTLINT_URL,
            data=text.encode("utf-8"),
            headers={"Content-Type": "text/plain"},
            method="POST",
        )
        with urllib.request.urlopen(req, timeout=TIMEOUT_SECONDS) as resp:
            return json.loads(resp.read())
    except (urllib.error.URLError, TimeoutError, json.JSONDecodeError):
        return None


def route_request(request_body: dict) -> dict:
    """
    Main entry point for ccproxy hook.

    Returns modified request with routed model, or original if scoring fails.
    """
    prompt_text = extract_prompt_text(request_body)

    # Skip very short prompts (not worth scoring)
    if len(prompt_text) < 20:
        return request_body

    result = score_prompt(prompt_text)
    if result is None:
        # Scoring failed, pass through unchanged
        return request_body

    complexity = result.get("complexity", "medium")
    suggested_model = MODEL_MAP.get(complexity, request_body.get("model", ""))

    # Override model in request
    request_body["model"] = suggested_model

    # Add metadata header for telemetry
    request_body["_promptlint"] = {
        "complexity": complexity,
        "suggested_model": result.get("suggested_model", ""),
        "words": result.get("words", 0),
        "action": result.get("action", ""),
    }

    return request_body
