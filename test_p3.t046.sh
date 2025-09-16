#!/bin/bash

# Test script for P3.T046 - Design Long Term Archives
# This script validates that all design deliverables have been created
# and meet the acceptance criteria specified in the task.

set -euo pipefail

echo "=== P3.T046 Long Term Archives Design Validation ==="
echo "Task: Design Long Term Archives"
echo "Feature: F018"
echo

# Test 1: Verify design document exists and contains required sections
echo "✓ Testing design document..."
DESIGN_FILE="docs/design/f018-design.md"
if [[ ! -f "$DESIGN_FILE" ]]; then
    echo "❌ Design document not found: $DESIGN_FILE"
    exit 1
fi

# Check for required sections
REQUIRED_SECTIONS=(
    "Executive Summary"
    "System Architecture"
    "TUI Design"
    "Data Models and Schema Design"
    "Performance Requirements"
    "Security Model"
)

for section in "${REQUIRED_SECTIONS[@]}"; do
    if ! grep -q "## $section" "$DESIGN_FILE" && ! grep -q "### $section" "$DESIGN_FILE"; then
        echo "❌ Missing required section: $section"
        exit 1
    fi
done

# Check for Mermaid diagrams
DIAGRAM_COUNT=$(grep -c '```mermaid' "$DESIGN_FILE" || echo "0")
if [[ $DIAGRAM_COUNT -lt 3 ]]; then
    echo "❌ Insufficient Mermaid diagrams: $DIAGRAM_COUNT - expected at least 3"
    exit 1
fi

# Check document length
LINE_COUNT=$(wc -l < "$DESIGN_FILE")
if [[ $LINE_COUNT -lt 800 ]]; then
    echo "❌ Design document too short: $LINE_COUNT lines - expected 800-1200"
    exit 1
fi

echo "  ✅ Design document structure validated ($LINE_COUNT lines)"

# Test 2: Verify OpenAPI specification exists and is valid
echo "✓ Testing OpenAPI specification..."
API_FILE="docs/api/f018-openapi.yaml"
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

# Check for key endpoints
REQUIRED_ENDPOINTS=(
    "/archives/config"
    "/archives/query"
    "/archives/schema"
    "/archives/retention"
    "/archives/export/status"
    "/archives/gdpr/deletion-requests"
)

for endpoint in "${REQUIRED_ENDPOINTS[@]}"; do
    if ! grep -q "$endpoint" "$API_FILE"; then
        echo "❌ Missing required endpoint: $endpoint"
        exit 1
    fi
done

echo "  ✅ OpenAPI specification validated"

# Test 3: Verify JSON Schema definitions exist and are valid
echo "✓ Testing JSON Schema definitions..."
SCHEMA_FILE="docs/schemas/f018-schema.json"
if [[ ! -f "$SCHEMA_FILE" ]]; then
    echo "❌ JSON Schema file not found: $SCHEMA_FILE"
    exit 1
fi

# Check for valid JSON format
if ! python3 -m json.tool "$SCHEMA_FILE" > /dev/null 2>&1; then
    echo "❌ Invalid JSON format in schema file"
    exit 1
fi

# Check for required schema definitions
REQUIRED_SCHEMAS=(
    "ArchiveConfig"
    "StorageBackend"
    "RetentionPolicy"
    "JobRecord"
    "QueryRequest"
    "QueryResponse"
)

for schema in "${REQUIRED_SCHEMAS[@]}"; do
    if ! grep -q "\"$schema\":" "$SCHEMA_FILE"; then
        echo "❌ Missing required schema definition: $schema"
        exit 1
    fi
done

echo "  ✅ JSON Schema definitions validated"

# Test 4: Verify technical approach matches requirements
echo "✓ Testing technical approach alignment..."

# Check for key technical components
REQUIRED_COMPONENTS=(
    "ClickHouse"
    "S3"
    "Parquet"
    "retention"
    "GDPR"
    "schema evolution"
    "export"
    "batch"
)

for component in "${REQUIRED_COMPONENTS[@]}"; do
    if ! grep -qi "$component" "$DESIGN_FILE"; then
        echo "❌ Missing technical component: $component"
        exit 1
    fi
done

echo "  ✅ Technical approach alignment validated"

# Summary
echo
echo "=== Test Results Summary ==="
echo "✅ All acceptance criteria passed"
echo "✅ Design document: $LINE_COUNT lines"
echo "✅ API specification: $(wc -l < "$API_FILE") lines"
echo "✅ Schema definitions: $(wc -l < "$SCHEMA_FILE") lines"
echo "✅ Mermaid diagrams: $DIAGRAM_COUNT found"
echo
echo "🎉 P3.T046 Long Term Archives design is complete and valid!"
echo
echo "Deliverables created:"
echo "  - $DESIGN_FILE"
echo "  - $API_FILE"
echo "  - $SCHEMA_FILE"
echo
echo "Ready for architect review and approval."