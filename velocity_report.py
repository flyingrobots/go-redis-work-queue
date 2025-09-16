import subprocess
import re
from datetime import datetime, timedelta

print("=== SLAPS VELOCITY REPORT ===\n")

# Current status
completed = len(subprocess.check_output("ls slaps-coordination/finished-tasks/ 2>/dev/null", shell=True).decode().split())
open_tasks = len(subprocess.check_output("ls slaps-coordination/open-tasks/ 2>/dev/null", shell=True).decode().split())
active = len(subprocess.check_output("find slaps-coordination/claude-* -name '*.json' 2>/dev/null", shell=True).decode().split())

print(f"Current Status (as of {datetime.now().strftime('%H:%M:%S')}):")
print(f"  ✅ Completed: {completed}/88 tasks ({completed*100//88}%)")
print(f"  📋 Open: {open_tasks} tasks")
print(f"  🔄 Active: {active} tasks")
print(f"  📊 Total Progress: {completed + active}/88 in flight\n")

# Git log analysis
print("Velocity Timeline (from auto-commits):")
commits = subprocess.check_output("git log --oneline --since='7 hours ago' | grep 'auto-sync'", shell=True).decode().split('\n')
commits = [c for c in commits if c]

prev_count = None
prev_time = None
velocities = []

for commit in commits[:10]:
    match = re.search(r'(\w+) feat\(slaps\): auto-sync progress - (\d+) tasks done', commit)
    if match:
        hash_id = match.group(1)
        count = int(match.group(2))
        
        # Get timestamp
        timestamp = subprocess.check_output(f"git show -s --format=%ci {hash_id}", shell=True).decode().strip()
        dt = datetime.strptime(timestamp.split()[0] + ' ' + timestamp.split()[1], '%Y-%m-%d %H:%M:%S')
        
        if prev_count and prev_time:
            time_diff = (prev_time - dt).total_seconds() / 60  # minutes
            task_diff = prev_count - count
            if time_diff > 0:
                velocity = task_diff / (time_diff / 60)  # tasks per hour
                velocities.append(velocity)
        
        time_ago = datetime.now() - dt
        hours = int(time_ago.total_seconds() / 3600)
        mins = int((time_ago.total_seconds() % 3600) / 60)
        print(f"  {hours:2d}h {mins:2d}m ago: {count:2d} tasks | {timestamp.split()[1][:5]}")
        
        prev_count = count
        prev_time = dt

if velocities:
    avg_velocity = sum(velocities) / len(velocities)
    max_velocity = max(velocities)
    print(f"\nVelocity Metrics:")
    print(f"  ⚡ Peak: {max_velocity:.1f} tasks/hour")
    print(f"  📈 Average: {avg_velocity:.1f} tasks/hour")
    print(f"  🕒 Current: {velocities[0]:.1f} tasks/hour (last interval)")
    
    if avg_velocity > 0:
        remaining = 88 - completed
        eta_hours = remaining / avg_velocity
        print(f"\nProjection:")
        print(f"  🎯 Remaining: {remaining} tasks")
        print(f"  ⏱️  ETA: ~{eta_hours:.1f} hours at average velocity")
        print(f"  🏁 Completion: ~{(datetime.now() + timedelta(hours=eta_hours)).strftime('%H:%M')}")

# Recent completions
print("\nRecent Completions:")
recent = subprocess.check_output("ls -lt slaps-coordination/finished-tasks/*.json 2>/dev/null | head -5", shell=True).decode()
for line in recent.split('\n')[:5]:
    if line:
        parts = line.split()
        if len(parts) >= 9:
            print(f"  {parts[5]} {parts[6]} {parts[7]:>5} - {parts[8]}")

print("\n🏆 SLAPS Status: OPERATIONAL")
