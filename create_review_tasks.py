import argparse
import json
import os
import re
from datetime import datetime, timezone
from typing import Iterable, List


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Generate rigorous code review tasks")
    parser.add_argument(
        "--limit",
        type=int,
        default=12,
        help="Maximum number of tasks to generate (default: 12)",
    )
    parser.add_argument(
        "--dir",
        dest="output_dir",
        default="slaps-coordination/open-tasks",
        help="Directory to write review task files (default: slaps-coordination/open-tasks)",
    )
    parser.add_argument(
        "--completed-dir",
        default="slaps-coordination/finished-tasks",
        help="Directory to read completed implementation tasks (default: slaps-coordination/finished-tasks)",
    )
    parser.add_argument(
        "--timestamp",
        help="ISO8601 or epoch timestamp to stamp tasks (default: now in UTC)",
    )
    return parser.parse_args()


def parse_timestamp(value: str | None) -> datetime:
    if value is None:
        return datetime.now(timezone.utc)
    try:
        epoch = float(value)
        return datetime.fromtimestamp(epoch, tz=timezone.utc)
    except (TypeError, ValueError):
        pass
    dt = datetime.fromisoformat(value)
    if dt.tzinfo is None:
        dt = dt.replace(tzinfo=timezone.utc)
    return dt.astimezone(timezone.utc)


def isoformat_z(dt: datetime) -> str:
    return dt.astimezone(timezone.utc).isoformat().replace("+00:00", "Z")


def load_completed_tasks(completed_dir: str) -> List[str]:
    if not os.path.isdir(completed_dir):
        return []
    tasks: List[str] = []
    for filename in os.listdir(completed_dir):
        lowered = filename.lower()
        if lowered.endswith(".json") and "duplicate" not in lowered:
            task_id = filename[:-5]
            if task_id.startswith(("P1.T", "P2.T", "P3.T", "P4.T")):
                try:
                    task_num = int(task_id.split(".T")[1])
                except (IndexError, ValueError) as exc:
                    print(f"Skipping {filename}: unable to parse task number ({exc})")
                    continue
                if task_num < 50:
                    tasks.append(task_id)
    return tasks


def numeric_key(task_id: str) -> tuple[int, str]:
    match = re.search(r"(\d+)$", task_id)
    if match:
        return int(match.group(1)), task_id
    return (0, task_id)


def write_tasks(tasks: Iterable[str], output_dir: str, created_at: datetime) -> None:
    os.makedirs(output_dir, exist_ok=True)
    review_tasks = []
    for i, task_id in enumerate(tasks, 1):
        review_task = {
            "task_id": f"REVIEW.{i:03d}",
            "created_at": isoformat_z(created_at),
            "task": {
                "id": f"REVIEW.{i:03d}",
                "feature_id": "CODE_REVIEW",
                "title": f"RIGOROUS Code Review: {task_id}",
                "description": (
                    f"Conduct EXTREMELY RIGOROUS code review of {task_id}. "
                    "Find ALL issues, no matter how small. FIX everything you find."
                ),
                "boundaries": {
                    "expected_complexity": {
                        "value": "Deep Review + Complete Fixes",
                        "breakdown": "Rigorous review (40%), Fix all issues (60%)",
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
                            "Code is production-ready",
                        ]
                    },
                    "severity_levels": {
                        "CRITICAL": "Security holes, data loss, crashes - FIX IMMEDIATELY",
                        "HIGH": "Race conditions, missing error handling - MUST FIX",
                        "MEDIUM": "Performance issues, bad patterns - SHOULD FIX",
                        "LOW": "Style issues, naming - FIX IF TIME",
                    },
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
                        "Accessibility: If UI, WCAG compliance issues?",
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
                        "12. Commit all fixes with comprehensive message",
                    ],
                    "expected_findings": "Find at least 5-10 issues per task - be EXTREMELY thorough",
                },
                "dependencies": [],
                "resources_required": ["source_code", "test_framework", "linter"],
            },
        }
        review_tasks.append(review_task)

    for task in review_tasks:
        filename = os.path.join(output_dir, f"{task['task_id']}.json")
        with open(filename, "w") as f:
            json.dump(task, f, indent=2)

    for task in review_tasks:
        print(f"Created: {task['task']['title']}")

    print(f"\n🔍 Created {len(review_tasks)} RIGOROUS code review tasks!")
    print("Workers will now tear apart the code and fix EVERYTHING!")


def main() -> None:
    args = parse_args()
    timestamp = parse_timestamp(args.timestamp)
    completed_tasks = load_completed_tasks(args.completed_dir)
    if not completed_tasks:
        print("No completed implementation tasks found; nothing to create.")
        return

    selected = sorted(completed_tasks, key=numeric_key)[: max(args.limit, 0)]
    if not selected:
        print("No tasks selected after applying limit/filter conditions.")
        return

    write_tasks(selected, args.output_dir, timestamp)


if __name__ == "__main__":
    main()
