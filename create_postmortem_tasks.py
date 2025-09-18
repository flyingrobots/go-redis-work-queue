import errno
import json
import os
import sys
from datetime import datetime, timezone

# Create post-mortem reflection tasks for each worker
workers = [
    "claude-001", "claude-002", "claude-003", "claude-004", "claude-005",
    "claude-006", "claude-007", "claude-008", "claude-009", "claude-010"
]

postmortem_tasks = []

for i, worker in enumerate(workers, 1):
    task = {
        "task_id": f"POSTMORTEM.{i:03d}",
        "created_at": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
        "task": {
            "id": f"POSTMORTEM.{i:03d}",
            "feature_id": "SLAPS_REFLECTION",
            "title": f"Worker {i:02d} Post-Mortem Reflection",
            "description": f"As Worker {i} ({worker}), write your personal post-mortem reflection on the SLAPS experiment.",
            "boundaries": {
                "expected_complexity": {
                    "value": "~500-1000 words",
                    "breakdown": "Personal reflection and insights"
                },
                "definition_of_done": {
                    "criteria": [
                        f"Write from YOUR perspective as Worker {i}",
                        "Describe your unique experiences during SLAPS",
                        "Share challenges you faced",
                        "Highlight your achievements",
                        "Discuss any conflicts or collaborations",
                        "Reflect on the chaos and how you adapted",
                        "Include specific examples from your tasks",
                        "Be honest about what worked and what didn't",
                        "Share any emergent behaviors you developed",
                        f"Save to docs/SLAPS/worker-reflections/{worker}-reflection.md"
                    ]
                }
            },
            "execution_guidance": {
                "reflection_prompts": [
                    "What was your most challenging moment?",
                    "How did you handle conflicts with other workers?",
                    "What task are you most proud of completing?",
                    "Did you develop any unique strategies or patterns?",
                    "How did you deal with the compilation/test conflicts?",
                    "What was it like working without central coordination?",
                    "Did you ever get confused or lost? How did you recover?",
                    "What would you do differently next time?",
                    "How did the rate limit pauses affect you?",
                    "What emergent behaviors did you develop?",
                    "Any memorable interactions with other workers?",
                    "How did you feel about the chaos?"
                ],
                "tone": "Personal, honest, reflective - this is YOUR story",
                "format": "First-person narrative from the worker's perspective"
            },
            "dependencies": [],
            "resources_required": ["reflection_time"],
            "output_path": f"docs/SLAPS/worker-reflections/{worker}-reflection.md"
        }
    }
    postmortem_tasks.append(task)

# Add final coordinator summary task
coordinator_task = {
    "task_id": "POSTMORTEM.FINAL",
    "created_at": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
    "task": {
        "id": "POSTMORTEM.FINAL",
        "feature_id": "SLAPS_REFLECTION",
        "title": "Coordinator Final Summary and Analysis",
        "description": "Collect all worker reflections and create the final SLAPS post-mortem document.",
        "boundaries": {
            "expected_complexity": {
                "value": "Comprehensive document",
                "breakdown": "Collect, analyze, synthesize, conclude"
            },
            "definition_of_done": {
                "criteria": [
                    "Read all 10 worker reflection documents",
                    "Include excerpts from each worker's perspective",
                    "Identify common themes and patterns",
                    "Highlight unique insights from individual workers",
                    "Add coordinator's perspective and analysis",
                    "Draw conclusions about distributed AI coordination",
                    "Document lessons learned",
                    "Create final comprehensive post-mortem",
                    "Save to docs/SLAPS/FINAL-POSTMORTEM.md"
                ]
            }
        },
        "execution_guidance": {
            "sections": [
                "Executive Summary",
                "Individual Worker Perspectives (excerpts from all 10)",
                "Common Themes Across Workers",
                "Unique Experiences and Edge Cases",
                "Technical Challenges and Solutions",
                "Emergent Behaviors Observed",
                "Coordinator's Analysis",
                "Lessons for Future Distributed AI Systems",
                "Conclusions"
            ]
        },
        "dependencies": [task["task_id"] for task in postmortem_tasks],
        "resources_required": ["all_worker_reflections"]
    }
}

def ensure_directory(path: str) -> None:
    try:
        os.makedirs(path, exist_ok=True)
    except OSError as err:
        if err.errno in (errno.EACCES, errno.EROFS):
            raise RuntimeError(f"insufficient permissions to create directory '{path}' ({err.strerror})") from err
        raise RuntimeError(f"failed to create directory '{path}': {err.strerror}") from err


# Write all tasks to open-tasks
try:
    ensure_directory('slaps-coordination/open-tasks')
    ensure_directory('docs/SLAPS/worker-reflections')
except RuntimeError as err:
    print(f"create_postmortem_tasks: {err}", file=sys.stderr)
    sys.exit(1)

for task in postmortem_tasks:
    filename = f"slaps-coordination/open-tasks/{task['task_id']}.json"
    with open(filename, 'w') as f:
        json.dump(task, f, indent=2)
    print(f"Created: {task['task']['title']}")

# Write coordinator task
filename = f"slaps-coordination/open-tasks/{coordinator_task['task_id']}.json"
with open(filename, 'w') as f:
    json.dump(coordinator_task, f, indent=2)
print(f"Created: {coordinator_task['task']['title']}")

print(f"\n📝 Created {len(postmortem_tasks) + 1} post-mortem reflection tasks!")
print("Workers will now share their unique perspectives on the SLAPS experiment.")
print("Final task will synthesize all reflections into comprehensive post-mortem.")
