#!/usr/bin/env python3
"""
Append YAML metadata to feature documents and generate DAG.json.
"""

import argparse
import json
import os
from typing import Dict, List

from dependency_analysis import features


def format_list(items: List[str], prefix: str) -> str:
    if not items:
        return f"{prefix}[]"
    return "\n".join(f"{prefix}{item}" for item in items)


def generate_yaml_metadata(feature_name: str, deps: Dict[str, List[str]]) -> str:
    return f"""
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


def append_metadata_for_features(ideas_dir: str) -> None:
    for feature_name, deps in features.items():
        file_path = os.path.join(ideas_dir, f"{feature_name}.md")
        if not os.path.exists(file_path):
            print(f"✗ File not found: {feature_name}.md")
            continue

        with open(file_path, "r", encoding="utf-8") as handle:
            content = handle.read()

        if content.endswith("---"):
            print(f"⚠ Metadata already exists in {feature_name}.md")
            continue

        yaml_metadata = generate_yaml_metadata(feature_name, deps)
        with open(file_path, "w", encoding="utf-8") as handle:
            handle.write(content)
            handle.write(yaml_metadata)
        print(f"✓ Appended metadata to {feature_name}.md")


def generate_dag(ideas_dir: str) -> Dict[str, List[Dict[str, str]]]:
    nodes: List[Dict[str, str]] = []
    edges: List[Dict[str, str]] = []
    node_map: Dict[str, str] = {}

    infrastructure_nodes = [
        "redis",
        "tui_framework",
        "internal_admin",
        "config_management",
        "metrics_system",
        "monitoring_system",
        "auth_middleware",
        "audit_logging",
        "opentelemetry_sdk",
        "scheduler_primitives",
        "plugin_runtime",
        "event_sourcing",
        "idempotency_keys",
        "fault_injection",
        "controller_runtime",
        "k8s_api",
        "worker_versioning",
        "routing_system",
        "http_client",
        "content_hashing",
        "tenant_labels",
        "graph_storage",
        "relationship_tracking",
        "storage_backend",
        "completed_stream",
        "clickhouse",
        "s3",
        "time_series_analysis",
        "namespace_separation",
        "rate_limiting",
        "circuit_breaker",
        "serialization",
        "git_integration",
        "retry_system",
        "ml_models",
        "storage_abstraction",
        "migration_system",
        "voice_recognition",
        "accessibility_framework",
        "lipgloss",
        "log_aggregation",
        "classification_engine",
        "policy_engine",
        "rate_limiter",
        "scheduling_system",
        "multiplexing",
        "json_editor",
        "schema_validation",
        "worker_management",
        "bubblezone",
    ]

    for infra_name in infrastructure_nodes:
        node_id = f"infra_{infra_name}"
        nodes.append({
            "node_id": node_id,
            "idea": infra_name,
            "type": "infrastructure",
        })
        node_map[infra_name] = node_id

    for feature_name in features:
        node_id = f"feature_{feature_name}"
        nodes.append(
            {
                "node_id": node_id,
                "idea": feature_name,
                "spec": f"docs/ideas/{feature_name}.md",
            }
        )
        node_map[feature_name] = node_id

    for feature_name, deps in features.items():
        feature_node = node_map[feature_name]
        for dep in deps.get("hard", []):
            if dep in node_map:
                edges.append(
                    {
                        "from": node_map[dep],
                        "to": feature_node,
                        "dependencyType": "hard",
                    }
                )
        for dep in deps.get("soft", []):
            if dep in node_map:
                edges.append(
                    {
                        "from": node_map[dep],
                        "to": feature_node,
                        "dependencyType": "soft",
                    }
                )

    return {"nodes": nodes, "edges": edges}


def write_dag(ideas_dir: str) -> None:
    dag = generate_dag(ideas_dir)
    dag_path = os.path.join(ideas_dir, "DAG.json")
    with open(dag_path, "w", encoding="utf-8") as handle:
        json.dump(dag, handle, indent=2)
    print(f"\n✓ Generated DAG.json with {len(dag['nodes'])} nodes and {len(dag['edges'])} edges")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Append metadata and build feature dependency DAG")
    parser.add_argument(
        "--ideas-dir",
        default=os.environ.get("IDEAS_DIR", "docs/ideas"),
        help="Directory containing idea markdown files (default: docs/ideas or IDEAS_DIR env)",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    ideas_dir = os.path.expanduser(args.ideas_dir)
    append_metadata_for_features(ideas_dir)
    write_dag(ideas_dir)


if __name__ == "__main__":
    main()
