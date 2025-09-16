#!/usr/bin/env python3
"""
SLAPS File-Based Coordinator
Coordinates Claude workers via filesystem instead of subprocesses
"""

import json
import os
import time
import shutil
from datetime import datetime, timezone
from pathlib import Path
from typing import Dict, List, Set, Optional
import sys

class FileBasedSLAPSCoordinator:
    """Coordinates tasks via filesystem"""

    def __init__(self, base_dir: str = "slaps-coordination"):
        self.base_dir = Path(base_dir)

        # Create directory structure
        self.dirs = {
            'open_tasks': self.base_dir / 'open-tasks',
            'claimed': self.base_dir / 'claimed',
            'finished': self.base_dir / 'finished-tasks',
            'failed': self.base_dir / 'failed-tasks',
            'help_me': self.base_dir / 'help-me',
            'stats': self.base_dir / 'stats'
        }

        # Create all directories
        for dir_path in self.dirs.values():
            dir_path.mkdir(parents=True, exist_ok=True)

        # Create subdirs for each potential Claude worker
        for i in range(1, 11):  # Support up to 10 Claude workers
            (self.base_dir / f'claude-{i:03d}').mkdir(exist_ok=True)

        # Load tasks and dependencies
        self.load_tasks()

        # Track state
        self.completed_tasks = set()
        self.failed_tasks = set()
        self.in_progress = {}  # task_id -> claude_id
        self.help_requests = {}

    def load_tasks(self):
        """Load tasks from v3 planning artifacts"""
        with open('docs/planning/TASKS/idea-features-v3/tasks.json') as f:
            self.tasks_data = json.load(f)

        with open('docs/planning/TASKS/idea-features-v3/dag.json') as f:
            self.dag = json.load(f)

        # Build dependency graph
        self.tasks = {t['id']: t for t in self.tasks_data['tasks']}
        self.dependencies = {t['id']: set() for t in self.tasks_data['tasks']}
        self.dependents = {t['id']: set() for t in self.tasks_data['tasks']}

        for edge in self.dag.get('dependencies', []):
            from_task = edge['from']
            to_task = edge['to']
            if to_task in self.dependencies:
                self.dependencies[to_task].add(from_task)
            if from_task in self.dependents:
                self.dependents[from_task].add(to_task)

        print(f"[LOADED] {len(self.tasks)} tasks with dependencies")

    def get_ready_tasks(self) -> List[str]:
        """Get tasks with satisfied dependencies"""
        ready = []
        for task_id, deps in self.dependencies.items():
            if task_id not in self.completed_tasks and \
               task_id not in self.failed_tasks and \
               task_id not in self.in_progress and \
               all(d in self.completed_tasks for d in deps):
                ready.append(task_id)
        return ready

    def create_task_file(self, task_id: str) -> Dict:
        """Create task JSON for workers"""
        task = self.tasks[task_id]

        # Add coordination metadata
        task_data = {
            'task_id': task_id,
            'created_at': datetime.now(timezone.utc).isoformat(),
            'task': task,
            'dependencies_completed': list(self.dependencies[task_id] & self.completed_tasks),
            'instructions': {
                'on_complete': f"Move this file to finished-tasks/{task_id}.json",
                'on_failure': f"Move this file to failed-tasks/{task_id}.json with error details",
                'on_help_needed': f"Move to help-me/{task_id}.json with 'help_request' field explaining the issue",
                'working_dir': f"claude-XXX/{task_id}.json (where XXX is your worker ID)"
            }
        }

        return task_data

    def publish_ready_tasks(self):
        """Publish tasks with satisfied dependencies to open-tasks"""
        ready = self.get_ready_tasks()
        published = 0

        for task_id in ready[:5]:  # Limit concurrent tasks
            task_file = self.dirs['open_tasks'] / f"{task_id}.json"
            if not task_file.exists():
                task_data = self.create_task_file(task_id)
                with open(task_file, 'w') as f:
                    json.dump(task_data, f, indent=2)
                published += 1
                print(f"[PUBLISHED] {task_id} -> open-tasks/")

        return published

    def scan_directories(self):
        """Scan directories for status updates"""

        # Check for completed tasks
        for task_file in self.dirs['finished'].glob('*.json'):
            task_id = task_file.stem
            if task_id not in self.completed_tasks:
                self.completed_tasks.add(task_id)
                if task_id in self.in_progress:
                    del self.in_progress[task_id]
                print(f"[COMPLETED] {task_id}")

                # Trigger dependent tasks
                for dep_id in self.dependents.get(task_id, []):
                    print(f"[UNLOCK] {dep_id} (dependency {task_id} completed)")

        # Check for failed tasks
        for task_file in self.dirs['failed'].glob('*.json'):
            task_id = task_file.stem
            if task_id not in self.failed_tasks:
                self.failed_tasks.add(task_id)
                if task_id in self.in_progress:
                    del self.in_progress[task_id]

                # Read failure reason
                try:
                    with open(task_file) as f:
                        data = json.load(f)
                        reason = data.get('error', 'Unknown')
                        print(f"[FAILED] {task_id}: {reason}")
                except:
                    print(f"[FAILED] {task_id}")

        # Check for help requests
        for task_file in self.dirs['help_me'].glob('*.json'):
            task_id = task_file.stem
            if task_id not in self.help_requests:
                try:
                    with open(task_file) as f:
                        data = json.load(f)
                        help_msg = data.get('help_request', 'Needs assistance')
                        self.help_requests[task_id] = help_msg
                        print(f"[HELP] {task_id}: {help_msg}")
                        print(f"[HELP] Waiting for coordinator intervention...")

                        # Move back to open tasks with updated plan
                        # (This is where the main Claude would intervene)
                        response = input(f"Update plan for {task_id}? (y/n/skip): ")
                        if response.lower() == 'y':
                            # Update the task data
                            data['updated_plan'] = input("Enter updated instructions: ")
                            data['retry_count'] = data.get('retry_count', 0) + 1

                            # Move back to open tasks
                            updated_file = self.dirs['open_tasks'] / f"{task_id}.json"
                            with open(updated_file, 'w') as f:
                                json.dump(data, f, indent=2)
                            task_file.unlink()
                            del self.help_requests[task_id]
                            print(f"[RETRY] {task_id} with updated plan")
                except Exception as e:
                    print(f"[ERROR] Processing help request: {e}")

        # Check for claimed tasks (in worker directories)
        for i in range(1, 11):
            claude_dir = self.base_dir / f'claude-{i:03d}'
            for task_file in claude_dir.glob('*.json'):
                task_id = task_file.stem
                if task_id not in self.in_progress:
                    self.in_progress[task_id] = f'claude-{i:03d}'
                    print(f"[CLAIMED] {task_id} by claude-{i:03d}")

    def write_stats(self):
        """Write current statistics"""
        stats = {
            'timestamp': datetime.now(timezone.utc).isoformat(),
            'total_tasks': len(self.tasks),
            'completed': len(self.completed_tasks),
            'failed': len(self.failed_tasks),
            'in_progress': len(self.in_progress),
            'ready': len(self.get_ready_tasks()),
            'help_requests': len(self.help_requests),
            'completion_rate': f"{len(self.completed_tasks)/len(self.tasks)*100:.1f}%"
        }

        with open(self.dirs['stats'] / 'current.json', 'w') as f:
            json.dump(stats, f, indent=2)

        return stats

    def run(self):
        """Main coordination loop"""
        print(f"[START] File-based SLAPS Coordinator")
        print(f"[BASE] {self.base_dir}")
        print(f"[TASKS] {len(self.tasks)} total tasks")

        # Initial task publication
        self.publish_ready_tasks()

        try:
            while len(self.completed_tasks) < len(self.tasks):
                # Scan for updates
                self.scan_directories()

                # Publish newly ready tasks
                published = self.publish_ready_tasks()

                # Write stats
                stats = self.write_stats()

                # Display progress
                if time.time() % 10 < 1:  # Every ~10 seconds
                    print(f"[STATUS] Complete: {stats['completed']}/{stats['total_tasks']} | "
                          f"Active: {stats['in_progress']} | Ready: {stats['ready']} | "
                          f"Failed: {stats['failed']} | Help: {stats['help_requests']}")

                # Brief sleep
                time.sleep(1)

        except KeyboardInterrupt:
            print("\n[STOP] Coordinator interrupted")

        # Final stats
        print(f"\n[FINAL] Completed: {len(self.completed_tasks)}/{len(self.tasks)}")
        print(f"[FINAL] Failed: {len(self.failed_tasks)}")

if __name__ == "__main__":
    import argparse
    parser = argparse.ArgumentParser(description="File-based SLAPS Coordinator")
    parser.add_argument("--dir", default="slaps-coordination", help="Coordination directory")
    parser.add_argument("--reset", action="store_true", help="Reset coordination directory")
    args = parser.parse_args()

    if args.reset and Path(args.dir).exists():
        shutil.rmtree(args.dir)
        print(f"[RESET] Removed {args.dir}")

    coordinator = FileBasedSLAPSCoordinator(args.dir)
    coordinator.run()