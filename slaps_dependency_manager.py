#!/usr/bin/env python3
"""
SLAPS Dependency Manager
Stages all tasks and releases them as dependencies are satisfied
"""

import json
import shutil
from pathlib import Path
from datetime import datetime
from typing import Dict, List, Set
import time

class DependencyManager:
    def __init__(self):
        self.base_dir = Path('slaps-coordination')
        self.staged_dir = self.base_dir / 'staged-tasks'
        self.open_dir = self.base_dir / 'open-tasks'
        self.finished_dir = self.base_dir / 'finished-tasks'
        self.tasks_json = Path('docs/planning/TASKS/idea-features-v3/tasks.json')

        # Load task data
        with open(self.tasks_json, 'r') as f:
            self.data = json.load(f)

        # Build dependency graph
        self.dependencies = {}
        self.task_map = {}

        for task in self.data['tasks']:
            task_id = task['id']
            self.task_map[task_id] = task

            # Extract dependencies from task ID pattern
            # P1.T002 depends on P1.T001, P2.T006 depends on P2.T005, etc.
            phase, num = task_id.split('.')
            num = int(num[1:])

            deps = []
            if task_id == 'P1.T002':  # Admin API Implementation
                deps = ['P1.T001']  # Depends on Design
            elif task_id == 'P1.T003':  # Admin API Test
                deps = ['P1.T002']  # Depends on Implementation
            elif task_id == 'P1.T004':  # Admin API Deploy
                deps = ['P1.T003']  # Depends on Test
            elif task_id == 'P2.T006':  # Multi-cluster Implementation
                deps = ['P2.T005']  # Depends on Design
            elif task_id == 'P2.T007':  # Multi-cluster Test
                deps = ['P2.T006']  # Depends on Implementation
            elif task_id == 'P2.T009':  # DAG Builder Implementation
                deps = ['P2.T008']  # Depends on Design
            elif task_id == 'P2.T010':  # DAG Builder Test
                deps = ['P2.T009']  # Depends on Implementation
            elif task_id == 'P1.T013':  # Tracing Test
                deps = ['P1.T012']  # Depends on Implementation
            # Add more dependency rules as needed

            self.dependencies[task_id] = deps

    def stage_all_tasks(self):
        """Create all task files in staging directory"""
        self.staged_dir.mkdir(parents=True, exist_ok=True)

        print("=== Staging all tasks ===")
        for task in self.data['tasks']:
            task_id = task['id']

            # Create full task specification
            task_data = {
                'task_id': task_id,
                'created_at': datetime.utcnow().isoformat() + 'Z',
                'task': task,  # Include FULL task spec
                'dependencies': self.dependencies.get(task_id, []),
                'instructions': {
                    'on_complete': f'Move this file to finished-tasks/{task_id}.json',
                    'on_failure': f'Move this file to failed-tasks/{task_id}.json',
                    'on_help_needed': f'Move to help-me/{task_id}.json',
                    'note': 'READ THE FULL TASK SPECIFICATION! All details are in the task field.',
                    'resource_locks': task.get('shared_resources', {})
                }
            }

            # Write to staging
            output_path = self.staged_dir / f'{task_id}.json'
            with open(output_path, 'w') as f:
                json.dump(task_data, f, indent=2)

        print(f"Staged {len(self.data['tasks'])} tasks")

    def get_completed_tasks(self) -> Set[str]:
        """Get set of completed task IDs"""
        completed = set()
        if self.finished_dir.exists():
            for f in self.finished_dir.glob('*.json'):
                task_id = f.stem
                completed.add(task_id)
        return completed

    def get_open_tasks(self) -> Set[str]:
        """Get set of currently open task IDs"""
        open_tasks = set()
        if self.open_dir.exists():
            for f in self.open_dir.glob('*.json'):
                open_tasks.add(f.stem)
        return open_tasks

    def get_ready_tasks(self) -> List[str]:
        """Get tasks whose dependencies are satisfied"""
        completed = self.get_completed_tasks()
        open_tasks = self.get_open_tasks()
        staged = set(f.stem for f in self.staged_dir.glob('*.json'))

        ready = []
        for task_id in staged:
            if task_id in open_tasks:
                continue  # Already published

            deps = self.dependencies.get(task_id, [])
            if all(dep in completed for dep in deps):
                ready.append(task_id)

        return ready

    def publish_ready_tasks(self):
        """Move ready tasks from staging to open"""
        ready = self.get_ready_tasks()

        if ready:
            print(f"\n=== Publishing {len(ready)} ready tasks ===")
            for task_id in ready:
                src = self.staged_dir / f'{task_id}.json'
                dst = self.open_dir / f'{task_id}.json'

                if src.exists():
                    shutil.move(str(src), str(dst))
                    print(f"Published: {task_id}")

        return len(ready)

    def monitor_loop(self):
        """Continuously monitor and publish tasks as dependencies are met"""
        print("\n=== Starting dependency monitor ===")

        while True:
            completed = self.get_completed_tasks()
            open_tasks = self.get_open_tasks()
            staged = set(f.stem for f in self.staged_dir.glob('*.json'))

            print(f"\n[{datetime.now().strftime('%H:%M:%S')}] Status:")
            print(f"  Completed: {len(completed)}")
            print(f"  Open: {len(open_tasks)}")
            print(f"  Staged: {len(staged)}")

            # Publish any newly ready tasks
            published = self.publish_ready_tasks()

            if published > 0:
                print(f"  -> Published {published} new tasks")

            # Check if all tasks are done
            if len(completed) == len(self.data['tasks']):
                print("\n🎉 All tasks completed!")
                break

            time.sleep(5)

if __name__ == '__main__':
    manager = DependencyManager()

    # Stage all tasks
    manager.stage_all_tasks()

    # Publish initial wave (tasks with no dependencies)
    print("\n=== Publishing initial wave ===")
    manager.publish_ready_tasks()

    # Start monitoring loop
    manager.monitor_loop()