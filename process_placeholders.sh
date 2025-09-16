#!/bin/bash

while true; do
    echo "[$(date)] Worker 3: Checking for tasks..."

    # Check for tasks in open-tasks
    TASKS=$(ls slaps-coordination/open-tasks/*.json 2>/dev/null)

    if [ ! -z "$TASKS" ]; then
        for TASK_PATH in $TASKS; do
            TASK_FILE=$(basename "$TASK_PATH")

            # Try to claim the task
            if mv "$TASK_PATH" "slaps-coordination/claude-003/$TASK_FILE" 2>/dev/null; then
                echo "Claimed $TASK_FILE"

                # Check if it's a placeholder
                if grep -q "F-TBD" "slaps-coordination/claude-003/$TASK_FILE"; then
                    echo "$TASK_FILE is a placeholder, moving to help-me"
                    jq '. + {"help_request": "Placeholder task with F-TBD. Needs proper specification."}' \
                        "slaps-coordination/claude-003/$TASK_FILE" > "/tmp/$TASK_FILE"
                    mv "/tmp/$TASK_FILE" "slaps-coordination/help-me/"
                else
                    echo "$TASK_FILE is a REAL TASK!"
                    echo "Task details:"
                    jq -r '.task.feature_id, .task.title, .task.description' "slaps-coordination/claude-003/$TASK_FILE"
                    exit 0  # Exit to process the real task
                fi
            fi
        done
    else
        echo "No tasks found. Sleeping 30 seconds..."
        sleep 30
    fi
done