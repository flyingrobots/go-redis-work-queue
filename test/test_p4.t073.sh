#!/bin/bash

# Test script for P4.T073 - Patterned Load Generator Design Validation
# Validates that the design deliverables meet all acceptance criteria

set -e

# Change to project directory
cd "/Users/james/git/go-redis-work-queue"

echo "P4.T073 Acceptance Test - Patterned Load Generator Design"
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

DESIGN_FILE="docs/design/f030-design.md"
API_FILE="docs/api/f030-openapi.yaml"
SCHEMA_FILE="docs/schemas/f030-schema.json"

# Test 1: Architecture document validation
echo
echo "Test 1: Architecture Document Validation"
echo "----------------------------------------"

if [ -f "$DESIGN_FILE" ]; then
    check_test 0 "Architecture document exists"

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

    # Check for pattern-specific content
    grep -qi "pattern.*generator\|pattern.*manager" "$DESIGN_FILE"
    check_test $? "Pattern generation system documented"

    grep -qi "sine.*burst.*ramp\|pattern.*type" "$DESIGN_FILE"
    check_test $? "Pattern types documented"

    grep -qi "guardrail\|safety.*limit\|rate.*limit" "$DESIGN_FILE"
    check_test $? "Guardrail system documented"

    grep -qi "visualization\|chart\|real.*time" "$DESIGN_FILE"
    check_test $? "Visualization system documented"

    # Check document size
    DESIGN_LINES=$(wc -l < "$DESIGN_FILE")
    if [ "$DESIGN_LINES" -gt 800 ]; then
        check_test 0 "Design document comprehensive ($DESIGN_LINES lines)"
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
    check_test 0 "OpenAPI specification exists"

    # Check OpenAPI version
    grep -q "openapi: 3\." "$API_FILE"
    check_test $? "OpenAPI 3.x format used"

    # Check for pattern execution endpoints
    grep -q "/load-test/patterns/start" "$API_FILE"
    check_test $? "Pattern start endpoint defined"

    grep -q "/load-test/patterns/{testId}/stop" "$API_FILE"
    check_test $? "Pattern stop endpoint defined"

    grep -q "/load-test/patterns/{testId}/status" "$API_FILE"
    check_test $? "Pattern status endpoint defined"

    grep -q "/load-test/patterns/{testId}/metrics" "$API_FILE"
    check_test $? "Pattern metrics endpoint defined"

    # Check for profile management endpoints
    grep -q "/load-test/profiles" "$API_FILE"
    check_test $? "Profile management endpoints defined"

    grep -q "/load-test/profiles/{profileId}/run" "$API_FILE"
    check_test $? "Profile execution endpoint defined"

    # Check for guardrail endpoints
    grep -q "/load-test/guardrails" "$API_FILE"
    check_test $? "Guardrail endpoints defined"

    grep -q "/load-test/guardrails/validate" "$API_FILE"
    check_test $? "Guardrail validation endpoint defined"

    # Check for streaming endpoint
    grep -q "/load-test/stream" "$API_FILE"
    check_test $? "WebSocket streaming endpoint defined"

    # Check for comprehensive schemas
    grep -q "PatternConfig" "$API_FILE"
    check_test $? "PatternConfig schema defined"

    grep -q "GuardrailConfig" "$API_FILE"
    check_test $? "GuardrailConfig schema defined"

    grep -q "LoadTestProfile" "$API_FILE"
    check_test $? "LoadTestProfile schema defined"

    # Check file size for comprehensiveness
    API_LINES=$(wc -l < "$API_FILE")
    if [ "$API_LINES" -gt 1500 ]; then
        check_test 0 "API specification comprehensive ($API_LINES lines)"
    else
        check_test 1 "API specification incomplete ($API_LINES lines, expected >1500)"
    fi
else
    check_test 1 "OpenAPI specification exists"
fi

# Test 3: JSON Schema validation
echo
echo "Test 3: JSON Schema Validation"
echo "------------------------------"

if [ -f "$SCHEMA_FILE" ]; then
    check_test 0 "JSON Schema file exists"

    # Validate JSON syntax
    if python3 -m json.tool "$SCHEMA_FILE" > /dev/null 2>&1; then
        check_test 0 "JSON Schema syntax valid"
    else
        check_test 1 "JSON Schema syntax valid"
    fi

    # Check for core pattern schemas
    grep -q "PatternConfig" "$SCHEMA_FILE"
    check_test $? "PatternConfig schema defined"

    grep -q "PatternType" "$SCHEMA_FILE"
    check_test $? "PatternType enum defined"

    grep -q "GuardrailConfig" "$SCHEMA_FILE"
    check_test $? "GuardrailConfig schema defined"

    grep -q "JobConfiguration" "$SCHEMA_FILE"
    check_test $? "JobConfiguration schema defined"

    # Check for specific pattern types
    grep -q "sine" "$SCHEMA_FILE"
    check_test $? "Sine pattern type defined"

    grep -q "burst" "$SCHEMA_FILE"
    check_test $? "Burst pattern type defined"

    grep -q "ramp" "$SCHEMA_FILE"
    check_test $? "Ramp pattern type defined"

    # Check for profile schemas
    grep -q "LoadTestProfile" "$SCHEMA_FILE"
    check_test $? "LoadTestProfile schema defined"

    grep -q "ProfileMetadata" "$SCHEMA_FILE"
    check_test $? "ProfileMetadata schema defined"

    # Check for execution schemas
    grep -q "PatternTestExecution" "$SCHEMA_FILE"
    check_test $? "PatternTestExecution schema defined"

    grep -q "TestProgress" "$SCHEMA_FILE"
    check_test $? "TestProgress schema defined"

    grep -q "GuardrailViolation" "$SCHEMA_FILE"
    check_test $? "GuardrailViolation schema defined"

    # Check schema completeness
    SCHEMA_LINES=$(wc -l < "$SCHEMA_FILE")
    if [ "$SCHEMA_LINES" -gt 1500 ]; then
        check_test 0 "Schema definitions comprehensive ($SCHEMA_LINES lines)"
    else
        check_test 1 "Schema definitions incomplete ($SCHEMA_LINES lines, expected >1500)"
    fi
else
    check_test 1 "JSON Schema file exists"
fi

# Test 4: Acceptance criteria validation
echo
echo "Test 4: Acceptance Criteria Coverage"
echo "-----------------------------------"

if [ -f "$DESIGN_FILE" ]; then
    # Sine, burst, ramp patterns; cancel/stop supported
    grep -qi "sine.*burst.*ramp\|pattern.*type.*sine" "$DESIGN_FILE"
    check_test $? "Sine, burst, ramp patterns designed"

    grep -qi "cancel.*stop\|stop.*pattern\|terminate.*test" "$DESIGN_FILE"
    check_test $? "Pattern cancellation/stop support designed"

    # Guardrails prevent runaway load
    grep -qi "guardrail.*prevent\|safety.*limit\|runaway.*load" "$DESIGN_FILE"
    check_test $? "Guardrail safety mechanisms designed"

    grep -qi "emergency.*stop\|circuit.*breaker\|rate.*limit" "$DESIGN_FILE"
    check_test $? "Emergency stop mechanisms designed"

    # Saved profiles can be reloaded
    grep -qi "profile.*save\|profile.*load\|profile.*persist" "$DESIGN_FILE"
    check_test $? "Profile persistence designed"

    grep -qi "reload.*profile\|profile.*management" "$DESIGN_FILE"
    check_test $? "Profile reloading functionality designed"

    # Implementation requirements from task
    grep -qi "implement.*sine.*burst.*ramp" "$DESIGN_FILE"
    check_test $? "Implementation approach for patterns documented"

    grep -qi "controls.*guardrail\|duration.*amplitude" "$DESIGN_FILE"
    check_test $? "Pattern controls and guardrails documented"
fi

# Test 5: Technical approach validation
echo
echo "Test 5: Technical Approach Validation"
echo "------------------------------------"

if [ -f "$DESIGN_FILE" ]; then
    # Pattern generators
    grep -qi "pattern.*generator\|sine.*generator\|burst.*generator" "$DESIGN_FILE"
    check_test $? "Pattern generators designed"

    # Controls for duration/amplitude
    grep -qi "duration.*control\|amplitude.*control\|rate.*control" "$DESIGN_FILE"
    check_test $? "Pattern controls designed"

    # Guardrails (max rate/total)
    grep -qi "max.*rate\|max.*total\|rate.*limit" "$DESIGN_FILE"
    check_test $? "Rate and total limits designed"

    # Overlay target vs actual
    grep -qi "target.*actual\|overlay.*chart\|visualization" "$DESIGN_FILE"
    check_test $? "Target vs actual visualization designed"

    # Profile persistence
    grep -qi "profile.*persistence\|save.*profile\|load.*profile" "$DESIGN_FILE"
    check_test $? "Profile persistence system designed"

    # Live charts and visualization
    grep -qi "live.*chart\|real.*time.*chart\|visualization.*engine" "$DESIGN_FILE"
    check_test $? "Live visualization system designed"
fi

# Test 6: Integration validation
echo
echo "Test 6: Integration and Safety Validation"
echo "----------------------------------------"

if [ -f "$DESIGN_FILE" ]; then
    # Integration with existing bench tool
    grep -qi "bench.*tool\|bench.*integration\|existing.*bench" "$DESIGN_FILE"
    check_test $? "Bench tool integration designed"

    # WebSocket for real-time updates
    grep -qi "websocket\|real.*time.*update\|streaming" "$DESIGN_FILE"
    check_test $? "Real-time streaming system designed"

    # Safety mechanisms
    grep -qi "safety.*mechanism\|emergency.*brake\|circuit.*breaker" "$DESIGN_FILE"
    check_test $? "Safety mechanisms designed"

    # Performance considerations
    grep -qi "performance.*requirement\|latency.*target\|throughput" "$DESIGN_FILE"
    check_test $? "Performance requirements specified"
fi

# Test 7: Completeness validation
echo
echo "Test 7: Completeness Validation"
echo "------------------------------"

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
    # Check that all files mention core pattern concepts
    DESIGN_PATTERN=$(grep -c -i "pattern" "$DESIGN_FILE" || echo 0)
    API_PATTERN=$(grep -c -i "pattern" "$API_FILE" || echo 0)
    SCHEMA_PATTERN=$(grep -c -i "pattern" "$SCHEMA_FILE" || echo 0)

    if [ "$DESIGN_PATTERN" -gt 50 ] && [ "$API_PATTERN" -gt 20 ] && [ "$SCHEMA_PATTERN" -gt 20 ]; then
        check_test 0 "Consistent pattern terminology across documents"
    else
        check_test 1 "Inconsistent terminology (D:$DESIGN_PATTERN A:$API_PATTERN S:$SCHEMA_PATTERN)"
    fi

    # Check for guardrail consistency
    DESIGN_GUARDRAIL=$(grep -c -i "guardrail\|safety\|limit" "$DESIGN_FILE" || echo 0)
    API_GUARDRAIL=$(grep -c -i "guardrail\|safety\|limit" "$API_FILE" || echo 0)
    SCHEMA_GUARDRAIL=$(grep -c -i "guardrail\|safety\|limit" "$SCHEMA_FILE" || echo 0)

    if [ "$DESIGN_GUARDRAIL" -gt 10 ] && [ "$API_GUARDRAIL" -gt 5 ] && [ "$SCHEMA_GUARDRAIL" -gt 5 ]; then
        check_test 0 "Consistent guardrail coverage across documents"
    else
        check_test 1 "Inconsistent guardrail coverage (D:$DESIGN_GUARDRAIL A:$API_GUARDRAIL S:$SCHEMA_GUARDRAIL)"
    fi
fi

# Test 8: User story validation
echo
echo "Test 8: User Story Validation"
echo "----------------------------"

if [ -f "$DESIGN_FILE" ]; then
    # "I can run predefined patterns and see accurate live charts"
    grep -qi "predefined.*pattern\|pattern.*template" "$DESIGN_FILE"
    check_test $? "Predefined patterns support designed"

    grep -qi "accurate.*chart\|live.*chart\|real.*time.*chart" "$DESIGN_FILE"
    check_test $? "Live charting accuracy designed"

    grep -qi "user.*story\|as.*user\|I can" "$DESIGN_FILE"
    check_test $? "User stories included in design"
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