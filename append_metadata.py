#!/usr/bin/env python3
"""
Append YAML metadata to feature documents and generate DAG.json
"""

import os
import json
from dependency_analysis import features

# Path to ideas directory
ideas_dir = "/Users/james/git/go-redis-work-queue/docs/ideas"

# Generate YAML metadata for each feature
def generate_yaml_metadata(feature_name, deps):
    yaml = f"""
---
feature: {feature_name}
dependencies:
  hard:
{format_list(deps.get('hard', []), '    - ')}
  soft:
{format_list(deps.get('soft', []), '    - ')}
enables:
{format_list(deps.get('enables', []), '  - ')}
provides:
{format_list(deps.get('provides', []), '  - ')}
---"""
    return yaml

def format_list(items, prefix):
    if not items:
        return f"{prefix}[]"
    return '\n'.join([f"{prefix}{item}" for item in items])

# Process each document
for feature_name, deps in features.items():
    file_path = os.path.join(ideas_dir, f"{feature_name}.md")

    if os.path.exists(file_path):
        # Read existing content
        with open(file_path, 'r') as f:
            content = f.read()

        # Check if metadata already exists
        if not content.endswith('---'):
            # Generate and append metadata
            yaml_metadata = generate_yaml_metadata(feature_name, deps)

            # Write back with metadata
            with open(file_path, 'w') as f:
                f.write(content)
                f.write(yaml_metadata)

            print(f"✓ Appended metadata to {feature_name}.md")
        else:
            print(f"⚠ Metadata already exists in {feature_name}.md")
    else:
        print(f"✗ File not found: {feature_name}.md")

# Generate DAG.json
def generate_dag():
    nodes = []
    edges = []
    node_id = 0
    node_map = {}

    # Add infrastructure nodes
    for infra_name in ["redis", "tui_framework", "internal_admin", "config_management",
                       "metrics_system", "monitoring_system", "auth_middleware", "audit_logging",
                       "opentelemetry_sdk", "scheduler_primitives", "plugin_runtime", "event_sourcing",
                       "idempotency_keys", "fault_injection", "controller_runtime", "k8s_api",
                       "worker_versioning", "routing_system", "http_client", "content_hashing",
                       "tenant_labels", "graph_storage", "relationship_tracking", "storage_backend",
                       "completed_stream", "clickhouse", "s3", "time_series_analysis",
                       "namespace_separation", "rate_limiting", "circuit_breaker", "serialization",
                       "git_integration", "retry_system", "ml_models", "storage_abstraction",
                       "migration_system", "voice_recognition", "accessibility_framework", "lipgloss",
                       "log_aggregation", "classification_engine", "policy_engine", "rate_limiter",
                       "scheduling_system", "multiplexing", "json_editor", "schema_validation",
                       "worker_management", "bubblezone"]:
        nodes.append({
            "node_id": f"infra_{infra_name}",
            "idea": infra_name,
            "type": "infrastructure"
        })
        node_map[infra_name] = f"infra_{infra_name}"

    # Add feature nodes
    for feature_name in features.keys():
        node_id_str = f"feature_{feature_name}"
        nodes.append({
            "node_id": node_id_str,
            "idea": feature_name,
            "spec": f"docs/ideas/{feature_name}.md"
        })
        node_map[feature_name] = node_id_str

    # Add edges for dependencies
    for feature_name, deps in features.items():
        feature_node = node_map[feature_name]

        # Hard dependencies
        for dep in deps.get('hard', []):
            if dep in node_map:
                edges.append({
                    "from": node_map[dep],
                    "to": feature_node,
                    "dependencyType": "hard"
                })

        # Soft dependencies
        for dep in deps.get('soft', []):
            if dep in node_map:
                edges.append({
                    "from": node_map[dep],
                    "to": feature_node,
                    "dependencyType": "soft"
                })

    return {
        "nodes": nodes,
        "edges": edges
    }

# Write DAG.json
dag = generate_dag()
dag_path = os.path.join(ideas_dir, "DAG.json")
with open(dag_path, 'w') as f:
    json.dump(dag, f, indent=2)

print(f"\n✓ Generated DAG.json with {len(dag['nodes'])} nodes and {len(dag['edges'])} edges")