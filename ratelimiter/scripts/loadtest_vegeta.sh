#!/usr/bin/env bash
# Phase 5 load test using vegeta (https://github.com/tsenart/vegeta).
# Install: go install github.com/tsenart/vegeta@latest
#
# Usage:
#   ./scripts/loadtest_vegeta.sh              # defaults: 200 rps, 30s
#   RATE=500 DURATION=60s ./scripts/loadtest_vegeta.sh
set -euo pipefail

HOST="${HOST:-http://localhost:8080}"
RATE="${RATE:-200}"       # requests/sec
DURATION="${DURATION:-30s}"
API_KEY="${API_KEY:-loadtest-client}"

echo "Target: ${HOST}/check | Rate: ${RATE}/s | Duration: ${DURATION} | Client: ${API_KEY}"

echo "GET ${HOST}/check
X-API-KEY: ${API_KEY}" | vegeta attack \
    -rate="${RATE}" \
    -duration="${DURATION}" \
  | tee /tmp/vegeta_results.bin \
  | vegeta report

echo
echo "--- Latency histogram ---"
vegeta report -type='hist[0,10ms,25ms,50ms,100ms,250ms,500ms,1s]' < /tmp/vegeta_results.bin

echo
echo "--- Status code breakdown ---"
vegeta report -type=json < /tmp/vegeta_results.bin | python3 -c "
import json, sys
data = json.load(sys.stdin)
print(data.get('status_codes', {}))
"
