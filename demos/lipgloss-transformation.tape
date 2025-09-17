# VHS Tape: LipGloss TUI Transformation Demo
# This demonstrates the before/after of applying professional styling to the Redis Work Queue TUI

# Settings
Set Theme "Catppuccin Mocha"
Set FontFamily "JetBrains Mono"
Set FontSize 14
Set Width 1200
Set Height 800
Set TypingSpeed 100ms
Set Shell "zsh"

# Title card
Type "# Redis Work Queue TUI - LipGloss Transformation Demo"
Enter
Sleep 1s

Type "echo 'Before: Basic terminal output...'"
Enter
Sleep 2s

# Show the BEFORE state - basic text output like in the screenshots
Type "cat << 'EOF'"
Enter
Type "? toggle help • q quit"
Enter
Enter
Type "Duck"
Enter
Type "Duck"
Enter
Type "Duck"
Enter
Type "Honk -> Goose"
Enter
Type "Duck"
Enter
Enter
Type "> |"
Enter
Type "EOF"
Enter

Sleep 3s
Clear

# Transition
Type "echo '✨ Transforming with LipGloss styling...'"
Enter
Sleep 2s

# Show the AFTER state - professional styled TUI mockup
Type "cat << 'EOF'"
Enter

# Professional header with branding
Type "┌─ 🔄 Redis Work Queue Monitor ─────────────────────────────────────┐"
Enter
Type "│ Production Environment • Desktop Mode                            │"
Enter
Type "└──────────────────────────────────────────────────────────────────┘"
Enter
Enter

# Status bar with indicators
Type "┌─ Status ─────────────────────────────────────────────────────────┐"
Enter
Type "│ ✓ Connected   Queues: 5 ↗   Workers: 12 ↗   ⚠ 3 DLQ           │"
Enter
Type "└──────────────────────────────────────────────────────────────────┘"
Enter
Enter

# Enhanced tabs
Type "▶ ⚡ Jobs    👥 Workers    💀 DLQ    ⏰ Debug    🔗 Events    ⚙️  Settings"
Enter
Type "────────────────────────────────────────────────────────────────────"
Enter
Enter

# Professional queue table
Type "┌─ 📋 Active Queues ──────────┐  ┌─ Quick Actions ────────────┐"
Enter
Type "│ Queue Name         Jobs     │  │ [Pause] [Resume] [Peek]    │"
Enter
Type "│ ──────────────────────────  │  │ [Drain]  [Scale]  [Logs]   │"
Enter
Type "│ ▶ prod-jobs        1,234 ✓  │  │                            │"
Enter
Type "│   async-emails       789 ✓  │  │ Selected: prod-jobs        │"
Enter
Type "│   data-processing     45 ○  │  │ Status: ✓ Active           │"
Enter
Type "│   notifications      156 ⚠  │  │ Rate: 45 jobs/min         │"
Enter
Type "│   cleanup-tasks       23 ○  │  └────────────────────────────┘"
Enter
Type "└─────────────────────────────┘"
Enter
Enter

# Charts and metrics
Type "┌─ 📊 Queue Metrics ──────────┐  ┌─ 👥 Worker Fleet ──────────┐"
Enter
Type "│ Processing Rate (jobs/min)   │  │ Fleet Status: ✓ Healthy    │"
Enter
Type "│      ▄▆█▆▄▃▂▁▂▃▄▆█▆▄        │  │ Active: 12/15 workers      │"
Enter
Type "│      ▁▂▃▄▆█▆▄▃▂▁▂▃▄▆        │  │ CPU: ▓▓▓▓▓░░░░░ 52%       │"
Enter
Type "│ Current: 245/min             │  │ Mem: ▓▓▓▓▓▓▓░░░ 68%       │"
Enter
Type "│ Peak: 487/min                │  │ Load: Normal               │"
Enter
Type "└──────────────────────────────┘  └────────────────────────────┘"
Enter
Enter

# Enhanced footer
Type "────────────────────────────────────────────────────────────────────"
Enter
Type "[Enter] Peek Queue • [Space] Pause/Resume • [Tab] Switch • [?] Help"
Enter

Type "EOF"
Enter
Sleep 3s

# Interactive demo section
Clear
Type "echo '🎬 Interactive Demo:'"
Enter
Sleep 1s

# Simulate running the enhanced TUI
Type "# Simulate starting the enhanced TUI"
Enter
Type "if [ -x ./bin/tui ]; then ./bin/tui --config config/config.yaml; else echo 'Skipping ./bin/tui demo (binary missing)'; fi"
Enter
Sleep 1s

# Show animated loading
Type "echo 'Starting Redis Work Queue Monitor...'"
Enter
Sleep 1s

# Show connection sequence
Type "echo '⠋ Connecting to Redis...'"
Enter
Sleep 500ms
Type "echo '⠙ Detecting queues...'"
Enter
Sleep 500ms
Type "echo '⠹ Loading worker status...'"
Enter
Sleep 500ms
Type "echo '⠸ Initializing dashboard...'"
Enter
Sleep 500ms
Type "echo '✓ Ready!'"
Enter
Sleep 1s

# Simulate the responsive behavior
Type "echo '📱 Responsive Design Demo:'"
Enter
Sleep 1s

Type "echo 'Terminal Width: 40 cols (Mobile Mode)'"
Enter
Type "cat << 'EOF'"
Enter
Type "┌─ Queue Monitor ─┐"
Enter
Type "│ ⚡ prod-jobs    │"
Enter
Type "│    1,234 queued │"
Enter
Type "│ ▓▓▓▓▓▓▓░░░ 70%  │"
Enter
Type "│                 │"
Enter
Type "│ 🔴 3 DLQ alerts │"
Enter
Type "│ 🟢 12 workers   │"
Enter
Type "│                 │"
Enter
Type "│ [Swipe for more]│"
Enter
Type "│ ◉ ○ ○ ○         │"
Enter
Type "└─────────────────┘"
Enter
Type "EOF"
Enter
Sleep 2s

Clear

Type "echo 'Terminal Width: 120 cols (Desktop Mode)'"
Enter
Type "cat << 'EOF'"
Enter
Type "┌─ Redis Work Queue Monitor ─────────────────────────────────────────┐"
Enter
Type "│ ▶ ⚡Jobs    👥Workers    💀DLQ    ⏰Debug    [Ctrl+P] Command      │"
Enter
Type "├─────────────────────────────────────────────────────────────────────┤"
Enter
Type "│ Queue Name      │Jobs │Rate   ║Worker Health║DLQ Status ║Actions  │"
Enter
Type "│ prod-jobs      │1,234│45/min ║▓▓▓▓▓▓▓▓░░80%║⚠ 3 fails ║[P][L]   │"
Enter
Type "│ async-emails   │ 789 │12/min ║▓▓▓▓▓▓▓▓▓▓100%║✓ Clean   ║[S][I]   │"
Enter
Type "├─────────────────────────────────────────────────────────────────────┤"
Enter
Type "│ 🎤 Voice: \"Hey Queue\" │ [Tab] Focus │ [?] Help │ [Q] Quit       │"
Enter
Type "└─────────────────────────────────────────────────────────────────────┘"
Enter
Type "EOF"
Enter
Sleep 3s

# Show key features
Clear
Type "echo '🚀 Key LipGloss Features Added:'"
Enter
Sleep 1s

Type "echo '• 🎨 Adaptive color schemes (light/dark mode)'"
Enter
Sleep 500ms
Type "echo '• 📱 Responsive layouts (mobile → desktop → ultrawide)'"
Enter
Sleep 500ms
Type "echo '• 🎯 Professional status indicators and badges'"
Enter
Sleep 500ms
Type "echo '• 📊 Enhanced progress bars and metrics'"
Enter
Sleep 500ms
Type "echo '• 🎪 Consistent borders, spacing, and typography'"
Enter
Sleep 500ms
Type "echo '• ⚡ Interactive elements with hover states'"
Enter
Sleep 500ms
Type "echo '• 🎭 Context-sensitive help and actions'"
Enter
Sleep 1s

# Closing
Clear
Type "echo '✨ From basic text to professional TUI in minutes!'"
Enter
Type "echo ''"
Enter
Type "echo '🔗 Learn more:'"
Enter
Type "echo '  • LipGloss: https://github.com/charmbracelet/lipgloss'"
Enter
Type "echo '  • Bubbletea: https://github.com/charmbracelet/bubbletea'"
Enter
Type "echo '  • Full TUI redesign: docs/design/TUI2-design.md'"
Enter
Sleep 2s

# Final frame
Type "if command -v figlet >/dev/null 2>&1; then figlet -f small 'Redis TUI'; else echo 'Redis TUI'; fi"
Enter
Type "echo 'Powered by LipGloss 💄'"
Enter

# Pause to show final screen
Sleep 5s
