# Bulk Requeue Helper Plugin
# An action plugin that helps requeue failed jobs in bulk with filtering

# Plugin state
failed_jobs = []
selected_jobs = []
current_queue = ""
filter_text = ""
confirmation_mode = False

def init():
    """Initialize the plugin"""
    log("info", "Bulk Requeue Helper plugin initialized")
    subscribe("selection")
    subscribe("key_events")

def start():
    """Start the plugin"""
    log("info", "Bulk Requeue Helper plugin started")
    refresh_failed_jobs()

def stop():
    """Stop the plugin"""
    log("info", "Bulk Requeue Helper plugin stopped")

def on_event(event):
    """Handle incoming events"""
    event_type = event.get("type", "")

    if event_type == "selection":
        handle_selection_event(event)
    elif event_type == "key_events":
        handle_key_event(event)

def handle_selection_event(event):
    """Handle selection changes"""
    selected = event.get("selected", {})
    if selected.get("type") == "queue":
        global current_queue
        current_queue = selected.get("id", "")
        refresh_failed_jobs()

def handle_key_event(event):
    """Handle keyboard input"""
    key = event.get("key", "")

    if key == "r":
        refresh_failed_jobs()
    elif key == "f":
        toggle_filter_mode()
    elif key == "space":
        toggle_job_selection()
    elif key == "a":
        select_all_jobs()
    elif key == "c":
        clear_selection()
    elif key == "q":
        if selected_jobs:
            enter_confirmation_mode()
    elif key == "y" and confirmation_mode:
        execute_requeue()
    elif key == "n" and confirmation_mode:
        cancel_requeue()

def refresh_failed_jobs():
    """Refresh the list of failed jobs"""
    global failed_jobs

    if not current_queue:
        queues = get_queues()
        if queues:
            current_queue = queues[0]

    if current_queue:
        # Get failed jobs from the current queue
        all_jobs = get_jobs(current_queue, 100)
        failed_jobs = []

        for job in all_jobs:
            if job.get("status") == "failed":
                failed_jobs.append(job)

        log("debug", "Found " + str(len(failed_jobs)) + " failed jobs in queue " + current_queue)

    render_panel()

def toggle_job_selection():
    """Toggle selection of the currently focused job"""
    # Implementation would depend on cursor position
    # For now, just log the action
    log("debug", "Toggle job selection")

def select_all_jobs():
    """Select all filtered jobs"""
    global selected_jobs
    selected_jobs = get_filtered_jobs()
    log("info", "Selected " + str(len(selected_jobs)) + " jobs")
    render_panel()

def clear_selection():
    """Clear all job selections"""
    global selected_jobs
    selected_jobs = []
    log("info", "Cleared selection")
    render_panel()

def get_filtered_jobs():
    """Get jobs that match the current filter"""
    if not filter_text:
        return failed_jobs

    filtered = []
    for job in failed_jobs:
        job_id = job.get("id", "")
        error = job.get("error", "")
        if filter_text in job_id or filter_text in error:
            filtered.append(job)

    return filtered

def enter_confirmation_mode():
    """Enter confirmation mode for bulk requeue"""
    global confirmation_mode

    if not selected_jobs:
        log("warn", "No jobs selected for requeue")
        return

    confirmation_mode = True
    log("info", "Entering confirmation mode for " + str(len(selected_jobs)) + " jobs")
    render_panel()

def execute_requeue():
    """Execute the bulk requeue operation"""
    global confirmation_mode, selected_jobs

    log("info", "Executing bulk requeue for " + str(len(selected_jobs)) + " jobs")

    success_count = 0
    error_count = 0

    for job in selected_jobs:
        job_id = job.get("id", "")
        try:
            requeue_job(job_id)
            success_count += 1
            log("debug", "Requeued job " + job_id)
        except:
            error_count += 1
            log("error", "Failed to requeue job " + job_id)

    # Show results
    message = "Requeued " + str(success_count) + " jobs"
    if error_count > 0:
        message += ", " + str(error_count) + " errors"

    show_dialog("Bulk Requeue Complete", message)

    # Reset state
    confirmation_mode = False
    selected_jobs = []
    refresh_failed_jobs()

def cancel_requeue():
    """Cancel the bulk requeue operation"""
    global confirmation_mode
    confirmation_mode = False
    log("info", "Bulk requeue cancelled")
    render_panel()

def toggle_filter_mode():
    """Toggle filter input mode"""
    # In a real implementation, this would activate text input
    log("debug", "Toggle filter mode")

def render():
    """Render the bulk requeue panel"""
    render_panel()

def render_panel():
    """Render the bulk requeue helper panel"""
    lines = []

    # Header
    lines.append("┌─ Bulk Requeue Helper ────────────────────────┐")

    if confirmation_mode:
        render_confirmation_panel(lines)
    else:
        render_main_panel(lines)

    # Footer with help
    lines.append("├──────────────────────────────────────────────┤")
    if confirmation_mode:
        lines.append("│ [Y]es  [N]o                                  │")
    else:
        lines.append("│ [R]efresh [F]ilter [Space]Select [A]ll [C]lear │")
        lines.append("│ [Q]ueue selected jobs                       │")
    lines.append("└──────────────────────────────────────────────┘")

    # Render each line
    for i, line in enumerate(lines):
        render(line)

def render_main_panel(lines):
    """Render the main panel interface"""
    # Queue info
    queue_line = "│ Queue: " + pad_right(current_queue, 36) + " │"
    lines.append(queue_line)

    # Filter info
    if filter_text:
        filter_line = "│ Filter: " + pad_right(filter_text, 35) + " │"
        lines.append(filter_line)

    # Stats
    filtered_jobs = get_filtered_jobs()
    stats_line = "│ Jobs: " + str(len(filtered_jobs)) + \
                 "  Selected: " + str(len(selected_jobs)) + \
                 pad_right("", 20) + " │"
    lines.append(stats_line)

    lines.append("├──────────────────────────────────────────────┤")

    # Job list
    if not filtered_jobs:
        lines.append("│  No failed jobs found                       │")
        lines.append("│                                              │")
    else:
        # Headers
        lines.append("│ Sel │ Job ID      │ Error                   │")
        lines.append("├─────┼─────────────┼─────────────────────────┤")

        # Show up to 10 jobs
        displayed_jobs = filtered_jobs[:10]
        for job in displayed_jobs:
            job_id = job.get("id", "")
            error = job.get("error", "")

            # Check if selected
            is_selected = job in selected_jobs
            sel_marker = "[X]" if is_selected else "[ ]"

            # Truncate text to fit
            job_id_short = truncate(job_id, 11)
            error_short = truncate(error, 23)

            job_line = "│ " + sel_marker + " │ " + \
                      pad_right(job_id_short, 11) + " │ " + \
                      pad_right(error_short, 23) + " │"
            lines.append(job_line)

        # Show more indicator
        if len(filtered_jobs) > 10:
            more_line = "│     │ ... " + str(len(filtered_jobs) - 10) + \
                       " more jobs" + pad_right("", 20) + " │"
            lines.append(more_line)

def render_confirmation_panel(lines):
    """Render the confirmation panel"""
    lines.append("│                                              │")
    lines.append("│           CONFIRMATION REQUIRED             │")
    lines.append("│                                              │")

    job_count = str(len(selected_jobs))
    confirm_line = "│  Requeue " + job_count + " failed jobs?"
    confirm_line += pad_right("", 44 - len(confirm_line) - 1) + " │"
    lines.append(confirm_line)

    lines.append("│                                              │")
    lines.append("│  This action cannot be undone.              │")
    lines.append("│                                              │")

    # Show first few job IDs
    if selected_jobs:
        lines.append("│  Jobs to requeue:                           │")
        for i, job in enumerate(selected_jobs[:3]):
            job_id = job.get("id", "")
            job_line = "│    " + truncate(job_id, 38) + \
                      pad_right("", 38 - len(truncate(job_id, 38))) + " │"
            lines.append(job_line)

        if len(selected_jobs) > 3:
            more_line = "│    ... and " + str(len(selected_jobs) - 3) + " more"
            more_line += pad_right("", 44 - len(more_line) - 1) + " │"
            lines.append(more_line)

    lines.append("│                                              │")

def truncate(text, max_len):
    """Truncate text to maximum length"""
    if len(text) <= max_len:
        return text
    return text[:max_len-2] + ".."

def pad_right(text, width):
    """Pad text to the right with spaces"""
    if len(text) >= width:
        return text[:width]
    return text + " " * (width - len(text))

# Plugin lifecycle hooks
init()