#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BINARY="$PROJECT_DIR/canopy"
GOLDEN="$PROJECT_DIR/testdata/golden/index.json"
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

echo "=== canopy E2E test ==="
echo "Using temp dir: $TMPDIR"

# Build
echo "1. Building..."
(cd "$PROJECT_DIR" && make build)

# Init
echo "2. Testing init..."
cd "$TMPDIR"
$BINARY init
[ -f .canopy/config.json ] || { echo "FAIL: config.json not created"; exit 1; }
[ -d .canopy/components ] || { echo "FAIL: components/ not created"; exit 1; }
[ -d .canopy/prompts ] || { echo "FAIL: prompts/ not created"; exit 1; }
echo "   PASS: init created .canopy/ structure"

# Import golden fixture
echo "3. Testing import..."
$BINARY import --force "$GOLDEN"
[ -f .canopy/index.json ] || { echo "FAIL: index.json not created"; exit 1; }
echo "   PASS: imported golden fixture"

# Validate
echo "4. Testing validate..."
$BINARY validate
echo "   PASS: validation passed"

# Serve and query
echo "5. Testing serve + queries..."
$BINARY serve --port 13451 &
SERVER_PID=$!
sleep 1

# Health check
HEALTH=$(curl -sf http://127.0.0.1:13451/health)
echo "$HEALTH" | grep -q '"ok"' || { echo "FAIL: health check"; kill $SERVER_PID; exit 1; }
echo "   PASS: /health"

# Context query - Customer controller
CONTEXT=$(curl -sf "http://127.0.0.1:13451/context?file=Customer/src/main/java/com/jmendoza/swa/hexagonal/customer/application/rest/controller/CustomerController.java")
echo "$CONTEXT" | grep -q '"customer-service"' || { echo "FAIL: context component"; kill $SERVER_PID; exit 1; }
echo "$CONTEXT" | grep -q '"customer-controller"' || { echo "FAIL: context archetype"; kill $SERVER_PID; exit 1; }
echo "   PASS: /context (Customer controller)"

# Components
COMPONENTS=$(curl -sf http://127.0.0.1:13451/components)
echo "$COMPONENTS" | grep -q '"customer-service"' || { echo "FAIL: components"; kill $SERVER_PID; exit 1; }
echo "$COMPONENTS" | grep -q '"order-service"' || { echo "FAIL: components"; kill $SERVER_PID; exit 1; }
echo "$COMPONENTS" | grep -q '"product-service"' || { echo "FAIL: components"; kill $SERVER_PID; exit 1; }
echo "   PASS: /components"

# Archetypes
ARCHETYPES=$(curl -sf http://127.0.0.1:13451/archetypes/controllers)
echo "$ARCHETYPES" | grep -q '"customer-controller"' || { echo "FAIL: archetypes"; kill $SERVER_PID; exit 1; }
echo "   PASS: /archetypes/controllers"

# Relationships
RELS=$(curl -sf "http://127.0.0.1:13451/relationships?symbol=customer-controller&direction=downstream")
echo "$RELS" | grep -q '"create-customer-service"' || { echo "FAIL: relationships"; kill $SERVER_PID; exit 1; }
echo "   PASS: /relationships"

# Flows
FLOWS=$(curl -sf "http://127.0.0.1:13451/flows?through=customer-controller")
echo "$FLOWS" | grep -q '"create-customer"' || { echo "FAIL: flows"; kill $SERVER_PID; exit 1; }
echo "   PASS: /flows"

# Cleanup
kill $SERVER_PID 2>/dev/null || true
wait $SERVER_PID 2>/dev/null || true

echo ""
echo "=== ALL E2E TESTS PASSED ==="
