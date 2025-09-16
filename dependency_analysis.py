#!/usr/bin/env python3
"""
Dependency analysis for feature documents
"""

# Feature dependency mapping based on document analysis
features = {
    "admin-api": {
        "hard": ["redis", "internal_admin", "auth_middleware"],
        "soft": ["metrics_system", "audit_logging"],
        "enables": ["multi_cluster_control", "visual_dag_builder", "plugin_panel_system", "rbac_and_tokens",
                   "event_hooks", "kubernetes_operator", "time_travel_debugger", "automatic_capacity_planning"],
        "provides": ["rest_api", "grpc_api", "stats_endpoints", "purge_endpoints", "bench_endpoints"]
    },
    "multi-cluster-control": {
        "hard": ["admin_api", "redis", "tui_framework", "config_management"],
        "soft": ["dlq_remediation_ui", "capacity_planning", "monitoring_system"],
        "enables": ["global_operations", "disaster_recovery", "federation", "cross_region_sync", "chaos_engineering"],
        "provides": ["cluster_switching", "compare_view", "multi_apply_actions"]
    },
    "visual-dag-builder": {
        "hard": ["admin_api", "tui_framework", "redis", "scheduler_primitives"],
        "soft": ["event_hooks", "distributed_tracing"],
        "enables": ["workflow_orchestration", "pipeline_execution", "compensation_patterns"],
        "provides": ["dag_editor", "workflow_runner", "dag_validation", "live_status"]
    },
    "distributed-tracing-integration": {
        "hard": ["opentelemetry_sdk", "redis"],
        "soft": ["admin_api", "event_hooks"],
        "enables": ["trace_drilldown_log_tail", "time_travel_debugger", "visual_dag_builder"],
        "provides": ["trace_context", "span_management", "trace_exemplars", "waterfall_view"]
    },
    "plugin-panel-system": {
        "hard": ["admin_api", "tui_framework", "plugin_runtime"],
        "soft": ["event_hooks", "rbac_and_tokens"],
        "enables": ["custom_visualizations", "third_party_integrations", "extensible_ui"],
        "provides": ["plugin_api", "panel_registry", "hot_reload", "plugin_marketplace"]
    },
    "time-travel-debugger": {
        "hard": ["admin_api", "event_sourcing", "redis"],
        "soft": ["distributed_tracing", "job_genealogy_navigator"],
        "enables": ["replay_debugging", "production_debugging", "state_comparison"],
        "provides": ["event_capture", "timeline_navigation", "state_snapshots", "diff_viewer"]
    },
    "exactly-once-patterns": {
        "hard": ["redis", "idempotency_keys"],
        "soft": ["admin_api", "event_hooks"],
        "enables": ["reliable_processing", "duplicate_prevention", "transactional_outbox"],
        "provides": ["dedup_sets", "idempotency_helpers", "outbox_pattern", "state_management"]
    },
    "rbac-and-tokens": {
        "hard": ["admin_api", "auth_middleware"],
        "soft": ["audit_logging", "tui_framework"],
        "enables": ["multi_tenant_isolation", "kubernetes_operator", "plugin_panel_system"],
        "provides": ["token_management", "role_system", "permission_scopes", "audit_trail"]
    },
    "chaos-harness": {
        "hard": ["admin_api", "workers", "fault_injection"],
        "soft": ["multi_cluster_control", "monitoring_system"],
        "enables": ["reliability_testing", "chaos_engineering", "failure_recovery"],
        "provides": ["fault_injectors", "chaos_scenarios", "recovery_validation"]
    },
    "anomaly-radar-slo-budget": {
        "hard": ["metrics_system", "redis"],
        "soft": ["admin_api", "monitoring_system"],
        "enables": ["sre_operations", "incident_detection", "slo_management"],
        "provides": ["anomaly_detection", "slo_tracking", "burn_rate_alerts", "threshold_monitoring"]
    },
    "automatic-capacity-planning": {
        "hard": ["admin_api", "metrics_history", "forecasting"],
        "soft": ["kubernetes_operator", "multi_cluster_control"],
        "enables": ["auto_scaling", "resource_optimization", "cost_reduction"],
        "provides": ["capacity_recommendations", "scaling_policies", "resource_predictions"]
    },
    "kubernetes-operator": {
        "hard": ["admin_api", "controller_runtime", "k8s_api"],
        "soft": ["rbac_and_tokens", "automatic_capacity_planning"],
        "enables": ["gitops_deployment", "auto_scaling", "declarative_config"],
        "provides": ["queue_crd", "worker_crd", "reconciliation", "k8s_integration"]
    },
    "canary-deployments": {
        "hard": ["worker_versioning", "routing_system", "metrics_system"],
        "soft": ["admin_api", "multi_cluster_control"],
        "enables": ["safe_rollouts", "gradual_deployment", "automatic_promotion"],
        "provides": ["traffic_splitting", "version_routing", "rollback_mechanism"]
    },
    "event-hooks": {
        "hard": ["admin_api", "http_client"],
        "soft": ["rbac_and_tokens", "distributed_tracing"],
        "enables": ["external_integrations", "real_time_notifications", "workflow_triggers"],
        "provides": ["webhook_delivery", "event_filtering", "retry_logic", "deep_links"]
    },
    "smart-payload-deduplication": {
        "hard": ["redis", "content_hashing"],
        "soft": ["admin_api", "monitoring_system"],
        "enables": ["memory_optimization", "cost_reduction", "scale_improvement"],
        "provides": ["dedup_engine", "compression", "reference_counting", "similarity_detection"]
    },
    "job-budgeting": {
        "hard": ["metrics_system", "tenant_labels"],
        "soft": ["admin_api", "multi_tenant_isolation", "advanced_rate_limiting"],
        "enables": ["cost_control", "resource_governance", "chargeback"],
        "provides": ["cost_tracking", "budget_enforcement", "spending_alerts", "usage_reports"]
    },
    "job-genealogy-navigator": {
        "hard": ["graph_storage", "relationship_tracking", "tui_framework"],
        "soft": ["admin_api", "distributed_tracing"],
        "enables": ["debugging_workflows", "impact_analysis", "root_cause_analysis"],
        "provides": ["tree_visualization", "relationship_graph", "blame_mode", "ancestry_tracking"]
    },
    "long-term-archives": {
        "hard": ["storage_backend", "completed_stream"],
        "soft": ["admin_api", "clickhouse", "s3"],
        "enables": ["historical_analysis", "compliance", "forensics"],
        "provides": ["data_export", "retention_policies", "query_interface", "archive_management"]
    },
    "forecasting": {
        "hard": ["metrics_history", "time_series_analysis"],
        "soft": ["admin_api", "automatic_capacity_planning"],
        "enables": ["predictive_operations", "capacity_planning", "proactive_scaling"],
        "provides": ["arima_models", "prophet_integration", "trend_analysis", "predictions"]
    },
    "multi-tenant-isolation": {
        "hard": ["redis", "namespace_separation", "rbac_and_tokens"],
        "soft": ["admin_api", "job_budgeting"],
        "enables": ["saas_deployment", "secure_multi_tenancy", "compliance"],
        "provides": ["tenant_namespacing", "quota_management", "encryption", "audit_trail"]
    },
    "producer-backpressure": {
        "hard": ["redis", "rate_limiting"],
        "soft": ["admin_api", "circuit_breaker"],
        "enables": ["reliability", "cascade_prevention", "system_stability"],
        "provides": ["adaptive_rate_limiting", "circuit_breakers", "priority_shedding", "sdk_hints"]
    },
    "queue-snapshot-testing": {
        "hard": ["redis", "serialization"],
        "soft": ["admin_api", "git_integration"],
        "enables": ["reproducible_testing", "regression_detection", "ci_integration"],
        "provides": ["snapshot_capture", "diff_engine", "test_helpers", "restore_capability"]
    },
    "smart-retry-strategies": {
        "hard": ["redis", "retry_system"],
        "soft": ["admin_api", "ml_models"],
        "enables": ["intelligent_recovery", "reduced_failures", "adaptive_behavior"],
        "provides": ["bayesian_retry", "ml_prediction", "error_classification", "retry_optimization"]
    },
    "storage-backends": {
        "hard": ["redis", "storage_abstraction"],
        "soft": ["admin_api", "migration_system"],
        "enables": ["backend_flexibility", "performance_optimization", "multi_cloud"],
        "provides": ["redis_streams", "keydb_support", "dragonfly_support", "kafka_bridge"]
    },
    "terminal-voice-commands": {
        "hard": ["voice_recognition", "tui_framework"],
        "soft": ["admin_api", "accessibility_framework"],
        "enables": ["hands_free_operation", "accessibility", "productivity"],
        "provides": ["voice_control", "command_recognition", "audio_feedback", "wake_words"]
    },
    "theme-playground": {
        "hard": ["tui_framework", "lipgloss"],
        "soft": ["admin_api"],
        "enables": ["customization", "accessibility", "user_preference"],
        "provides": ["theme_system", "color_picker", "wcag_validation", "theme_export"]
    },
    "trace-drilldown-log-tail": {
        "hard": ["distributed_tracing", "log_aggregation", "admin_api"],
        "soft": ["time_travel_debugger"],
        "enables": ["deep_debugging", "correlated_logs", "incident_response"],
        "provides": ["trace_viewer", "log_correlation", "span_details", "waterfall_timeline"]
    },
    "dlq-remediation-ui": {
        "hard": ["admin_api", "tui_framework", "redis"],
        "soft": ["job_genealogy_navigator", "json_payload_studio"],
        "enables": ["error_recovery", "dlq_management", "operational_efficiency"],
        "provides": ["dlq_viewer", "bulk_operations", "requeue_actions", "pattern_analysis"]
    },
    "dlq-remediation-pipeline": {
        "hard": ["admin_api", "redis", "classification_engine"],
        "soft": ["dlq_remediation_ui", "event_hooks", "ml_models"],
        "enables": ["automated_recovery", "intelligent_remediation", "self_healing"],
        "provides": ["auto_classification", "remediation_rules", "transformation_pipeline", "recovery_actions"]
    },
    "patterned-load-generator": {
        "hard": ["admin_api", "redis"],
        "soft": ["json_payload_studio", "monitoring_system"],
        "enables": ["load_testing", "performance_validation", "capacity_testing"],
        "provides": ["traffic_patterns", "load_profiles", "benchmark_tools", "stress_testing"]
    },
    "policy-simulator": {
        "hard": ["admin_api", "policy_engine"],
        "soft": ["forecasting", "automatic_capacity_planning"],
        "enables": ["policy_testing", "what_if_analysis", "safe_changes"],
        "provides": ["simulation_engine", "impact_preview", "policy_validation", "dry_run"]
    },
    "advanced-rate-limiting": {
        "hard": ["redis", "rate_limiter"],
        "soft": ["admin_api", "producer_backpressure"],
        "enables": ["fair_queuing", "priority_management", "resource_protection"],
        "provides": ["token_bucket", "sliding_window", "priority_fairness", "global_limits"]
    },
    "calendar-view": {
        "hard": ["tui_framework", "scheduling_system"],
        "soft": ["admin_api", "visual_dag_builder"],
        "enables": ["schedule_visualization", "job_planning", "recurring_jobs"],
        "provides": ["calendar_ui", "schedule_management", "cron_visualization", "drag_drop_scheduling"]
    },
    "collaborative-session": {
        "hard": ["tui_framework", "multiplexing"],
        "soft": ["admin_api", "rbac_and_tokens"],
        "enables": ["team_debugging", "pair_operations", "training"],
        "provides": ["session_sharing", "read_only_mode", "control_handoff", "cursor_sharing"]
    },
    "json-payload-studio": {
        "hard": ["tui_framework", "json_editor"],
        "soft": ["admin_api", "schema_validation"],
        "enables": ["payload_creation", "testing", "validation"],
        "provides": ["json_editor", "template_system", "snippet_expansion", "schema_validation"]
    },
    "worker-fleet-controls": {
        "hard": ["admin_api", "worker_management"],
        "soft": ["multi_cluster_control", "kubernetes_operator"],
        "enables": ["fleet_management", "rolling_updates", "drain_patterns"],
        "provides": ["worker_control", "pause_resume", "drain_operations", "health_monitoring"]
    },
    "right-click-context-menus": {
        "hard": ["tui_framework", "bubblezone"],
        "soft": ["admin_api"],
        "enables": ["improved_ux", "quick_actions", "contextual_operations"],
        "provides": ["context_menus", "mouse_integration", "action_shortcuts", "menu_system"]
    }
}

# Base infrastructure that doesn't have documents but is referenced
infrastructure = {
    "redis": "Base Redis infrastructure",
    "tui_framework": "Bubble Tea TUI framework",
    "internal_admin": "Internal admin package",
    "config_management": "Configuration system",
    "metrics_system": "Prometheus metrics",
    "monitoring_system": "General monitoring",
    "auth_middleware": "Authentication middleware",
    "audit_logging": "Audit logging system"
}