#!/usr/bin/env python3
"""
Claude Worker - Claims and executes tasks from the coordination directory
Run this in a Claude instance to become a worker
"""

import json
import os
import time
from enum import Enum
from pathlib import Path
from datetime import datetime, timezone
from typing import Callable, Optional, Tuple
import random
import sys

class TaskStatus(Enum):
    COMPLETED = "completed"
    HELP = "help"
    FAILED = "failed"


class ClaudeWorker:
    def __init__(
        self,
        worker_id: int,
        base_dir: str = "slaps-coordination",
        executor: Optional[Callable[[dict], Tuple[TaskStatus, str]]] = None,
    ):
        self.worker_id = worker_id
        self.worker_name = f"claude-{worker_id:03d}"
        self.base_dir = Path(base_dir)
        self.executor = executor or self._default_executor

        # Directory paths
        self.open_tasks_dir = self.base_dir / 'open-tasks'
        self.my_dir = self.base_dir / self.worker_name
        self.finished_dir = self.base_dir / 'finished-tasks'
        self.failed_dir = self.base_dir / 'failed-tasks'
        self.help_dir = self.base_dir / 'help-me'

        # Ensure required directories exist
        for path in (
            self.open_tasks_dir,
            self.my_dir,
            self.finished_dir,
            self.failed_dir,
            self.help_dir,
        ):
            path.mkdir(parents=True, exist_ok=True)

        print(f"[WORKER] {self.worker_name} initialized")
        print(f"[WORKER] Watching: {self.open_tasks_dir}")

    def claim_task(self) -> Optional[Path]:
        """Try to claim an available task"""
        try:
            # List available tasks
            tasks = list(self.open_tasks_dir.glob('*.json'))
            if not tasks:
                return None

            # Try to claim a random task (avoid collision)
            random.shuffle(tasks)
            for task_file in tasks:
                try:
                    # Atomic move to claim
                    my_task_file = self.my_dir / task_file.name
                    task_file.rename(my_task_file)
                    print(f"[CLAIMED] {task_file.stem}")
                    return my_task_file
                except FileNotFoundError:
                    # Someone else got it
                    continue
                except Exception as err:
                    print(f"[ERROR] Claiming {task_file.name}: {err}")

        except Exception as err:
            print(f"[ERROR] Scanning tasks: {err}")

        return None

    def execute_task(self, task_file: Path) -> bool:
        """Execute the claimed task"""
        try:
            # Load task
            with open(task_file) as handle:
                task_data = json.load(handle)

            task_id = task_data['task_id']
            task = task_data['task']

            print(f"\n{'='*60}")
            print(f"[EXECUTE] Task {task_id}: {task['title']}")
            print(f"[DESC] {task['description']}")
            status, message = self.executor(task)

            if status is TaskStatus.COMPLETED:
                task_data['completed_at'] = datetime.now(timezone.utc).isoformat()
                task_data['completed_by'] = self.worker_name
                task_data['completion_summary'] = message or ""
                destination = self.finished_dir / task_file.name
                self._persist_task(destination, task_data)
                task_file.unlink(missing_ok=True)
                print(f"[COMPLETED] {task_id}")
                return True

            if status is TaskStatus.HELP:
                task_data['help_request'] = message or "Assistance requested"
                task_data['help_requested_at'] = datetime.now(timezone.utc).isoformat()
                task_data['help_requested_by'] = self.worker_name
                destination = self.help_dir / task_file.name
                self._persist_task(destination, task_data)
                task_file.unlink(missing_ok=True)
                print(f"[HELP] Logged request for {task_id}")
                return False

            if status is TaskStatus.FAILED:
                task_data['error'] = message or "Task execution failed"
                task_data['failed_at'] = datetime.now(timezone.utc).isoformat()
                task_data['failed_by'] = self.worker_name
                destination = self.failed_dir / task_file.name
                self._persist_task(destination, task_data)
                task_file.unlink(missing_ok=True)
                print(f"[FAILED] {task_id}: {task_data['error']}")
                return False

            print(f"[WARN] Unknown status {status} for task {task_id}; moving to help queue")
            return False

        except (json.JSONDecodeError, OSError) as err:
            print(f"[ERROR] Reading task file {task_file.name}: {err}")
            return False
        except Exception as err:
            print(f"[ERROR] Executing task: {err}")
            try:
                task_data = task_data if 'task_data' in locals() else {"task_id": task_file.stem}
                task_data['error'] = str(err)
                task_data['failed_at'] = datetime.now(timezone.utc).isoformat()
                task_data['failed_by'] = self.worker_name
                destination = self.failed_dir / task_file.name
                self._persist_task(destination, task_data)
                task_file.unlink(missing_ok=True)
            except OSError as persist_err:
                print(f"[ERROR] Writing failure record: {persist_err}")
            return False

    def run(self):
        """Main worker loop"""
        print(f"[START] {self.worker_name} worker loop")
        print(f"[START] Polling every 3 seconds for tasks...")

        try:
            while True:
                # Check for available tasks
                task_file = self.claim_task()

                if task_file:
                    # Execute the task
                    success = self.execute_task(task_file)
                    print(f"[READY] Looking for next task...")
                else:
                    # No tasks available, wait
                    print(".", end="", flush=True)
                    time.sleep(3)

        except KeyboardInterrupt:
            print(f"\n[STOP] {self.worker_name} shutting down")

    def _persist_task(self, destination: Path, payload: dict) -> None:
        destination.parent.mkdir(parents=True, exist_ok=True)
        with open(destination, "w", encoding="utf-8") as handle:
            json.dump(payload, handle, indent=2)

    def _default_executor(self, task: dict) -> Tuple[TaskStatus, str]:
        return TaskStatus.HELP, "Manual intervention required (no executor configured)"

if __name__ == "__main__":
    import argparse
    parser = argparse.ArgumentParser(description="Claude Worker")
    parser.add_argument("--id", type=int, required=True, help="Worker ID (1-10)")
    parser.add_argument("--dir", default="slaps-coordination", help="Coordination directory")
    args = parser.parse_args()

    if args.id < 1 or args.id > 10:
        print("Worker ID must be between 1 and 10")
        sys.exit(1)

    worker = ClaudeWorker(args.id, args.dir)
    worker.run()
