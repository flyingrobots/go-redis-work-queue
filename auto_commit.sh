#!/bin/bash
# Auto-commit and push script for SLAPS coordination

while true; do
    echo "[$(date +%H:%M:%S)] Running auto-commit..."
    
    # Get current stats
    OPEN=$(ls slaps-coordination/open-tasks/ 2>/dev/null | wc -l | tr -d ' ')
    DONE=$(ls slaps-coordination/finished-tasks/ 2>/dev/null | wc -l | tr -d ' ')
    HELP=$(ls slaps-coordination/help-me/ 2>/dev/null | wc -l | tr -d ' ')
    
    # Stage all changes
    git add -A
    
    # Create commit message
    git commit -m "$(cat <<COMMIT_MSG
feat(slaps): auto-sync progress - $DONE tasks done, $OPEN open

Stats:
- Completed: $DONE tasks
- Open: $OPEN tasks  
- Help needed: $HELP tasks

🤖 Generated with [Claude Code](https://claude.ai/code)

Co-Authored-By: Claude <noreply@anthropic.com>
COMMIT_MSG
    )" 2>/dev/null
    
    if [ $? -eq 0 ]; then
        echo "[$(date +%H:%M:%S)] Committed changes"
        
        # Push to remote
        git push origin feat/ideas 2>&1 | grep -E "To |Everything up-to-date"
        
        if [ $? -eq 0 ]; then
            echo "[$(date +%H:%M:%S)] Pushed to remote"
        else
            echo "[$(date +%H:%M:%S)] Push failed (might need to set upstream)"
        fi
    else
        echo "[$(date +%H:%M:%S)] No changes to commit"
    fi
    
    # Wait 5 minutes before next commit
    sleep 300
done
