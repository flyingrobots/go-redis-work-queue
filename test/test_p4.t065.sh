#!/bin/bash

# Test script for P4.T065 - Theme Playground Design Validation
# Validates that the design deliverables meet all acceptance criteria

set -e

# Change to project directory
cd "/Users/james/git/go-redis-work-queue"

echo "P4.T065 Acceptance Test - Theme Playground Design"
echo "================================================"

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

DESIGN_FILE="docs/design/f026-design.md"
API_FILE="docs/api/f026-openapi.yaml"
SCHEMA_FILE="docs/schemas/f026-schema.json"

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

    # Check for theme-specific content
    grep -qi "theme.*system\|theme.*manager\|theme.*registry" "$DESIGN_FILE"
    check_test $? "Theme system architecture documented"

    grep -qi "accessibility\|wcag\|contrast" "$DESIGN_FILE"
    check_test $? "Accessibility features documented"

    grep -qi "playground\|preview\|customization" "$DESIGN_FILE"
    check_test $? "Theme playground features documented"

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

    # Check for theme management endpoints
    grep -q "/themes" "$API_FILE"
    check_test $? "Theme management endpoints defined"

    grep -q "/themes/{themeId}/apply" "$API_FILE"
    check_test $? "Theme application endpoint defined"

    grep -q "/themes/{themeId}/preview" "$API_FILE"
    check_test $? "Theme preview endpoint defined"

    grep -q "/themes/{themeId}/validate" "$API_FILE"
    check_test $? "Accessibility validation endpoint defined"

    grep -q "/accessibility" "$API_FILE"
    check_test $? "Accessibility endpoints defined"

    grep -q "/user/preferences" "$API_FILE"
    check_test $? "User preferences endpoints defined"

    # Check for import/export functionality
    grep -q "/themes/import" "$API_FILE"
    check_test $? "Theme import endpoint defined"

    grep -q "/themes/{themeId}/export" "$API_FILE"
    check_test $? "Theme export endpoint defined"

    # Check for comprehensive schemas
    grep -q "ColorPalette" "$API_FILE"
    check_test $? "ColorPalette schema defined"

    grep -q "ComponentStyles" "$API_FILE"
    check_test $? "ComponentStyles schema defined"

    grep -q "AccessibilityReport" "$API_FILE"
    check_test $? "AccessibilityReport schema defined"

    # Check file size for comprehensiveness
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
    check_test 0 "JSON Schema file exists"

    # Validate JSON syntax
    if python3 -m json.tool "$SCHEMA_FILE" > /dev/null 2>&1; then
        check_test 0 "JSON Schema syntax valid"
    else
        check_test 1 "JSON Schema syntax valid"
    fi

    # Check for core theme schemas
    grep -q "Theme\":" "$SCHEMA_FILE"
    check_test $? "Theme schema defined"

    grep -q "ColorPalette" "$SCHEMA_FILE"
    check_test $? "ColorPalette schema defined"

    grep -q "ComponentStyles" "$SCHEMA_FILE"
    check_test $? "ComponentStyles schema defined"

    grep -q "ThemeMetadata" "$SCHEMA_FILE"
    check_test $? "ThemeMetadata schema defined"

    # Check for accessibility schemas
    grep -q "AccessibilityInfo" "$SCHEMA_FILE"
    check_test $? "AccessibilityInfo schema defined"

    grep -q "AccessibilityReport" "$SCHEMA_FILE"
    check_test $? "AccessibilityReport schema defined"

    grep -q "ContrastTest" "$SCHEMA_FILE"
    check_test $? "ContrastTest schema defined"

    # Check for user preference schemas
    grep -q "UserThemePreferences" "$SCHEMA_FILE"
    check_test $? "UserThemePreferences schema defined"

    grep -q "AccessibilityPreferences" "$SCHEMA_FILE"
    check_test $? "AccessibilityPreferences schema defined"

    # Check for import/export schemas
    grep -q "ThemeExportData" "$SCHEMA_FILE"
    check_test $? "ThemeExportData schema defined"

    grep -q "ThemeImportData" "$SCHEMA_FILE"
    check_test $? "ThemeImportData schema defined"

    # Check schema completeness
    SCHEMA_LINES=$(wc -l < "$SCHEMA_FILE")
    if [ "$SCHEMA_LINES" -gt 1000 ]; then
        check_test 0 "Schema definitions comprehensive ($SCHEMA_LINES lines)"
    else
        check_test 1 "Schema definitions incomplete ($SCHEMA_LINES lines, expected >1000)"
    fi
else
    check_test 1 "JSON Schema file exists"
fi

# Test 4: Acceptance criteria validation
echo
echo "Test 4: Acceptance Criteria Coverage"
echo "-----------------------------------"

if [ -f "$DESIGN_FILE" ]; then
    # Theme registry with dark/light/high-contrast
    grep -qi "theme.*registry\|registry.*theme" "$DESIGN_FILE"
    check_test $? "Theme registry system designed"

    grep -qi "dark.*light.*contrast\|contrast.*theme\|accessibility.*theme" "$DESIGN_FILE"
    check_test $? "Accessibility themes addressed"

    # Settings UI to preview/apply themes
    grep -qi "settings.*ui\|playground.*ui\|preview.*ui" "$DESIGN_FILE"
    check_test $? "Settings UI for theme preview designed"

    grep -qi "apply.*theme\|theme.*application" "$DESIGN_FILE"
    check_test $? "Theme application mechanism designed"

    grep -qi "preview.*theme\|live.*preview" "$DESIGN_FILE"
    check_test $? "Theme preview functionality designed"

    # Persistence across sessions
    grep -qi "persist\|storage\|session" "$DESIGN_FILE"
    check_test $? "Theme persistence mechanism designed"

    # Theme struct and registry (from task acceptance checks)
    grep -qi "theme.*struct\|struct.*theme" "$DESIGN_FILE"
    check_test $? "Theme data structure designed"

    # Replace hardcoded colors with theme lookups
    grep -qi "hardcoded.*color\|theme.*lookup\|lip.*gloss" "$DESIGN_FILE"
    check_test $? "Color system integration designed"
fi

# Test 5: Technical approach validation
echo
echo "Test 5: Technical Approach Validation"
echo "------------------------------------"

if [ -f "$DESIGN_FILE" ]; then
    # Theme core features
    grep -qi "palette.*border.*emphasis\|color.*palette" "$DESIGN_FILE"
    check_test $? "Color palette system designed"

    grep -qi "adaptive.*color\|terminal.*detect" "$DESIGN_FILE"
    check_test $? "Adaptive color system designed"

    # Playground features
    grep -qi "settings.*tab\|preview.*tile\|toggle.*key" "$DESIGN_FILE"
    check_test $? "Playground UI features designed"

    grep -qi "live.*apply\|animation.*flicker" "$DESIGN_FILE"
    check_test $? "Live application system designed"

    # Persistence features
    grep -qi "xdg.*config\|theme\.json\|state.*file" "$DESIGN_FILE"
    check_test $? "Configuration persistence designed"

    grep -qi "no_color\|minimal.*style" "$DESIGN_FILE"
    check_test $? "NO_COLOR compliance designed"

    # Accessibility features
    grep -qi "contrast.*check\|wcag.*heuristic" "$DESIGN_FILE"
    check_test $? "Accessibility validation designed"

    grep -qi "monochrome.*theme\|limited.*terminal" "$DESIGN_FILE"
    check_test $? "Fallback themes designed"
fi

# Test 6: Integration and completeness
echo
echo "Test 6: Integration and Completeness"
echo "-----------------------------------"

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
    # Check that all files mention core theme concepts
    DESIGN_THEME=$(grep -c -i "theme" "$DESIGN_FILE" || echo 0)
    API_THEME=$(grep -c -i "theme" "$API_FILE" || echo 0)
    SCHEMA_THEME=$(grep -c -i "theme" "$SCHEMA_FILE" || echo 0)

    if [ "$DESIGN_THEME" -gt 50 ] && [ "$API_THEME" -gt 20 ] && [ "$SCHEMA_THEME" -gt 20 ]; then
        check_test 0 "Consistent theme terminology across documents"
    else
        check_test 1 "Inconsistent terminology (D:$DESIGN_THEME A:$API_THEME S:$SCHEMA_THEME)"
    fi

    # Check for accessibility consistency
    DESIGN_ACCESSIBILITY=$(grep -c -i "accessibility\|wcag\|contrast" "$DESIGN_FILE" || echo 0)
    API_ACCESSIBILITY=$(grep -c -i "accessibility\|wcag\|contrast" "$API_FILE" || echo 0)
    SCHEMA_ACCESSIBILITY=$(grep -c -i "accessibility\|wcag\|contrast" "$SCHEMA_FILE" || echo 0)

    if [ "$DESIGN_ACCESSIBILITY" -gt 10 ] && [ "$API_ACCESSIBILITY" -gt 5 ] && [ "$SCHEMA_ACCESSIBILITY" -gt 5 ]; then
        check_test 0 "Consistent accessibility coverage across documents"
    else
        check_test 1 "Inconsistent accessibility coverage (D:$DESIGN_ACCESSIBILITY A:$API_ACCESSIBILITY S:$SCHEMA_ACCESSIBILITY)"
    fi
fi

# Test 7: Design quality validation
echo
echo "Test 7: Design Quality Validation"
echo "--------------------------------"

# Check for comprehensive component coverage
if [ -f "$DESIGN_FILE" ] || [ -f "$SCHEMA_FILE" ]; then
    # Core UI components should be covered
    COMPONENTS_MENTIONED=0

    for component in "button" "input" "table" "card" "modal" "navigation"; do
        if grep -qi "$component" "$DESIGN_FILE" "$SCHEMA_FILE" 2>/dev/null; then
            ((COMPONENTS_MENTIONED++))
        fi
    done

    if [ "$COMPONENTS_MENTIONED" -ge 5 ]; then
        check_test 0 "Comprehensive component coverage ($COMPONENTS_MENTIONED/6 components)"
    else
        check_test 1 "Insufficient component coverage ($COMPONENTS_MENTIONED/6 components)"
    fi
fi

# Check for user story coverage
if [ -f "$DESIGN_FILE" ]; then
    grep -qi "user.*story\|as.*user\|I can" "$DESIGN_FILE"
    check_test $? "User stories included in design"

    grep -qi "switch.*theme.*persist\|persist.*choice" "$DESIGN_FILE"
    check_test $? "Theme switching user story covered"

    grep -qi "high.*contrast.*readable\|accessible.*theme" "$DESIGN_FILE"
    check_test $? "Accessibility user story covered"
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