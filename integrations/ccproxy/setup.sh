#!/bin/bash
# Setup promptlint + ccproxy integration
# Usage: ./setup.sh

set -e

echo "=== promptlint + ccproxy setup ==="

# Check dependencies
if ! command -v promptlint &> /dev/null; then
    echo "Installing promptlint..."
    go install github.com/mikeshogin/promptlint/cmd/promptlint@latest
fi

if ! command -v pip &> /dev/null; then
    echo "Error: pip required for ccproxy"
    exit 1
fi

# Install ccproxy
echo "Installing ccproxy..."
pip install ccproxy 2>/dev/null || pip3 install ccproxy

# Create config directory
CCPROXY_DIR="${HOME}/.ccproxy"
mkdir -p "${CCPROXY_DIR}"

# Copy rule file
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cp "${SCRIPT_DIR}/promptlint_rule.py" "${CCPROXY_DIR}/"

# Generate ccproxy config if not exists
if [ ! -f "${CCPROXY_DIR}/ccproxy.yaml" ]; then
    cat > "${CCPROXY_DIR}/ccproxy.yaml" << 'YAML'
# ccproxy configuration with promptlint routing
# See: https://github.com/mikeshogin/promptlint

port: 3456

rules:
  - name: "promptlint_scoring"
    module: "promptlint_rule"
    function: "route_request"
    endpoint: "http://localhost:8090/analyze"
YAML
    echo "Created ${CCPROXY_DIR}/ccproxy.yaml"
fi

echo ""
echo "=== Setup complete ==="
echo ""
echo "To start routing:"
echo "  1. Start promptlint:  promptlint serve 8090"
echo "  2. Start ccproxy:     ccproxy start"
echo "  3. Configure Claude:  export ANTHROPIC_BASE_URL=http://localhost:3456"
echo "  4. Run Claude Code:   claude"
echo ""
echo "Requests will be automatically scored and routed to the cheapest sufficient model."
