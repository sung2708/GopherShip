#!/bin/bash
# scripts/test-automation.sh
# Verifies GopherShip gs-ctl automation compliance (AC5)

set -e

# Build the CLI
echo "Building gs-ctl..."
go build -o gs-ctl ./cmd/gs-ctl

# 1. Verify JSON output
echo "Checking JSON output..."
./gs-ctl --mock-zone 0 --output json status > status.json
grep "GREEN" status.json
echo "✅ JSON check passed"

# 2. Verify YAML output
echo "Checking YAML output..."
./gs-ctl --mock-zone 0 --output yaml status > status.yaml
grep "Somatic Zone: GREEN" status.yaml
echo "✅ YAML check passed"

# 3. Verify exit codes (simulated transitions - AC5)
echo "Verifying exit code transitions..."
# We use set +e because we expect non-zero exit codes for Yellow/Red
set +e

# Green Zone (Expect 0)
./gs-ctl --mock-zone 0 status > /dev/null
EXIT_CODE=$?
if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ Green zone (Exit 0) passed"
else
    echo "❌ Green zone failed (Got: $EXIT_CODE)"
    exit 1
fi

# Yellow Zone (Expect 1)
./gs-ctl --mock-zone 1 status > /dev/null
EXIT_CODE=$?
if [ $EXIT_CODE -eq 1 ]; then
    echo "✅ Yellow zone (Exit 1) passed"
else
    echo "❌ Yellow zone failed (Got: $EXIT_CODE)"
    exit 1
fi

# Red Zone (Expect 2)
./gs-ctl --mock-zone 2 status > /dev/null
EXIT_CODE=$?
if [ $EXIT_CODE -eq 2 ]; then
    echo "✅ Red zone (Exit 2) passed"
else
    echo "❌ Red zone failed (Got: $EXIT_CODE)"
    exit 1
fi

# 4. Verify precise error codes
echo "Verifying precise error codes..."
# Connection error (trying a port that is likely closed)
# We expect exit 129 for connection failures.
./gs-ctl status --addr localhost:1 > /dev/null 2>&1
EXIT_CODE=$?
if [ $EXIT_CODE -eq 129 ]; then
    echo "✅ Precise error code check passed (Connection Error: 129)"
else
    echo "❌ Unexpected error code: $EXIT_CODE (Expected 129)"
    exit 1
fi

set -e

echo "🚀 ALL AUTOMATION CHECKS PASSED"
rm status.json status.yaml gs-ctl
