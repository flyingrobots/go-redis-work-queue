#!/bin/bash

while true; do
    echo "[$(date)] Worker 3: Checking for open tasks..."

    # List tasks in open-tasks directory
    TASKS=$(ls slaps-coordination/open-tasks/*.json 2>/dev/null)

    if [ ! -z "$TASKS" ]; then
        for TASK_PATH in $TASKS; do
            TASK_FILE=$(basename "$TASK_PATH")
            echo "Found task: $TASK_FILE"

            # Try to claim the task
            if mv "$TASK_PATH" "slaps-coordination/claude-003/$TASK_FILE" 2>/dev/null; then
                echo "Successfully claimed $TASK_FILE"
                echo "Exiting monitor to process task..."
                exit 0
            else
                echo "Failed to claim $TASK_FILE (another worker got it)"
            fi
        done
    else
        echo "No open tasks found. Sleeping for 30 seconds..."
    fi

    sleep 30
done