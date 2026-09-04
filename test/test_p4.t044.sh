#!/bin/bash

# Test script for P4.T044 - Job Genealogy Navigator Design Validation
# Validates that the design deliverables meet all acceptance criteria

set -e

# Change to project directory
cd "/Users/james/git/go-redis-work-queue"

TEST_DIR="/Users/james/git/go-redis-work-queue"
DESIGN_FILE="docs/design/f017-design.md"
API_FILE="docs/api/f017-openapi.yaml"
SCHEMA_FILE="docs/schemas/f017-schema.json"

echo "P4.T044 Acceptance Test - Job Genealogy Navigator Design"
echo "========================================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

PASSED=0
FAILED=0

check_test() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ PASS${NC}: $2"
        ((PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}: $2"
        ((FAILED++))
    fi
}

# Test 1: Architecture document exists and contains required sections
echo
echo "Test 1: Architecture Document Validation"
echo "----------------------------------------"

if [ -f "$DESIGN_FILE" ]; then
    # Check for required sections
    grep -q "System Architecture" "$DESIGN_FILE"
    check_test $? "System Architecture section present"

    grep -q "API Specification" "$DESIGN_FILE"
    check_test $? "API Specification section present"

    grep -q "Data Models" "$DESIGN_FILE"
    check_test $? "Data Models section present"

    grep -q "Security Model" "$DESIGN_FILE"
    check_test $? "Security Model section present"

    grep -q "Performance Requirements" "$DESIGN_FILE"
    check_test $? "Performance Requirements section present"

    grep -q "Testing Strategy" "$DESIGN_FILE"
    check_test $? "Testing Strategy section present"

    # Check for Mermaid diagrams
    grep -q "\`\`\`mermaid" "$DESIGN_FILE"
    check_test $? "Mermaid diagrams included"

    # Check document size (should be substantial)
    DESIGN_LINES=$(wc -l < "$DESIGN_FILE")
    if [ "$DESIGN_LINES" -gt 800 ]; then
        check_test 0 "Design document has sufficient detail ($DESIGN_LINES lines)"
    else
        check_test 1 "Design document too brief ($DESIGN_LINES lines, expected >800)"
    fi
else
    check_test 1 "Architecture document exists"
fi

# Test 2: OpenAPI specification validation
echo
echo "Test 2: OpenAPI Specification Validation"
echo "----------------------------------------"

if [ -f "$API_FILE" ]; then
    # Check OpenAPI version
    grep -q "openapi: 3\." "$API_FILE"
    check_test $? "OpenAPI 3.x format used"

    # Check for key endpoints
    grep -q "/genealogy/tree" "$API_FILE"
    check_test $? "Tree endpoint defined"

    grep -q "/genealogy/blame" "$API_FILE"
    check_test $? "Blame analysis endpoint defined"

    grep -q "/genealogy/impact" "$API_FILE"
    check_test $? "Impact analysis endpoint defined"

    grep -q "/genealogy/relationships" "$API_FILE"
    check_test $? "Relationships endpoint defined"

    # Check for security definitions
    grep -q "security:" "$API_FILE"
    check_test $? "Security definitions present"

    # Check for comprehensive schemas
    grep -q "JobGenealogy" "$API_FILE"
    check_test $? "JobGenealogy schema defined"

    grep -q "JobRelationship" "$API_FILE"
    check_test $? "JobRelationship schema defined"

    # Check file size indicates comprehensive spec
    API_LINES=$(wc -l < "$API_FILE")
    if [ "$API_LINES" -gt 1000 ]; then
        check_test 0 "API specification comprehensive ($API_LINES lines)"
    else
        check_test 1 "API specification incomplete ($API_LINES lines, expected >1000)"
    fi
else
    check_test 1 "OpenAPI specification exists"
fi

# Test 3: JSON Schema validation
echo
echo "Test 3: JSON Schema Validation"
echo "------------------------------"

if [ -f "$SCHEMA_FILE" ]; then
    # Validate JSON syntax
    if python3 -m json.tool "$SCHEMA_FILE" > /dev/null 2>&1; then
        check_test 0 "JSON Schema syntax valid"
    else
        check_test 1 "JSON Schema syntax valid"
    fi

    # Check for key schema definitions
    grep -q "JobRelationship" "$SCHEMA_FILE"
    check_test $? "JobRelationship schema defined"

    grep -q "JobGenealogy" "$SCHEMA_FILE"
    check_test $? "JobGenealogy schema defined"

    grep -q "GenealogyNode" "$SCHEMA_FILE"
    check_test $? "GenealogyNode schema defined"

    grep -q "GenealogyEdge" "$SCHEMA_FILE"
    check_test $? "GenealogyEdge schema defined"

    grep -q "BlameAnalysis" "$SCHEMA_FILE"
    check_test $? "BlameAnalysis schema defined"

    grep -q "ImpactAnalysis" "$SCHEMA_FILE"
    check_test $? "ImpactAnalysis schema defined"

    # Check for relationship types
    grep -q "retry" "$SCHEMA_FILE"
    check_test $? "Retry relationship type defined"

    grep -q "spawn" "$SCHEMA_FILE"
    check_test $? "Spawn relationship type defined"

    # Check schema completeness
    SCHEMA_LINES=$(wc -l < "$SCHEMA_FILE")
    if [ "$SCHEMA_LINES" -gt 500 ]; then
        check_test 0 "Schema definitions comprehensive ($SCHEMA_LINES lines)"
    else
        check_test 1 "Schema definitions incomplete ($SCHEMA_LINES lines, expected >500)"
    fi
else
    check_test 1 "JSON Schema file exists"
fi

# Test 4: Integration points documentation
echo
echo "Test 4: Integration Points Validation"
echo "------------------------------------"

if [ -f "$DESIGN_FILE" ]; then
    # Check for Redis integration
    grep -qi "redis" "$DESIGN_FILE"
    check_test $? "Redis integration documented"

    # Check for TUI integration
    grep -qi "TUI" "$DESIGN_FILE"
    check_test $? "TUI integration documented"

    # Check for performance considerations
    grep -qi "performance" "$DESIGN_FILE"
    check_test $? "Performance considerations documented"

    # Check for graph algorithms
    grep -qi "graph" "$DESIGN_FILE"
    check_test $? "Graph algorithms documented"

    # Check for ASCII art specifications
    grep -qi "ascii" "$DESIGN_FILE"
    check_test $? "ASCII art specifications documented"
fi

# Test 5: Acceptance criteria validation
echo
echo "Test 5: Acceptance Criteria Coverage"
echo "-----------------------------------"

if [ -f "$DESIGN_FILE" ]; then
    # Complete genealogy capture
    grep -qi "genealogy.*captur" "$DESIGN_FILE"
    check_test $? "Genealogy capture mechanism designed"

    # Tree rendering for 100+ nodes
    grep -qi "100.*node\|node.*100" "$DESIGN_FILE"
    check_test $? "Large tree rendering addressed"

    # Navigation performance
    grep -qi "50ms\|navigation.*performance" "$DESIGN_FILE"
    check_test $? "Navigation performance requirements specified"

    # TTL pruning
    grep -qi "TTL\|prune\|cleanup" "$DESIGN_FILE"
    check_test $? "Data pruning mechanism designed"

    # Relationship schema
    grep -qi "relationship.*schema\|schema.*relationship" "$DESIGN_FILE"
    check_test $? "Relationship schema designed"
fi

# Test 6: Design quality validation
echo
echo "Test 6: Design Quality Validation"
echo "--------------------------------"

TOTAL_FILES=0
EXISTING_FILES=0

for file in "$DESIGN_FILE" "$API_FILE" "$SCHEMA_FILE"; do
    ((TOTAL_FILES++))
    if [ -f "$file" ]; then
        ((EXISTING_FILES++))
    fi
done

if [ "$EXISTING_FILES" -eq "$TOTAL_FILES" ]; then
    check_test 0 "All required deliverables created ($EXISTING_FILES/$TOTAL_FILES)"
else
    check_test 1 "All required deliverables created ($EXISTING_FILES/$TOTAL_FILES)"
fi

# Check for consistent terminology across files
if [ -f "$DESIGN_FILE" ] && [ -f "$API_FILE" ] && [ -f "$SCHEMA_FILE" ]; then
    # Sample consistency check - both should mention genealogy
    DESIGN_GENEALOGY=$(grep -c -i "genealogy" "$DESIGN_FILE" || echo 0)
    API_GENEALOGY=$(grep -c -i "genealogy" "$API_FILE" || echo 0)
    SCHEMA_GENEALOGY=$(grep -c -i "genealogy" "$SCHEMA_FILE" || echo 0)

    if [ "$DESIGN_GENEALOGY" -gt 10 ] && [ "$API_GENEALOGY" -gt 5 ] && [ "$SCHEMA_GENEALOGY" -gt 5 ]; then
        check_test 0 "Consistent terminology across documents"
    else
        check_test 1 "Consistent terminology across documents (D:$DESIGN_GENEALOGY A:$API_GENEALOGY S:$SCHEMA_GENEALOGY)"
    fi
fi

# Summary
echo
echo "Test Summary"
echo "============"
echo -e "Tests Passed: ${GREEN}$PASSED${NC}"
echo -e "Tests Failed: ${RED}$FAILED${NC}"
echo -e "Total Tests: $((PASSED + FAILED))"

if [ "$FAILED" -eq 0 ]; then
    echo -e "${GREEN}✓ ALL TESTS PASSED - Design meets acceptance criteria${NC}"
    exit 0
else
    echo -e "${RED}✗ TESTS FAILED - Design needs revision${NC}"
    exit 1
fi