#!/bin/bash
set -Eeuo pipefail
IFS=$'\n\t'

# Auto-commit and push script for SLAPS coordination
MAX_ITERATIONS=${MAX_ITERATIONS:-0}
SLEEP_SECONDS=${SLEEP_SECONDS:-300}
STOP_REQUESTED=0

handle_signal() {
  echo "[$(date +%H:%M:%S)] Stop signal received; exiting after current iteration." >&2
  STOP_REQUESTED=1
}

trap handle_signal INT TERM

iteration=0

count_files() {
  local dir="$1"
  if [[ -d "$dir" ]]; then
    find "$dir" -mindepth 1 -maxdepth 1 -type f 2>/dev/null | wc -l | tr -d ' '
  else
    echo 0
  fi
}

while true; do
  if (( STOP_REQUESTED )); then
    echo "[$(date +%H:%M:%S)] Stop requested; exiting loop." >&2
    break
  fi

  if (( MAX_ITERATIONS > 0 && iteration >= MAX_ITERATIONS )); then
    echo "[$(date +%H:%M:%S)] Reached MAX_ITERATIONS=$MAX_ITERATIONS; exiting." >&2
    break
  fi

  iteration=$((iteration + 1))
  echo "[$(date +%H:%M:%S)] Running auto-commit (iteration $iteration)..."

  OPEN=$(count_files "slaps-coordination/open-tasks/")
  DONE=$(count_files "slaps-coordination/finished-tasks/")
  HELP=$(count_files "slaps-coordination/help-me/")

  git add -A

  if git diff --cached --quiet; then
    echo "[$(date +%H:%M:%S)] No staged changes; skipping commit."
  else
    commit_message=$(cat <<COMMIT_MSG
chore(slaps): auto-sync progress - $DONE done / $OPEN open

Stats:
- Completed: $DONE tasks
- Open: $OPEN tasks
- Help needed: $HELP tasks
COMMIT_MSG
)
    if git commit -m "$commit_message"; then
      echo "[$(date +%H:%M:%S)] Committed changes"
      current_branch=$(git rev-parse --abbrev-ref HEAD)
      if git rev-parse --symbolic-full-name @{u} >/dev/null 2>&1; then
        if git push origin "$current_branch"; then
          echo "[$(date +%H:%M:%S)] Pushed $current_branch to origin"
        else
          echo "[$(date +%H:%M:%S)] Push failed for $current_branch" >&2
        fi
      else
        if git push --set-upstream origin "$current_branch"; then
          echo "[$(date +%H:%M:%S)] Set upstream and pushed $current_branch"
        else
          echo "[$(date +%H:%M:%S)] Failed to set upstream for $current_branch" >&2
        fi
      fi
    else
      echo "[$(date +%H:%M:%S)] Commit failed" >&2
    fi
  fi

  if (( STOP_REQUESTED )); then
    echo "[$(date +%H:%M:%S)] Stop requested after commit; exiting."
    break
  fi

  echo "[$(date +%H:%M:%S)] Sleeping for ${SLEEP_SECONDS}s"
  sleep "$SLEEP_SECONDS" || true
  if (( STOP_REQUESTED )); then
    echo "[$(date +%H:%M:%S)] Stop requested during sleep; exiting."
    break
  fi

done
