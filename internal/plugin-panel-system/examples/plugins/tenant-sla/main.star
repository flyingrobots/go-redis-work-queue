# Tenant SLA Monitor Plugin
# A read-only plugin that displays SLA metrics for different tenants

# Plugin state
sla_data = {}
last_update = 0

def init():
    """Initialize the plugin"""
    log("info", "Tenant SLA Monitor plugin initialized")
    subscribe("stats")
    subscribe("timer")

def start():
    """Start the plugin"""
    log("info", "Tenant SLA Monitor plugin started")
    refresh_data()

def stop():
    """Stop the plugin"""
    log("info", "Tenant SLA Monitor plugin stopped")

def on_event(event):
    """Handle incoming events"""
    event_type = event.get("type", "")

    if event_type == "stats":
        handle_stats_event(event)
    elif event_type == "timer":
        if event.get("name") == "refresh":
            refresh_data()

def handle_stats_event(event):
    """Process stats event to extract SLA data"""
    global sla_data, last_update

    # Extract tenant data from stats
    tenants = event.get("tenants", {})

    for tenant_id, stats in tenants.items():
        if tenant_id not in sla_data:
            sla_data[tenant_id] = {
                "success_rate": 0.0,
                "avg_response_time": 0.0,
                "error_count": 0,
                "total_jobs": 0
            }

        # Calculate SLA metrics
        total = stats.get("total_jobs", 0)
        successful = stats.get("successful_jobs", 0)
        failed = stats.get("failed_jobs", 0)
        avg_time = stats.get("avg_response_time_ms", 0)

        if total > 0:
            sla_data[tenant_id]["success_rate"] = (successful / total) * 100
            sla_data[tenant_id]["avg_response_time"] = avg_time
            sla_data[tenant_id]["error_count"] = failed
            sla_data[tenant_id]["total_jobs"] = total

    last_update = event.get("timestamp", 0)
    render_panel()

def refresh_data():
    """Refresh SLA data from the system"""
    stats = get_stats()

    # Process the stats to update SLA data
    if stats:
        log("debug", "Refreshed SLA data for " + str(len(sla_data)) + " tenants")
        render_panel()

def render():
    """Render the SLA dashboard panel"""
    render_panel()

def render_panel():
    """Render the tenant SLA panel"""
    lines = []

    # Header
    lines.append("┌─ Tenant SLA Dashboard ──────────────────────┐")
    lines.append("│                                              │")

    if not sla_data:
        lines.append("│  No tenant data available                   │")
        lines.append("│  Waiting for stats...                       │")
    else:
        # Column headers
        lines.append("│ Tenant     │ SLA%  │ Avg RT │ Errors │ Jobs │")
        lines.append("├────────────┼───────┼────────┼────────┼──────┤")

        # Sort tenants by success rate (worst first)
        sorted_tenants = sorted(sla_data.items(),
                               key=lambda x: x[1]["success_rate"])

        for tenant_id, metrics in sorted_tenants:
            sla_pct = metrics["success_rate"]
            avg_rt = metrics["avg_response_time"]
            errors = metrics["error_count"]
            total = metrics["total_jobs"]

            # Color coding based on SLA
            if sla_pct >= 99.9:
                status = "🟢"
            elif sla_pct >= 99.0:
                status = "🟡"
            else:
                status = "🔴"

            # Format the row
            tenant_str = truncate(tenant_id, 8)
            sla_str = format_number(sla_pct, 1) + "%"
            rt_str = format_number(avg_rt, 0) + "ms"
            error_str = str(errors)
            job_str = str(total)

            line = "│ " + status + " " + pad_right(tenant_str, 7) + " │ " + \
                   pad_left(sla_str, 5) + " │ " + \
                   pad_left(rt_str, 6) + " │ " + \
                   pad_left(error_str, 6) + " │ " + \
                   pad_left(job_str, 4) + " │"
            lines.append(line)

    # Footer
    lines.append("│                                              │")
    if last_update > 0:
        update_str = "Last updated: " + format_timestamp(last_update)
        lines.append("│ " + pad_right(update_str, 44) + " │")
    lines.append("└──────────────────────────────────────────────┘")

    # Render each line
    for i, line in enumerate(lines):
        render(line)

def format_number(num, decimals):
    """Format a number with specified decimal places"""
    if decimals == 0:
        return str(int(num))
    else:
        # Simple decimal formatting
        factor = 10 ** decimals
        return str(int(num * factor) / factor)

def format_timestamp(ts):
    """Format timestamp for display"""
    # Simplified timestamp formatting
    return "now"

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

def pad_left(text, width):
    """Pad text to the left with spaces"""
    if len(text) >= width:
        return text[:width]
    return " " * (width - len(text)) + text

# Plugin lifecycle hooks
init()