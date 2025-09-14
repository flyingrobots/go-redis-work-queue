#!/usr/bin/env python3
"""
Claude Worker - Claims and executes tasks from the coordination directory
Run this in a Claude instance to become a worker
"""

import json
import os
import time
import shutil
from pathlib import Path
from datetime import datetime, timezone
import random
import sys

class ClaudeWorker:
    def __init__(self, worker_id: int, base_dir: str = "slaps-coordination"):
        self.worker_id = worker_id
        self.worker_name = f"claude-{worker_id:03d}"
        self.base_dir = Path(base_dir)

        # Directory paths
        self.open_tasks_dir = self.base_dir / 'open-tasks'
        self.my_dir = self.base_dir / self.worker_name
        self.finished_dir = self.base_dir / 'finished-tasks'
        self.failed_dir = self.base_dir / 'failed-tasks'
        self.help_dir = self.base_dir / 'help-me'

        # Ensure my directory exists
        self.my_dir.mkdir(parents=True, exist_ok=True)

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
                except Exception as e:
                    print(f"[ERROR] Claiming {task_file.name}: {e}")

        except Exception as e:
            print(f"[ERROR] Scanning tasks: {e}")

        return None

    def execute_task(self, task_file: Path) -> bool:
        """Execute the claimed task"""
        try:
            # Load task
            with open(task_file) as f:
                task_data = json.load(f)

            task_id = task_data['task_id']
            task = task_data['task']

            print(f"\n{'='*60}")
            print(f"[EXECUTE] Task {task_id}: {task['title']}")
            print(f"[DESC] {task['description']}")

            # Show requirements
            if 'boundaries' in task:
                print(f"\n[REQUIREMENTS]")
                for criterion in task['boundaries']['definition_of_done']['criteria'][:3]:
                    print(f"  - {criterion}")
                print(f"\n[STOP WHEN] {task['boundaries']['definition_of_done']['stop_when']}")

            print(f"\n[INSTRUCTION] Please execute this task now.")
            print(f"[INSTRUCTION] When complete, type 'done'")
            print(f"[INSTRUCTION] If you need help, type 'help'")
            print(f"[INSTRUCTION] If task fails, type 'failed'")
            print(f"{'='*60}\n")

            # Wait for user (Claude) to complete the task
            while True:
                response = input(f"[{self.worker_name}] Status (done/help/failed): ").strip().lower()

                if response == 'done':
                    # Mark as completed
                    task_data['completed_at'] = datetime.now(timezone.utc).isoformat()
                    task_data['completed_by'] = self.worker_name

                    # Get summary of what was done
                    summary = input("Brief summary of what you did: ")
                    task_data['completion_summary'] = summary

                    # Move to finished
                    finished_file = self.finished_dir / task_file.name
                    with open(finished_file, 'w') as f:
                        json.dump(task_data, f, indent=2)
                    task_file.unlink()

                    print(f"[COMPLETED] {task_id}")
                    return True

                elif response == 'help':
                    # Request help
                    help_msg = input("What help do you need? ")
                    task_data['help_request'] = help_msg
                    task_data['help_requested_at'] = datetime.now(timezone.utc).isoformat()
                    task_data['help_requested_by'] = self.worker_name

                    # Move to help directory
                    help_file = self.help_dir / task_file.name
                    with open(help_file, 'w') as f:
                        json.dump(task_data, f, indent=2)
                    task_file.unlink()

                    print(f"[HELP] Request sent for {task_id}")
                    return False

                elif response == 'failed':
                    # Mark as failed
                    error_msg = input("Error details: ")
                    task_data['error'] = error_msg
                    task_data['failed_at'] = datetime.now(timezone.utc).isoformat()
                    task_data['failed_by'] = self.worker_name

                    # Move to failed
                    failed_file = self.failed_dir / task_file.name
                    with open(failed_file, 'w') as f:
                        json.dump(task_data, f, indent=2)
                    task_file.unlink()

                    print(f"[FAILED] {task_id}: {error_msg}")
                    return False

                else:
                    print(f"[INVALID] Please enter 'done', 'help', or 'failed'")

        except Exception as e:
            print(f"[ERROR] Executing task: {e}")
            # Move to failed on error
            try:
                task_data['error'] = str(e)
                task_data['failed_at'] = datetime.now(timezone.utc).isoformat()
                failed_file = self.failed_dir / task_file.name
                with open(failed_file, 'w') as f:
                    json.dump(task_data, f, indent=2)
                task_file.unlink()
            except:
                pass
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