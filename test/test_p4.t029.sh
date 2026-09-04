#!/bin/bash

# Test script for P4.T029 - Design Anomaly Radar SLO Budget
# This script validates that all design deliverables have been created
# and meet the acceptance criteria specified in the task.

set -euo pipefail

echo "=== P4.T029 Anomaly Radar SLO Budget Design Validation ==="
echo "Task: Design Anomaly Radar SLO Budget"
echo "Feature: F010"
echo

# Test 1: Verify design document exists and contains required sections
echo "✓ Testing design document..."
DESIGN_FILE="docs/design/f010-design.md"
if [[ ! -f "$DESIGN_FILE" ]]; then
    echo "❌ Design document not found: $DESIGN_FILE"
    exit 1
fi

# Check for required sections based on task specification
REQUIRED_SECTIONS=(
    "Executive Summary"
    "System Architecture"
    "TUI Design"
    "Data Models and Schema Design"
    "Performance Requirements"
    "Security Model"
    "Testing Strategy"
    "Implementation Roadmap"
)

for section in "${REQUIRED_SECTIONS[@]}"; do
    if ! grep -q "## $section" "$DESIGN_FILE" && ! grep -q "### $section" "$DESIGN_FILE"; then
        echo "❌ Missing required section: $section"
        exit 1
    fi
done

# Check for Mermaid diagrams (architecture requirement)
DIAGRAM_COUNT=$(grep -c '```mermaid' "$DESIGN_FILE" || echo "0")
if [[ $DIAGRAM_COUNT -lt 3 ]]; then
    echo "❌ Insufficient Mermaid diagrams: $DIAGRAM_COUNT - expected at least 3"
    exit 1
fi

# Check document length (should be substantial design)
LINE_COUNT=$(wc -l < "$DESIGN_FILE")
if [[ $LINE_COUNT -lt 800 ]]; then
    echo "❌ Design document too short: $LINE_COUNT lines - expected 800-1200"
    exit 1
fi

echo "  ✅ Design document structure validated ($LINE_COUNT lines, $DIAGRAM_COUNT diagrams)"

# Test 2: Verify OpenAPI specification exists and is valid
echo "✓ Testing OpenAPI specification..."
API_FILE="docs/api/f010-openapi.yaml"
if [[ ! -f "$API_FILE" ]]; then
    echo "❌ OpenAPI specification not found: $API_FILE"
    exit 1
fi

# Check for OpenAPI 3.0 format
if ! grep -q "openapi: 3.0" "$API_FILE"; then
    echo "❌ Not a valid OpenAPI 3.0 specification"
    exit 1
fi

# Check for required API sections
if ! grep -q "info:" "$API_FILE"; then
    echo "❌ Missing info section in API spec"
    exit 1
fi
if ! grep -q "paths:" "$API_FILE"; then
    echo "❌ Missing paths section in API spec"
    exit 1
fi

# Check for key SLO monitoring endpoints
REQUIRED_ENDPOINTS=(
    "/slo/config"
    "/slo/metrics"
    "/slo/budget"
    "/slo/alerts"
    "/slo/thresholds"
    "/slo/anomalies"
    "/slo/health"
)

for endpoint in "${REQUIRED_ENDPOINTS[@]}"; do
    if ! grep -q "$endpoint" "$API_FILE"; then
        echo "❌ Missing required endpoint: $endpoint"
        exit 1
    fi
done

# Check for authentication and security
if ! grep -q "security:" "$API_FILE"; then
    echo "❌ Missing security configuration in API spec"
    exit 1
fi

echo "  ✅ OpenAPI specification validated"

# Test 3: Verify JSON Schema definitions exist and are valid
echo "✓ Testing JSON Schema definitions..."
SCHEMA_FILE="docs/schemas/f010-schema.json"
if [[ ! -f "$SCHEMA_FILE" ]]; then
    echo "❌ JSON Schema file not found: $SCHEMA_FILE"
    exit 1
fi

# Check for valid JSON format
if ! python3 -m json.tool "$SCHEMA_FILE" > /dev/null 2>&1; then
    echo "❌ Invalid JSON format in schema file"
    exit 1
fi

# Check for required schema definitions for SLO monitoring
REQUIRED_SCHEMAS=(
    "SLOConfig"
    "SLOMetrics"
    "ErrorBudget"
    "BurnRate"
    "SLOAlert"
    "ThresholdConfig"
    "Anomaly"
    "HealthStatus"
)

for schema in "${REQUIRED_SCHEMAS[@]}"; do
    if ! grep -q "\"$schema\":" "$SCHEMA_FILE"; then
        echo "❌ Missing required schema definition: $schema"
        exit 1
    fi
done

# Check for JSON Schema draft-07 format
if ! grep -q "\"http://json-schema.org/draft-07/schema#\"" "$SCHEMA_FILE"; then
    echo "❌ Not a valid JSON Schema draft-07 format"
    exit 1
fi

echo "  ✅ JSON Schema definitions validated"

# Test 4: Verify SLO-specific technical approach
echo "✓ Testing SLO technical approach..."

# Check for key SLO concepts mentioned in design
SLO_CONCEPTS=(
    "burn rate"
    "error budget"
    "SLO"
    "SLI"
    "threshold"
    "rolling window"
    "percentile"
    "anomaly detection"
)

for concept in "${SLO_CONCEPTS[@]}"; do
    if ! grep -qi "$concept" "$DESIGN_FILE"; then
        echo "❌ Missing SLO concept: $concept"
        exit 1
    fi
done

# Check for performance requirements (lightweight footprint)
if ! grep -qi "cpu\|memory\|performance\|lightweight" "$DESIGN_FILE"; then
    echo "❌ Missing performance/resource requirements"
    exit 1
fi

# Check for TUI integration
if ! grep -qi "tui\|widget\|display\|color" "$DESIGN_FILE"; then
    echo "❌ Missing TUI integration details"
    exit 1
fi

echo "  ✅ SLO technical approach validated"

# Test 5: Verify acceptance criteria alignment
echo "✓ Testing acceptance criteria alignment..."

# Acceptance criteria from task:
# - Backlog growth, failure rate, and p95 displayed with thresholds
if ! grep -qi "backlog.*growth\|failure.*rate\|p95.*threshold" "$DESIGN_FILE"; then
    echo "❌ Missing backlog growth, failure rate, or P95 threshold requirements"
    exit 1
fi

# - SLO config and budget burn shown; alert when burning too fast
if ! grep -qi "slo.*config\|budget.*burn\|alert.*burn" "$DESIGN_FILE"; then
    echo "❌ Missing SLO config, budget burn, or burn rate alert requirements"
    exit 1
fi

# - Lightweight CPU/memory footprint
if ! grep -qi "lightweight\|cpu.*memory\|footprint\|resource" "$DESIGN_FILE"; then
    echo "❌ Missing lightweight resource footprint requirements"
    exit 1
fi

echo "  ✅ Acceptance criteria alignment validated"

# Test 6: Verify user story implementation
echo "✓ Testing user story implementation..."

# User story: "I can see whether we're inside SLO and how fast we're burning budget"
if ! grep -qi "inside.*slo\|slo.*status\|budget.*burn.*rate" "$DESIGN_FILE"; then
    echo "❌ User story not addressed: SLO status and burn rate visibility"
    exit 1
fi

# Technical approach mentions from task spec
if ! grep -qi "rolling.*rates.*percentiles" "$DESIGN_FILE"; then
    echo "❌ Missing rolling rates and percentiles implementation"
    exit 1
fi

if ! grep -qi "configurable.*slo.*target.*window" "$DESIGN_FILE"; then
    echo "❌ Missing configurable SLO target and window"
    exit 1
fi

echo "  ✅ User story implementation validated"

# Test 7: Verify security and testing strategy
echo "✓ Testing security and testing strategy..."

# Check for security considerations
if ! grep -qi "security\|authentication\|authorization\|rbac" "$DESIGN_FILE"; then
    echo "❌ Missing security considerations"
    exit 1
fi

# Check for testing strategy
if ! grep -qi "test\|testing\|validation\|unit.*test" "$DESIGN_FILE"; then
    echo "❌ Missing testing strategy"
    exit 1
fi

echo "  ✅ Security and testing strategy validated"

# Test 8: Check file sizes and completeness
echo "✓ Testing file completeness..."

API_LINE_COUNT=$(wc -l < "$API_FILE")
SCHEMA_LINE_COUNT=$(wc -l < "$SCHEMA_FILE")

# Verify substantial API specification
if [[ $API_LINE_COUNT -lt 500 ]]; then
    echo "❌ API specification too short: $API_LINE_COUNT lines - expected >500"
    exit 1
fi

# Verify comprehensive schema definitions
if [[ $SCHEMA_LINE_COUNT -lt 200 ]]; then
    echo "❌ Schema definitions too short: $SCHEMA_LINE_COUNT lines - expected >200"
    exit 1
fi

echo "  ✅ File completeness validated"

# Summary
echo
echo "=== Test Results Summary ==="
echo "✅ All acceptance criteria passed"
echo "✅ Design document: $LINE_COUNT lines with $DIAGRAM_COUNT Mermaid diagrams"
echo "✅ API specification: $API_LINE_COUNT lines with $(grep -c 'operationId:' "$API_FILE") endpoints"
echo "✅ Schema definitions: $SCHEMA_LINE_COUNT lines with $(grep -c '\"[A-Z][a-zA-Z]*\":' "$SCHEMA_FILE") schemas"
echo
echo "🎯 P4.T029 Anomaly Radar SLO Budget design is complete and valid!"
echo
echo "Key Features Validated:"
echo "  • SLO target configuration and monitoring"
echo "  • Error budget tracking with burn rate alerts"
echo "  • Real-time metrics with rolling windows"
echo "  • Threshold-based colorization for TUI"
echo "  • Anomaly detection capabilities"
echo "  • Comprehensive API for SRE operations"
echo "  • Lightweight performance requirements"
echo
echo "Deliverables created:"
echo "  - $DESIGN_FILE (Architecture + Mermaid diagrams)"
echo "  - $API_FILE (OpenAPI 3.0 specification)"
echo "  - $SCHEMA_FILE (JSON Schema definitions)"
echo
echo "Ready for architect review and approval."