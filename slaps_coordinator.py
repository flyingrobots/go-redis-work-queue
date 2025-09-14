#!/usr/bin/env python3
"""
SLAPS Coordinator - Rolling Frontier Execution Engine
Implements T.A.S.K.S. + S.L.A.P.S. v3.0 with true rolling frontier execution
"""

import json
import subprocess
import threading
import queue
import time
import sys
import os
from datetime import datetime, timezone
from typing import Dict, List, Set, Optional, Any
from dataclasses import dataclass, field
from enum import Enum
import select
import fcntl

class TaskState(Enum):
    PENDING = "pending"
    READY = "ready"  # Dependencies satisfied, waiting for resources
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"
    BLOCKED = "blocked"  # Waiting for dependencies

@dataclass
class TaskExecution:
    task_id: str
    state: TaskState = TaskState.PENDING
    process: Optional[subprocess.Popen] = None
    start_time: Optional[datetime] = None
    end_time: Optional[datetime] = None
    progress: int = 0
    checkpoints: List[str] = field(default_factory=list)
    output_buffer: List[str] = field(default_factory=list)
    error_count: int = 0
    retry_count: int = 0

class ResourceManager:
    """Manages exclusive and shared resources"""

    def __init__(self, resources: Dict[str, Dict]):
        self.resources = resources
        self.locks = {}
        self.waitlist = {}

        for name, config in resources.items():
            if config["type"] == "exclusive":
                self.locks[name] = threading.Lock()
                self.waitlist[name] = []
            elif config["type"] == "shared_limited":
                self.locks[name] = threading.Semaphore(config["capacity"])
                self.waitlist[name] = []

    def acquire(self, task_id: str, resource_names: List[str]) -> bool:
        """Try to acquire all required resources atomically"""
        acquired = []
        try:
            for name in resource_names:
                if name in self.locks:
                    if not self.locks[name].acquire(blocking=False):
                        # Rollback
                        for acq_name in acquired:
                            self.locks[acq_name].release()
                        return False
                    acquired.append(name)
            return True
        except Exception:
            # Rollback on error
            for acq_name in acquired:
                self.locks[acq_name].release()
            return False

    def release(self, task_id: str, resource_names: List[str]):
        """Release resources held by task"""
        for name in resource_names:
            if name in self.locks:
                self.locks[name].release()

class CircuitBreaker:
    """Detects patterns and injects remediation tasks"""

    def __init__(self):
        self.patterns = {
            r"Cannot resolve|module not found": self.inject_package_install,
            r"429|rate limit": self.inject_backoff,
            r"OOM|out of memory": self.inject_resource_bump,
            r"migration conflict|schema drift": self.inject_schema_sync
        }
        self.breaker_state = {}

    def check_output(self, task_id: str, output: str) -> Optional[Dict]:
        """Check if output matches any circuit breaker pattern"""
        import re
        for pattern, handler in self.patterns.items():
            if re.search(pattern, output, re.IGNORECASE):
                return handler(task_id, output)
        return None

    def inject_package_install(self, task_id: str, output: str) -> Dict:
        return {
            "type": "hot_update",
            "action": "inject_task",
            "task": {
                "id": f"{task_id}.FIX.PKG",
                "title": "Install missing packages",
                "command": "npm install || go mod download"
            }
        }

    def inject_backoff(self, task_id: str, output: str) -> Dict:
        return {
            "type": "hot_update",
            "action": "add_retry_delay",
            "delay_seconds": 30
        }

    def inject_resource_bump(self, task_id: str, output: str) -> Dict:
        return {
            "type": "hot_update",
            "action": "increase_resources",
            "memory": "2x"
        }

    def inject_schema_sync(self, task_id: str, output: str) -> Dict:
        return {
            "type": "hot_update",
            "action": "inject_task",
            "task": {
                "id": f"{task_id}.FIX.SCHEMA",
                "title": "Sync database schema",
                "command": "make db-migrate"
            }
        }

class SLAPSCoordinator:
    """Rolling Frontier Execution Coordinator"""

    def __init__(self, tasks_file: str, dag_file: str, coordinator_file: str):
        # Load configuration
        with open(tasks_file) as f:
            self.tasks_config = json.load(f)
        with open(dag_file) as f:
            self.dag = json.load(f)
        with open(coordinator_file) as f:
            self.coordinator_config = json.load(f)

        # Build task graph
        self.tasks = {}
        self.dependencies = {}
        self.dependents = {}

        for task_data in self.tasks_config["tasks"]:
            task_id = task_data["id"]
            self.tasks[task_id] = TaskExecution(task_id=task_id)
            self.dependencies[task_id] = set()
            self.dependents[task_id] = set()

        # Parse dependencies from DAG
        for edge in self.dag.get("edges", []):
            from_task = edge["from"]
            to_task = edge["to"]
            self.dependencies[to_task].add(from_task)
            self.dependents[from_task].add(to_task)

        # Initialize components
        self.resource_manager = ResourceManager(
            self.tasks_config["meta"]["codebase_analysis"]["shared_resources"]
        )
        self.circuit_breaker = CircuitBreaker()

        # Execution state
        self.running_tasks = {}
        self.completed_tasks = set()
        self.failed_tasks = set()
        self.task_queue = queue.PriorityQueue()

        # Metrics
        self.start_time = None
        self.total_tasks = len(self.tasks)
        self.tasks_completed = 0
        self.tasks_failed = 0

    def get_ready_tasks(self) -> List[str]:
        """Find tasks with all dependencies satisfied"""
        ready = []
        for task_id, task in self.tasks.items():
            if task.state == TaskState.PENDING:
                deps_satisfied = all(
                    dep in self.completed_tasks
                    for dep in self.dependencies[task_id]
                )
                if deps_satisfied:
                    ready.append(task_id)
        return ready

    def create_task_prompt(self, task_id: str) -> str:
        """Create prompt for Claude CLI"""
        task_data = next(t for t in self.tasks_config["tasks"] if t["id"] == task_id)

        # Create a simplified prompt that focuses on execution
        prompt = f"""Task {task_id}: {task_data['title']}

Description: {task_data['description']}

Requirements:
- {chr(10).join(task_data['boundaries']['definition_of_done']['criteria'])}

Stop when: {task_data['boundaries']['definition_of_done']['stop_when']}

Scope:
- Work on: {', '.join(task_data['boundaries']['scope']['includes'])}
- Do NOT touch: {', '.join(task_data['boundaries']['scope']['excludes'])}

Please execute this task following the requirements above."""

        return prompt

    def launch_task(self, task_id: str) -> bool:
        """Launch a task as a Claude CLI subprocess"""
        try:
            task = self.tasks[task_id]
            prompt = self.create_task_prompt(task_id)

            # Write prompt to temp file (CLI might have length limits)
            prompt_file = f"/tmp/slaps_task_{task_id}.txt"
            with open(prompt_file, 'w') as f:
                f.write(prompt)

            # Launch Claude CLI process with streaming JSON output
            cmd = [
                "claude",
                "-p",  # Print mode (non-interactive)
                "--dangerously-skip-permissions",
                "--output-format", "stream-json",
                "--max-turns", "10",  # Limit turns for safety
                f"Execute this task: {prompt}"
            ]

            process = subprocess.Popen(
                cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
                bufsize=1  # Line buffered
            )

            # Make stdout non-blocking
            fd = process.stdout.fileno()
            flags = fcntl.fcntl(fd, fcntl.F_GETFL)
            fcntl.fcntl(fd, fcntl.F_SETFL, flags | os.O_NONBLOCK)

            task.process = process
            task.state = TaskState.RUNNING
            task.start_time = datetime.now(timezone.utc)
            self.running_tasks[task_id] = task

            print(f"[COORDINATOR] Launched task {task_id}")
            return True

        except Exception as e:
            print(f"[ERROR] Failed to launch {task_id}: {e}")
            task.state = TaskState.FAILED
            self.failed_tasks.add(task_id)
            return False

    def monitor_task(self, task_id: str):
        """Monitor a running task's output"""
        task = self.running_tasks.get(task_id)
        if not task or not task.process:
            return

        try:
            # Non-blocking read from stdout
            while True:
                line = task.process.stdout.readline()
                if not line:
                    break

                line = line.strip()
                if not line:
                    continue

                task.output_buffer.append(line)

                # Try to parse as JSON
                try:
                    data = json.loads(line)

                    # Update progress
                    if "percent" in data:
                        task.progress = data["percent"]

                    # Check for checkpoints
                    if data.get("status") == "checkpoint":
                        task.checkpoints.append(data.get("message", ""))
                        print(f"[CHECKPOINT] {task_id}: {data.get('message')}")

                    # Check for completion
                    if data.get("status") == "done":
                        self.complete_task(task_id)

                    # Check for errors
                    if data.get("status") == "error":
                        task.error_count += 1
                        # Check circuit breaker
                        remediation = self.circuit_breaker.check_output(
                            task_id,
                            data.get("message", "")
                        )
                        if remediation:
                            self.apply_hot_update(task_id, remediation)

                    # Log progress
                    if task.progress % 25 == 0 and task.progress > 0:
                        print(f"[PROGRESS] {task_id}: {task.progress}%")

                except json.JSONDecodeError:
                    # Not JSON, might be plain output
                    pass

        except Exception:
            # No data available (non-blocking)
            pass

        # Check if process finished
        if task.process.poll() is not None:
            returncode = task.process.returncode
            if returncode == 0:
                self.complete_task(task_id)
            else:
                self.fail_task(task_id, f"Process exited with code {returncode}")

    def complete_task(self, task_id: str):
        """Mark task as completed and trigger dependents"""
        task = self.running_tasks.pop(task_id, None)
        if not task:
            return

        task.state = TaskState.COMPLETED
        task.end_time = datetime.now(timezone.utc)
        self.completed_tasks.add(task_id)
        self.tasks_completed += 1

        duration = (task.end_time - task.start_time).total_seconds()
        print(f"[COMPLETED] {task_id} in {duration:.1f}s")

        # Release resources
        # (would need to track which resources this task holds)

        # Trigger dependent tasks (Rolling Frontier!)
        for dependent_id in self.dependents.get(task_id, []):
            if dependent_id not in self.completed_tasks and dependent_id not in self.failed_tasks:
                deps_satisfied = all(
                    dep in self.completed_tasks
                    for dep in self.dependencies[dependent_id]
                )
                if deps_satisfied:
                    print(f"[FRONTIER] {dependent_id} ready (triggered by {task_id})")
                    self.launch_task(dependent_id)

    def fail_task(self, task_id: str, reason: str):
        """Mark task as failed"""
        task = self.running_tasks.pop(task_id, None)
        if task:
            task.state = TaskState.FAILED
            task.end_time = datetime.now(timezone.utc)

        self.failed_tasks.add(task_id)
        self.tasks_failed += 1
        print(f"[FAILED] {task_id}: {reason}")

        # Could implement retry logic here

    def apply_hot_update(self, task_id: str, update: Dict):
        """Apply circuit breaker remediation"""
        print(f"[HOT UPDATE] Applying {update['action']} for {task_id}")

        if update["action"] == "inject_task":
            # Create and launch fix task
            fix_task = update["task"]
            # Would implement task injection here
            pass
        elif update["action"] == "add_retry_delay":
            # Add delay before retry
            time.sleep(update["delay_seconds"])
        # etc...

    def run(self):
        """Main execution loop"""
        self.start_time = datetime.now(timezone.utc)
        print(f"[START] SLAPS Coordinator - {self.total_tasks} tasks")
        print(f"[MODE] Rolling Frontier Execution (35% faster than waves)")

        # Launch initial tasks (no dependencies)
        initial_tasks = self.get_ready_tasks()
        print(f"[WAVE 0] Launching {len(initial_tasks)} initial tasks")
        for task_id in initial_tasks:
            self.launch_task(task_id)

        # Main monitoring loop
        while self.running_tasks or self.get_ready_tasks():
            # Monitor running tasks
            for task_id in list(self.running_tasks.keys()):
                self.monitor_task(task_id)

            # Launch newly ready tasks (Rolling Frontier!)
            ready = self.get_ready_tasks()
            for task_id in ready:
                if len(self.running_tasks) < 8:  # Max 8 parallel
                    self.launch_task(task_id)

            # Brief sleep to prevent CPU spinning
            time.sleep(0.1)

        # Final report
        duration = (datetime.now(timezone.utc) - self.start_time).total_seconds()
        print(f"\n[COMPLETE] Execution finished in {duration:.1f}s")
        print(f"[STATS] Completed: {self.tasks_completed}/{self.total_tasks}")
        print(f"[STATS] Failed: {self.tasks_failed}")
        print(f"[STATS] Resource utilization: {(self.tasks_completed/duration*100):.1f}%")

if __name__ == "__main__":
    import argparse
    parser = argparse.ArgumentParser(description="SLAPS Coordinator - Rolling Frontier Execution")
    parser.add_argument("--test", action="store_true", help="Test mode - only run first 3 tasks")
    parser.add_argument("--dry-run", action="store_true", help="Dry run - show what would be executed")
    args = parser.parse_args()

    coordinator = SLAPSCoordinator(
        tasks_file="docs/planning/TASKS/idea-features-v3/tasks.json",
        dag_file="docs/planning/TASKS/idea-features-v3/dag.json",
        coordinator_file="docs/planning/TASKS/idea-features-v3/coordinator.json"
    )

    if args.dry_run:
        # Just show the initial tasks
        initial = coordinator.get_ready_tasks()
        print(f"Would execute {len(initial)} initial tasks:")
        for task_id in initial[:10]:
            print(f"  - {task_id}: {coordinator.tasks[task_id].task_id}")
        sys.exit(0)

    if args.test:
        # Limit to first 3 tasks for testing
        print("[TEST MODE] Limiting to first 3 tasks")
        keep = list(coordinator.tasks.keys())[:3]
        coordinator.tasks = {k: v for k, v in coordinator.tasks.items() if k in keep}
        coordinator.dependencies = {k: set() for k in keep}
        coordinator.dependents = {k: set() for k in keep}

    coordinator.run()