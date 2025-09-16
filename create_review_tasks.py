import json
import os
import re
from datetime import datetime

# Get list of completed implementation tasks to review
completed_dir = 'slaps-coordination/finished-tasks'
completed_tasks = []

for f in os.listdir(completed_dir):
    if f.endswith('.json') and not 'duplicate' in f:
        task_id = f[:-5]
        # Review P1, P2, P3, P4 implementation tasks
        if task_id.startswith(('P1.T', 'P2.T', 'P3.T', 'P4.T')):
            try:
                task_num = int(task_id.split('.T')[1])
                # Focus on implementation tasks (lower numbers tend to be implementations)
                if task_num < 50:
                    completed_tasks.append(task_id)
            except:
                pass

def _task_numeric_key(task_id: str) -> tuple[int, str]:
    match = re.search(r"(\d+)$", task_id.replace(".json", ""))
    if match:
        return int(match.group(1)), task_id
    return (0, task_id)


# Sort and take first 12 for review using numeric ordering to keep tasks in sequence
completed_tasks = sorted(completed_tasks, key=_task_numeric_key)[:12]

# Create rigorous code review tasks
review_tasks = []
for i, task_id in enumerate(completed_tasks, 1):
    review_task = {
        "task_id": f"REVIEW.{i:03d}",
        "created_at": datetime.now().isoformat() + "Z",
        "task": {
            "id": f"REVIEW.{i:03d}",
            "feature_id": "CODE_REVIEW",
            "title": f"RIGOROUS Code Review: {task_id}",
            "description": f"Conduct EXTREMELY RIGOROUS code review of {task_id}. Find ALL issues, no matter how small. FIX everything you find.",
            "boundaries": {
                "expected_complexity": {
                    "value": "Deep Review + Complete Fixes",
                    "breakdown": "Rigorous review (40%), Fix all issues (60%)"
                },
                "definition_of_done": {
                    "criteria": [
                        "Found and documented ALL bugs",
                        "Fixed every single issue discovered",
                        "Eliminated ALL security vulnerabilities", 
                        "Removed ALL race conditions",
                        "Fixed ALL error handling gaps",
                        "Refactored ALL problematic code",
                        "Achieved 90%+ test coverage",
                        "Zero linting warnings remain",
                        "Performance is optimal",
                        "Code is production-ready"
                    ]
                },
                "severity_levels": {
                    "CRITICAL": "Security holes, data loss, crashes - FIX IMMEDIATELY",
                    "HIGH": "Race conditions, missing error handling - MUST FIX",
                    "MEDIUM": "Performance issues, bad patterns - SHOULD FIX",
                    "LOW": "Style issues, naming - FIX IF TIME"
                }
            },
            "execution_guidance": {
                "review_checklist": [
                    "Security: SQL injection, XSS, path traversal, secrets exposed?",
                    "Concurrency: Race conditions, deadlocks, data races?",
                    "Resources: Memory leaks, unclosed files, connection leaks?",
                    "Errors: Unhandled errors, panics, nil pointers?",
                    "Validation: Input validation, bounds checking, type safety?",
                    "Performance: O(n²) algorithms, N+1 queries, inefficient loops?",
                    "Testing: Missing tests, untested edge cases, low coverage?",
                    "Patterns: Anti-patterns, code smells, violations of DRY/SOLID?",
                    "Dependencies: Outdated, vulnerable, or unnecessary dependencies?",
                    "Documentation: Missing docs, outdated comments, unclear code?",
                    "Logging: Too much/little logging, sensitive data in logs?",
                    "Configuration: Hardcoded values, missing config validation?",
                    "Compatibility: Breaking changes, version conflicts?",
                    "Maintainability: Complex code, magic numbers, unclear logic?",
                    "Accessibility: If UI, WCAG compliance issues?"
                ],
                "instructions": [
                    f"1. Locate and read ALL code for {task_id}",
                    "2. Run static analysis tools (go vet, golangci-lint, etc.)",
                    "3. Check test coverage (must be 80%+ or add tests)",
                    "4. Review against the 15-point checklist above",
                    "5. Document EVERY issue found, no matter how minor",
                    "6. Fix CRITICAL issues first (security, crashes)",
                    "7. Fix HIGH priority issues (correctness, reliability)",
                    "8. Fix MEDIUM issues (performance, patterns)",
                    "9. Fix LOW issues if any remain",
                    "10. Re-test everything after fixes",
                    "11. Create detailed review report",
                    "12. Commit all fixes with comprehensive message"
                ],
                "expected_findings": "Find at least 5-10 issues per task - be EXTREMELY thorough"
            },
            "dependencies": [],
            "resources_required": ["source_code", "test_framework", "linter"]
        }
    }
    review_tasks.append(review_task)

# Write review tasks to open-tasks
os.makedirs('slaps-coordination/open-tasks', exist_ok=True)
for task in review_tasks:
    filename = f"slaps-coordination/open-tasks/{task['task_id']}.json"
    with open(filename, 'w') as f:
        json.dump(task, f, indent=2)
    print(f"Created: {task['task']['title']}")

print(f"\n🔍 Created {len(review_tasks)} RIGOROUS code review tasks!")
print("Workers will now tear apart the code and fix EVERYTHING!")
