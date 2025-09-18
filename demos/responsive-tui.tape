# VHS Tape: Responsive TUI Breakpoints Demo
# Shows how the TUI adapts across different terminal sizes

Set Theme "Tokyo Night"
Set FontFamily "Fira Code"
Set FontSize 12
Set Width 1400
Set Height 900
Set TypingSpeed 80ms
Set Shell "zsh"

# Introduction
Type "# 📱➡️🖥️ Responsive Terminal UI Demo"
Enter
Type "# Redis Work Queue TUI adapts to any screen size"
Enter
Clear

# Mobile Mode Demo (≤40 cols)
Type "echo '📱 MOBILE MODE (≤40 cols)'"
Enter
Type "echo 'Stack panels vertically, card-based UI'"
Enter
Enter
Type "echo 'Note: COLUMNS override is cosmetic; the real TUI toggles modes via breakpoints.'"
Enter
Type "export COLUMNS=35"
Enter
Type "cat << 'EOF'"
Enter
Type "┌─ Queue Monitor ─┐"
Enter
Type "│                 │"
Enter
Type "│ ⚡ Current Queue │"
Enter
Type "│ ┌─────────────┐ │"
Enter
Type "│ │ prod-jobs   │ │"
Enter
Type "│ │ 1,234 jobs  │ │"
Enter
Type "│ │ ▓▓▓▓▓▓▓░░   │ │"
Enter
Type "│ │ 📈 +15/min  │ │"
Enter
Type "│ └─────────────┘ │"
Enter
Type "│                 │"
Enter
Type "│ ⚠️  Alerts      │"
Enter
Type "│ • 3 DLQ jobs    │"
Enter
Type "│ • Worker #2 slow│"
Enter
Type "│                 │"
Enter
Type "│ 🎤 \"Hey Queue\"   │"
Enter
Type "│ for voice cmds  │"
Enter
Type "│                 │"
Enter
Type "│ ◉ ○ ○ ○         │"
Enter
Type "│ Swipe → for more│"
Enter
Type "└─────────────────┘"
Enter
Type "EOF"
Enter
Clear

# Tablet Mode Demo (41-80 cols)
Type "echo '📟 TABLET MODE (41-80 cols)'"
Enter
Type "echo 'Two-column layout with tabs'"
Enter
Enter
Type "echo 'Note: COLUMNS override is cosmetic; the actual TUI reflows via breakpoints.'"
Enter
Type "export COLUMNS=75"
Enter
Type "cat << 'EOF'"
Enter
Type "┌─ Redis Work Queue Monitor ───────────────────────────────────────┐"
Enter
Type "│ ▶ Jobs    Workers    DLQ    Debug           Connected ✓ 12 ↗   │"
Enter
Type "├───────────────────────────────────────────────────────────────────┤"
Enter
Type "│ ┌─ Queues ──────────────────┐ ┌─ Quick Actions ─────────────────┐│"
Enter
Type "│ │ Name           Jobs Status│ │                                 ││"
Enter
Type "│ │ ────────────────────────  │ │  [Pause All]  [Resume All]     ││"
Enter
Type "│ │ ▶ prod-jobs     1,234  ✓ │ │  [Peek Jobs]  [Clear DLQ]      ││"
Enter
Type "│ │   emails          789  ✓ │ │                                 ││"
Enter
Type "│ │   processing       45  ○ │ │ Selected: prod-jobs             ││"
Enter
Type "│ │   cleanup          23  ○ │ │ Rate: 45 jobs/min               ││"
Enter
Type "│ └───────────────────────────┘ │ Workers: 12 active              ││"
Enter
Type "│                               └─────────────────────────────────┘│"
Enter
Type "│ ┌─ System Health ─────────────┐ ┌─ Recent Activity ─────────────┐│"
Enter
Type "│ │ Redis: ✓ Connected          │ │ 14:23 Worker #3 started       ││"
Enter
Type "│ │ Memory: ▓▓▓▓░░░ 45%         │ │ 14:22 Queue drained safely    ││"
Enter
Type "│ │ CPU: ▓▓▓░░░░░░ 28%          │ │ 14:21 DLQ job requeued        ││"
Enter
Type "│ │ Network: ▓▓▓▓▓▓░ 62%        │ │ 14:20 Batch job completed     ││"
Enter
Type "│ └─────────────────────────────┘ └───────────────────────────────┘│"
Enter
Type "├───────────────────────────────────────────────────────────────────┤"
Enter
Type "│ [Enter] Peek • [/] Filter • [Tab] Switch • [Ctrl+P] Commands      │"
Enter
Type "└───────────────────────────────────────────────────────────────────┘"
Enter
Type "EOF"
Enter
Clear

# Desktop Mode Demo (81-120 cols)
Type "echo '🖥️  DESKTOP MODE (81-120 cols)'"
Enter
Type "echo 'Multi-panel dashboard with full features'"
Enter
Enter
Type "echo 'Note: Demo uses COLUMNS for visuals; desktop layout is handled by TUI breakpoints.'"
Enter
Type "export COLUMNS=115"
Enter
Type "cat << 'EOF'"
Enter
Type "┌─ 🔄 Redis Work Queue Monitor ─────────────────────────────────────────────────────────────────────┐"
Enter
Type "│ Production Environment • Connected to redis://prod:6379 • Desktop Mode                            │"
Enter
Type "└───────────────────────────────────────────────────────────────────────────────────────────────────┘"
Enter
Type "┌─ System Status ───────────────────────────────────────────────────────────────────────────────────┐"
Enter
Type "│ ✓ Connected   Queues: 5 ↗   Workers: 12 ↗   DLQ: ⚠ 3   Throughput: 245/min ↗   Updated: now    │"
Enter
Type "└───────────────────────────────────────────────────────────────────────────────────────────────────┘"
Enter
Type "▶ ⚡ Jobs         👥 Workers        💀 DLQ           ⏰ Debug         🔗 Events       ⚙️  Settings"
Enter
Type "═══════════════════════════════════════════════════════════════════════════════════════════════════════"
Enter
Type "┌─ 📋 Active Queues ────────────────────┐ ┌─ 📊 Processing Metrics ─────────────────────────────────┐"
Enter
Type "│ Queue Name           Jobs    Rate     │ │ Throughput (jobs/minute)                                │"
Enter
Type "│ ──────────────────────────────────────│ │      ▃▄▆█▆▄▃▂▁▂▃▄▆█▆▄▃▂▁▂▃▄▆█▆▄▃▂▁                   │"
Enter
Type "│ ▶ prod-jobs          1,234   45/min ✓│ │      ▁▂▃▄▆█▆▄▃▂▁▂▃▄▆█▆▄▃▂▁▂▃▄▆█▆                     │"
Enter
Type "│   async-emails         789   12/min ✓│ │                                                         │"
Enter
Type "│   data-processing       45    8/min ○│ │ Current: 245 j/min    Peak: 487 j/min                  │"
Enter
Type "│   notifications        156   22/min ⚠│ │ Avg Response: 1.2s    P95: 3.4s                        │"
Enter
Type "│   cleanup-tasks         23    3/min ○│ └─────────────────────────────────────────────────────────┘"
Enter
Type "└───────────────────────────────────────┘ ┌─ 👥 Worker Fleet Status ────────────────────────────────┐"
Enter
Type "┌─ Quick Actions ───────────────────────┐ │ Fleet Health: ✓ All Systems Operational                │"
Enter
Type "│  [Pause All]  [Resume All]  [Scale]   │ │ Active Workers: 12/15 (80% utilization)                │"
Enter
Type "│  [Peek Queue] [Drain Safe]   [Debug]  │ │                                                         │"
Enter
Type "│                                       │ │ Resource Usage:                                         │"
Enter
Type "│ 🎯 Selected: prod-jobs                │ │ CPU:    ▓▓▓▓▓▓░░░░░░░░░ 42% (avg across fleet)        │"
Enter
Type "│ 📈 Processing: 45 jobs/min            │ │ Memory: ▓▓▓▓▓▓▓▓░░░░░░░ 56% (12.8GB / 23GB)           │"
Enter
Type "│ ⚡ Status: Active & Healthy           │ │ Network:▓▓▓▓▓▓▓░░░░░░░░ 48% (245MB/s)                  │"
Enter
Type "│ 🎤 Voice: \"Hey Queue, show status\"    │ │ Disk I/O:▓▓▓░░░░░░░░░░░░ 18% (low activity)             │"
Enter
Type "└───────────────────────────────────────┘ └─────────────────────────────────────────────────────────┘"
Enter
Type "┌─ 📝 Recent Activity ──────────────────┐ ┌─ ❤️  System Health ─────────────────────────────────────┐"
Enter
Type "│ 14:23:45 Worker #3 started processing │ │ Overall Status: ✓ HEALTHY                              │"
Enter
Type "│ 14:23:32 Queue 'emails' drained safely│ │ Redis Connection: ✓ Connected (2ms latency)            │"
Enter
Type "│ 14:23:18 DLQ job requeued successfully│ │ Worker Fleet: ✓ 12/15 active (no failures)            │"
Enter
Type "│ 14:23:05 Batch processing completed   │ │ Queue Backlog: ✓ Under control (avg 45 jobs)          │"
Enter
Type "│ 14:22:51 Auto-scaling triggered       │ │ Error Rate: ✓ Low (0.02% failure rate)                 │"
Enter
Type "└───────────────────────────────────────┘ └─────────────────────────────────────────────────────────┘"
Enter
Type "═══════════════════════════════════════════════════════════════════════════════════════════════════════"
Enter
Type "[Enter] Peek Queue • [Space] Pause/Resume • [D] Drain • [/] Filter • [Tab] Focus • [Ctrl+P] Commands"
Enter
Type "EOF"
Enter
Clear

# Ultrawide Mode Demo (121+ cols)
Type "echo '🖥️🖥️ ULTRAWIDE MODE (121+ cols)'"
Enter
Type "echo 'Mission control layout with side-by-side multi-view'"
Enter
Enter
Type "echo 'Note: COLUMNS export here is illustrative; production UI switches layouts internally.'"
Enter
Type "export COLUMNS=140"
Enter
Type "cat << 'EOF'"
Enter
Type "┌─ 🔄 Redis Work Queue Monitor ─ Mission Control Mode ─────────────────────────────────────────────────────────────────────────────┐"
Enter
Type "│ Production Environment • Multi-Cluster • Connected to 3 Redis instances • Ultrawide Mode (140 cols)                              │"
Enter
Type "└───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘"
Enter
Type "┌─ Multi-Cluster Status ────────────────────────────────────────────────────────────────────────────────────────────────────────────┐"
Enter
Type "│ 🟢 us-east-1 (✓ 245/min)  🟢 us-west-2 (✓ 189/min)  🟡 eu-west-1 (⚠ 89/min)  📊 Global: 523 jobs/min  🎯 SLA: 99.97%        │"
Enter
Type "└───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘"
Enter
Type "▶ ⚡ Multi-Queue    👥 Fleet Control    💀 Global DLQ    ⏰ Time Travel    🌍 Clusters    📊 Analytics    🔧 Operations"
Enter
Type "═══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════"
Enter
Type "┌─ 📋 Queue Dashboard ──────────────────────────────┐│┌─ 🎛️  Fleet Operations ────────────────────────────────────────────────────┐"
Enter
Type "│ Cluster: us-east-1         Cluster: us-west-2    ││││ Global Worker Status                                                       │"
Enter
Type "│ ┌─────────────────┐        ┌─────────────────┐   │││├─────────────────────────────────────────────────────────────────────────────┤"
Enter
Type "│ │▶prod-jobs  1,234│        │▶prod-jobs    892│   │││ Region      │Workers│Status  │CPU   │Memory │Throughput │Actions         │"
Enter
Type "│ │ emails       789│        │ emails       456│   │││ us-east-1   │  15   │✓ Healthy│ 45%  │  67%  │  245/min  │[Scale][Drain]  │"
Enter
Type "│ │ processing    45│        │ processing    23│   │││ us-west-2   │  12   │✓ Healthy│ 38%  │  52%  │  189/min  │[Scale][Drain]  │"
Enter
Type "│ │ cleanup       23│        │ cleanup       12│   │││ eu-west-1   │   8   │⚠ Degraded│ 72%  │  81%  │   89/min  │[⚠Add][Debug]   │"
Enter
Type "│ └─────────────────┘        └─────────────────┘   │││ ap-south-1  │   6   │○ Idle   │ 12%  │  23%  │   12/min  │[Scale][Status] │"
Enter
Type "├───────────────────────────────────────────────────┤││                                                                        │"
Enter
Type "│ Cluster: eu-west-1         Cluster: ap-south-1   │││ 🚀 Fleet Actions:                                                      │"
Enter
Type "│ ┌─────────────────┐        ┌─────────────────┐   │││ [Auto-Scale All] [Emergency Drain] [Rolling Restart] [Health Check]  │"
Enter
Type "│ │ prod-jobs    234│        │ prod-jobs     45│   │││                                                                        │"
Enter
Type "│ │⚠emails       156│        │○emails         8│   │││ 🎯 Quick Actions:                                                      │"
Enter
Type "│ │ processing    12│        │ processing     2│   │││ • Scale us-east-1 to handle eu-west-1 overflow                       │"
Enter
Type "│ │ cleanup        5│        │ cleanup        1│   │││ • Investigate eu-west-1 performance degradation                      │"
Enter
Type "│ └─────────────────┘        └─────────────────┘   │││ • Add workers to ap-south-1 for upcoming traffic                     │"
Enter
Type "└───────────────────────────────────────────────────┘│└────────────────────────────────────────────────────────────────────────────┘"
Enter
Type "┌─ 📊 Global Analytics ─────────────────────────────┐│┌─ ⏰ Time Travel Debugger ──────────────────────────────────────────────────┐"
Enter
Type "│ Traffic Pattern (24h)                             ││││ 🔍 Job Execution Replay                                                    │"
Enter
Type "│ ╭─────────────────────────────────────────────────╮│││ ┌─────────────┐ job_id: abc123                                          │"
Enter
Type "│ │        ▆▆██                          ▆█         │││ │ ⏮  ⏪  ⏸  ⏩ │ status: investigating failure                        │"
Enter
Type "│ │     ▁▃███████▅▃▁                ▃▅████████       │││ └─────────────┘                                                        │"
Enter
Type "│ │  ▁▂██████████████▇▅▂          ▂▆█████████████    │││ Timeline:                                                              │"
Enter
Type "│ │▂▄███████████████████▇▄▁    ▁▃▆███████████████▅   │││ 14:23:01 ✓ Job enqueued (us-east-1)                                   │"
Enter
Type "│ ╰─────────────────────────────────────────────────╯│││ 14:23:02 ✓ Worker assigned (worker-15)                               │"
Enter
Type "│ Peak: 1,247 j/min  Current: 523 j/min            ││││ 14:23:05 ⚠ External API timeout (retry 1/3)                          │"
Enter
Type "│ Prediction: Traffic spike at 16:00 EST           ││││ 14:23:08 ⚠ External API timeout (retry 2/3)                          │"
Enter
Type "│ Recommendation: Pre-scale us-east-1 by +5 workers│││ 14:23:11 ✗ Job failed, moved to DLQ                                  │"
Enter
Type "└───────────────────────────────────────────────────┘│└────────────────────────────────────────────────────────────────────────────┘"
Enter
Type "═══════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════════"
Enter
Type "[F1] Help • [F2] Clusters • [F3] Debug • [Ctrl+P] Command Palette • [Ctrl+V] Voice • [Tab] Focus • [Esc] Menu"
Enter
Type "EOF"
Enter
Clear

# Feature summary
Type "echo '✨ Responsive TUI Features:'"
Enter
Type "echo ''"
Enter
Type "echo '📱 Mobile (≤40 cols):'"
Enter
Type "echo '   • Card-based UI with swipe navigation'"
Enter
Type "echo '   • Essential info only, large touch targets'"
Enter
Type "echo '   • Voice commands prominent'"
Enter
Type "echo ''"
Type "echo '📟 Tablet (41-80 cols):'"
Enter
Type "echo '   • Two-column layouts with tabs'"
Enter
Type "echo '   • Medium information density'"
Enter
Type "echo '   • Mixed mouse/keyboard interaction'"
Enter
Type "echo ''"
Type "echo '🖥️  Desktop (81-120 cols):'"
Enter
Type "echo '   • Multi-panel dashboard'"
Enter
Type "echo '   • Rich data visualization'"
Enter
Type "echo '   • Full feature access'"
Enter
Type "echo ''"
Type "echo '🖥️🖥️ Ultrawide (121+ cols):'"
Enter
Type "echo '   • Mission control layout'"
Enter
Type "echo '   • Multi-cluster comparison'"
Enter
Type "echo '   • Advanced analytics'"
Enter

Type "echo ''"
Type "if command -v figlet >/dev/null 2>&1; then figlet -f small 'Responsive TUI'; else echo 'Responsive TUI'; fi"
Enter
Type "echo '🎯 Implementation: Progressive Enhancement'"
Enter
Type "echo '   Base features work everywhere → Enhanced features for larger screens'"
Enter
Type "echo ''"
Type "echo '🔧 Built with: LipGloss + Bubbletea + Responsive Design Principles'"
Enter

