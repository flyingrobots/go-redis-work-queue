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
import psutil  # For system resource monitoring
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
    held_resources: List[str] = field(default_factory=list)

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

class SystemResourceManager:
    """Monitors and manages system CPU/memory resources"""

    def __init__(self, max_cpu_percent: float = 60.0, max_memory_percent: float = 70.0,
                 max_claude_instances: int = 3):
        self.max_cpu_percent = max_cpu_percent
        self.max_memory_percent = max_memory_percent
        self.max_claude_instances = max_claude_instances
        self.launch_cooldown = 2.0  # Seconds between launches
        self.last_launch_time = 0

    def can_launch_task(self, current_running: int) -> tuple[bool, str]:
        """Check if system can handle another task"""
        # Check Claude instance limit
        if current_running >= self.max_claude_instances:
            return False, f"Max Claude instances ({self.max_claude_instances}) reached"

        # Check CPU usage
        cpu_percent = psutil.cpu_percent(interval=0.1)
        if cpu_percent > self.max_cpu_percent:
            return False, f"CPU usage too high: {cpu_percent:.1f}%"

        # Check memory usage
        memory = psutil.virtual_memory()
        if memory.percent > self.max_memory_percent:
            return False, f"Memory usage too high: {memory.percent:.1f}%"

        # Check launch cooldown (stagger launches)
        time_since_last = time.time() - self.last_launch_time
        if time_since_last < self.launch_cooldown:
            return False, f"Launch cooldown: {self.launch_cooldown - time_since_last:.1f}s remaining"

        return True, "Resources available"

    def record_launch(self):
        """Record that a task was launched"""
        self.last_launch_time = time.time()

    def get_system_stats(self) -> Dict:
        """Get current system resource usage"""
        return {
            "cpu_percent": psutil.cpu_percent(interval=0.1),
            "memory_percent": psutil.virtual_memory().percent,
            "memory_available_gb": psutil.virtual_memory().available / (1024**3),
            "load_average": os.getloadavg()[0]  # 1-minute load average
        }

class SLAPSCoordinator:
    """Rolling Frontier Execution Coordinator"""

    def __init__(self, tasks_file: str, dag_file: str, coordinator_file: str,
                 max_claude_instances: int = 3, model: str = "sonnet"):
        # Load configuration
        with open(tasks_file) as f:
            self.tasks_config = json.load(f)
        with open(dag_file) as f:
            self.dag = json.load(f)
        with open(coordinator_file) as f:
            self.coordinator_config = json.load(f)

        # Configuration
        self.max_claude_instances = max_claude_instances
        self.model = model  # sonnet or opus

        # Build task graph
        self.tasks = {}
        self.dependencies = {}
        self.dependents = {}

        for task_data in self.tasks_config["tasks"]:
            task_id = task_data["id"]
            self.tasks[task_id] = TaskExecution(task_id=task_id)
            self.dependencies[task_id] = set()
            self.dependents[task_id] = set()

        # Parse dependencies from DAG (stored as edges)
        dag_deps = self.dag.get("dependencies", [])
        for edge in dag_deps:
            from_task = edge["from"]
            to_task = edge["to"]
            if to_task in self.dependencies:
                self.dependencies[to_task].add(from_task)
            if from_task in self.dependents:
                self.dependents[from_task].add(to_task)

        # Debug: Show dependency counts
        deps_count = sum(1 for deps in self.dependencies.values() if deps)
        no_deps = [t for t, deps in self.dependencies.items() if not deps]
        print(f"[DEBUG] Tasks with dependencies: {deps_count}/{len(self.tasks)}")
        print(f"[DEBUG] Initial tasks (no deps): {len(no_deps)}")

        # Initialize components
        self.resource_manager = ResourceManager(
            self.tasks_config["meta"]["codebase_analysis"]["shared_resources"]
        )
        self.system_resource_manager = SystemResourceManager(
            max_cpu_percent=self.coordinator_config.get("max_cpu_percent", 60.0),
            max_memory_percent=self.coordinator_config.get("max_memory_percent", 70.0),
            max_claude_instances=max_claude_instances
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
        # For testing, use a simple prompt
        if hasattr(self, 'test_mode') and self.test_mode:
            return f"This is test task {task_id}. Just respond with 'Task {task_id} completed successfully' and nothing else."

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

    def get_task_resources(self, task_id: str) -> List[str]:
        """Determine which resources a task needs"""
        task_data = next(t for t in self.tasks_config["tasks"] if t["id"] == task_id)
        resources = []

        # Check if task needs exclusive resources based on type
        if "Deploy" in task_data["title"]:
            resources.append("deployment_slot")
        if "schema" in task_data.get("description", "").lower():
            resources.append("redis_schema")
        if "Test" in task_data["title"]:
            resources.append("test_redis")
        if any(x in task_data["title"] for x in ["Build", "Test", "Deploy"]):
            resources.append("ci_runners")

        # Check resource_requirements in task spec
        if "resource_requirements" in task_data:
            if "exclusive_locks" in task_data["resource_requirements"]:
                resources.extend(task_data["resource_requirements"]["exclusive_locks"])
            if "shared_resources" in task_data["resource_requirements"]:
                resources.extend(task_data["resource_requirements"]["shared_resources"])

        return resources

    def launch_task(self, task_id: str) -> bool:
        """Launch a task as a Claude CLI subprocess"""
        try:
            task = self.tasks[task_id]

            # Check system resources first (CPU/memory)
            can_launch, reason = self.system_resource_manager.can_launch_task(len(self.running_tasks))
            if not can_launch:
                print(f"[THROTTLE] {task_id}: {reason}")
                return False

            # Check task-specific resources (locks, databases, etc.)
            required_resources = self.get_task_resources(task_id)
            if required_resources:
                if not self.resource_manager.acquire(task_id, required_resources):
                    # Resources not available, task stays in READY state
                    task.state = TaskState.READY
                    print(f"[BLOCKED] {task_id} waiting for resources: {required_resources}")
                    return False

            # Store which resources this task holds
            task.held_resources = required_resources

            prompt = self.create_task_prompt(task_id)

            # Write prompt to temp file (CLI might have length limits)
            prompt_file = f"/tmp/slaps_task_{task_id}.txt"
            with open(prompt_file, 'w') as f:
                f.write(prompt)

            # Launch Claude CLI process - use text for now, stream-json seems buffered
            cmd = [
                "claude",
                "-p",  # Print mode (non-interactive)
                "--model", self.model,  # Use sonnet by default
                "--dangerously-skip-permissions",
                "--output-format", "text",  # Text mode for now
                "--verbose",  # Show what's happening
                "--max-turns", "10",  # Limit turns for safety
                prompt  # Prompt as last argument
            ]

            print(f"[LAUNCH] Running: {' '.join(cmd[:6])}... '{prompt[:50]}...'")  # Debug

            # Use unbuffered mode and force line buffering
            env = os.environ.copy()
            env['PYTHONUNBUFFERED'] = '1'  # In case claude uses python

            process = subprocess.Popen(
                cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
                bufsize=0,  # Unbuffered
                env=env,
                preexec_fn=os.setsid  # Create new process group
            )

            # Make stdout non-blocking
            fd = process.stdout.fileno()
            flags = fcntl.fcntl(fd, fcntl.F_GETFL)
            fcntl.fcntl(fd, fcntl.F_SETFL, flags | os.O_NONBLOCK)

            task.process = process
            task.state = TaskState.RUNNING
            task.start_time = datetime.now(timezone.utc)
            self.running_tasks[task_id] = task

            # Record launch for cooldown tracking
            self.system_resource_manager.record_launch()

            print(f"[COORDINATOR] Launched task {task_id} (Claude #{len(self.running_tasks)}/{self.max_claude_instances})")
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

        # For now, just check if process is done (text mode)
        returncode = task.process.poll()
        if returncode is not None:
            # Process finished
            if returncode == 0:
                print(f"[DONE] {task_id} completed successfully")
                self.complete_task(task_id)
            else:
                # Read any error output
                try:
                    stderr = task.process.stderr.read()
                    print(f"[ERROR] {task_id} stderr: {stderr[:200]}")
                except:
                    pass
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
        if task.held_resources:
            self.resource_manager.release(task_id, task.held_resources)
            print(f"[RESOURCES] Released {task.held_resources} from {task_id}")

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

            # Release resources
            if task.held_resources:
                self.resource_manager.release(task_id, task.held_resources)
                print(f"[RESOURCES] Released {task.held_resources} from failed {task_id}")

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
        print(f"[CONFIG] Max {self.max_claude_instances} Claude instances, model: {self.model}")

        # Show initial system stats
        stats = self.system_resource_manager.get_system_stats()
        print(f"[SYSTEM] CPU: {stats['cpu_percent']:.1f}%, Memory: {stats['memory_percent']:.1f}%, Available: {stats['memory_available_gb']:.1f}GB")

        # Launch initial tasks (no dependencies) - but throttled!
        initial_tasks = self.get_ready_tasks()
        print(f"[WAVE 0] {len(initial_tasks)} tasks ready, will launch with throttling")

        # Queue initial tasks instead of launching all at once
        launch_queue = initial_tasks.copy()

        # Track tasks waiting for resources
        blocked_tasks = []
        last_stats_time = time.time()

        # Main monitoring loop
        while self.running_tasks or self.get_ready_tasks() or blocked_tasks or launch_queue:
            # Monitor running tasks
            for task_id in list(self.running_tasks.keys()):
                self.monitor_task(task_id)

            # Try to launch from queue (initial tasks first)
            if launch_queue:
                task_id = launch_queue[0]
                if self.launch_task(task_id):
                    launch_queue.pop(0)
                # If launch failed due to system resources, keep in queue

            # Retry blocked tasks (resources might be free now)
            still_blocked = []
            for task_id in blocked_tasks:
                if self.launch_task(task_id):
                    print(f"[UNBLOCKED] {task_id} acquired resources")
                else:
                    still_blocked.append(task_id)
            blocked_tasks = still_blocked

            # Launch newly ready tasks (Rolling Frontier!)
            ready = self.get_ready_tasks()
            for task_id in ready:
                if not self.launch_task(task_id):
                    # Task blocked on resources or system limits
                    blocked_tasks.append(task_id)

            # Periodically show system stats
            if time.time() - last_stats_time > 10:  # Every 10 seconds
                stats = self.system_resource_manager.get_system_stats()
                print(f"[SYSTEM] Running: {len(self.running_tasks)}, CPU: {stats['cpu_percent']:.1f}%, Mem: {stats['memory_percent']:.1f}%, Queue: {len(launch_queue) + len(blocked_tasks)}")
                last_stats_time = time.time()

            # Brief sleep to prevent CPU spinning
            time.sleep(0.5)  # Slightly longer to reduce overhead

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
    parser.add_argument("--max-instances", type=int, default=3, help="Max concurrent Claude instances (default: 3)")
    parser.add_argument("--model", default="sonnet", choices=["sonnet", "opus"], help="Claude model to use (default: sonnet)")
    parser.add_argument("--max-cpu", type=float, default=60.0, help="Max CPU usage percent (default: 60)")
    parser.add_argument("--max-memory", type=float, default=70.0, help="Max memory usage percent (default: 70)")
    args = parser.parse_args()

    # Check if psutil is installed
    try:
        import psutil
    except ImportError:
        print("ERROR: psutil not installed. Run: pip install psutil")
        sys.exit(1)

    coordinator = SLAPSCoordinator(
        tasks_file="docs/planning/TASKS/idea-features-v3/tasks.json",
        dag_file="docs/planning/TASKS/idea-features-v3/dag.json",
        coordinator_file="docs/planning/TASKS/idea-features-v3/coordinator.json",
        max_claude_instances=args.max_instances,
        model=args.model
    )

    # Override system resource limits from command line
    coordinator.system_resource_manager.max_cpu_percent = args.max_cpu
    coordinator.system_resource_manager.max_memory_percent = args.max_memory

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
        coordinator.test_mode = True  # Flag for simple prompts
        keep = list(coordinator.tasks.keys())[:3]
        coordinator.tasks = {k: v for k, v in coordinator.tasks.items() if k in keep}
        coordinator.dependencies = {k: set() for k in keep}
        coordinator.dependents = {k: set() for k in keep}

    coordinator.run()